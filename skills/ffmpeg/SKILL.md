---
name: ffmpeg
description: >
  Use when working with ffmpeg CLI argument lexing or completion — the stream-oriented
  command structure, option syntax, stream specifiers, filter graph DSL, value types,
  option scopes, and stream mapping. Triggers on: "ffmpeg", "ffmpeg cli", "ffmpeg arguments",
  "ffmpeg options", "ffmpeg flags", "stream specifier", "filter_complex", "ffmpeg filter",
  "ffmpeg map", "ffmpeg lexer", "ffmpeg completion", "ffmpeg value types", "ffmpeg syntax", "ffplay", "ffprobe",
  "ffmpeg stream", "ffmpeg pipeline", "ffmpeg transcoding", "-filter_complex",
  "stream_specifier", "-map", "ffmpeg quoting".
user-invocable: true
---

# ffmpeg CLI In-Depth Reference

ffmpeg's command-line is a **stream-oriented pipeline**, not a traditional flag tree. Arguments form a linear sequence of inputs, per-file/per-stream options, and outputs — where options apply to the *next* file and "flags" are more like transformers bound to a preceding input. This skill provides the reference material needed to write a lexer for ffmpeg CLI arguments.

## Data Flow

```
global_options
  → [input_file_options -i input_url]...
  → [filter_complex_options]?
  → [output_file_options output_url]...
```

Options are positional: each option applies to the next `-i` (input) or the next output URL. Global options must come first. Do not mix input and output options — all inputs precede all outputs.

## Sub-Resources

Load the reference that matches your task. When in doubt, load multiple references.

| Keywords | Reference |
|----------|----------|
| stream model, pipeline, demuxer, decoder, filter, encoder, muxer, elementary stream, transcoding, streamcopy, flow, data flow | [references/stream-model.md](references/stream-model.md) |
| option syntax, flag, dash, double dash, boolean flag, positional option, stream specifier suffix, colon syntax, option argument, quoting, escaping | [references/option-syntax.md](references/option-syntax.md) |
| option scope, global, per-file, per-stream, input-only, output-only, option application order, positional | [references/option-scopes.md](references/option-scopes.md) |
| stream specifier, stream_spec, stream type, :v, :a, :s, :d, :t, program, group, metadata, disposition, usable | [references/stream-specifier.md](references/stream-specifier.md) |
| filter graph, filtergraph, simple filter, complex filter, -vf, -af, -filter_complex, link label, pad, filter chain, filter option, lavfi | [references/filtergraph.md](references/filtergraph.md) |
| -map, stream mapping, negative map, optional map, view specifier, linklabel mapping, loopback decoder | [references/mapping.md](references/mapping.md) |
| value types, time duration, video size, video rate, ratio, pixel format, sample format, channel layout, color, codec, format, boolean, integer, float, string, expression | [references/value-types.md](references/value-types.md) |
| ffplay, player, display, fullscreen, show mode, sync, vulkan renderer, playback control | [references/ffplay.md](references/ffplay.md) |
| ffprobe, probe, inspector, show streams, show format, output format, print format, select streams, read intervals, sections | [references/ffprobe.md](references/ffprobe.md) |

## Quick Guide

- **How does the ffmpeg pipeline work end-to-end?** → [references/stream-model.md](references/stream-model.md)
- **How do I parse an option name and its stream specifier suffix?** → [references/option-syntax.md](references/option-syntax.md) and [references/stream-specifier.md](references/stream-specifier.md)
- **Which options apply to which scope (global, per-file, per-stream)?** → [references/option-scopes.md](references/option-scopes.md)
- **How do I lex a stream specifier like `:a:1` or `:m:language:eng`?** → [references/stream-specifier.md](references/stream-specifier.md)
- **How do I parse a filter graph string?** → [references/filtergraph.md](references/filtergraph.md)
- **How does `-map` select streams and what are its sub-formats?** → [references/mapping.md](references/mapping.md)
- **What value types does ffmpeg use for option arguments?** → [references/value-types.md](references/value-types.md)
- **How does quoting/escaping work in ffmpeg option values?** → [references/option-syntax.md](references/option-syntax.md)
- **How do complex filtergraphs connect inputs and outputs?** → [references/filtergraph.md](references/filtergraph.md) and [references/mapping.md](references/mapping.md)
- **How does ffplay's CLI model differ from ffmpeg?** → [references/ffplay.md](references/ffplay.md)
- **How does ffprobe's CLI model differ from ffmpeg?** → [references/ffprobe.md](references/ffprobe.md)

## Cross-Project References

- For **shell quoting rules** (bash/zsh/fish escaping that wraps ffmpeg arguments), see the **bash**, **zsh**, or **fish** skills.
- For **carapace completion framework** internals, see the **carapace** skill.