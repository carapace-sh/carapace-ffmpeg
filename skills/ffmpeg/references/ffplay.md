# ffplay CLI Model

How ffplay's command-line differs from ffmpeg — the simplified scope model, player-specific options, and input handling.

> **Source of truth**: <https://ffmpeg.org/ffplay.html>.

## Invocation Syntax

```
ffplay [options] [input_url]
```

Unlike ffmpeg which requires `-i` before each input and always has outputs, ffplay is a **media player** with a single input and no output section.

### Input URL

ffplay accepts the input URL in two ways:

```bash
ffplay input.mp4          # positional (no -i)
ffplay -i input.mp4       # explicit -i flag
```

Both forms are valid. Unlike ffmpeg, `-i` is optional and there is **only one input** — multiple `-i` flags are not supported.

## Pipeline

ffplay's data flow is a **read-and-play** pipeline — no encoding, no muxing, no output:

```
input_url ─► demuxer ─► decoder ─► filter? ─► renderer (SDL/Vulkan)
```

|| Component | Role | Created By |
|-----------|------|-----------|
| **Demuxer** | Reads container, splits into elementary streams | Each input | 
| **Decoder** | Decodes packets to raw frames | Automatic or `-c:*` |
| **Filter** | Optional simple filtergraph (`-vf`/`-af`) | `-vf`/`-af` |
| **Renderer** | Displays video, plays audio | SDL or Vulkan |

There is no encoder, no muxer, and no output file.

## Simplified Scope Model

ffplay uses a **subset** of ffmpeg's positional scope model:

```
ffplay [global_options] [input_options] [input_url]
```

|| Scope | Description |
|-------|-------------|
| **Global** | Entire ffplay process (`-loglevel`, `-hide_banner`, `-report`, etc.) |
| **Input** | Options before the input URL (`-f`, `-ss`, `-t`, `-an`, `-vn`, etc.) |

**No output section** — ffplay never transitions to an output scope. This means:

- No `-map`, no `-metadata`, no output-only options
- No output URL (non-option arguments are the input URL)
- Per-stream options (`-c:*`, `-an`, `-vn`) apply to the input
- The scope never advances past `ScopeInputFile`

### Scope Transition

```
GLOBAL ──option──► INPUT ──input_url──► (done)
```

Unlike ffmpeg's `GLOBAL → INPUT → OUTPUT`, ffplay stops at INPUT.

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

### Main Options (ffplay-specific)

|| Option | Short | Type | Value Type | AcceptsSpec | Description |
|--------|-------|------|------------|-------------|-------------|
| `x` | `x` | Value | int | | Force displayed width |
| `y` | `y` | Value | int | | Force displayed height |
| `fs` | `fs` | Boolean | | | Start in fullscreen mode |
| `an` | `an` | Boolean | | | Disable audio |
| `vn` | `vn` | Boolean | | | Disable video |
| `sn` | `sn` | Boolean | | | Disable subtitles |
| `ss` | `ss` | Value | duration | | Seek to position |
| `t` | `t` | Value | duration | | Play only duration seconds |
| `bytes` | `bytes` | Boolean | | | Seek by bytes |
| `seek_interval` | `seek_interval` | Value | float | | Custom seek interval (default: 10s) |
| `nodisp` | `nodisp` | Boolean | | | Disable graphical display |
| `noborder` | `noborder` | Boolean | | | Borderless window |
| `alwaysontop` | `alwaysontop` | Boolean | | | Window always on top |
| `volume` | `volume` | Value | int | | Startup volume (0-100) |
| `f` | `f` | Value | format | | Force format |
| `window_title` | `window_title` | Value | string | | Set window title |
| `left` | `left` | Value | int | | Window x position |
| `top` | `top` | Value | int | | Window y position |
| `loop` | `loop` | Value | int | | Loop count (0 = forever) |
| `showmode` | `showmode` | Value | show_mode | | Display mode: 0/video, 1/waves, 2/rdft |
| `vf` | `vf` | Value | filter | | Video filtergraph (ImplicitSpec: "v") |
| `af` | `af` | Value | filter | | Audio filtergraph (ImplicitSpec: "a") |
| `i` | `i` | Value | file_url | | Input URL |

### Advanced Options (ffplay-specific)

|| Option | Short | Type | Value Type | AcceptsSpec | Description |
|--------|-------|------|------------|-------------|-------------|
| `stats` | `stats` | Boolean | | | Print playback statistics |
| `fast` | `fast` | Boolean | | | Non-spec-compliant optimizations |
| `genpts` | `genpts` | Boolean | | | Generate PTS |
| `sync` | `sync` | Value | sync_type | | Master clock: audio, video, ext |
| `ast` | `ast` | Value | stream_spec | | Select audio stream by specifier |
| `vst` | `vst` | Value | stream_spec | | Select video stream by specifier |
| `sst` | `sst` | Value | stream_spec | | Select subtitle stream by specifier |
| `autoexit` | `autoexit` | Boolean | | | Exit when playback finishes |
| `exitonkeydown` | `exitonkeydown` | Boolean | | | Exit on any key press |
| `exitonmousedown` | `exitonmousedown` | Boolean | | | Exit on any mouse click |
| `codec` | `c` | Value | codec | Yes | Force decoder by media specifier (a/v/s) |
| `acodec` | `acodec` | Value | codec | | Force audio decoder (ImplicitSpec: "a") |
| `vcodec` | `vcodec` | Value | codec | | Force video decoder (ImplicitSpec: "v") |
| `scodec` | `scodec` | Value | codec | | Force subtitle decoder (ImplicitSpec: "s") |
| `autorotate` | `autorotate` | Boolean | | | Auto-rotate video by metadata (default: enabled) |
| `framedrop` | `framedrop` | Boolean | | | Drop video frames when out of sync |
| `infbuf` | `infbuf` | Boolean | | | Don't limit input buffer size |
| `filter_threads` | `filter_threads` | Value | int | | Threads for filter pipeline (0=auto) |
| `enable_vulkan` | `enable_vulkan` | Boolean | | | Use Vulkan renderer |
| `vulkan_params` | `vulkan_params` | Value | vulkan_params | | Vulkan configuration (key:value pairs) |
| `hwaccel` | `hwaccel` | Boolean | | | Use HW accelerated decoding |
| `video_bg` | `video_bg` | Value | string | | Video background: color, tiles (default), none |

## Differences from ffmpeg Summary

|| Aspect | ffmpeg | ffplay |
|--------|--------|--------|
| Invocation | `ffmpeg [g_opts] {-i in}... {out}...` | `ffplay [opts] [input_url]` |
| Input | Multiple `-i` inputs | Single input (positional or `-i`) |
| Output | One or more output URLs | None (player only) |
| Scope model | Global → Input → Output | Global → Input only |
| Encoding | Full encoding pipeline | Decoding only |
| Filter | Simple (`-vf`/`-af`) + Complex (`-filter_complex`) | Simple only (`-vf`/`-af`) |
| Mapping | `-map` for stream routing | No mapping |
| Display | N/A | Window options, fullscreen, Vulkan renderer |
| Playback | N/A | `-ss`, `-t`, `-loop`, `-autoexit`, `-sync` |
| Stream selection | `-map`, stream specifiers on `-c` | `-ast`, `-vst`, `-sst` selectors |

## Edge Cases

- **Positional input without `-i`**: `ffplay input.mp4` — the non-option argument is the input URL, not an output URL (unlike ffmpeg where positional args are outputs)
- **Multiple `-i`**: Undefined behavior — ffplay is designed for a single input
- **`-y` conflict**: In ffmpeg `-y` means "overwrite"; in ffplay `-y` means "force display height". The option namespace differs between tools.
- **`-codec` scope**: ffplay's `-codec` only selects decoders (no encoding context); the specifier is limited to `a`, `v`, `s`
- **`-vf`/`-af` are simple filtergraphs**: No `-filter_complex` option exists in ffplay
