# Option Syntax

How ffmpeg option names are structured: dashes, stream specifier suffixes, boolean flags, and quoting.

> **Source of truth**: <https://ffmpeg.org/ffmpeg.html> and <https://ffmpeg.org/ffmpeg-utils.html>.

## Option Name Patterns

ffmpeg uses a single-dash prefix for all options — there are **no double-dash `--` long options** (unlike GNU conventions). The `--help` form is accepted but is just an alias for `-h`.

### Core Patterns

| Pattern | Example | Description |
|---------|---------|-------------|
| `-x` | `-y`, `-n`, `-vn` | Single-letter boolean flag |
| `-x value` | `-f mp4`, `-t 60` | Single-letter option with value |
| `-x:spec value` | `-c:v libx264`, `-b:a:1 128k` | Option with stream specifier suffix |
| `-word value` | `-metadata title=foo` | Multi-letter (word) option with value |
| `-word:spec value` | `-metadata:s:a title=foo` | Word option with stream specifier |

### Boolean Flags

Boolean flags have no value argument — their presence enables the feature:

```bash
-y          # overwrite output files
-n          # never overwrite
-vn         # disable video
-an         # disable audio
-sn         # disable subtitle
-hide_banner
-report
```

Some flags have inverse forms:

| Flag | Inverse | Meaning |
|------|---------|---------|
| `-y` | `-n` | overwrite vs. never overwrite |

## Stream Specifier Suffix

The stream specifier is appended to the option name after a **colon**:

```
option_name[:stream_specifier] value
```

The colon is part of the option name token — the lexer must split `c:v:1` into option `c`, specifier `v:1`. See [stream-specifier.md](stream-specifier.md) for the full specifier grammar.

### Lexing Implication

The lexer must handle compound tokens like `-c:v:1` as a single option token where:
- `-c` is the option base name
- `:v:1` is the stream specifier

The colon is the delimiter between the option name and the stream specifier, but colons *also* appear *within* the specifier itself (e.g., `-c:v:1`, `-disposition:default+forced`).

## Option Name Variants

Many options have short and long forms that are aliases:

| Short | Long | Notes |
|-------|------|-------|
| `-c` | `-codec` | `-codec` documented as alias for `-c` |
| `-vf` | `-filter:v` | Derived: `-vf` = `-filter:v` |
| `-af` | `-filter:a` | Derived: `-af` = `-filter:a` |
| `-vcodec` | `-c:v` | `-vcodec` is alias for `-c:v` |
| `-acodec` | `-c:a` | `-acodec` is alias for `-c:a` |
| `-scodec` | `-c:s` | `-scodec` is alias for `-c:s` |
| `-dcodec` | `-c:d` | `-dcodec` is alias for `-c:d` |
| `-ab` | `-b:a` | `-ab` is alias for `-b:a` |
| `-b` | (contextual) | Alone means "video bitrate", deprecated in favor of `-b:v` |
| `-h` | `-help`, `--help`, `-?` | All aliases |

For lexing, these aliases expand to the same semantic — the lexer should normalize them.

## Quoting and Escaping

ffmpeg uses its own quoting rules for option values, **separate from shell quoting**. The shell strips its quoting first, then ffmpeg applies its own.

### ffmpeg Quoting Rules

1. **Single quotes** `'...'` — include contents literally; the `'` itself cannot be quoted inside
2. **Backslash escaping** `\x` — escapes any special character `x`
3. **Leading/trailing whitespace** is stripped unless quoted or escaped

### Interaction with Shell Quoting

There are **two layers** of quoting: the shell layer, then the ffmpeg layer. A typical pipeline input:

```bash
# Shell sees: single-quoted string, passes verbatim to ffmpeg
ffmpeg -filter_complex '[0:v]scale=1280:720[out]'
```

Inside that string, ffmpeg sees its own quoting. The `[0:v]` and `[out]` are filter labels, not shell constructs.

### Special Characters in Values

| Context | Special Characters | Escape |
|---------|-------------------|--------|
| General option values | `:`, `'`, `\` | `\` or quoting |
| Filter option values | `=`, `:`, `'`, `\` | `\` or quoting |
| Metadata values (`key=value`) | `=` (delimiter), `:` | `\` or quoting |
| Stream specifier `m:key:value` | `:` in key or value | `\:` |

### Example: Metadata with Colons

```bash
# Metadata key containing a colon must be escaped
-metadata 'title\:subtitle=foo'
```

## Value-less vs Value-taking Options

The lexer needs to classify each option as either:

1. **Boolean/flag** — no argument follows
2. **Value-taking** — next token is the argument

This classification must be known from a static option definition. Some options are context-dependent:

- `-t` takes a duration value
- `-y` takes no value
- `-f` takes a format name
- `-ss` takes a time offset

However, some ambiguous cases exist:
- `-b` alone takes a bitrate value
- `-b:v` (with specifier) also takes a bitrate value
- `-vn` takes no value (boolean "no video")

## Edge Cases

- **Option-like output URLs**: `ffmpeg -i input -output.mp4` — here `-output.mp4` could be mistaken for an option. The lexer must determine when we're past all options and into outputs.
- **Negative map**: `-map -0:a:1` — the `-` before `0` is part of the map value, not a new option
- **Option value starting with dash**: `-probesize -1` — the value `-1` is a number, not a flag
- **Consecutive boolean flags**: `-y -hide_banner -nostats` — three flags in a row with no values
