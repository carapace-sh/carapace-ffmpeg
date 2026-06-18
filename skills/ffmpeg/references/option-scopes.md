# Option Scopes

How ffmpeg classifies options by scope: global, per-file, and per-stream, and the positional application rules.

> **Source of truth**: <https://ffmpeg.org/ffmpeg.html>. The `-h full` output groups options by scope.

## Scope Categories

Every ffmpeg option belongs to exactly one scope category:

| Scope | Applies To | Reset Between Files | Examples |
|-------|-----------|-------------------|----------|
| **Global** | Entire ffmpeg process | N/A (no reset) | `-y`, `-v`, `-report`, `-benchmark`, `-filter_complex` |
| **Per-file (input+output)** | Next input or output file | Yes | `-f`, `-t`, `-ss`, `-bitexact` |
| **Per-file (input-only)** | Next input file only | Yes | `-sseof`, `-re`, `-readrate`, `-isync` |
| **Per-file (output-only)** | Next output file only | Yes | `-metadata`, `-map`, `-shortest`, `-fs` |
| **Per-stream** | Specific streams within a file | Yes | `-c`, `-b`, `-r`, `-pix_fmt`, `-vf`, `-af` |

## Positional Application

The fundamental rule: **options apply to the next specified file**.

```
ffmpeg [global_opts] [input_opts1] -i input1 [input_opts2] -i input2 [output_opts1] output1 [output_opts2] output2
```

- `input_opts1` → applies to `input1`
- `input_opts2` → applies to `input2`
- `output_opts1` → applies to `output1`
- `output_opts2` → applies to `output2`

Options are **reset** between files — each file starts with default values.

### Example

```bash
ffmpeg -r 25 -i input.m2v -r 30 -c:v libx264 output.mp4
```

| Option | Scope | Applies To |
|--------|-------|-----------|
| `-r 25` | Per-file (input) | `input.m2v` |
| `-r 30` | Per-file (output) | `output.mp4` |
| `-c:v libx264` | Per-stream (output video) | `output.mp4` |

## Per-Stream Options and Stream Specifiers

Per-stream options can optionally include a stream specifier to target specific streams:

```bash
# All audio streams
-c:a aac

# Second audio stream only
-c:a:1 aac

# All streams (no specifier)
-c copy

# All video streams with specific bitrate
-b:v 2M
```

Without a specifier, the option applies to **all streams** of the appropriate type in the file.

See [stream-specifier.md](stream-specifier.md) for the full specifier grammar.

## Global Options

Global options affect the entire ffmpeg process and must appear **before** any `-i` or output:

| Option | Type | Description |
|--------|------|-------------|
| `-y` | Boolean | Overwrite output files |
| `-n` | Boolean | Never overwrite |
| `-v` / `-loglevel` | Value | Set logging level |
| `-report` | Boolean | Generate report |
| `-hide_banner` | Boolean | Suppress banner |
| `-benchmark` | Boolean | Show timing |
| `-filter_complex` | Value | Define complex filtergraph (global!) |
| `-stats` | Boolean | Show progress stats |
| `-cpuflags` | Value | Set CPU flags |
| `-max_alloc` | Value | Max allocation size |

### Special Case: `-filter_complex` is Global

`-filter_complex` is a **global** option because a complex filtergraph connects multiple inputs and outputs — it cannot belong to a single file. It can appear multiple times; each instance creates a new independent filtergraph.

## Per-File Input Options

These only affect the next `-i` file:

| Option | Type | Description |
|--------|------|-------------|
| `-f fmt` | Value | Force input format |
| `-t duration` | Value | Limit reading duration |
| `-ss position` | Value | Seek to position |
| `-sseof position` | Value | Seek from EOF |
| `-re` | Boolean | Read at native framerate |
| `-readrate speed` | Value | Read at specified rate |
| `-isync ref` | Value | Sync reference |
| `-stream_loop count` | Value | Loop input |

## Per-File Output Options

These only affect the next output URL:

| Option | Type | Description |
|--------|------|-------------|
| `-f fmt` | Value | Force output format |
| `-t duration` | Value | Limit output duration |
| `-to position` | Value | Stop at position |
| `-ss position` | Value | Seek (output) |
| `-map spec` | Value | Stream mapping |
| `-metadata[:spec] kv` | Value | Set metadata |
| `-shortest` | Boolean | Finish with shortest input |
| `-fs size` | Value | Limit file size |
| `-map_metadata` | Value | Metadata mapping |
| `-map_chapters` | Value | Chapter mapping |

## Per-Stream Options

These apply to specific streams within a file, selected by stream specifier:

| Option | Specifier | Description |
|--------|-----------|-------------|
| `-c[:spec] codec` | Stream | Codec selection |
| `-b[:spec] bitrate` | Stream | Bitrate |
| `-r[:spec] rate` | Stream | Framerate |
| `-pix_fmt[:spec] fmt` | Stream (video) | Pixel format |
| `-ar[:spec] rate` | Stream (audio) | Sample rate |
| `-ac[:spec] channels` | Stream (audio) | Channel count |
| `-filter[:spec] graph` | Stream | Simple filtergraph |
| `-vf graph` | Stream (video) | Alias for `-filter:v` |
| `-af graph` | Stream (audio) | Alias for `-filter:a` |
| `-frames[:spec] count` | Stream | Frame limit |
| `-bsf[:spec] filters` | Stream | Bitstream filters |
| `-disposition[:spec]` | Stream | Stream disposition |
| `-tag[:spec] fourcc` | Stream | Codec tag |

## Scope and the Lexer

For a lexer, the scope determines:

1. **Whether the option takes a value** — global booleans like `-y` don't; per-stream options like `-c:v` always do
2. **How to resolve the same option name in different positions** — `-r` before `-i` is input; after `-i` is output
3. **When to reset state** — per-file options reset between files
4. **Where stream specifiers apply** — only per-stream options accept `:spec` suffixes

The lexer must track a **file position context**: are we in global, input N, or output N space? This context determines how to interpret each option.