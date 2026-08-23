package engine

import (
	"testing"
	"time"
)

func TestStatsFlushHistoryIsSnapshot(t *testing.T) {
	e, err := Open(testCfg(t))
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()

	if _, err := e.IngestDocument(IngestDocReq{
		Collection: "default",
		Content:    "统计结果应当是调用时刻的独立快照。",
	}); err != nil {
		t.Fatal(err)
	}
	if err := e.Flush("manual"); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(2 * time.Second)
	var first string
	for time.Now().Before(deadline) {
		stats := e.Stats()
		if len(stats.FlushHistory) > 0 {
			first = stats.FlushHistory[0].Reason
			stats.FlushHistory[0].Reason = "client annotation"
			break
		}
		time.Sleep(time.Millisecond)
	}
	if first == "" {
		t.Fatal("flush history was not published")
	}

	got := e.Stats().FlushHistory[0].Reason
	if got != first {
		t.Fatalf("a caller mutated later stats: got reason %q, want %q", got, first)
	}
}
