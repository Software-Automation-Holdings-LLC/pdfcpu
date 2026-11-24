/*
Copyright 2018 The pdfcpu Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pdfcpu

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/pdfcpu/pdfcpu/internal/contextutil"
	"github.com/pdfcpu/pdfcpu/pkg/log"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// ReadFile reads a PDF file and builds the internal model context holding its cross reference table.
// If the passed Go context is canceled, reading will be interrupted.
func ReadFile(c context.Context, inFile string, conf *model.Configuration) (ctx *model.Context, err error) {
	if err := contextutil.Check(c); err != nil {
		return nil, err
	}

	if log.InfoEnabled() {
		log.Info.Printf("reading %s..\n", inFile)
	}

	f, err := os.Open(inFile)
	if err != nil {
		return nil, fmt.Errorf("can't open %q: %w", inFile, err)
	}

	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close %q: %w", inFile, closeErr))
		}
	}()

	return Read(c, f, conf)
}
