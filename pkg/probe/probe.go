package probe

import (
	"encoding/json"
	"os/exec"
	"strconv"
)

// StreamInfo holds metadata for a single stream probed from a media file.
type StreamInfo struct {
	Index       int               `json:"index"`
	CodecName   string            `json:"codec_name"`
	CodecType   string            `json:"codec_type"`
	SampleFmt   string            `json:"sample_fmt,omitempty"`
	PixFmt      string            `json:"pix_fmt,omitempty"`
	Disposition map[string]int    `json:"disposition"`
	Tags        map[string]string `json:"tags,omitempty"`
}

type ffprobeOutput struct {
	Streams []StreamInfo `json:"streams"`
}

// Probe runs ffprobe on a local file and returns stream metadata.
// Returns nil and no error if ffprobe is unavailable or the file cannot be probed.
func Probe(inputURL string) ([]StreamInfo, error) {
	if !isLocalFile(inputURL) {
		return nil, nil
	}

	cmd := exec.Command("ffprobe",
		"-hide_banner",
		"-show_streams",
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
	return result.Streams, nil
}

// MetadataValues extracts unique values for a metadata key from the given streams.
func MetadataValues(streams []StreamInfo, key string) []string {
	seen := make(map[string]bool)
	var vals []string
	for _, s := range streams {
		if s.Tags != nil {
			if v, ok := s.Tags[key]; ok && v != "" && !seen[v] {
				seen[v] = true
				vals = append(vals, v)
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
