# Improvement: make image bounds, private directories and errors explicit

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

Five G304 findings remain deliberately visible; this is not a green security gate:

| File and baseline line                  | Authority                                                                 |
| --------------------------------------- | ------------------------------------------------------------------------- |
| `pkg/pdfcpu/read.go:149`                | Public ReadFile accepts caller-selected input file.                       |
| `pkg/pdfcpu/model/xreftable.go:524`     | NewStreamDictForFile embeds a caller-selected file.                       |
| `pkg/pdfcpu/model/xreftable.go:832`     | NewEmbeddedFileStreamDict embeds a caller-selected file.                  |
| `pkg/pdfcpu/model/configuration.go:616` | Configuration loading opens the selected root's config.yml.               |
| `pkg/pdfcpu/model/configuration.go:662` | Legacy configuration initialization opens the selected root's config.yml. |

No trusted sandbox root exists in these APIs. Cleaning paths, opening os.Root at
`/`, moving calls, or adding fake validation would not constrain authority. A real
sandbox would require a library/CLI contract change and consumer migration. A
reviewed path-policy exception would instead require an explicit gate-policy
decision. Neither change belongs in this narrowly scoped dependency patch.
No suppression, hook bypass, policy change, or parser rewrite is included.

## Proof

`TestImageSampleConversionBounds` goes through RenderImage and decodes PNG/TIFF
results for all changed component depths and indexed base spaces, with soft masks.
Existing CMYK depth/stride, malformed-input and aggregate-budget tests remain the
regression contract. Configuration tests check every directory-creating public
entrypoint and preservation of existing modes. The repository's full test gate is
`mise run test` (isolated HOME/XDG, serialized packages). Final evidence is reported
with the checkpoint; publication remains blocked if the unchanged security gate
reports the intentional path APIs.
