# SAH explicit decoded-buffer budget

Maintainer: Software Automation Holdings LLC. Base: upstream pdfcpu v0.16.4.
This fork adds `ResourceLimits.MaxTotalDecodeBytes`, a positive explicit shared
budget across ordinary XRefTable decoding and lazy object-stream decoding. Zero
preserves upstream's legacy separate cumulative allowances. Original Apache-2.0
license and copyright remain unchanged. No customer fixtures are included.

The account API pins this fork by immutable Go module revision, not a branch or
local replace. It configures 32 MiB combined retained decoding and 8 MiB per stream.
Upgrade ownership includes re-running mixed-stream and public API multi-object
stream regressions before replacing the pinned revision. This patch has not been
accepted upstream. Upstream's contribution policy disallows substantially
AI-authored code submissions; no upstream issue or pull request was submitted.

Focused proof: `go test ./pkg/pdfcpu ./pkg/pdfcpu/model -run TestExplicitTotalDecodeBudget`.
Regression: `go test ./internal/streamlimit ./pkg/filter`.
