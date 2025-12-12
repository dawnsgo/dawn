package stat_test

import (
	"testing"

	"github.com/dawnsgo/dawn/v2/core/stat"
)

func TestStat(t *testing.T) {
	fi, err := stat.Stat("stat_linux.go")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(fi.CreateTime())
}
