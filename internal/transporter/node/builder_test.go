package node_test

import (
	"context"
	"testing"

	"github.com/dawnsgo/dawn/cluster"
	"github.com/dawnsgo/dawn/core/buffer"
	"github.com/dawnsgo/dawn/internal/transporter/node"
	"github.com/dawnsgo/dawn/utils/xuuid"
)

func TestBuilder(t *testing.T) {
	builder := node.NewBuilder(&node.Options{
		InsID:   xuuid.UUID(),
		InsKind: cluster.Gate,
	})

	client, err := builder.Build("127.0.0.1:49898")
	if err != nil {
		t.Fatal(err)
	}

	err = client.Deliver(context.Background(), 1, 2, buffer.NewNocopyBuffer([]byte("hello world")))
	if err != nil {
		t.Fatal(err)
	}
}
