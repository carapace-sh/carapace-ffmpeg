# Value Types

The data types ffmpeg uses for option arguments — how to recognize and parse each type.

> **Source of truth**: <https://ffmpeg.org/ffmpeg-utils.html> and `ffmpeg -h full` output.

## Overview

Each ffmpeg option expects a specific value type. The `-h full` output annotates options with type information:

```
-b <int64>           Set bitrate (in bits/s)
-ar <int>            Set audio sampling rate
-strict <int>        How strictly to follow the standards
-pix_fmt <fmt>       Set pixel format
-r <rate>            Set frame rate
-ss <time>           Set start time
```

## Type Reference

### Boolean

```
true | false | 1 | 0
```

Used by flag options. Some flags are presence-only (just `-y` enables), others accept explicit true/false.

Examples: `-y`, `-n`, `-vn`, `-an`, `-sn`, `-report`, `-hide_banner`, `-shortest`.

### Integer (`<int>`)

Signed or unsigned integer. Some have explicit ranges:

```
-strict <int>        (from INT_MIN to INT_MAX) (default 0)
-threads <int>       (from 0 to INT_MAX) (default 1)
```

Some integer options have named constant values (enums):

```
-strict
     very            2
     strict          1
     normal          0
     unofficial      -1
     experimental    -2
```

### Integer 64-bit (`<int64>`)

Same as integer but 64-bit range. Used for large values like bitrates and file sizes.

```
-b <int64>           (from 0 to I64_MAX) (default 200000)
-fs <limit_size>     Set the limit file size in bytes
```

### Float / Double (`<float>`, `<double>`)

Floating-point numbers. May have range restrictions:

```
-b_qoffset <float>   (from -FLT_MAX to FLT_MAX) (default 1.25)
```

### String (`<string>`)

Free-form text. Subject to ffmpeg quoting rules (see [option-syntax.md](option-syntax.md)).

```
-metadata <string>    key=value
-filter <string>      filter graph description
```

### Time Duration

Two accepted syntaxes:

```
[-][<HH>:]<MM>:<SS>[.<m>...]       # 12:03:45.500
[-]<S>+[.<m>...][s|ms|us]          # 55, 0.2, 200ms, 200000us
```

| Example | Value |
|---------|-------|
| `55` | 55 seconds |
| `0.2` | 0.2 seconds |
| `200ms` | 200 milliseconds |
| `200000us` | 200000 microseconds |
| `12:03:45` | 12 hours, 3 minutes, 45 seconds |
| `23.189` | 23.189 seconds |
| `-0.5` | Negative 0.5 seconds |

Used by: `-t`, `-ss`, `-to`, `-itsoffset`, stats period, etc.

### Time / Timestamp

Absolute timestamp format:

```
[(YYYY-MM-DD|YYYYMMDD)[T|t| ]]((HH:MM:SS[.m...]]])|(HHMMSS[.m...]]]))[Z]
now
```

Used by: `-timestamp`, progress reporting.

### Video Size

```
widthxheight    # 1920x1080
abbreviation    # hd1080, vga, ntsc, etc.
```

Standard abbreviations (partial list):

| Abbreviation | Size |
|--------------|------|
| `ntsc` | 720x480 |
| `pal` | 720x576 |
| `vga` | 640x480 |
| `svga` | 800x600 |
| `xga` | 1024x768 |
| `hd720` | 1280x720 |
| `hd1080` | 1920x1080 |
| `4k` | 4096x2160 (DCI 4K) |
| `uhd2160` | 3840x2160 |
| `sqcif` | 128x96 |
| `qcif` | 176x144 |
| `cif` | 352x288 |

Used by: `-canvas_size`, `-video_size`, `scale` filter `w`/`h`.

### Video Rate (Frame Rate)

```
frame_rate_num/frame_rate_den    # 30000/1001
integer_number                   # 25
float_number                     # 29.97
abbreviation                     # ntsc, pal, film, etc.
```

| Abbreviation | Rate |
|--------------|------|
| `ntsc` | 30000/1001 |
| `pal` | 25/1 |
| `film` | 24/1 |
| `ntsc-film` | 24000/1001 |

Used by: `-r`, `-fpsmax`.

### Ratio

```
numerator:denominator    # 16:9
expression               # 1.7777
```

Can be a fraction or decimal. `0:0` represents undefined. Infinite (1/0) and negative values are valid.

Used by: `-aspect`, `-sar`, time_base, etc.

### Pixel Format

Name string from the set reported by `-pix_fmts`:

```
yuv420p, yuv422p, yuv444p, rgb24, bgr24, rgba, nv12, nv21, ...
```

Each pixel format has flags: input-supported (I), output-supported (O), hardware (H), paletted (P), bitstream (B).

Used by: `-pix_fmt`, `format` filter, encoder options.

### Sample Format

Name string from the set reported by `-sample_fmts`:

```
u8, s16, s32, flt, dbl, u8p, s16p, s32p, fltp, dblp, s64, s64p
```

The `p` suffix means planar (interleaved vs. separate planes). The prefix indicates type and bit depth.

Used by: `-sample_fmt`, `aformat` filter.

### Channel Layout

Name string from the set reported by `-layouts`. Two forms:

1. **Standard layout names**: `mono`, `stereo`, `5.1`, `7.1`, etc.
2. **Individual channels**: `FL+FR+FC+LFE` (front left + front right + front center + low frequency)

| Layout | Decomposition |
|--------|-------------|
| `mono` | FC |
| `stereo` | FL+FR |
| `5.1` | FL+FR+FC+LFE+BL+BR |
| `7.1` | FL+FR+FC+LFE+BL+BR+SL+SR |

Used by: `-ac`, `aformat` filter, audio encoder options.

### Color

```
color_name                         # Red, AliceBlue, etc.
[0x|#]RRGGBB[AA][@alpha]          # 0xFF0000, #FF0000, 0xFF000080
random                             # Random color
```

Alpha can be:
- Hex: `0x80` (0x00-0xFF)
- Decimal: `0.5` (0.0-1.0)
- Omitted: defaults to `0xFF` (fully opaque)

Used by: `color` source filter, various drawing filters.

### Codec

Codec name string from `-encoders` or `-decoders`. Special value: `copy` (streamcopy without decoding/encoding).

```
libx264, libx265, aac, libmp3lame, copy, ...
```

Codec names can have library prefixes: `libx264`, `libaom-av1`, `libvpx-vp9`.

Used by: `-c`, `-c:v`, `-c:a`, `-vcodec`, `-acodec`.

### Format (Container)

Format name string from `-formats` or `-muxers`/`-demuxers`:

```
mp4, mkv, avi, mov, flv, webm, mp3, ogg, ...
```

Used by: `-f`.

### Flags / Bitmask

A set of named constants that can be combined with `+`:

```
-err_detect crccheck+bitstream+buffer
-debug pict+rc+mb_type
```

Each flag option lists its named values with short descriptions.

### Dictionary (`key=value`)

Key-value pairs, typically for metadata:

```
-metadata title="My Video"
-metadata:s:a:0 language=eng
```

### Expression

Some options accept mathematical expressions with a full expression language including variables, functions, and operators. Common in filter options:

```
scale=w='2*iw':h='2*ih'
drawtext=x='w-text_w-10':y='h-text_h-10'
```

Expression functions include: `abs`, `ceil`, `floor`, `round`, `max`, `min`, `pow`, `sqrt`, `log`, `sin`, `cos`, `rand`, `mod`, etc.

## Type Detection for the Lexer

The `-h full` output encodes type information in each option line:

```
  -b_qfactor         <float>      E..V....... QP factor between P- and B-frames
  -b_qoffset         <float>      E..V....... QP offset between P- and B-frames
  -err_detect        <flags>      ED.VAS..... set error detection flags
  -threads           <int>        ED.VA...... set the number of threads
  -pix_fmt           <fmt>        E..V....... set pixel format
```

The type is in angle brackets: `<float>`, `<int>`, `<flags>`, `<fmt>`, `<string>`, `<duration>`, `<ratio>`, etc.

### Flags Column

After the type, the flags column indicates which contexts the option applies to:

```
E = Encoding
D = Decoding
V = Video
A = Audio
S = Subtitle
```

```
  -strict <int> ED.VA......   (applies to Encoding/Decoding, Video/Audio)
  -pix_fmt <fmt> E..V.......  (applies to Encoding, Video only)
```

## ffplay Value Types

### Show Mode (ffplay)

Used by `-showmode`:

```
0 | video       # Display video (default)
1 | waves       # Show audio waveform
2 | rdft        # Show audio frequency data (RDFT)
```

Both numeric and string forms are accepted.

### Sync Type (ffplay)

Used by `-sync`:

```
audio   # Audio clock is master (default)
video   # Video clock is master
ext     # External clock is master
```

### Vulkan Parameters (ffplay)

Used by `-vulkan_params`. Colon-separated key=value pairs:

```
key1=value1:key2=value2
```

Specific keys depend on the libplacebo version. The value is a free-form string with colon-separated key=value pairs.

## ffprobe Value Types

### Probe Output Format (ffprobe)

Used by `-of` / `-print_format` / `-output_format`:

```
writer_name[=writer_options]
```

|| Writer | Description |
|--------|-------------|
| `default` | Human-readable key=value (default) |
| `compact` | Compact one-line format |
| `csv` | CSV format |
| `flat` | Flat key=value with dot-separated paths |
| `ini` | INI-style sections |
| `json` | JSON format |
| `xml` | XML format |

Writer options follow after `=`: `-of json=compact=1`.

### Data Dump Format (ffprobe)

Used by `-data_dump_format`:

```
xxd      # Hex+ASCII dump (default)
base64   # Base64-encoded
```

### Show Optional Fields (ffprobe)

Used by `-show_optional_fields`:

```
always | 1   # Always print, even if invalid
never  | 0   # Never print invalid fields
auto   | -1  # Print only if valid (default)
```

### Read Intervals (ffprobe)

Used by `-read_intervals`:

```
[START|+START_OFFSET][%[END|+END_OFFSET]]
```

Multiple intervals are separated by `,`.

## Edge Cases

- **Numeric with suffix**: `128k` = 128 kilobits, `2M` = 2 megabits — the suffix is part of the value
- **Negative numbers**: `-1`, `-0.5` — these can look like new options to a naïve parser
- **Fraction strings**: `30000/1001` — slash-separated numerator/denominator
- **Named constants**: `auto`, `default`, `all`, `unknown` — these are enum values, not free-form strings
- **Codec-specific options**: Each codec may add its own private options with their own types; these are only visible with `-h encoder=NAME` or `-h decoder=NAME`