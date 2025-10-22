// package coupons contins
package coupons

import (
	"bufio"
	_ "embed"
	"fmt"
	"io"
)

//go:embed data
var DB string

// Store is the interface for interacting with coupon data.
type Store interface {
	Has(code string) bool
}

var _ Store = Mem{}

// NewMem creates a [Mem] store reading data from r.
//
// r is expected to contain newline separated coupon codes.
func NewMem(r io.Reader) (Store, error) {
	m := Mem{}
	s := bufio.NewScanner(r)
	for s.Scan() {
		m[s.Text()] = struct{}{}
	}
	if err := s.Err(); err != nil {
		return nil, fmt.Errorf("failed to fill coupons store: %w", err)
	}
	return m, nil
}

// Mem is a [Store] implemented on an in memory map.
type Mem map[string]struct{}

// Has implements [Store.Has]
func (m Mem) Has(code string) bool {
	_, has := m[code]
	return has
}
