package cost

import (
	"math"
	"sync"
	"testing"

	"github.com/xavskye/gorag/internal/model"
)

func approxEq(a, b float64) bool {
	return math.Abs(a-b) <= 1e-6*math.Max(math.Abs(a), math.Abs(b))
}

func TestRecordConcurrentSpentConsistency(t *testing.T) {
	const goroutines = 64
	const perG = 100
	unit := 0.01
	l := New(float64(goroutines*perG) * unit) // exact budget, no overflow expected

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < perG; j++ {
				l.Record(model.CostRecord{Provider: "p", CNY: unit, OK: true})
			}
		}()
	}
	wg.Wait()

	want := float64(goroutines*perG) * unit
	if !approxEq(l.Spent(), want) {
		t.Fatalf("spent mismatch: got %v want %v (lost updates)", l.Spent(), want)
	}
	items := l.List()
	var sum float64
	for _, r := range items {
		sum += r.CNY
	}
	if len(items) != goroutines*perG {
		t.Fatalf("item count: got %d want %d", len(items), goroutines*perG)
	}
	// Critical invariant: cumulative spent must equal the detail sum.
	// Without holding the lock across the increment, concurrent writers drop
	// updates and spent ends up strictly less than the item sum.
	if !approxEq(sum, l.Spent()) {
		t.Fatalf("sum(items)=%v != spent=%v (accumulator drifted from detail)", sum, l.Spent())
	}
}

func TestAllowBlocksAtLimitAfterConcurrentRecords(t *testing.T) {
	const n = 50
	unit := 0.02
	// budget exactly equals sum of all records, so the next Allow must fail
	l := New(float64(n)*unit + unit - 1e-9)

	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			l.Record(model.CostRecord{Provider: "p", CNY: unit, OK: true})
		}()
	}
	wg.Wait()

	if err := l.Allow(unit); err == nil {
		t.Fatalf("expected budget exceeded after %v spent, got nil; spent=%v", float64(n)*unit, l.Spent())
	}
}
