package vars

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVarsSet(t *testing.T) {
	t.Parallel()

	v := Vars{}
	v.Set("foo", "bar")
	v.Set("baz", "biz")

	assert.Equal(t, "bar", map[string]string(v)["foo"])
	assert.Equal(t, "biz", map[string]string(v)["baz"])
}
