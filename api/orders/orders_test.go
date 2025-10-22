package orders

import (
	"testing"

	"github.com/google/uuid"
	"github.com/matgreaves/kart-challenge/api/coupons"
	"github.com/matgreaves/kart-challenge/api/products"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testReq() OrderReq {
	return OrderReq{
		Items: []OrderItem{
			{
				ProductID: "1",
				Quantity:  1,
			},
		},
	}
}

func TestCreate(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		req := testReq()
		req.CouponCode = "OVER9000"
		ps := products.NewSlice(products.SampleData)

		o, err := Create(t.Context(), req, NewMem(), ps, coupons.Mem{"OVER9000": struct{}{}})
		require.NoError(t, err)

		assert.NoError(t, uuid.Validate(o.ID))
		assert.Equal(t, req.Items, o.Items)
		wantProduct, err := ps.Get(t.Context(), o.Items[0].ProductID)
		require.NoError(t, err)
		assert.Equal(t, []products.Product{wantProduct}, o.Products)
	})

	t.Run("no items", func(t *testing.T) {
		t.Parallel()
		req := testReq()
		req.Items = nil
		_, err := Create(t.Context(), req, nil, nil, nil)
		assert.ErrorContains(t, err, "at least one item is required")
	})

	t.Run("item missing productID", func(t *testing.T) {
		t.Parallel()
		req := testReq()
		req.Items[0].ProductID = ""
		_, err := Create(t.Context(), req, nil, nil, nil)
		assert.ErrorContains(t, err, "productId is required")
	})

	t.Run("coupon not in list", func(t *testing.T) {
		t.Parallel()
		req := testReq()
		req.CouponCode = "UNDER9000"
		_, err := Create(t.Context(), req, nil, nil, coupons.Mem{})
		assert.ErrorContains(t, err, "invalid couponCode specified")
	})

	t.Run("product doesn't exist", func(t *testing.T) {
		t.Parallel()
		req := testReq()
		req.Items[0].ProductID = "9999"
		_, err := Create(t.Context(), req, nil, products.NewSlice(products.SampleData), nil)
		assert.ErrorContains(t, err, "invalid product specified")
	})
}
