package micro

import (
    "context"
    "testing"
)

func TestNewServer(t *testing.T) {
    srv := NewServer(context.Background())
    srv.Run()
}

