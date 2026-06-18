# AGENTS.md

## Project Overview

Go library for parsing ffmpeg CLI argument streams, stream specifiers, filter graphs, and map values into ASTs, with completion support. Part of the [carapace-sh](https://github.com/carapace-sh) ecosystem (shell completion framework). The module path is `github.com/carapace-sh/carapace-ffmpeg`.

## Commands

```sh
go test ./...                              # run all tests
go test ./pkg/streamspec/                   # run streamspec tests only
go test ./pkg/filtergraph/                  # run filtergraph tests only
go test ./pkg/mapvalue/                     # run mapvalue tests only
go test ./pkg/argstream/                    # run argstream tests only
go build ./...                              # build all packages
go run . streamspec "a:1"                  # parse stream specifier, output AST as JSON
go run . streamspec-complete "a:"           # stream specifier completion context as JSON
go run . filtergraph "scale=1280:720"       # parse filtergraph, output AST as JSON
go run . filtergraph-complete "sca"         # filtergraph completion context as JSON
go run . mapvalue "0:v"                     # parse -map value, output AST as JSON
go run . mapvalue-complete "0:"             # -map value completion context as JSON
go run . argstream -- -i input.mp4 -c:v libx264 output.mp4  # parse ffmepg arg stream, output AST as JSON
go run . argstream-complete -- -i input.mp4 -c:v            # argstream completion context as JSON
```

No Makefile, no linter config.

## Architecture

Cobra-based CLI (`cmd/`) wrapping four independent parsers with completion context support that wire to carapace (`pkg/actions/tools/ffmpeg/`).

### CLI (`cmd/`)

- **`root.go`** — Root cobra command with `carapace.Gen(rootCmd).Standalone()` + spec registration
- **`streamspec.go`** — `streamspec` and `streamspec-complete` subcommands
- **`filtergraph.go`** — `filtergraph` and `filtergraph-complete` subcommands
- **`mapvalue.go`** — `mapvalue` and `mapvalue-complete` subcommands
- **`argstream.go`** — `argstream` and `argstream-complete` subcommands

Entry point is `main.go` at `cmd/carapace-ffmpeg/` which calls `cmd.Execute()`.

### Stream Specifier (`pkg/streamspec/`)

- **`parser.go`** — Full parser. `Parse()` → `*Specifier` AST with spans. Parses the stream specifier grammar: `v`, `a:1`, `g:0`, `m:language:eng`, `disp:default+forced`, `u`, `#0x1F3`, `i:0x1F3`, `p:0xa`, and their compositions with `:`-separated additional specifiers.
- **`completion_parser.go`** — Completion parser. `ParseForCompletion(input)` → `*CompletionContext` describing what is expected at cursor.
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

### Argument Stream (`pkg/argstream/`)

- **`parser.go`** — Full parser. `Parse(args)` → `*Program` AST. Tokenizes an ffmpeg argument list into global options, input options, input files, output options, and output files. Tracks scope (global → input → output) based on `-i` markers and option definitions.
- **`completion_parser.go`** — Completion parser. `ParseForCompletion(args)` → `*CompletionContext` with scope info, current option context, and expected tokens.
- **`completion.go`** — Completion context types (`ExpectedToken`, `Scope`, `OptionContext`, `CompletionContext`).
- **`ast.go`** — AST node types (`Token`, `TokenKind`, `Scope`, `InputFile`, `OutputFile`, `Program`).
- **`options.go`** — Static option definitions (`OptionDef`, `OptionScope`, `OptionType`, `ValueType`, `OptionIndex`). Covers common ffmpeg options with their scope, type, value type, and stream specifier support. `ParseOptionName()` splits `-c:v:1` into `("c", "v:1")`.
- **`span.go`** — `Span` type.
- **`argstream_test.go`** / **`completion_test.go`** — Tests.

### Actions (`pkg/actions/tools/ffmpeg/`)

- **`value_types.go`** — Carapace completion actions for ffmpeg value types (codecs, encoders, decoders, formats, pixel formats, sample formats, channel layouts, filters, video sizes, frame rates, log levels, dispositions, booleans). These shell out to `ffmpeg` to get dynamic lists.
- **`helpers.go`** — Parsing helpers for ffmpeg command output (`splitLines`, `extractCodecName`, `extractFormatName`, `extractFilterName`).

### Key Patterns & Gotchas

#### ffmpeg's positional argument model

Unlike traditional CLI flag trees, ffmpeg arguments form a **linear stream** where options apply to the *next* `-i` (input) or output URL. The argstream parser tracks a state machine: `GLOBAL → INPUT → OUTPUT`. Same option name (`-r`, `-f`) has different meaning depending on position.

#### Stream specifier colon splitting

The option `-c:v:1` is a single token where `c` is the option base name and `v:1` is the stream specifier. The colon is ambiguous — it can separate option name from specifier, or components within the specifier, or be part of a value. `ParseOptionName()` splits on the first colon only.

#### `disp:` must be checked before stream type `d`

In the stream specifier grammar, `d` is both the stream type letter for "data streams" and the start of `disp:` (disposition). The parser checks `disp:` first.

#### Two layers of parsing

The argstream parser handles the overall command structure (options, `-i`, URLs). Sub-parsers handle:
- **Stream specifier** (`pkg/streamspec/`): the `v:1` part after `-c:`
- **Map value** (`pkg/mapvalue/`): the `-map` argument value
- **Filter graph** (`pkg/filtergraph/`): the `-vf`/`-af`/`-filter_complex` argument value

#### Option definitions are iterative

The `options.go` file contains the ~60 most common options. This should be expanded over time by scraping `ffmpeg -h full` output.