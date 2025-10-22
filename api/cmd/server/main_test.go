// main contains a series of blackbox tests that test our full running application.
//
// This guarantees as much as possible that the application will work as expected in production
// at least as far as we've guaranteed in our tests, see [The Beyoncé Rule](https://abseil.io/resources/swe-book/html/ch11.html)
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/matgreaves/kart-challenge/api/orders"
	"github.com/matgreaves/kart-challenge/api/products"
	"github.com/matgreaves/kart-challenge/api/server"
	grun "github.com/matgreaves/run"
	exp "github.com/matgreaves/run/exp"
	"github.com/matgreaves/run/exp/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListProduct(t *testing.T) {
	t.Parallel()

	addr, close := startServer(t)
	defer noErr(t, close)

	res, err := http.Get("http://" + addr + "/product")
	require.NoError(t, err)
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	b, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	var got []products.Product
	err = json.Unmarshal(b, &got)
	require.NoError(t, err)
	expected, err := products.NewSlice(products.SampleData).List(t.Context(), 0, 100)
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestGetProduct(t *testing.T) {
	t.Parallel()

	// useful for validating product respones
	productStore := products.NewSlice(products.SampleData)
	t.Run("product exists", func(t *testing.T) {
		t.Parallel()

		addr, close := startServer(t)
		defer noErr(t, close)

		res, err := http.Get("http://" + addr + "/product/1")
		require.NoError(t, err)
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)

		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		var got products.Product
		err = json.Unmarshal(b, &got)
		require.NoError(t, err)
		want, err := productStore.Get(t.Context(), "1")
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("product missing", func(t *testing.T) {
		t.Parallel()

		addr, close := startServer(t)
		defer noErr(t, close)

		res, err := http.Get("http://" + addr + "/product/9999")
		require.NoError(t, err)
		defer res.Body.Close()
		assert.Equal(t, http.StatusNotFound, res.StatusCode)

		var se server.ServerError
		err = json.NewDecoder(res.Body).Decode(&se)
		require.NoError(t, err)
		assert.Equal(t, server.ServerError{Code: server.ErrCodeNotFound, Message: "product 9999 not found"}, se)
	})
}

func TestCreateOrder(t *testing.T) {
	t.Parallel()

	t.Run("no token", func(t *testing.T) {
		t.Parallel()
		addr, close := startServer(t)
		defer noErr(t, close)

		res, err := http.Post("http://"+addr+"/order", "application/json", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)

		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.Len(t, b, 0)
	})

	t.Run("expired token", func(t *testing.T) {
		t.Parallel()
		addr, close := startServer(t)
		defer noErr(t, close)

		req, err := http.NewRequest(http.MethodPost, "http://"+addr+"/order", nil)
		require.NoError(t, err)
		req.Header.Set(server.APIKeyHeader, "toolate")
		require.NoError(t, err)
		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, res.StatusCode)

		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.Len(t, b, 0)
	})

	t.Run("token not valid yet", func(t *testing.T) {
		t.Parallel()
		addr, close := startServer(t)
		defer noErr(t, close)

		req, err := http.NewRequest(http.MethodPost, "http://"+addr+"/order", nil)
		require.NoError(t, err)
		req.Header.Set(server.APIKeyHeader, "tooearly")
		require.NoError(t, err)
		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, res.StatusCode)

		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.Len(t, b, 0)
	})

	t.Run("token missing required scope", func(t *testing.T) {
		t.Parallel()
		addr, close := startServer(t)
		defer noErr(t, close)

		req, err := http.NewRequest(http.MethodPost, "http://"+addr+"/order", nil)
		require.NoError(t, err)
		req.Header.Set(server.APIKeyHeader, "noscope")
		require.NoError(t, err)
		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, res.StatusCode)

		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.Len(t, b, 0)
	})

	productStore := products.NewSlice(products.SampleData)
	t.Run("no coupon", func(t *testing.T) {
		addr, close := startServer(t)
		defer noErr(t, close)

		req, err := http.NewRequest(http.MethodPost, "http://"+addr+"/order", bytes.NewReader(goodOrderBytes(t)))
		req.Header.Set(server.APIKeyHeader, "apitest")
		require.NoError(t, err)
		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)

		var o orders.Order
		err = json.NewDecoder(res.Body).Decode(&o)
		require.NoError(t, err)

		assert.NotEmpty(t, o.ID)
		assert.Equal(t, o.Items, goodOrder().Items)
		prod, err := productStore.Get(t.Context(), goodOrder().Items[0].ProductID)
		require.NoError(t, err)
		assert.Equal(t, []products.Product{prod}, o.Products)
	})

	t.Run("valid coupon", func(t *testing.T) {
		addr, close := startServer(t)
		defer noErr(t, close)

		or := goodOrder()
		or.CouponCode = "OVER9000"
		b, err := json.Marshal(or)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "http://"+addr+"/order", bytes.NewReader(b))
		require.NoError(t, err)
		req.Header.Set(server.APIKeyHeader, "apitest")
		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)

		var o orders.Order
		err = json.NewDecoder(res.Body).Decode(&o)
		require.NoError(t, err)

		assert.NotEmpty(t, o.ID)
		assert.Equal(t, o.Items, goodOrder().Items)
		prod, err := productStore.Get(t.Context(), goodOrder().Items[0].ProductID)
		require.NoError(t, err)
		assert.Equal(t, []products.Product{prod}, o.Products)
	})

	t.Run("coupon too short", func(t *testing.T) {
		addr, close := startServer(t)
		defer noErr(t, close)

		or := goodOrder()
		or.CouponCode = "123"
		b, err := json.Marshal(or)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "http://"+addr+"/order", bytes.NewReader(b))
		require.NoError(t, err)
		req.Header.Set(server.APIKeyHeader, "apitest")
		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()
		assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)

		var se server.ServerError
		err = json.NewDecoder(res.Body).Decode(&se)
		require.NoError(t, err)
		assert.Equal(t, server.ServerError{Code: server.ErrCodeConstraint, Message: "invalid couponCode specified"}, se)
	})

	t.Run("missing productID", func(t *testing.T) {
		addr, close := startServer(t)
		defer noErr(t, close)

		or := goodOrder()
		or.Items[0].ProductID = ""
		b, err := json.Marshal(or)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "http://"+addr+"/order", bytes.NewReader(b))
		require.NoError(t, err)
		req.Header.Set(server.APIKeyHeader, "apitest")
		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()
		assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)

		var se server.ServerError
		err = json.NewDecoder(res.Body).Decode(&se)
		require.NoError(t, err)
		assert.Equal(t, server.ServerError{Code: server.ErrCodeConstraint, Message: "item[0] productId is required"}, se)
	})

	t.Run("product doesn't exist", func(t *testing.T) {
		addr, close := startServer(t)
		defer noErr(t, close)

		or := goodOrder()
		or.Items[0].ProductID = "9999"
		b, err := json.Marshal(or)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "http://"+addr+"/order", bytes.NewReader(b))
		require.NoError(t, err)
		req.Header.Set(server.APIKeyHeader, "apitest")
		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()
		assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)

		var se server.ServerError
		err = json.NewDecoder(res.Body).Decode(&se)
		require.NoError(t, err)
		assert.Equal(t, server.ServerError{Code: server.ErrCodeConstraint, Message: "invalid product specified"}, se)
	})

	t.Run("quantity < 0", func(t *testing.T) {
		addr, close := startServer(t)
		defer noErr(t, close)

		or := goodOrder()
		or.Items[0].Quantity = -1
		b, err := json.Marshal(or)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "http://"+addr+"/order", bytes.NewReader(b))
		require.NoError(t, err)
		req.Header.Set(server.APIKeyHeader, "apitest")
		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()
		assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)

		var se server.ServerError
		err = json.NewDecoder(res.Body).Decode(&se)
		require.NoError(t, err)
		assert.Equal(t, server.ServerError{Code: server.ErrCodeConstraint, Message: "item[0] quantity cannot be less than zero"}, se)
	})
}

func goodOrder() orders.OrderReq {
	return orders.OrderReq{
		Items: []orders.OrderItem{
			{
				ProductID: "1",
				Quantity:  1,
			},
		},
	}
}

func goodOrderBytes(t *testing.T) []byte {
	t.Helper()
	b, err := json.Marshal(goodOrder())
	require.NoError(t, err)
	return b
}

func startServer(t *testing.T) (addr string, close func() error) {
	t.Helper()
	addr, err := ports.Random(t.Context())
	require.NoError(t, err)
	err, close = grun.Start(t.Context(), toRun([]string{"-a", addr}), exp.Poller(addr, exp.PollHTTP))
	require.NoError(t, err)
	return addr, close
}

func toRun(args []string) grun.Runner {
	return grun.Func(func(ctx context.Context) error {
		return run(ctx, args)
	})
}

func noErr(t *testing.T, f func() error) {
	if err := f(); err != nil {
		t.Error(err)
	}
}
