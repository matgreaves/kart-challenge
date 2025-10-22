package products

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/matgreaves/kart-challenge/api/apperr"
)

//go:embed data.json
var SampleData []byte

// Product defines model for Product.
type Product struct {
	Category string  `json:"category,omitempty"`
	ID       string  `json:"id,omitempty"`
	Name     string  `json:"name,omitempty"`
	Price    float32 `json:"price,omitempty"`
	// note: demo server responses include an image field. Leaving off to match
	// the OpenAPI spec but might be missing.
}

// Store contains the methods for interacting with a store of product data.
type Store interface {
	Get(ctx context.Context, id string) (Product, error)
	List(ctx context.Context, page, pageSize int) ([]Product, error)
}

// NewSlice creates a [Slice] from r where b contains a JSON encoded
// list of slices. Useful for test data.
//
// Panics if b does not contain the expected data, fails tests fast and saves on
// assertions.
func NewSlice(b []byte) Slice {
	s := Slice{}
	if err := json.Unmarshal(b, &s); err != nil {
		panic(err)
	}
	return s
}

var _ Store = Slice{}

// Slice is a [Store] backed by a static slice of [Product]. Useful for testing.
type Slice []Product

// Get implements [Store.Get].
func (s Slice) Get(_ context.Context, id string) (Product, error) {
	for _, v := range s {
		if v.ID == id {
			return v, nil
		}
	}
	return Product{}, apperr.NewError(apperr.CodeNotFound, fmt.Errorf("product %s not found", id))
}

// List implements [Store.List].
func (s Slice) List(_ context.Context, page, pageSize int) ([]Product, error) {
	if page < 0 {
		return nil, apperr.Error{
			Code:  apperr.CodeValidation,
			Cause: errors.New("page must be zero or greater"),
		}
	}
	if pageSize < 1 {
		return nil, apperr.Error{
			Code:  apperr.CodeValidation,
			Cause: errors.New("pageSize must be greater than zero"),
		}
	}
	return s[min(pageSize*page, len(s)):min((pageSize*page)+pageSize, len(s))], nil
}
