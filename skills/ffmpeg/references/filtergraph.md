# Filtergraph Syntax

How to parse ffmpeg filtergraph strings — the DSL used by `-vf`, `-af`, and `-filter_complex`.

> **Source of truth**: <https://ffmpeg.org/ffmpeg-filters.html> — Filtergraph syntax section.

## Overview

A filtergraph is a text description of a directed graph of filters. It uses a compact DSL with:

- **Filter chains** — linear sequences of filters connected by commas
- **Links** — connections between chains using labeled pads `[label]`
- **Filter options** — key=value pairs separated by colons within each filter

```
filtergraph := chain (';' chain)*
chain := link_label? filter (',' filter)* link_label?
filter := name ('=' option_list)?
option_list := option (':' option)*
option := key '=' value | value    (positional if no key)
```

## Simple vs Complex Filtergraphs

### Simple Filtergraph

One input, one output, same stream type. Created with `-vf` (video) or `-af` (audio). Associated with a specific output stream.

```bash
ffmpeg -i input.mp4 -vf "yadif,scale=1280:720" output.mp4
```

- One video input, one video output
- Tied to a specific output stream
- `-vf` = `-filter:v`, `-af` = `-filter:a`

### Complex Filtergraph

Arbitrary number of inputs/outputs, potentially different types. Created with `-filter_complex`. A global option that is not tied to any single file.

```bash
ffmpeg -i video.mkv -i image.png -filter_complex "[0:v][1:v]overlay[out]" -map "[out]" output.mkv
```

- Multiple inputs, multiple outputs
- Can mix audio and video
- Global scope — references streams from any input

## Link Labels

Labels mark input and output pads of the filter graph:

```
[input_label] filter_chain [output_label] ; [input_label] filter_chain [output_label]
```

### Label Syntax

- Enclosed in square brackets: `[label]`
- Label characters: alphanumeric, underscore, and most printable ASCII
- Labels come **before** a chain (input) or **after** a chain (output)

```bash
[0:v]scale=1280:720[out_v]
```

### Input Labels (Complex Filtergraph)

| Syntax | Meaning |
|--------|---------|
| `[file_index:stream_specifier]` | Connect input stream (same as `-map` syntax) |
| `[dec:dec_idx]` | Connect loopback decoder |
| `[label]` | Connect output from another filtergraph |
| (unlabeled) | First unused input stream of matching type |

```bash
# Explicit input from file 0, video stream
[0:v]scale=1280:720[out]

# Implicit: first unused video input
overlay[out]
```

### Output Labels

- Used with `-map '[label]'` to connect to an output file
- Unlabeled outputs are added to the first output file automatically
- Each output label must be mapped exactly once

## Filter Chain

A chain is a comma-separated sequence of filters where each filter's output feeds the next:

```
filter1,filter2,filter3
```

is equivalent to:

```
filter1 → filter2 → filter3
```

### Chain Example

```bash
-vf "yadif=mode=send_frame,scale=1280:720,format=yuv420p"
```

This chains: deinterlace → resize → set pixel format.

## Filter Options

Each filter accepts options in two forms:

### Key=Value Form

```
filter=key1=value1:key2=value2
```

```bash
scale=w=1280:h=720
overlay=x=10:y=20:format=auto
```

### Positional Form

```
filter=value1:value2
```

```bash
scale=1280:720
overlay=10:20
```

The positional form assigns values in the order the filter defines its options. Mixing positional and key=value is allowed:

```bash
scale=1280:720:force_original_aspect_ratio=decrease
```

### Special Characters in Values

| Character | Must Be Escaped | Context |
|-----------|----------------|---------|
| `:` | Yes (in option values) | Delimits filter options |
| `=` | Yes (in option values) | Delimits key from value |
| `'` | Yes | FFmpeg quoting |
| `\` | Yes | Escape character |
| `;` | Yes (in filter graph) | Delimits chains |
| `,` | Yes (in chains) | Delimits filters in a chain |
| `[` `]` | Yes (outside labels) | Label delimiters |

To use these literally in a value, either:
- Escape with backslash: `\;`
- Wrap in single quotes: `'[0:v]'`

### Example: Escaping

```bash
# A drawtext filter with semicolons in the text
drawtext=text='Hello World\; Part 2':x=10:y=20
```

## Filter Introspection

Each filter has typed inputs and outputs. The `-filters` listing shows:

```
 .. scale             V->V       Scale the input video size and/or convert the image format.
 .. aformat           A->A       Convert the input audio to one of the specified formats.
 .. amerge            N->A       Merge two or more audio streams into a single multi-channel stream.
```

| Column | Meaning |
|--------|---------|
| First 2 chars | `T` = Timeline support, `S` = Slice threading |
| Arrow format `V->V` | Input type → Output type |
| `N` | Dynamic number of inputs/outputs |

### Filter-Specific Options

Each filter defines its own option set. Available via:

```bash
ffmpeg -h filter=FILTER_NAME
```

This shows all options with their types, ranges, and defaults. See [value-types.md](value-types.md) for how to parse filter option values.

## Lexer Implications

### Nested Quoting

A filtergraph is typically single-quoted at the shell level, then ffmpeg parsing applies:

```bash
# Shell: single quotes pass the string verbatim
ffmpeg -filter_complex '[0:v]scale=1280:720[out]'

# Shell: double quotes require escaping
ffmpeg -filter_complex "[0:v]scale=1280:720[out]"
```

Inside the filtergraph string, the DSL has its own escaping. There are effectively 2-3 layers:

1. **Shell layer** — bash/zsh/fish quoting rules
2. **ffmpeg layer** — `'` and `\` quoting
3. **filter option layer** — `:`, `=`, `,`, `;` as delimiters that can be escaped with `\`

### Parsing Order

The filtergraph DSL should be parsed in this order:

1. Split on `;` (top-level chain separator) — respecting quotes and escapes
2. For each chain, split on `,` (filter separator) — respecting quotes and escapes
3. For each filter, split on `=` (first one only) to get name and options
4. For options, split on `:` (option separator) — respecting quotes and escapes
5. For each option, split on `=` (key=value) or treat as positional

### Ambiguity: Colons in Values

The `:` character serves triple duty:
- Stream specifier delimiter: `-c:v:1`
- Filter option delimiter: `scale=1280:720`
- Time separator: `12:03:45`

In a filtergraph context, `:` is the option delimiter. To include a literal `:` in a value, escape it: `\:`.

### Ambiguity: Brackets

`[` and `]` are label delimiters in the filtergraph DSL. They also appear in:
- Stream references: `[0:v]`
- Output labels: `[out_v]`

Outside of labels, brackets must be escaped.