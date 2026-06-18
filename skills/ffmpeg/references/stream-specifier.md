# Stream Specifiers

The grammar for stream specifiers — the suffix that targets specific streams within a file.

> **Source of truth**: <https://ffmpeg.org/ffmpeg.html> — Stream specifiers section.

## Overview

Stream specifiers are appended to per-stream option names after a colon. They select which stream(s) the option applies to. A specifier can match one stream or multiple streams.

```
-option_name[:stream_specifier] value
```

Example: `-codec:a:1 ac3` → option `codec`, specifier `a:1`, value `ac3` → apply `ac3` codec to the second audio stream.

An **empty specifier** (or no colon) matches all streams:
- `-codec copy` or `-codec: copy` → copy all streams

## Grammar

```
stream_specifier := stream_index
                  | stream_type [':' additional_specifier]
                  | 'g:' group_specifier [':' additional_specifier]
                  | 'p:' program_id [':' additional_specifier]
                  | '#' stream_id
                  | 'i:' stream_id
                  | 'm:' key [':' value]
                  | 'disp:' dispositions [':' additional_specifier]
                  | 'u'

additional_specifier := stream_index
                      | stream_type [':' additional_specifier]
                      | 'g:' group_specifier [':' additional_specifier]
                      | 'p:' program_id [':' additional_specifier]
                      | '#' stream_id
                      | 'i:' stream_id
                      | 'm:' key [':' value]
                      | 'disp:' dispositions [':' additional_specifier]
                      | 'u'

group_specifier := group_index
                 | '#' group_id
                 | 'i:' group_id
```

## Specifier Forms in Detail

### `stream_index`

Matches the stream with this index (0-based). When used as an **additional** specifier after a type, selects the Nth stream *of that type*.

```bash
-threads:1 4          # second stream overall (index 1)
-c:a:1 aac            # second audio stream (type 'a', then index 1 within audio)
```

### `stream_type[:additional_specifier]`

| Type Letter | Matches | Notes |
|------------|---------|-------|
| `v` | All video streams | Including attached pictures, thumbnails, cover art |
| `V` | Video streams excluding attached pictures | Only "real" video streams |
| `a` | Audio streams | |
| `s` | Subtitle streams | |
| `d` | Data streams | |
| `t` | Attachment streams | |

With an additional specifier, matches streams that are both the given type AND match the additional specifier.

```bash
-c:v libx264          # all video streams
-c:V libx264          # only non-thumbnail video streams
-b:a 128k             # all audio streams
-c:a:1 aac            # second audio stream
```

### `g:group_specifier[:additional_specifier]`

Matches streams in a stream group.

| group_specifier | Meaning |
|----------------|---------|
| `group_index` | Group by index (0-based) |
| `#group_id` | Group by ID |
| `i:group_id` | Group by ID (alternate syntax) |

```bash
-c:g:0 copy           # all streams in group 0
```

### `p:program_id[:additional_specifier]`

Matches streams in the program with ID `program_id`.

```bash
-c:p:0x1 libx264     # streams in program with ID 0x1
```

### `#stream_id` or `i:stream_id`

Match by stream ID (e.g., PID in MPEG-TS).

```bash
-c:#0x1F3 copy       # stream with PID 0x1F3
```

### `m:key[:value]`

Match streams that have a metadata tag `key` with the given `value` (or any value if `value` is omitted). Colons in key or value must be backslash-escaped.

```bash
-c:m:language:eng copy   # English audio tracks
-metadata:m:language:eng title="English"   # metadata on English streams
```

### `disp:dispositions[:additional_specifier]`

Match by stream disposition. Dispositions are joined with `+`.

Available dispositions (from `ffmpeg -dispositions`): `default`, `dub`, `original`, `comment`, `lyrics`, `karaoke`, `forced`, `hearing_impaired`, `visual_impaired`, `clean_effects`, `attached_pic`, `timed_thumbnails`, `non_diegetic`, `captions`, `descriptions`, `metadata`, `dependent`, `still_image`, `multilayer`.

```bash
-disp:default+forced   # match streams with both default and forced dispositions
```

### `u`

Matches streams with **usable configuration** — codec must be defined and essential info (video dimensions, audio sample rate) is present.

```bash
-c:u copy            # copy all usable streams
```

## Lexer Implications

### Token Splitting

Given an option like `-c:a:1`, the lexer must split on colons from left to right:

1. Option base: `c`
2. Specifier starts: `a:1`
3. Within specifier: `a` is the stream type, `1` is the additional index

But colons are also used as **value delimiters** in certain contexts (e.g., `-metadata:s:a title=foo` where `s:a` is the metadata specifier, not `s` + `a` as separate tokens). The specifier grammar is nested and ambiguous — the lexer must parse it recursively.

### Ambiguity: Specifier vs. Value

The colon after the option base is always the specifier boundary. The challenge is parsing the specifier correctly:

```
-b:a:1 128k    → option 'b', specifier 'a:1', value '128k'
-metadata:s:a key=val  → option 'metadata', specifier 's:a', value 'key=val'
```

The `m:key[:value]` form is particularly tricky — after `m:`, the next colon *might* be a key/value delimiter or *might* be an escaped colon within the key.

### Recursive Specifiers

The `additional_specifier` allows recursion:
- `a:1` = type `a`, index `1`
- `a:g:0:1` = type `a`, group `0`, index `1`
- `p:0xa:1` = program `0xa`, index `1`

The grammar is potentially unbounded in depth, though in practice 2-3 levels is the maximum used.

## Edge Cases

- **Empty specifier**: `-c: copy` (colon present but no specifier) matches all streams
- **Specifier matching multiple streams**: `-c:a aac` matches *all* audio streams; the option is applied to each
- **Specifier matching no streams**: Silently ignored (option has no effect) unless `?` is used in `-map` (see [mapping.md](mapping.md))
- **Specifier on non-per-stream option**: Undefined behavior — the specifier is typically ignored or causes an error
