package engine

import (
	"testing"

	"github.com/xavskye/gorag/internal/config"
	"github.com/xavskye/gorag/internal/model"
)

func TestOpenEngineCanCreateCollection(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	cfg.DataDir = t.TempDir()

	eng, err := Open(cfg)
	if err != nil {
		t.Fatalf("open engine: %v", err)
	}
	defer eng.Close()

	if err := eng.CreateCollection(model.Collection{
		Name:      "articles",
		Dim:       model.DefaultDim,
		Metric:    model.MetricCosine,
		IndexType: model.IndexHNSW,
	}); err != nil {
		t.Fatalf("create collection after successful open: %v", err)
	}
}
