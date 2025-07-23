package sync

import "testing"

func TestNewSyncService(t *testing.T) {
	s := NewSyncService(nil, nil)
	if s == nil {
		t.Fatal("nil service")
	}
}
