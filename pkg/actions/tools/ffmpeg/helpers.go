package ffmpeg

import (
	"bytes"
	"strings"
)

func splitLines(output []byte) []string {
	return strings.Split(string(bytes.TrimSpace(output)), "\n")
}

func splitFields(line string) []string {
	return strings.Fields(line)
}

// extractCodecName parses a codec line like " D.VSI.S..... libx264   Libx264" and returns "libx264".
func extractCodecName(line string) string {
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return ""
	}
	// First field is the flags (D.VSI etc.), second is the name
	if len(fields[0]) < 2 {
		return ""
	}
	return fields[1]
}

// extractFormatName parses a format line like " D  mp4   MP4" and returns "mp4".
func extractFormatName(line string) string {
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return ""
	}
	return fields[1]
}

// extractFilterName parses a filter line like " .. scale             V->V       Scale" and returns "scale".
func extractFilterName(line string) string {
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return ""
	}
	// First field: flags (2 chars + dots), second: name
	return fields[1]
}
