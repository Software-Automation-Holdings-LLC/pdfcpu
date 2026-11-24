package model

import (
	"errors"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/filter"
)

func TestExplicitTotalDecodeBudget(t *testing.T) {
	conf := NewStatelessConfiguration()
	conf.Limits.MaxDecodeBytes = 1024
	conf.Limits.MaxTotalDecodeBytes = 1024
	table := newXRefTable(conf)
	first := decodeLimitTestStream(t, true, false)
	if err := table.DecodeStream(&first); err != nil {
		t.Fatal(err)
	}
	used := table.decodedStreamBudget.used
	if used <= 0 || used > 1024 {
		t.Fatalf("charged=%d", used)
	}
	// Cached content must not be charged twice.
	if err := table.DecodeStream(&first); err != nil || table.decodedStreamBudget.used != used {
		t.Fatal("double charged cached bytes")
	}
	second := decodeLimitTestStream(t, true, false)
	// Leave less than the decoded stream's actual byte count.
	conf.Limits.MaxTotalDecodeBytes = used + 31
	if err := table.DecodeStreamWithinBudget(&second, 1024); !errors.Is(err, filter.ErrDecodeLimitExceeded) {
		t.Fatalf("shared stream budget: %v", err)
	}
	if second.Content != nil {
		t.Fatal("retained over-budget bytes")
	}
	// Existing default behavior is unchanged without an explicit total limit.
	conf.Limits.MaxTotalDecodeBytes = 0
	if err := table.DecodeStream(&second); err != nil {
		t.Fatalf("legacy default: %v", err)
	}
}
