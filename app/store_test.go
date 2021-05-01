package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_WriteRead(t *testing.T) {
	st := newStore()
	st.Write("foo", "bar", "bee")
	v, ok := st.Read("foo", "bar")
	assert.True(t, ok)
	assert.Equal(t, "bee", v)
}

func Test_ReadAll(t *testing.T) {
	st := newStore()
	st.Write("foo", "bar", "bee")
	st.Write("foo", "bar2", "bee2")
	st.Write("foo", "bar3", "bee3")
	table, ok := st.ReadAll("foo")
	assert.True(t, ok)
	assert.Len(t, table, 3)
	assert.Contains(t, table, "bar")
	assert.Contains(t, table, "bar2")
	assert.Contains(t, table, "bar3")
}
