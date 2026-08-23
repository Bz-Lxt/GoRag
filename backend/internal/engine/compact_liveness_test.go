package engine

import (
	"fmt"
	"testing"
	"time"
)

func TestCompactReturnsWithMultipleLiveSegments(t *testing.T) {
	e, err := Open(testCfg(t))
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 2; i++ {
		if _, err := e.IngestDocument(IngestDocReq{
			Collection: "default",
			Content:    fmt.Sprintf("live document %d for segment compaction", i),
		}); err != nil {
			t.Fatal(err)
		}
		if err := e.Flush("test"); err != nil {
			t.Fatal(err)
		}
	}

	deadline := time.Now().Add(2 * time.Second)
	for len(e.Seg.List()) < 2 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if got := len(e.Seg.List()); got < 2 {
		t.Fatalf("expected two persisted segments, got %d", got)
	}

	done := make(chan error, 1)
	go func() { done <- e.Compact() }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("compact: %v", err)
		}
		e.Close()
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Compact did not return")
	}
}
