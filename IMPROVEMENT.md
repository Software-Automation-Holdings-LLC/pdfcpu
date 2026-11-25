# Improvement: own caller-selected filesystem lifecycles and preserve parser limits

## Embedded-file close errors

`NewEmbeddedFileStreamDict` owns the opened file, including its deferred close. Its named error result now joins a close error with any existing stat/stream error, following the existing `ReadFile` and configuration-file boundary pattern. Public parameter/result types, returned stream reference, streaming operation, allocation behavior and existing parser limits are unchanged. Focused public-path tests cover success, missing files and stream-read failure; errcheck verifies the unchecked close is gone. A real read-only regular-file close failure cannot be forced deterministically through this public API without adding a new injection seam, so fault injection is not claimed.

## Intent

Resolve the 23 whole-changed-file gosec findings without changing the shared
decode budget, suppressing warnings, or restricting intentionally arbitrary-path
library APIs. The original budget worktree and its commits remain intact.

## Inventory and ownership

- Eleven G115 findings in `pkg/pdfcpu/writeImage.go`: eight unnecessary int-to-byte
  shift-count conversions, two CMYK sample conversions, and one packed sample mask.
  Renderers already validate supported depths (1, 2, 4, 8, 16 for CMYK; 1, 2, 4, 8
  for indexed samples). Shift counts remain ints. Exact bit-depth masks and integer
  scale factors preserve CMYK sample values; the packed mask is now uint16-native.
  These are representational proof gaps, not evidence of a demonstrated exploit.
- Five G301 findings in `pkg/pdfcpu/model/configuration.go`: configuration, font
  and certificate directory creation used 0755. New directories now use 0700;
  MkdirAll leaves existing directory modes unchanged. Per-user resources need not
  be readable by other users. Deliberately sharing newly created configuration
  trees now requires the owner to grant access explicitly.
- Two G104 findings: `ReadFile` now joins close errors with read errors, and blank
  page insertion checks SetValid before appending the new reference. Existing
  error return forms are retained; previously dropped operational errors surface.

## Unchanged path-authority contract

The five caller-selected path operations now live in explicit filesystem lifecycle
owners, using the platform's existing canonical `facade.go` policy. The policy
allows G304 only at that boundary; all other security rules still apply. There is
no policy change, suppression, path restriction, alias or OS-call-only wrapper.

| Owner                        | Complete operation                                                                                                                                                   |
| ---------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `pkg/pdfcpu/facade.go`       | `ReadFile`: cancellation check, logging, open, parse and close, retaining joined close errors.                                                                       |
| `pkg/pdfcpu/model/facade.go` | `NewStreamDictForFile`: read file and construct stream; `NewEmbeddedFileStreamDict`: open, metadata, construct embedded stream and close.                            |
| `pkg/pdfcpu/model/facade.go` | Configuration file loading and initialization, read-only/default/isolated resource binding, private directories, resource setup and legacy loader-state publication. |

Public names, parameters, results, errors and resource-state transitions remain
unchanged. Reader parsing, cross-reference object construction and decode budgets
stay with their existing owners. Callers resolve the same package symbols
directly, without migration shims. The configuration lifecycle moves as a unit
so initialization, resource binding and compatibility publication retain one
owner. `pkg/pdfcpu/LAYOUT.md` records that boundary.

No trusted sandbox root exists in these APIs. The caller retains deliberate
filesystem authority; cleaning paths or opening `os.Root` at `/` would not
constrain it. This change makes that authority explicit, not safer by validation.

The inventory is five path sinks across three old files, now owned by two facade
files: one PDF input lifecycle, two file-to-stream lifecycles, and two opens in
the configuration tree lifecycle. Existing success, missing/invalid configuration,
read-only loading, reset, private permissions, cancellation and merge tests cover
the unchanged contracts. No new uncertain lifecycle behavior is introduced, so no
synthetic OS abstraction or redundant regression test is added.

## Proof

`TestImageSampleConversionBounds` goes through RenderImage and decodes PNG/TIFF
results for all changed component depths and indexed base spaces, with soft masks.
Existing CMYK depth/stride, malformed-input and aggregate-budget tests remain the
regression contract. Configuration tests check every directory-creating public
entrypoint and preservation of existing modes. The repository's full test gate is
`mise run test` (isolated HOME/XDG, serialized packages). Final evidence is reported
with the checkpoint. Publication requires a fresh review and the unchanged security
gate, including its existing canonical filesystem-facade policy.
