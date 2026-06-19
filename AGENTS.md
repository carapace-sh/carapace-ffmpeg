# AGENTS.md

## Project Overview

Go library for parsing ffmpeg CLI argument streams, stream specifiers, filter graphs, and map values into ASTs, with completion support. Part of the [carapace-sh](https://github.com/carapace-sh) ecosystem (shell completion framework). The module path is `github.com/carapace-sh/carapace-ffmpeg`.

## Commands

```sh
go test ./...                              # run all tests
go test ./pkg/streamspec/                   # streamspec tests only
go test ./pkg/filtergraph/                  # filtergraph tests only
go test ./pkg/mapvalue/                     # mapvalue tests only
go test ./pkg/argstream/                    # argstream tests only
go build ./...                              # build all packages
```

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

### Completer CLI (`cmd/carapace-ffmpeg/`)

```sh
go run ./cmd/carapace-ffmpeg _carapace spec                    # generate carapace spec
go run ./cmd/carapace-ffmpeg _carapace bash '' ''              # complete at empty position
go run ./cmd/carapace-ffmpeg _carapace bash '-c:v' '' '-c:v' 'libx'  # complete codec value
```

No Makefile, no linter config.

## Architecture

Two CLIs and four independent parser packages with carapace completion actions.

### Completer CLI (`cmd/carapace-ffmpeg/`)

Standalone carapace completer for the `ffmpeg` command. Uses `DisableFlagParsing` + `PositionalAnyCompletion` with `argstream.ParseForCompletion()` to dispatch context-aware completions.

- **`root.go`** — Root cobra command (`Use: "ffmpeg"`) with `carapace.Gen(rootCmd).Standalone()`. `PositionalAnyCompletion` callback parses args with `argstream.ParseForCompletion()` and dispatches to `actionOptions()`, `actionOptionValue()`, `actionStreamSpecifiers()`, `actionFilterValue()`, `actionMapValue()`.
- **`root_test.go`** — Tests for `contextToArgs()` helper and argstream integration.
- **`main.go`** — Entry point calling `cmd.Execute()`.

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

### Argument Stream (`pkg/argstream/`)

- **`parser.go`** — Full parser. `Parse(args)` → `*Program` AST. Tokenizes an ffmpeg argument list into global options, input options, input files, output options, and output files. Tracks scope (global → input → output) based on `-i` markers and option definitions.
- **`completion_parser.go`** — Completion parser. `ParseForCompletion(args, trailingSpace)` → `*CompletionContext` with scope info, current option context, and expected tokens. The `trailingSpace` bool is critical: `true` means the cursor is at a new blank position after the last token; `false` means mid-token completion.
- **`completion.go`** — Completion context types: `ExpectedToken` (enum), `Scope`, `OptionContext`, `CompletionContext`. JSON-serializable with `json` tags.
- **`ast.go`** — AST node types (`Token`, `TokenKind`, `Scope`, `InputFile`, `OutputFile`, `Program`). Also JSON-serializable.
- **`options.go`** — Static option definitions (`OptionDef`, `OptionScope`, `OptionType`, `ValueType`, `OptionIndex`). `ParseOptionName()` splits `-c:v:1` into `("c", "v:1")`. `LookupOption()` looks up by base name. `buildOptionIndex()` is called from `init()`.
- **`span.go`** — `Span` type.
- **`argstream_test.go`** / **`completion_test.go`** — Tests.

### Actions (`pkg/actions/tools/ffmpeg/`)

- **`value_types.go`** — Carapace completion actions for ffmpeg value types. Dynamic actions use `carapace.ActionExecCommand("ffmpeg", "-hide_banner", ...)` to shell out to `ffmpeg` (codecs, encoders, decoders, formats, pixel formats, sample formats, channel layouts, filters, hwaccels, bitstream filters). Static actions use `carapace.ActionValues`/`ActionValuesDescribed` (video sizes, frame rates, log levels, booleans, bitrates, dispositions, targets, etc.). All actions use `.Tag()` and `.Uid()`/`.UidF()` for caching/identification.
- **`helpers.go`** — Empty file (placeholder for parsing helpers; currently unused).
- **`uid.go`** — `Uid()` factory function that builds `ffmpeg://` scheme UIDs for carapace's action deduplication system. Used via `.UidF(Uid("host"))` or `.Uid("ffmpeg", "host")`. Accepts optional key-value pairs as query parameters.

### Man Pages (`man/ffmpeg/`)

YAML files providing extended descriptions for completion value types. Each subdirectory contains a single YAML file keyed by value name. Used by the carapace spec system to augment completion descriptions. Example: `man/ffmpeg/codec/codec.yaml` contains `copy: |` with extended help text.

### Skills (`skills/ffmpeg/`)

A compound skill with a `SKILL.md` routing table and `references/` directory containing deep reference documentation on ffmpeg's CLI model. Not part of the Go codebase — used by AI agents working on this repo. References cover: stream model, option syntax, option scopes, stream specifiers, filtergraphs, mapping, and value types.

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

Getting this wrong produces wrong `ExpectedTokens`. The `contextToArgs()` function in the completer CLI handles the conversion from carapace's `Context.Args`/`Context.Value` to the `(args, trailingSpace)` pair.

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

When the user is typing `-c:v` as a single token (no space after the colon), the completer detects this with `isMidTokenOptionWithSpec()` and switches to `ActionMultiParts(":")` to handle the colon-separated parts within the single token. This is separate from the shell-split case where `:` arrives as a separate argument.

### UIDs use `ffmpeg://` scheme

All completion actions use `ffmpeg://` UIDs (defined in `pkg/actions/tools/ffmpeg/uid.go`) for carapace's action deduplication. Dynamic actions use `UidF(Uid("host"))` (function form), static actions use `Uid("ffmpeg", "host")` (direct form). Both produce URLs like `ffmpeg://codec` or `ffmpeg://codec?scope=input`.

### Codec completion differs by scope

`actionCodec()` in `root.go` returns different results depending on scope:
- **Global/Input scope**: codecs + decoders (decoding context)
- **Output scope**: codecs + encoders (encoding context)

### `actionMapValue()` is a stub

The map value completion action in `root.go` currently returns an empty `carapace.ActionValues()`. The `pkg/mapvalue/` parser exists but its completion support isn't wired into the completer CLI yet.

## Code Conventions

- **Standard library only for parsers**: The four parser packages (`streamspec`, `filtergraph`, `mapvalue`, `argstream`) use only Go standard library. No external dependencies in these packages.
- **Carapace + Cobra for CLIs and actions**: External deps (`carapace`, `carapace-spec`, `cobra`) are only in `cmd/` and `pkg/actions/`.
- **Test style**: Table-driven tests with `testing.T` only. No testify or other assertion libraries. Helper functions like `assertHasExpected()` defined locally in test files.
- **No test files for actions or debug CLI**: `pkg/actions/tools/ffmpeg/` and `cmd/carapace-ffmpeg-debug/` have no test files.
- **JSON serialization**: AST and completion context types implement `MarshalText()` or have `json` struct tags for the debug CLI's JSON output.
- **Parser pattern**: Each parser package follows the same structure: `parser.go` (full parser), `completion_parser.go` (completion parser), `completion.go` (context types), `ast.go` (AST types), `span.go` (spans), optional `format.go` (AST→string).
