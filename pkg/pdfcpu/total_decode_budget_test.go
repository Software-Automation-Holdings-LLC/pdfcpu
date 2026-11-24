package pdfcpu

import (
	"bytes"
	"errors"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/filter"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func TestExplicitTotalDecodeBudget(t *testing.T) {
	conf := model.NewStatelessConfiguration()
	conf.Limits.MaxDecodeBytes = 1024
	conf.Limits.MaxTotalDecodeBytes = 1024
	table := &model.XRefTable{Conf: conf}
	budget := newObjectStreamDecodeBudget(conf)
	budget.document = table
	stream := func() types.StreamDict {
		sd := types.StreamDict{Dict: types.NewDict(), Content: bytes.Repeat([]byte("x"), 600), FilterPipeline: []types.PDFFilter{{Name: filter.Flate}}}
		if err := sd.Encode(); err != nil {
			t.Fatal(err)
		}
		sd.Content = nil
		return sd
	}
	first := stream()
	if err := table.DecodeStream(&first); err != nil {
		t.Fatal(err)
	}
	second := stream()
	if err := budget.decode(&second, 1024); !errors.Is(err, filter.ErrDecodeLimitExceeded) {
		t.Fatalf("object streams did not share ordinary-stream budget: %v", err)
	}
	if second.Content != nil {
		t.Fatal("retained rejected content")
	}
	// Reverse ordering also shares the same document budget.
	table = &model.XRefTable{Conf: conf}
	budget.document = table
	first = stream()
	second = stream()
	if err := budget.decode(&first, 1024); err != nil {
		t.Fatal(err)
	}
	if err := table.DecodeStream(&second); !errors.Is(err, filter.ErrDecodeLimitExceeded) {
		t.Fatalf("ordinary stream bypassed object budget: %v", err)
	}
}
