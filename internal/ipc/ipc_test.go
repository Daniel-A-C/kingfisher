package ipc

import (
	"path/filepath"
	"testing"
	"time"
)

func TestIPCRoundTrip(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "kf.sock")
	received := make(chan Message, 1)

	srv, err := StartServer(sock, func(m Message) {
		received <- m
	})
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	want := Message{Action: "open", Slug: "corvidae", Path: "/tmp/corvidae"}
	if err := Send(sock, want); err != nil {
		t.Fatal(err)
	}

	select {
	case got := <-received:
		if got != want {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for IPC message")
	}
}
