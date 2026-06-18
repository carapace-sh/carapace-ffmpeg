# Stream Mapping

How `-map` selects input streams and routes them to outputs, including negative maps, optional maps, and filtergraph link labels.

> **Source of truth**: <https://ffmpeg.org/ffmpeg.html> — Advanced options section, `-map` entry.

## Overview

Without `-map`, ffmpeg uses automatic stream selection (first video, highest-channel-count audio, etc.). Adding `-map` disables automatic selection — only explicitly mapped streams appear in the output.

`-map` is an **output-only** option that creates one or more streams in the output file.

## Syntax

```
-map [-]input_file_id[:stream_specifier][:view_specifier][:?] | [linklabel]
```

## Two Forms

### Form 1: Select from Input File

```
-map input_file_id[:stream_specifier][:view_specifier][:?]
```

| Component | Required | Description |
|-----------|----------|-------------|
| `input_file_id` | Yes | 0-based index matching an `-i` input |
| `stream_specifier` | No | See [stream-specifier.md](stream-specifier.md); defaults to all streams |
| `view_specifier` | No | For multiview video (see below) |
| `?` | No | Makes the map optional — no error if no streams match |

### Form 2: Select from Filtergraph

```
-map [linklabel]
```

`linklabel` must correspond to an output pad label in a `-filter_complex` graph. The label is in square brackets.

## Negative Mapping

Prefixing with `-` **excludes** matching streams from already-created mappings:

```bash
# Map all streams except the second audio
ffmpeg -i INPUT -map 0 -map -0:a:1 OUTPUT
```

Negative maps can only **subtract** from existing positive maps — they cannot be used alone.

## Optional Mapping

A trailing `?` makes the map optional — if no streams match, the map is silently ignored instead of failing:

```bash
# Map video and optionally audio (ignore if no audio exists)
ffmpeg -i INPUT -map 0:v -map 0:a? OUTPUT
```

Without `?`, an unmatched map causes ffmpeg to exit with an error.

## View Specifiers (Multiview Video)

For multiview/free-viewpoint video, a view specifier can follow the stream specifier:

| Syntax | Meaning |
|--------|---------|
| `view:view_id` | Select by view ID; `all` for all views interleaved |
| `vidx:view_idx` | Select by view index (0 = base view) |
| `vpos:position` | Select by display position (`left`, `right`) |

Default for transcoding: `vidx:0` (base view only). Not supported for streamcopy.

```bash
ffmpeg -i INPUT -map 0:v:0:view:all OUTPUT
```

## `-map` Examples

```bash
# Map all streams from first input
ffmpeg -i INPUT -map 0 output

# Map second stream from first input
ffmpeg -i INPUT -map 0:1 out.wav

# Map stream 2 from input a.mov and stream 6 from b.mov
ffmpeg -i a.mov -i b.mov -c copy -map 0:2 -map 1:6 out.mov

# Map all video + third audio stream
ffmpeg -i INPUT -map 0:v -map 0:a:2 OUTPUT

# Map all except second audio (negative map)
ffmpeg -i INPUT -map 0 -map -0:a:1 OUTPUT

# Map from filtergraph output
ffmpeg -i video.mkv -i image.png -filter_complex "[0:v][1:v]overlay[out]" -map "[out]" output.mkv

# Map English audio by metadata
ffmpeg -i INPUT -map 0:m:language:eng OUTPUT
```

## Loopback Decoders

The `-dec` directive creates a loopback decoder that decodes an encoder's output and feeds it back to a filtergraph:

```bash
ffmpeg -i input.mkv \
  -filter_complex '[0:v]scale=size=hd1080,split=outputs=2[for_enc][orig_scaled]' \
  -c:v libx264 -map '[for_enc]' output.mkv \
  -dec 0:0 \
  -filter_complex '[dec:0][orig_scaled]hstack[stacked]' \
  -map '[stacked]' -c:v ffv1 comparison.mkv
```

- `-dec 0:0` creates loopback decoder 0 that decodes output stream 0:0
- `[dec:0]` in filtergraph links to that decoder
- Each `-dec` creates a new loopback decoder with successive indices (0, 1, 2, ...)
- Decoding AVOptions can be placed **before** `-dec`

## Automatic Stream Selection (No `-map`)

When `-map` is not used, ffmpeg selects streams automatically:

| Stream Type | Selection Rule |
|-------------|---------------|
| Video | First stream with highest resolution |
| Audio | First stream with most channels |
| Subtitle | First subtitle stream (if format supports it) |

Automatic selection is **disabled** for any output that uses `-map`.

## Stream Creation Order

Output streams are created in the order `-map` options appear on the command line. Each `-map` creates one output stream (or more, if the specifier matches multiple streams without the optional `?`).

## Mapping and Filtergraphs

When a complex filtergraph is present:

- **Unlabeled outputs** from the filtergraph are automatically added to the output
- **Labeled outputs** are added via `-map '[label]'`
- If a filtergraph produces output for a stream type, automatic selection for that type is disabled for downstream outputs
- Each output label can only be mapped once — mapping the same label to multiple outputs causes an error

## Edge Cases

- **Map to non-existent input**: `ffmpeg -i input -map 1:0 output` → fails (input 1 doesn't exist)
- **Duplicate label mapping**: `-map '[out]' -map '[out]'` → fails
- **Unmapped streams**: Without `-map`, automatically selected; with `-map`, only mapped streams appear
- **Metadata mapping**: `-map_metadata` is a separate option from `-map` — it copies metadata, not streams
