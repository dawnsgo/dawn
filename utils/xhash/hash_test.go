package xhash_test

import (
	"github.com/dawnsgo/dawn/v2/utils/xhash"
	"testing"
)

func TestSHA256(t *testing.T) {
	t.Log(xhash.SHA256("abc"))
}
