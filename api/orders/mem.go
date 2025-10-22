package orders

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

var _ Store = Mem{}

// NewMem creates a new [Mem] store.
func NewMem() Store {
	mu := sync.RWMutex{}
	return Mem{
		data: map[string]Order{},
		mu:   &mu,
	}
}

// Mem is a [Store] that stores created orders is a simple in memory map.
type Mem struct {
	data map[string]Order
	mu   *sync.RWMutex
}

// Create implements [Store.Create].
//
// Orders products are denormalised and stored alongside the order for better traceability though
// another possible option would be to fill at read time to allow fixing of product data.
func (m Mem) Create(ctx context.Context, o Order) (Order, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return Order{}, fmt.Errorf("orders failed to create uuid: %w", err)
	}
	o.ID = id.String()
	m.mu.Lock()
	m.data[o.ID] = o
	m.mu.Unlock()
	return o, nil
}
