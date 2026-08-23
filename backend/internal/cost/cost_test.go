package cost

import (
	"sync"
	"testing"

	"github.com/xavskye/gorag/internal/model"
)

func TestBudgetBlocks(t *testing.T) {
	l := New(1.0)
	if err := l.Allow(0.2); err != nil {
		t.Fatal(err)
	}
	l.Record(model.CostRecord{Provider: "openai", CNY: 0.9, OK: true})
	if err := l.Allow(0.2); err == nil {
		t.Fatal("expected budget error")
	}
}

func TestConcurrentRecordMaintainsBudget(t *testing.T) {
	const (
		workers   = 24
		perWorker = 20
	)
	l := New(workers * perWorker)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < perWorker; j++ {
				l.Record(model.CostRecord{Provider: "concurrent", CNY: 1, OK: true})
			}
		}()
	}
	wg.Wait()

	want := float64(workers * perWorker)
	if got := l.Spent(); got != want {
		t.Fatalf("spent = %.0f, want %.0f", got, want)
	}
	if err := l.Allow(1); err == nil {
		t.Fatal("expected the next request to be rejected after reaching the budget")
	}
}
