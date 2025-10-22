package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	glog "log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/matgreaves/kart-challenge/api/coupons"
	"github.com/matgreaves/kart-challenge/api/monitoring"
	"github.com/matgreaves/kart-challenge/api/orders"
	"github.com/matgreaves/kart-challenge/api/products"
	"github.com/matgreaves/kart-challenge/api/server"
	"go.opentelemetry.io/otel"
)

const (
	// Address for the http server to listen on.
	DefaultAddress  = "0.0.0.0:8080"
	DefaultLogLevel = slog.LevelInfo
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	if err := run(ctx, os.Args[1:]); err != nil {
		glog.Fatal("run exited with error:", err)
	}
}

func run(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("", flag.ExitOnError)
	addr := flags.String("a", DefaultAddress, "host:port to listen on")
	if err := flags.Parse(args); err != nil {
		return err
	}

	// an example of setting up tracing here we're just discarding it but we will still create traces
	// and we can therefore see the traces in our logs.
	tp, err := monitoring.WriterTracerProvider(io.Discard)
	if err != nil {
		return fmt.Errorf("failed to contruct trace provider: %w", err)
	}
	otel.SetTracerProvider(tp)

	ps := products.NewSlice(products.SampleData)
	cs, err := coupons.NewMem(strings.NewReader(coupons.DB))
	ors := orders.NewMem()
	if err != nil {
		return err
	}
	return server.Server{
		Logger:   monitoring.NewJSONLogger(os.Stdout, DefaultLogLevel),
		Products: ps,
		Orders:   ors,
		Coupons:  cs,
		Auth:     server.TestAuth(),
		Addr:     *addr,
	}.Run(ctx)
}
