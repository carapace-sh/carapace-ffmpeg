# AGENTS.md

## Project Overview

Go library for parsing ffmpeg CLI argument streams, stream specifiers, filter graphs, and map values into ASTs, with completion support. Part of the [carapace-sh](https://github.com/carapace-sh) ecosystem (shell completion framework). The module path is `github.com/carapace-sh/carapace-ffmpeg`.

## Commands

### Build & Test

```sh
go test ./...                              # run all tests
go test ./pkg/streamspec/                   # streamspec tests only
go test ./pkg/filtergraph/                  # filtergraph tests only
go test ./pkg/mapvalue/                     # mapvalue tests only
go test ./pkg/argstream/                    # argstream tests only
go test ./pkg/completer/                    # completer tests only
go test ./pkg/probe/                        # probe tests only
go test ./pkg/actions/tools/ffmpeg/         # action tests only (requires ffmpeg on PATH)
go build ./...                              # build all packages
```

### CI Checks (mirrors `.github/workflows/go.yml`)

```sh
go build -v ./...                                       # build
go test -v -coverprofile=profile.cov ./...               # test with coverage
[ "$(gofmt -d -s . | tee -a /dev/stderr)" = "" ]         # format check (fails if any diffs)
staticcheck ./...                                        # lint
```

Both `gofmt` and `staticcheck` are enforced in CI. Do not skip them.

### Test Data Generation

```sh
go generate ./testdata/                     # generate test media files (requires ffmpeg on PATH)
```

Creates 7 files in `testdata/`: `multistream.mkv`, `subtitles.mkv`, `surround.mkv`, `pixfmt.mkv`, `audio_only.wav`, `attachment.mkv`, `tagged_audio.flac`. Tests that use these files will skip if they're missing.

### Debug CLI (`cmd/carapace-ffmpeg-debug/`)

```sh
go run ./cmd/carapace-ffmpeg-debug streamspec "a:1"                  # parse stream specifier, output AST as JSON
go run ./cmd/carapace-ffmpeg-debug streamspec-complete "a:"           # stream specifier completion context as JSON
go run ./cmd/carapace-ffmpeg-debug filtergraph "scale=1280:720"       # parse filtergraph, output AST as JSON
go run ./cmd/carapace-ffmpeg-debug filtergraph-complete "sca"         # filtergraph completion context as JSON
go run ./cmd/carapace-ffmpeg-debug mapvalue "0:v"                     # parse -map value, output AST as JSON
go run ./cmd/carapace-ffmpeg-debug mapvalue-complete "0:"             # -map value completion context as JSON
go run ./cmd/carapace-ffmpeg-debug argstream -- -i input.mp4 -c:v libx264 output.mp4  # parse arg stream as JSON
go run ./cmd/carapace-ffmpeg-debug argstream-complete -- -i input.mp4 -c:v            # argstream completion context as JSON
```

### Completer CLIs

```sh
# ffmpeg
go run ./cmd/carapace-ffmpeg _carapace spec                    # generate carapace spec YAML
go run ./cmd/carapace-ffmpeg _carapace export '' ''              # complete at empty position (JSON)
go run ./cmd/carapace-ffmpeg _carapace export '-c:v' '' '-c:v' 'libx'  # complete codec value (JSON)

# ffplay
go run ./cmd/carapace-ffplay _carapace spec
go run ./cmd/carapace-ffplay _carapace export '' ''

# ffprobe
go run ./cmd/carapace-ffprobe _carapace spec
go run ./cmd/carapace-ffprobe _carapace export '' ''
```

The `_carapace spec` command generates YAML that references the `man/` directory for extended descriptions. The spec + man pages together form the completion definition consumed by carapace.

## Architecture

Four CLIs, a shared completer package, a probe package, and four independent parser packages with carapace completion actions.

```
cmd/carapace-ffmpeg/          Completer CLI for ffmpeg
cmd/carapace-ffplay/          Completer CLI for ffplay
cmd/carapace-ffprobe/         Completer CLI for ffprobe
cmd/carapace-ffmpeg-debug/    Debug/diagnostic CLI (JSON output)
pkg/argstream/                 Argument stream parser (options, -i, URLs, scope tracking)
pkg/streamspec/                Stream specifier parser (v, a:1, disp:default, etc.)
pkg/filtergraph/               Filter graph parser (chains, filters, options, link labels)
pkg/mapvalue/                  -map value parser (0:v, -0:a:1, [out], etc.)
pkg/completer/                 Shared completion dispatch logic
pkg/probe/                     ffprobe wrapper for stream-aware completion
pkg/actions/tools/ffmpeg/     Carapace action functions for ffmpeg value types
man/ffmpeg/                    YAML descriptions for completion value types
skills/ffmpeg/                 AI agent reference documentation (not compiled Go)
testdata/                      Generated media files for integration tests
```

### Completer CLIs (`cmd/carapace-ffmpeg/`, `cmd/carapace-ffplay/`, `cmd/carapace-ffprobe/`)

Standalone carapace completers for the `ffmpeg`, `ffplay`, and `ffprobe` commands. Each uses `DisableFlagParsing` + `PositionalAnyCompletion` with `argstream.ParseForCompletionWithProfile()` and a tool-specific `ToolProfile` to dispatch context-aware completions.

- **`root.go`** — Root cobra command with `carapace.Gen(rootCmd).Standalone()`. `PositionalAnyCompletion` callback parses args with `argstream.ParseForCompletionWithProfile()` and dispatches to shared actions from `pkg/completer/`.
- **`root_test.go`** — Tests for `completer.ContextToArgs()` helper and argstream integration (ffmpeg CLI only).
- **`main.go`** — Entry point calling `cmd.Execute()`.

The ffplay completer uses `DefaultFFplayProfile` (no output section, decoder-only codec). The ffprobe completer uses `DefaultFFprobeProfile` (no output section, no filter completion, decoder-only codec).

### Debug CLI (`cmd/carapace-ffmpeg-debug/`)

Testing/debug CLI exposing raw parser output as JSON. Each subcommand has a `-complete` variant that outputs the completion context instead of the AST.

- **`root.go`** — Root cobra command with `carapace.Gen(rootCmd)` + `spec.Register(rootCmd)`
- **`streamspec.go`** — `streamspec` and `streamspec-complete` subcommands
- **`filtergraph.go`** — `filtergraph` and `filtergraph-complete` subcommands
- **`mapvalue.go`** — `mapvalue` and `mapvalue-complete` subcommands
- **`argstream.go`** — `argstream` and `argstream-complete` subcommands

### Stream Specifier (`pkg/streamspec/`)

- **`parser.go`** — Full parser. `Parse()` → `*Specifier` AST with spans. Also exposes `IsSpecifier(text)` for validation. Parses the stream specifier grammar: `v`, `a:1`, `g:0`, `m:language:eng`, `disp:default+forced`, `u`, `#0x1F3`, `i:0x1F3`, `p:0xa`, and their compositions with `:`-separated additional specifiers.
- **`completion_parser.go`** — Completion parser. `ParseForCompletion(input)` → `*CompletionContext`.
- **`completion.go`** — Completion context types (`ExpectedToken`, `SpecifierForm`, `CompletionContext`).
- **`ast.go`** — AST node types (`SpecifierKind`, `StreamType`, `StreamTypeExpr`, `GroupExpr`, etc.).
- **`span.go`** — `Span` and `Pos` types.
- **`format.go`** — `Format()` for AST → string round-tripping.
- **`streamspec_test.go`** / **`completion_test.go`** — Tests.

### Filter Graph (`pkg/filtergraph/`)

- **`parser.go`** — Full parser. `Parse()` → `*Filtergraph` AST. Handles chains separated by `;`, filters separated by `,`, option lists separated by `:`, key=value pairs, link labels `[label]`, and quoting/escaping.
- **`completion_parser.go`** — Completion parser. `ParseForCompletion(input)` → `*CompletionContext`.
- **`completion.go`** — Completion context types (`ExpectedToken`, `FilterContext`, `CompletionContext`).
- **`ast.go`** — AST node types (`Filtergraph`, `Chain`, `Filter`, `FilterOption`).
- **`span.go`** — `Span` and `Pos` types.
- **`format.go`** — `Format()` for AST → string round-tripping.
- **`filtergraph_test.go`** / **`completion_test.go`** — Tests.

### Map Value (`pkg/mapvalue/`)

- **`parser.go`** — Full parser. `Parse()` → `*MapValue` AST. Parses `-map` values including negative maps (`-0:a:1`), stream specifiers (`0:v`, `0:m:language:eng`), view specifiers (`0:v:0:view:all`), optional maps (`0:a?`), and link labels (`[out]`).
- **`completion_parser.go`** — Completion parser. `ParseForCompletion(input)` → `*CompletionContext`.
- **`completion.go`** — Completion context types (`ExpectedToken`, `CompletionContext`).
- **`mapvalue_test.go`** — Tests.

Note: `mapvalue` is the only parser package that lacks `ast.go` and `format.go` — the `MapValue` struct is defined directly in `parser.go` since the AST is a flat structure (no nested nodes).

### Argument Stream (`pkg/argstream/`)

- **`parser.go`** — Full parser. `Parse(args)` → `*Program` AST (uses `DefaultFFmpegProfile`). `ParseWithProfile(args, profile)` allows ffplay/ffprobe profiles. Tokenizes an argument list into global options, input options, input files, output options, and output files. Tracks scope based on `-i` markers, option definitions, and `ToolProfile.HasOutputSection`.
- **`completion_parser.go`** — Completion parser. `ParseForCompletion(args, trailingSpace)` → `*CompletionContext` (uses `DefaultFFmpegProfile`). `ParseForCompletionWithProfile(args, trailingSpace, profile)` allows ffplay/ffprobe profiles. With `HasOutputSection=false`, positional non-option args are treated as input URLs and output-related expected tokens are never emitted.
- **`completion.go`** — Completion context types: `ExpectedToken` (enum), `Scope`, `OptionContext`, `CompletionContext`. JSON-serializable with `json` tags.
- **`ast.go`** — AST node types (`Token`, `TokenKind`, `Scope`, `InputFile`, `OutputFile`, `Program`). Also JSON-serializable.
- **`options.go`** — Static option definitions for ffmpeg (`OptionDef`, `OptionScope`, `OptionType`, `ValueType`, `OptionIndex`). `ParseOptionName()` splits `-c:v:1` into `("c", "v:1")`. `LookupOption()` looks up by base name. `buildOptionIndex()` is called from `init()`. `buildIndexFromOptions()` is the shared index builder used by all three tool option files.
- **`ffplay_options.go`** — Static option definitions for ffplay. `buildFFplayOptionIndex()` returns the ffplay option index.
- **`ffprobe_options.go`** — Static option definitions for ffprobe. `buildFFprobeOptionIndex()` returns the ffprobe option index.
- **`profile.go`** — `ToolProfile` struct with `Name`, `HasOutputSection`, and `OptionIndex`. Defines `DefaultFFmpegProfile`, `DefaultFFplayProfile`, `DefaultFFprobeProfile`. `LookupOption()` method looks up options in the profile's index.
- **`profile_test.go`** — Tests for profile properties, option isolation, and completion context behavior.
- **`span.go`** — `Span` type.
- **`argstream_test.go`** / **`completion_test.go`** — Tests.

### Completer (`pkg/completer/`)

- **`completer.go`** — Shared completion actions used by all three completer CLIs. Key functions:
  - `ContextToArgs(c carapace.Context) (args []string, trailingSpace bool)` — converts carapace context to argstream input.
  - `IsMidTokenOptionWithSpec(value, profile)` — detects mid-token `-c:v` style partial input.
  - `ActionOptions` / `ActionOptionNames` / `ActionOptionNamesWithSpecSuffix` — option name completions (spec-accepting options get `Suffix(":")` + `NoSpace(':')`).
  - `ActionOptionValue(ctx, codecAction, filterValue)` — giant switch on `ValueType` dispatching to the correct ffmpeg action.
  - `ActionStreamSpecifier` / `ActionStreamSpecifierWithStreams` — stream specifier completion; `WithStreams` variant uses probed stream info for index-aware completion.
  - `ActionFilterValue(value, isComplex, filterOpts)` — filtergraph-aware completion (filter names, options, values, link labels, chain separators).
  - `FilterOptsFromContext(ctx)` — derives audio/video filter scope from stream specifier or implicit spec.
  - `ActionDecoderOnlyCodec(ctx)` — decoder-only codec completions (for ffplay/ffprobe).
  - `ProbeAll(ctx)` — probes all `InputURLs` via `ffprobe` and merges stream info for stream-aware completion.
  - `ActionStreamSpecifierParts` / `ActionStreamSpecifierPartsWithStreams` — mid-token `ActionMultiParts(":")` path for specifier parts.
- **`completer_test.go`** — Unit tests for internal helpers + integration tests using `testdata/` media files (probing, stream indexing, metadata).

### Probe (`pkg/probe/`)

Wraps `ffprobe` CLI to extract stream metadata from local media files for stream-aware completion.

- **`probe.go`** — `Probe(inputURL)` runs `ffprobe -hide_banner -show_streams -show_format -of json=c=1 -- <file>`, parses output, merges format-level tags into stream-level tags. Returns `nil` (no error) for non-local files or failures. `isLocalFile()` rejects URLs (`http://`, `pipe:`, `lavfi:`, `-`, etc.) — only probes local files.
- **`StreamInfo`** struct: `Index`, `CodecName`, `CodecType`, `SampleFmt`, `PixFmt`, `Disposition` (map), `Tags` (map).
- **`MetadataValues(streams, key)`** — case-insensitive tag key lookup, returns unique values.
- **`StreamIndices(streams, codecType)`** — returns index strings filtered by codec type.
- **`ActiveDispositions(streams)`** — returns disposition names that are non-zero in any stream.
- **`probe_test.go`** — Unit tests for helper functions + integration tests that generate tiny media files on-the-fly via `ffmpeg -f lavfi` in `t.TempDir()`, then probe them.

### Actions (`pkg/actions/tools/ffmpeg/`)

Carapace completion actions for ffmpeg value types. All actions use `.Tag()` and `.Uid()`/`.UidF()` for caching/identification.

- **`value_types.go`** — Exported action functions. **Dynamic actions** shell out to `ffmpeg -hide_banner` (codecs, encoders, decoders, formats, pixel/sample formats, channel layouts, filters, hwaccels, bitstream filters, demuxers, muxers, protocols, devices). **Static actions** use `carapace.ActionValues`/`ActionValuesDescribed` (video sizes, frame rates, log levels, booleans, bitrates, dispositions, targets, etc.). Most dynamic actions accept an `*Opts` struct to filter by stream type/audio/video/muxing/demuxing.
- **`value_types_test.go`** — Tests using carapace's `sandbox.Action()` framework. Disposition tests use `Expect` (exact positive assertions on static values). Filter/BSF tests use `ExpectNot` (negative assertions) because result sets are large/dynamic from live `ffmpeg` output.
- **`filter_options.go`** — `parseFilterHelp(output string)` parses `ffmpeg -h filter=<name>` output into `[]FilterOption` (name, type, description, enum values). Used by filter option completion.
- **`filter_options_test.go`** — Tests with hardcoded `ffmpeg -h filter=<name>` output as const string fixtures.
- **`helpers.go`** — Empty file (placeholder; currently unused).
- **`uid.go`** — `Uid(host string, opts ...string)` returns a closure that builds `ffmpeg://` scheme URLs for carapace's action deduplication. Dynamic actions use `UidF(Uid("host"))`, static actions use `Uid("ffmpeg", "host")`.

#### Opts structs pattern

Dynamic action functions accept an `*Opts` struct with boolean filter fields. Each has a `Default()` method that sets all fields to `true`:

| Struct | Fields | Used by |
|--------|---------|----------|
| `CodecOpts` | Attachment, Audio, Data, Subtitle, Video | `ActionCodecs`, `ActionEncodableCodecs`, `ActionDecodableCodecs` |
| `DecoderOpts` | Audio, Subtitle, Video | `ActionDecoders` |
| `EncoderOpts` | Audio, Subtitle, Video | `ActionEncoders` |
| `FormatOpts` | Demuxing, Muxing | `ActionFormats` |
| `FilterOpts` | Audio, Video | `ActionFilters` |
| `BsfOpts` | Audio, Video, Subtitle | `ActionBitstreamFilters` (stream type inferred from BSF name prefix) |
| `DispositionOpts` | Audio, Video, Subtitle | `ActionDispositions` (per-disposition applicability) |
| `DeviceOpts` | Demuxing, Muxing | `ActionDevices` |

### Man Pages (`man/ffmpeg/`)

YAML files providing extended descriptions for completion value types. Each subdirectory contains a single YAML file keyed by value name. Used by the carapace spec system to augment completion descriptions. Example: `man/ffmpeg/codec/codec.yaml` contains `copy: |` with extended help text.

Descriptions are researched from online sources (Wikipedia, ffmpeg wiki, multimedia wiki, codec documentation) rather than raw ffmpeg output. Each description explains the value's purpose, history, and typical usage in under 200 words. Entries cover: codecs (580), decoders (557), encoders (228), formats (290), filters (561), muxers (182), demuxers (372), pixel formats (267), bitstream filters (49), protocols (40), channel layouts (38), sample formats (12), devices (13), and various static value types (booleans, dispositions, log levels, etc.).

#### YAML format

```yaml
key: |
  Multi-line description here. Can use `backticks` and **bold**.
"numeric_key": |
  Keys starting with digits must be quoted.
special_chars: |
  Keys with :{}[],&*#?|-@!%`'"\ must be quoted.
```

#### Key quoting rules

- Keys starting with digits MUST be quoted (e.g., `"012v": |`, `"3gp": |`)
- Keys with special characters (`:{}[],&*#?|-@!%`'"\`) must be quoted
- Multi-line descriptions in block scalar (`|`) need each continuation line indented with 2 spaces

#### When adding new man page entries

- Research descriptions from online sources, not from `ffmpeg` command output
- Keep descriptions under 200 words
- Explain what the value is, what it's used for, and any relevant context
- For hardware-accelerated variants (nvenc, qsv, vaapi, amf, v4l2m2m, cuda, vulkan), describe what hardware/GPU they target
- Validate YAML: `python3 -c "import yaml; yaml.safe_load(open('path'))"`
- Check for duplicate keys by round-tripping: the YAML parser silently drops duplicates

### Skills (`skills/ffmpeg/`)

A compound skill with a `SKILL.md` routing table and `references/` directory containing deep reference documentation on ffmpeg's CLI model. Not part of the Go codebase — used by AI agents working on this repo. References cover: stream model, option syntax, option scopes, stream specifiers, filtergraphs, mapping, value types, ffplay, and ffprobe.

## Key Patterns & Gotchas

### ffmpeg's positional argument model

Unlike traditional CLI flag trees, ffmpeg arguments form a **linear stream** where options apply to the *next* `-i` (input) or output URL. The argstream parser tracks a state machine: `GLOBAL → INPUT → OUTPUT`. Same option name (`-r`, `-f`) has different meaning depending on position.

### Stream specifier colon splitting

The option `-c:v:1` is a single token where `c` is the option base name and `v:1` is the stream specifier. The colon is ambiguous — it can separate option name from specifier, or components within the specifier, or be part of a value. `ParseOptionName()` splits on the first colon only.

### `disp:` must be checked before stream type `d`

In the stream specifier grammar, `d` is both the stream type letter for "data streams" and the start of `disp:` (disposition). The parser checks `disp:` first.

### Two layers of parsing

The argstream parser handles the overall command structure (options, `-i`, URLs). Sub-parsers handle:
- **Stream specifier** (`pkg/streamspec/`): the `v:1` part after `-c:`
- **Map value** (`pkg/mapvalue/`): the `-map` argument value
- **Filter graph** (`pkg/filtergraph/`): the `-vf`/`-af`/`-filter_complex` argument value

### Colons in option tokens

The option `-c:v:1` is a single token where the colon separates the option base name from the stream specifier. In bash, `:` is in `COMP_WORDBREAKS`, which causes bash to split the token at colons. However, **carapace handles this**: it re-lexes `COMP_LINE` using `carapace-shlex` and merges the wordbreak tokens back into their neighboring words via `Words()`. So `Context.Value` and `Context.Args` contain the **full merged token** (`-c:v`), not the shell-split fragments (`-c`, `:v`).

The completion parser in `argstream/completion_parser.go` also has code paths (lines 51–84) that handle colon-prefixed fragments like `:v` — these exist as a safety net for non-bash callers (other shells, the debug CLI, manual invocation) that might not do the re-lexing merge. When modifying the completion parser, be aware both paths exist: the merged-token path (primary, via carapace) and the colon-split path (defensive, for other integration scenarios).

### `trailingSpace` is critical for completion

`ParseForCompletion(args, trailingSpace)` behaves differently based on `trailingSpace`:
- `true` — cursor is at a new blank position after the last token; expect new options/URLs.
- `false` — cursor is mid-token within the last argument; the last arg is the partial text being completed.

Getting this wrong produces wrong `ExpectedTokens`. The `ContextToArgs()` function in the completer package handles the conversion from carapace's `Context.Args`/`Context.Value` to the `(args, trailingSpace)` pair.

### Implicit stream specifiers for aliases

Options like `-vcodec`, `-acodec`, `-vf`, `-af`, `-ab` have `ImplicitSpec` set (`"v"`, `"a"`, etc.) in their `OptionDef`. This means:
- `-vcodec libx264` is equivalent to `-c:v libx264` — the `:v` is implied, not typed.
- The completion parser skips the stream specifier step for these (no `ExpectedStreamSpecifier`).
- The `OptionContext.AcceptsSpec` is set to `false` for aliases with `ImplicitSpec`, since the user doesn't type the specifier explicitly.

### Option index is built in `init()`

`OptionIndex` is populated by `buildOptionIndex()` called from `init()`. All option names (canonical, short, and aliases) are registered in the map. When adding new options, add them to the `options` slice in `buildOptionIndex()` and ensure alias handling in the loop at the bottom registers them with correct `ImplicitSpec` values.

### Option definitions are iterative

The `options.go` file contains the ~100 most common options. This should be expanded over time by scraping `ffmpeg -h full` output. When adding options, pay attention to:
- **Scope**: `ScopeGlobalOpt`, `ScopePerFileOpt`, `ScopeInputOnlyOpt`, `ScopeOutputOnlyOpt`, `ScopePerStreamOpt`
- **Type**: `TypeBoolean` or `TypeValue`
- **ValueType**: Must match one of the defined `ValueType` constants; new value types need a corresponding action in `pkg/actions/tools/ffmpeg/value_types.go`
- **AcceptsSpec**: Whether a stream specifier suffix is valid
- **ImplicitSpec**: For alias options that imply a specific stream type

### Completer uses `DisableFlagParsing` + `PositionalAnyCompletion`

The completer CLI (`cmd/carapace-ffmpeg/cmd/root.go`) does NOT use cobra's flag parsing. It sets `DisableFlagParsing: true` so cobra hands all arguments through as positional args. Completion is handled entirely by `carapace.Gen(rootCmd).PositionalAnyCompletion(...)`, which manually parses via `argstream.ParseForCompletion()`.

### Mid-token option+specifier completion

When the user is typing `-c:v` as a single token (no space after the colon), the completer detects this with `IsMidTokenOptionWithSpec()` and switches to `ActionMultiParts(":")` to handle the colon-separated parts within the single token. This is separate from the shell-split case where `:` arrives as a separate argument.

### UIDs use `ffmpeg://` scheme

All completion actions use `ffmpeg://` UIDs (defined in `pkg/actions/tools/ffmpeg/uid.go`) for carapace's action deduplication. Dynamic actions use `UidF(Uid("host"))` (function form), static actions use `Uid("ffmpeg", "host")` (direct form). Both produce URLs like `ffmpeg://codec` or `ffmpeg://codec?scope=input`.

### Codec completion differs by scope

`actionCodec()` in the completer dispatches different results depending on scope:
- **Global/Input scope**: codecs + decoders (decoding context)
- **Output scope**: codecs + encoders (encoding context)

Within each scope, `FilterOptsFromContext()` derives audio/video filter scope from the stream specifier or implicit spec on the current option.

### Probe powers stream-aware completion

`ProbeAll()` in `pkg/completer/` calls `probe.Probe()` on each `InputURL` from the completion context. The resulting `[]probe.StreamInfo` is passed to `ActionStreamSpecifierWithStreams()` to offer index-aware completions (e.g., `a:0`, `a:1` for a file with two audio streams) and metadata-based completions (language tags, disposition names). Probe only runs on local files — URLs like `http://`, `pipe:`, `lavfi:` are silently skipped (returns `nil`, no error).

### `actionMapValue()` is a stub

The map value completion action in the completer currently returns an empty `carapace.ActionValues()`. The `pkg/mapvalue/` parser exists but its completion support isn't wired into the completer CLI yet.

### Completion dispatch flow

```
carapace.Context
  → ContextToArgs() → (args, trailingSpace)
  → argstream.ParseForCompletionWithProfile(args, trailingSpace, profile)
  → argstream.CompletionContext {ExpectedTokens, Scope, CurrentOption, ...}
  → ProbeAll(ctx) → []probe.StreamInfo
  → switch on ExpectedToken:
      Expected*Option → ActionOptions
      ExpectedInputURL/OutputURL → ActionFiles
      ExpectedOptionValue → ActionOptionValue (switch on ValueType)
      ExpectedStreamSpecifier → ActionStreamSpecifierWithStreams
      ExpectedFilterValue → ActionFilterValue
      ExpectedMapValue → (stub)
```

## Code Conventions

- **Standard library only for parsers**: The four parser packages (`streamspec`, `filtergraph`, `mapvalue`, `argstream`) use only Go standard library. No external dependencies in these packages.
- **Carapace + Cobra for CLIs and actions**: External deps (`carapace`, `carapace-spec`, `cobra`) are only in `cmd/` and `pkg/actions/`.
- **Test style**: Table-driven tests with `testing.T` only. No testify or other assertion libraries. Helper functions like `assertHasExpected()` defined locally in test files.
- **Action test style**: Tests in `pkg/actions/tools/ffmpeg/` use carapace's `sandbox.Action()` framework. Static value tests use `Expect` (exact positive assertions). Dynamic action tests use `ExpectNot` (negative assertions — verify absent values) because the live `ffmpeg` output set varies by installation.
- **Integration test media**: `pkg/completer/` and `pkg/probe/` tests use `testdata/` files generated by `go generate ./testdata/`. Tests skip if files are missing. `pkg/probe/` tests can also generate tiny files on-the-fly in `t.TempDir()` via `ffmpeg -f lavfi`.
- **No test files for debug CLI**: `cmd/carapace-ffmpeg-debug/` has no test files.
- **JSON serialization**: AST and completion context types implement `MarshalText()` or have `json` struct tags for the debug CLI's JSON output.
- **Parser package pattern**: Each parser package follows the same structure: `parser.go` (full parser), `completion_parser.go` (completion parser), `completion.go` (context types), `ast.go` (AST types), `span.go` (spans), optional `format.go` (AST→string). Exception: `mapvalue` lacks `ast.go` and `format.go` (flat AST struct in `parser.go`).

## Release

GoReleaser builds 4 binaries: `carapace-ffmpeg`, `carapace-ffmpeg-debug`, `carapace-ffplay`, `carapace-ffprobe`. Distribution channels: Homebrew tap (`rsteube/homebrew-tap`), Scoop bucket, AUR (`carapace-ffmpeg-bin`), nfpm (apk/deb/rpm/termux.deb), Gemfury. Releases are triggered by tag pushes in CI.
