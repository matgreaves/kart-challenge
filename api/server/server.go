// package server contains functionality related exposing our application as a http server.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/matgreaves/kart-challenge/api/coupons"
	"github.com/matgreaves/kart-challenge/api/orders"
	"github.com/matgreaves/kart-challenge/api/products"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const (
	// Time given to inflight requests to complete before the server hard shuts down.
	DefaultShutdownTimeout = 5 * time.Second
)

type Server struct {
	Addr     string
	Auth     StaticAuthProvider
	Logger   *slog.Logger
	Products products.Store
	Orders   orders.Store
	Coupons  coupons.Store
}

// Run starts s waiting for ctx to be cancelled before shutting down gracefully.
func (s Server) Run(ctx context.Context) error {
	server := http.Server{
		Addr:    s.Addr,
		Handler: s.Handler(),

		// sane default values to prevent common resource exhaustion attacks
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
	}

	var serr = make(chan error)
	go func() {
		serr <- server.ListenAndServe()
	}()

	s.Logger.InfoContext(ctx, "listening on "+s.Addr)

	select {
	case err := <-serr:
		return fmt.Errorf("HTTPServer server exited with error: %w", err)
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), DefaultShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	if !errors.Is(ctx.Err(), context.Canceled) {
		return ctx.Err()
	}
	return nil
}

func (s Server) Handler() http.Handler {
	m := &http.ServeMux{}
	m.Handle("GET /product", s.listProducts())
	m.Handle("GET /product/{productID}", s.getProduct())
	m.Handle("POST /order", ScopedHandler(s.Logger, "order:create", s.createOrder()))
	ah := AuthenticatedHandler(s.Auth, s.Logger, m, "/product")
	return otelhttp.NewHandler(LoggedHandler(s.Logger, ah), "req")
}

func (s Server) listProducts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// for now we don't support pagination in the API, set it to a reasonable default for "all"
		p, err := s.Products.List(r.Context(), 0, 100)
		if err != nil {
			s.handleErr(r.Context(), w, err)
		}
		if err := json.NewEncoder(w).Encode(p); err != nil {
			s.Logger.ErrorContext(r.Context(), "failed to write listProducts response to client")
		}
	}
}

func (s Server) getProduct() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("productID")

		p, err := s.Products.Get(r.Context(), id)
		if err != nil {
			s.handleErr(r.Context(), w, err)
			return
		}

		if err := json.NewEncoder(w).Encode(p); err != nil {
			s.Logger.ErrorContext(r.Context(), "failed to write getProduct response to client")
		}
	}
}

func (s Server) createOrder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req orders.OrderReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.handleErr(r.Context(), w, ServerError{
				Code:    ErrCodeBadRequest,
				Message: fmt.Sprintf("invalid request payload: %s", err.Error()),
			})
			return
		}
		order, err := orders.Create(r.Context(), req, s.Orders, s.Products, s.Coupons)
		if err != nil {
			s.handleErr(r.Context(), w, err)
		}

		if err := json.NewEncoder(w).Encode(order); err != nil {
			s.Logger.ErrorContext(r.Context(), "failed to write createOrder response to client")
		}
	}
}

// handleErr implements standard route error handling including logging and obfuscation.
func (s Server) handleErr(ctx context.Context, w http.ResponseWriter, err error) {
	// log the original error before we possible obscure it as an iternal sever error.
	s.Logger.ErrorContext(ctx, err.Error())
	var se ServerError
	if !errors.As(err, &se) {
		se = appErrToServer(err)
	}
	w.WriteHeader(se.StatusCode())
	if err := json.NewEncoder(w).Encode(se); err != nil {
		s.Logger.ErrorContext(ctx, "failed to write error response: "+err.Error())
	}
}
