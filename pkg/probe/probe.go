package probe

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// StreamInfo holds metadata for a single stream probed from a media file.
type StreamInfo struct {
	Index       int               `json:"index"`
	ID          string            `json:"id,omitempty"`
	CodecName   string            `json:"codec_name"`
	CodecType   string            `json:"codec_type"`
	SampleFmt   string            `json:"sample_fmt,omitempty"`
	PixFmt      string            `json:"pix_fmt,omitempty"`
	Disposition map[string]int    `json:"disposition"`
	Tags        map[string]string `json:"tags,omitempty"`
}

type ffprobeOutput struct {
	Streams []StreamInfo      `json:"streams"`
	Format  ffprobeFormatTags `json:"format"`
}

type ffprobeFormatTags struct {
	Tags map[string]string `json:"tags"`
}

// expandPath expands a leading ~/ in the path to the user's home directory.
func expandPath(path string) string {
	if len(path) >= 2 && path[0] == '~' && path[1] == '/' {
		if home, err := os.UserHomeDir(); err == nil {
			return home + path[1:]
		}
	}
	return path
}

// Probe runs ffprobe on a local file and returns stream metadata.
// Returns nil and no error if ffprobe is unavailable or the file cannot be probed.
func Probe(inputURL string) ([]StreamInfo, error) {
	inputURL = expandPath(inputURL)
	if !isLocalFile(inputURL) {
		return nil, nil
	}

	cmd := exec.Command("ffprobe",
		"-hide_banner",
		"-show_streams",
		"-show_format",
		"-of", "json=c=1",
		"--",
		inputURL,
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, nil
	}

	var result ffprobeOutput
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, nil
	}

	// Merge format-level tags into each stream's tags.
	// Stream-level tags take precedence over format-level tags.
	formatTags := result.Format.Tags
	for i := range result.Streams {
		if len(formatTags) > 0 {
			if result.Streams[i].Tags == nil {
				result.Streams[i].Tags = make(map[string]string)
			}
			for k, v := range formatTags {
				if _, exists := result.Streams[i].Tags[k]; !exists {
					result.Streams[i].Tags[k] = v
				}
			}
		}
	}
	return result.Streams, nil
}

// MetadataValues extracts unique values for a metadata key from the given streams.
// Key matching is case-insensitive, matching ffmpeg's behavior for -m:KEY: lookups.
func MetadataValues(streams []StreamInfo, key string) []string {
	keyLower := strings.ToLower(key)
	seen := make(map[string]bool)
	var vals []string
	for _, s := range streams {
		if s.Tags != nil {
			for k, v := range s.Tags {
				if strings.ToLower(k) == keyLower && v != "" && !seen[v] {
					seen[v] = true
					vals = append(vals, v)
				}
			}
		}
	}
	return vals
}

// StreamIndices returns unique stream indices for streams matching the given codec type.
// codecType should be "video", "audio", "subtitle", "data", or "attachment".
// If codecType is empty, all stream indices are returned.
func StreamIndices(streams []StreamInfo, codecType string) []string {
	seen := make(map[int]bool)
	var vals []string
	for _, s := range streams {
		if codecType != "" && s.CodecType != codecType {
			continue
		}
		if !seen[s.Index] {
			seen[s.Index] = true
			vals = append(vals, strconv.Itoa(s.Index))
		}
	}
	return vals
}

// StreamIDs returns unique stream IDs for use with the # and i: stream specifier forms.
// Each ID is formatted as both hex (0xNNNN) and decimal, since ffmpeg accepts either.
// Only streams with a non-empty ID are included.
func StreamIDs(streams []StreamInfo) []string {
	seen := make(map[string]bool)
	var vals []string
	for _, s := range streams {
		if s.ID == "" || seen[s.ID] {
			continue
		}
		seen[s.ID] = true
		id := s.ID
		vals = append(vals, id)
		if idHex := formatIDAsHex(id); idHex != "" {
			vals = append(vals, idHex)
		}
	}
	return vals
}

// formatIDAsHex converts a decimal stream ID string to hex format (e.g. "512" -> "0x200").
// Returns empty string if the ID is not a valid decimal integer or is already hex.
func formatIDAsHex(id string) string {
	if strings.HasPrefix(id, "0x") || strings.HasPrefix(id, "0X") {
		return ""
	}
	n, err := strconv.Atoi(id)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("0x%X", n)
}

// ActiveDispositions returns disposition names that are set (non-zero) in any stream.
func ActiveDispositions(streams []StreamInfo) []string {
	seen := make(map[string]bool)
	var vals []string
	for _, s := range streams {
		for name, val := range s.Disposition {
			if val != 0 && !seen[name] {
				seen[name] = true
				vals = append(vals, name)
			}
		}
	}
	return vals
}

// isLocalFile returns true if the URL looks like a local file path.
func isLocalFile(url string) bool {
	if url == "" {
		return false
	}
	if url == "-" || url == "/dev/stdin" {
		return false
	}
	for _, prefix := range []string{"http://", "https://", "ftp://", "rtmp://", "rtmps://", "rtp://", "udp://", "tcp://", "sctp://", "pipe:", "concat:", "lavfi:", "fd://"} {
		if len(url) >= len(prefix) && url[:len(prefix)] == prefix {
			return false
		}
	}
	return true
}
