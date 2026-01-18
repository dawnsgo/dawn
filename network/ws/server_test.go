// Package ws_test 提供 WebSocket 服务器功能的测试。
package ws_test

import (
	"net/http"
	"testing"

	"github.com/dawnsgo/dawn/network/ws"
	"github.com/dawnsgo/dawn/core/buffer"
	"github.com/dawnsgo/dawn/log"
	"github.com/dawnsgo/dawn/network"
	"github.com/dawnsgo/dawn/packet"
	"github.com/dawnsgo/dawn/utils/xcall"
)

func TestServer(t *testing.T) {
	server := ws.NewServer()
	server.OnStart(func() {
		t.Logf("server is started")
	})
	server.OnConnect(func(conn network.Conn) {
		t.Logf("connection is opened, connection id: %d", conn.ID())
	})
	server.OnDisconnect(func(conn network.Conn) {
		t.Logf("connection is closed, connection id: %d", conn.ID())
	})
	server.OnReceive(func(conn network.Conn, buf buffer.Buffer) {
		defer buf.Release()

		message, err := packet.UnpackMessage(buf.Bytes())
		if err != nil {
			t.Error(err)
			return
		}

		t.Logf("receive msg from client, connection id: %d, seq: %d, route: %d, msg: %s", conn.ID(), message.Seq, message.Route, string(message.Buffer))

		msg, err := packet.PackMessage(&packet.Message{
			Seq:    1,
			Route:  1,
			Buffer: []byte("I'm fine~~"),
		})
		if err != nil {
			t.Fatal(err)
		}

		if err = conn.Push(msg); err != nil {
			t.Error(err)
		}
	})
	server.OnUpgrade(func(w http.ResponseWriter, r *http.Request) (allowed bool) {
		return true
	})

	if err := server.Start(); err != nil {
		t.Fatal(err)
	}

	xcall.Go(func() {
		err := http.ListenAndServe(":8089", nil)
		if err != nil {
			log.Errorf("pprof server start failed: %v", err)
		}
	})

	select {}
}

func TestServer_Benchmark(t *testing.T) {
	server := ws.NewServer()
	server.OnStart(func() {
		t.Logf("server is started")
	})
	server.OnReceive(func(conn network.Conn, buf buffer.Buffer) {
		defer buf.Release()

		_, err := packet.UnpackMessage(buf.Bytes())
		if err != nil {
			t.Error(err)
			return
		}

		msg, err := packet.PackMessage(&packet.Message{
			Seq:    1,
			Route:  1,
			Buffer: []byte("I'm fine~~"),
		})
		if err != nil {
			t.Fatal(err)
		}

		if err = conn.Push(msg); err != nil {
			t.Error(err)
		}
	})
	server.OnUpgrade(func(w http.ResponseWriter, r *http.Request) (allowed bool) {
		return true
	})

	if err := server.Start(); err != nil {
		t.Fatal(err)
	}

	xcall.Go(func() {
		err := http.ListenAndServe(":8089", nil)
		if err != nil {
			log.Errorf("pprof server start failed: %v", err)
		}
	})

	select {}
}
