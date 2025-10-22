package products

import (
	"encoding/json"
	"testing"

	"github.com/matgreaves/kart-challenge/api/apperr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testProducts = []Product{
	{
		ID:       "1",
		Category: "Cake",
		Name:     "Black Forest",
		Price:    7.5,
	},
	{
		ID:       "2",
		Category: "Cake",
		Name:     "Carrot",
		Price:    8,
	},
	{
		ID:       "3",
		Category: "Cake",
		Name:     "Red Velvet",
		Price:    5,
	},
}

func testSlice(t *testing.T, products []Product) Slice {
	t.Helper()
	b, err := json.Marshal(products)
	require.NoError(t, err)
	return NewSlice(b)
}

func TestSlice_Get(t *testing.T) {
	t.Parallel()
	t.Run("has product", func(t *testing.T) {
		t.Parallel()

		s := testSlice(t, testProducts)
		p, err := s.Get(t.Context(), testProducts[0].ID)
		require.NoError(t, err)
		assert.Equal(t, testProducts[0], p)
	})

	t.Run("no product", func(t *testing.T) {
		t.Parallel()

		s := NewSlice(SampleData)
		_, err := s.Get(t.Context(), "9001")
		ae, ok := err.(apperr.Error)
		require.True(t, ok, "err must be an app error")
		assert.Equal(t, apperr.CodeNotFound, ae.Code)
		assert.ErrorContains(t, err, "product 9001 not found")
	})
}

func TestSlice_List(t *testing.T) {
	t.Parallel()

	t.Run("initial page", func(t *testing.T) {
		t.Parallel()
		s := testSlice(t, testProducts)
		p, err := s.List(t.Context(), 0, 1)
		require.NoError(t, err)
		assert.Len(t, p, 1)
		assert.Equal(t, testProducts[0:1], p)
	})

	t.Run("subsequent page", func(t *testing.T) {
		t.Parallel()
		s := testSlice(t, testProducts)
		p, err := s.List(t.Context(), 1, 1)
		require.NoError(t, err)
		assert.Len(t, p, 1)
		assert.Equal(t, testProducts[1:2], p)
	})

	t.Run("pageSize > len(items)", func(t *testing.T) {
		t.Parallel()
		s := testSlice(t, testProducts)
		p, err := s.List(t.Context(), 0, len(testProducts)+1)
		require.NoError(t, err)
		assert.Len(t, p, len(testProducts))
		assert.Equal(t, testProducts[0:], p)
	})

	t.Run("page and pageSize > len(items)", func(t *testing.T) {
		t.Parallel()
		s := testSlice(t, testProducts)
		p, err := s.List(t.Context(), len(testProducts)+1, len(testProducts)+1)
		require.NoError(t, err)
		assert.Len(t, p, 0)
		assert.Equal(t, []Product{}, p)
	})
}
