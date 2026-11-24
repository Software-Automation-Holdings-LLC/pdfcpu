package pdfcpu

import (
	"bytes"
	"errors"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/filter"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func TestExplicitTotalDecodeBudgetImages(t *testing.T) {
	conf := model.NewStatelessConfiguration()
	conf.Limits.MaxDecodeBytes = 1024
	conf.Limits.MaxTotalDecodeBytes = 1024
	stream := func() types.StreamDict {
		sd := types.StreamDict{Dict: types.NewDict(), Content: bytes.Repeat([]byte("x"), 600), FilterPipeline: []types.PDFFilter{{Name: filter.Flate}}}
		sd.InsertInt("Width", 600)
		sd.InsertInt("Height", 1)
		sd.InsertInt("BitsPerComponent", 8)
		sd.InsertName("ColorSpace", model.DeviceGrayCS)
		if err := sd.Encode(); err != nil {
			t.Fatal(err)
		}
		sd.Content = nil
		return sd
	}
	for _, image := range []bool{false, true} {
		table := &model.XRefTable{Conf: conf}
		first, second := stream(), stream()
		var err error
		if image {
			err = decodeImage(&model.Context{XRefTable: table}, &first, filter.Flate, filter.Flate, 1)
		} else {
			_, err = streamBytes(table, &first, "synthetic support")
		}
		if err != nil {
			t.Fatal(err)
		}
		budget := newObjectStreamDecodeBudget(conf)
		budget.document = table
		if err = budget.decode(&second, 1024); !errors.Is(err, filter.ErrDecodeLimitExceeded) {
			t.Fatalf("image=%v bypassed total budget: %v", image, err)
		}
	}
}
