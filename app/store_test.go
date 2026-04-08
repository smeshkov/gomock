package app //nolint:testpackage // testing unexported store internals

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_WriteRead(t *testing.T) {
	t.Parallel()

	store := newStore()
	store.Write("foo", "bar", "bee")

	val, ok := store.Read("foo", "bar")
	assert.True(t, ok)
	assert.Equal(t, "bee", val)
}

func Test_ReadAll(t *testing.T) {
	t.Parallel()

	store := newStore()
	store.Write("foo", "bar", "bee")
	store.Write("foo", "bar2", "bee2")
	store.Write("foo", "bar3", "bee3")

	table, ok := store.ReadAll("foo")
	assert.True(t, ok)
	assert.Len(t, table, 3)
	assert.Contains(t, table, "bar")
	assert.Contains(t, table, "bar2")
	assert.Contains(t, table, "bar3")
}
