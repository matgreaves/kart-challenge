package orders

import (
	"context"
	"errors"
	"fmt"

	"github.com/matgreaves/kart-challenge/api/apperr"
	"github.com/matgreaves/kart-challenge/api/coupons"
	"github.com/matgreaves/kart-challenge/api/products"
)

// Store is the interface for interacting with order data.
type Store interface {
	Create(ctx context.Context, req Order) (Order, error)
}

// Order defines the data model for an order.
type Order struct {
	ID       string             `json:"id,omitempty"`
	Items    []OrderItem        `json:"items,omitempty"`
	Products []products.Product `json:"products,omitempty"`
	// The example servers includes this field but I've removed it as it doesn't exist
	// in the OpenAPI spec.
	// CouponCode string `json:"couponCode,omitempty"`
}

// OrderItem is a single product within an [Order].
type OrderItem struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

// OrderReq Place a new order
type OrderReq struct {
	// CouponCode Optional promo code applied to the order
	CouponCode string      `json:"couponCode,omitempty"`
	Items      []OrderItem `json:"items"`
}

// Validate checks whether o is well formed.
func (o *OrderReq) Validate() error {
	ve := []error{}
	if len(o.Items) == 0 {
		ve = append(ve, errors.New("at least one item is required"))
	}
	for i, v := range o.Items {
		if v.ProductID == "" {
			ve = append(ve, fmt.Errorf("item[%d] productId is required", i))
		}
		// NOTE: similar error message as example server, but the example server returns that
		// error on quantity == 0 and not on < 0. Using logic in the spirit of the error message
		// rather than the observed behaviour.
		if v.Quantity < 0 {
			ve = append(ve, fmt.Errorf("item[%d] quantity cannot be less than zero", i))
		}
	}
	if len(ve) > 0 {
		return errors.Join(ve...)
	}
	return nil
}

// Create takes an [OrderReq], validates it, and persists it returning the persisted [Order].
func Create(ctx context.Context, req OrderReq, os Store, ps products.Store, cs coupons.Store) (Order, error) {
	if err := req.Validate(); err != nil {
		return Order{}, apperr.NewError(apperr.CodeConstraint, err)
	}
	if req.CouponCode != "" && !cs.Has(req.CouponCode) {
		return Order{}, apperr.NewError(apperr.CodeConstraint, errors.New("invalid couponCode specified"))
	}
	order := Order{
		Items: req.Items,
	}
	var err error
	order.Products, err = productsForItems(ctx, order.Items, ps)
	if err != nil {
		return Order{}, err
	}
	return os.Create(ctx, order)
}

func productsForItems(ctx context.Context, items []OrderItem, ps products.Store) ([]products.Product, error) {
	p := make([]products.Product, 0, len(items))
	for _, v := range items {
		prod, err := ps.Get(ctx, v.ProductID)
		if err != nil {
			fmt.Println("error")
			var ae apperr.Error
			if errors.As(err, &ae) && ae.Code == apperr.CodeNotFound {
				return nil, apperr.NewError(apperr.CodeConstraint, errors.New("invalid product specified"))
			}
			return nil, fmt.Errorf("failed to fill products: %w", err)
		}
		p = append(p, prod)
	}
	return p, nil
}
