package coupons

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMem_Has(t *testing.T) {
	m, err := NewMem(strings.NewReader(DB))
	require.NoError(t, err)
	assert.True(t, m.Has("OVER9000"))
	assert.False(t, m.Has("UNDER9000"))
}
