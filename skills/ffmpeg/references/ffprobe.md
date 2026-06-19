# ffprobe CLI Model

How ffprobe's command-line differs from ffmpeg — the simplified scope model, probe-specific options, and output format control.

> **Source of truth**: <https://ffmpeg.org/ffprobe.html>.

## Invocation Syntax

```
ffprobe [options] input_url
```

ffprobe is a **probe/inspection tool** — it reads a single input and prints metadata, stream information, or packet/frame data. There is no encoding, no muxing, and no output file (except via `-o` to redirect text output).

### Input URL

Like ffplay, ffprobe accepts the input in two ways:

```bash
ffprobe input.mp4          # positional (no -i)
ffprobe -i input.mp4      # explicit -i flag
```

Both forms are valid. There is **only one input** — multiple `-i` flags are not supported.

### Output

By default, ffprobe writes to stdout. The `-o` option redirects to a file:

```bash
ffprobe -o output.txt -show_streams input.mp4
```

This is **text output** (the probe result), not a media output file.

## Pipeline

ffprobe's data flow is a **read-and-inspect** pipeline — no encoding, no rendering:

```
input_url ─► demuxer ─► decoder? ─► inspector (print metadata/frames/packets)
```

|| Component | Role | Created By |
|-----------|------|-----------|
| **Demuxer** | Reads container, splits into elementary streams | Each input |
| **Decoder** | Optional — only if `-show_frames` or `-show_log` is used | `-c:*` or automatic |
| **Inspector** | Formats and prints stream/packet/frame information | `-show_*` options |

Decoding is lazy — ffprobe does not decode frames unless asked to via `-show_frames`, `-show_log`, or `-analyze_frames`. Otherwise it only reads container-level metadata.

## Simplified Scope Model

ffprobe uses a **subset** of ffmpeg's positional scope model:

```
ffprobe [global_options] [input_options] input_url
```

|| Scope | Description |
|-------|-------------|
| **Global** | Entire ffprobe process (`-loglevel`, `-hide_banner`, `-of`, etc.) |
| **Input** | Options before the input URL (`-f`, `-ss`, `-c:*`, `-select_streams`, etc.) |

**No output section** — there is no encoding or muxing scope. This means:

- No `-map`, no `-metadata`, no output-only codec/bitrate options
- No output URL (non-option arguments are the input URL)
- Per-stream options apply to the input (decoders only)
- The scope never advances past `ScopeInputFile`

### Scope Transition

```
GLOBAL ──option──► INPUT ──input_url──► (done)
```

Same as ffplay — the scope stops at INPUT.

## Option Catalog

### Global Options (shared with ffmpeg)

These are the generic options common to all ff\* tools:

|| Option | Short | Type | Value Type | Description |
|--------|-------|------|------------|-------------|
| `L` | `L` | Boolean | | Show license |
| `h` | `h` | Value | string | Show help (optional arg: `long`, `full`, `decoder=`, etc.) |
| `?` | `?` | Value | string | Alias for `-h` |
| `version` | `version` | Boolean | | Show version |
| `buildconf` | `buildconf` | Boolean | | Show build configuration |
| `formats` | `formats` | Boolean | | Show available formats |
| `demuxers` | `demuxers` | Boolean | | Show available demuxers |
| `muxers` | `muxers` | Boolean | | Show available muxers |
| `devices` | `devices` | Boolean | | Show available devices |
| `codecs` | `codecs` | Boolean | | Show available codecs |
| `decoders` | `decoders` | Boolean | | Show available decoders |
| `encoders` | `encoders` | Boolean | | Show available encoders |
| `bsfs` | `bsfs` | Boolean | | Show available bitstream filters |
| `protocols` | `protocols` | Boolean | | Show available protocols |
| `filters` | `filters` | Boolean | | Show available filters |
| `pix_fmts` | `pix_fmts` | Boolean | | Show available pixel formats |
| `sample_fmts` | `sample_fmts` | Boolean | | Show available sample formats |
| `layouts` | `layouts` | Boolean | | Show channel layouts |
| `dispositions` | `dispositions` | Boolean | | Show stream dispositions |
| `colors` | `colors` | Boolean | | Show recognized color names |
| `sources` | `sources` | Value | device | List sources of input device |
| `sinks` | `sinks` | Value | device | List sinks of output device |
| `loglevel` | `v` | Value | loglevel | Set logging level |
| `report` | `report` | Boolean | | Generate report |
| `hide_banner` | `hide_banner` | Boolean | | Suppress banner |
| `cpuflags` | `cpuflags` | Value | string | Force CPU flags |
| `cpucount` | `cpucount` | Value | int | Force CPU count |
| `max_alloc` | `max_alloc` | Value | int | Max allocation size |

### Main Options (ffprobe-specific)

|| Option | Short / Aliases | Type | Value Type | AcceptsSpec | Description |
|--------|-----------------|------|------------|-------------|-------------|
| `f` | `f` | Value | format | | Force format |
| `unit` | `unit` | Boolean | | | Show unit of displayed values |
| `prefix` | `prefix` | Boolean | | | Use SI prefixes for values |
| `byte_binary_prefix` | `byte_binary_prefix` | Boolean | | | Force binary prefixes for bytes |
| `sexagesimal` | `sexagesimal` | Boolean | | | Use HH:MM:SS.MICROSECONDS format |
| `pretty` | `pretty` | Boolean | | | Shorthand for `-unit -prefix -byte_binary_prefix -sexagesimal` |
| `output_format` | `of`, `print_format` | Value | probe_output_format | | Output printing format (json, xml, csv, etc.) |
| `sections` | `sections` | Boolean | | | Print section structure and exit |
| `select_streams` | `select_streams` | Value | stream_spec | | Select streams by specifier |
| `show_data` | `show_data` | Boolean | | | Show payload data as hex+ASCII |
| `show_data_hash` | `show_data_hash` | Value | string | | Hash algorithm for payload data |
| `data_dump_format` | `data_dump_format` | Value | data_dump_format | | Format for `-show_data`: xxd (default), base64 |
| `show_error` | `show_error` | Boolean | | | Show probe error info |
| `show_format` | `show_format` | Boolean | | | Show container format info |
| `show_entries` | `show_entries` | Value | string | | Set entries to show (section=key1,key2:...) |
| `show_packets` | `show_packets` | Boolean | | | Show packet info |
| `show_frames` | `show_frames` | Boolean | | | Show frame and subtitle info |
| `show_log` | `show_log` | Value | loglevel | | Show decoder log per frame (requires `-show_frames`) |
| `show_streams` | `show_streams` | Boolean | | | Show stream info |
| `show_programs` | `show_programs` | Boolean | | | Show program info |
| `show_stream_groups` | `show_stream_groups` | Boolean | | | Show stream group info |
| `show_chapters` | `show_chapters` | Boolean | | | Show chapter info |
| `count_frames` | `count_frames` | Boolean | | | Count frames per stream |
| `count_packets` | `count_packets` | Boolean | | | Count packets per stream |
| `read_intervals` | `read_intervals` | Value | string | | Read only specified intervals |
| `show_private_data` | `private` | Boolean | | | Show private data (default: enabled) |
| `show_program_version` | `show_program_version` | Boolean | | | Show program version |
| `show_library_versions` | `show_library_versions` | Boolean | | | Show library versions |
| `show_versions` | `show_versions` | Boolean | | | Shorthand for `-show_program_version -show_library_versions` |
| `show_pixel_formats` | `show_pixel_formats` | Boolean | | | Show pixel format info |
| `show_optional_fields` | `show_optional_fields` | Value | show_optional_fields | | Print invalid/non-applicable fields: always/1, never/0, auto/-1 |
| `analyze_frames` | `analyze_frames` | Boolean | | | Analyze frames for additional stream info (requires `-show_streams`) |
| `bitexact` | `bitexact` | Boolean | | | Force bitexact output |
| `i` | `i` | Value | file_url | | Input URL |
| `o` | `o` | Value | file_url | | Output file (default: stdout) |
| `codec` | `c` | Value | codec | Yes | Force decoder by media specifier (a/v/s/d) |

## Differences from ffmpeg Summary

|| Aspect | ffmpeg | ffprobe |
|--------|--------|---------|
| Invocation | `ffmpeg [g_opts] {-i in}... {out}...` | `ffprobe [opts] input_url` |
| Input | Multiple `-i` inputs | Single input (positional or `-i`) |
| Output | One or more output URLs | Text output to stdout or `-o` file |
| Scope model | Global → Input → Output | Global → Input only |
| Encoding | Full encoding pipeline | None |
| Filtering | `-vf`/`-af`/`-filter_complex` | None |
| Mapping | `-map` for stream routing | `-select_streams` for display filtering |
| Show options | N/A | Extensive `-show_*` and `-count_*` options |
| Output format | Container format (`-f`) | Text format (`-of`/`-print_format`): json, xml, csv, etc. |
| Decoding | Always (unless streamcopy) | Optional (only with `-show_frames`/`-show_log`) |

## Output Formats

The `-of`/`-print_format`/`-output_format` option controls how ffprobe formats its output:

|| Writer | Description |
|--------|-------------|
| `default` | Human-readable key=value format |
| `compact` | One line per property |
| `csv` | CSV format |
| `flat` | Flat key=value with dot-separated paths |
| `ini` | INI-style format |
| `json` | JSON format |
| `xml` | XML format |

Writers accept options via `=` suffix: `-of json=compact=1`.

## Read Intervals

The `-read_intervals` option limits which portions of the file are probed:

```
[START|+START_OFFSET][%[END|+END_OFFSET]]
```

|| Example | Meaning |
|---------|---------|
| `00:01:00%00:02:00` | Read from 1:00 to 2:00 |
| `+5%+15` | Read from offset 5s to offset 15s |
| `%+30` | Read first 30 seconds |

Multiple intervals are separated by `,`.

## Edge Cases

- **Positional input without `-i`**: `ffprobe input.mp4` — the non-option argument is the input URL (unlike ffmpeg where positional args are outputs)
- **`-o` is not a media output**: It redirects the text probe output, not a transcoded file
- **`-codec` is decoder-only**: ffprobe's `-c:v` selects a decoder, not an encoder
- **No `-filter_complex`**: ffprobe cannot apply filtergraphs
- **`-select_streams` vs `-map`**: `-select_streams` filters which streams appear in the output report; `-map` (ffmpeg) routes streams to output files. They serve different purposes.
- **`-show_log` requires `-show_frames`**: The decoder log is per-frame, so frames must be shown
- **`-analyze_frames` requires `-show_streams`**: Extra stream-level info comes from frame analysis
