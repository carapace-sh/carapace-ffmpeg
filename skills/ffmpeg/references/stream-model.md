# Stream Model

How ffmpeg's transcoding pipeline works: the flow of data from inputs through filters to outputs.

> **Source of truth**: <https://ffmpeg.org/ffmpeg.html>. See also the Detailed Description section of the ffmpeg man page.

## Overview

ffmpeg is a **stream-oriented pipeline processor**. Unlike typical CLI tools where flags form a tree of subcommands and options, ffmpeg arguments form a **linear sequence** where options are positional transformers that apply to the next input or output in the stream.

The fundamental invocation syntax is:

```
ffmpeg [global_options] {[input_file_options] -i input_url}... {[output_file_options] output_url}...
```

## Pipeline Components

Data flows through these components in order:

```
                         ┌───────────┐   ┌───────────┐   ┌───────────┐   ┌───────────┐
 input_url ─► demuxer ─►│  decoder  │─►│  filter   │─►│  encoder  │─►│  muxer   │─► output_url
                         └───────────┘   └───────────┘   └───────────┘   └───────────┘
```

Each component:

| Component | Role | Created By | Direction |
|-----------|------|-----------|-----------|
| **Demuxer** | Reads container, splits into elementary streams (encoded packets) | Each `-i` option | Input |
| **Decoder** | Decodes packets to raw frames | Automatic or `-c` selection | Input |
| **Filter** | Transforms raw frames (scale, overlay, mix, etc.) | `-vf`/`-af` or `-filter_complex` | Middle |
| **Encoder** | Encodes raw frames to packets | `-c` or `-c:v`/`-c:a` | Output |
| **Muxer** | Interleaves streams, writes container | Output URL | Output |

## Elementary Streams

Each input file contains one or more elementary streams of different types:

| Type | Letter | Description |
|------|--------|-------------|
| Video | `v` / `V` | Video frames; `V` excludes attached pictures/thumbnails |
| Audio | `a` | Audio samples |
| Subtitle | `s` | Subtitle data |
| Data | `d` | Data tracks |
| Attachment | `t` | Attached files (fonts, cover art) |

A single input file might contain (for example) one video stream, two audio streams, and one subtitle stream. These are indexed within the file: `0:0`, `0:1`, `0:2`, `0:3`.

## Two Paths

### Transcoding Path

```
demuxer → decoder → filter → encoder → muxer
```

Full decode-filter-encode cycle. Used when you need to change codecs, apply filters, or change parameters.

### Streamcopy Path

```
demuxer → muxer  (packets copied as-is, no decode/encode)
```

Selected with `-c copy`. Fast, lossless, but no filtering possible. Used to change container format, add/remove streams, or modify metadata.

## Positional Option Application

**This is the key difference from traditional CLIs**: options are not global flags — they apply positionally to the next file:

```bash
ffmpeg -r 25 -i input.m2v -r 30 output.mp4
```

- The first `-r 25` applies to `input.m2v` (input file option: sets input framerate)
- The second `-r 30` applies to `output.mp4` (output file option: sets output framerate)

Options are **reset between files**. The same option can appear multiple times, each applying to a different file.

## Strict Ordering Rules

1. **Global options** come first (before any `-i` or output)
2. **All inputs** come before all outputs — do not interleave
3. **Per-file options** must immediately precede their target file (or follow its `-i`)
4. **Per-stream options** with stream specifiers can be placed anywhere for their file group

In practice, the convention is:

```
ffmpeg [global] [input_opts] -i input1 [input_opts] -i input2 [filter_complex] [output_opts] output1 [output_opts] output2
```

## Input/Output Asymmetry

- **Inputs** are explicitly introduced by `-i`
- **Outputs** are positional — anything that is not an option is an output URL
- This means output URLs can look like option values to a naïve parser; the lexer must track whether it is in input or output context

## Multiple Inputs, Multiple Outputs

A single ffmpeg invocation can have multiple inputs and multiple outputs:

```bash
ffmpeg -i input1.mkv -i input2.aac -map 0:v -map 1:a -c:v libx264 -c:a aac output.mp4
```

Each output gets its own set of options. Options for output N appear between output N-1's URL and output N's URL.

## Complex Filtergraph as a Middle Layer

`-filter_complex` sits between inputs and outputs. It's a **global** option that can reference any input stream and feed any output. It does not belong to any single input or output. See [filtergraph.md](filtergraph.md) for details.

## Edge Cases

- **No inputs**: `-filter_complex` can use lavfi sources (e.g. `color=c=red`) without any `-i` inputs
- **No outputs**: Not valid — ffmpeg requires at least one output
- **Same option, different scope**: `-r` is both an input and output option; the same flag name applies to different files depending on position
- **Output URL looks like option**: a filename starting with `-` can be ambiguous — use `./-filename` to disambiguate
- **Stream count limits**: Container formats restrict which stream types and counts they support

## ffplay Pipeline: Read and Play

ffplay is a media player — a simplified pipeline with no encoding or output:

```
input_url ─► demuxer ─► decoder ─► filter? ─► renderer (SDL/Vulkan)
```

Key differences from ffmpeg:
- **Single input only** (no multiple `-i` inputs)
- **No encoder or muxer** — no output file
- **No complex filtergraphs** — only simple `-vf`/`-af` filters
- **Renderer instead of muxer** — SDL or Vulkan display
- **No `-map`** — stream selection via `-ast`/`-vst`/`-sst`

See [ffplay.md](ffplay.md) for the full ffplay CLI model.

## ffprobe Pipeline: Read and Inspect

ffprobe is a probe/inspection tool — it reads metadata without (by default) decoding:

```
input_url ─► demuxer ─► decoder? ─► inspector (print metadata/frames/packets)
```

Key differences from ffmpeg:
- **Single input only** (no multiple `-i` inputs)
- **No encoding, no filtering** — no `-vf`/`-af`/`-filter_complex`
- **Decoding is optional** — only happens with `-show_frames`/`-show_log`/`-analyze_frames`
- **No `-map`** — display filtering via `-select_streams`
- **Output is text** — formatted by `-of`/`-print_format` (json, xml, csv, etc.)

See [ffprobe.md](ffprobe.md) for the full ffprobe CLI model.
