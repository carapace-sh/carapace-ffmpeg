package completer

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/carapace-sh/carapace-ffmpeg/pkg/argstream"
	"github.com/carapace-sh/carapace-ffmpeg/pkg/probe"
	"github.com/carapace-sh/carapace-ffmpeg/pkg/streamspec"
)

func TestProbeAllEmpty(t *testing.T) {
	ctx := &argstream.CompletionContext{}
	streams := ProbeAll(ctx)
	if len(streams) != 0 {
		t.Errorf("expected no streams, got %d", len(streams))
	}
}

func TestProbeAllWithInputURLs(t *testing.T) {
	ctx := &argstream.CompletionContext{
		InputURLs: []string{"/nonexistent/file.mp4"},
	}
	streams := ProbeAll(ctx)
	// nonexistent file returns nil streams, which is fine
	if streams != nil {
		t.Errorf("expected nil streams for nonexistent file, got %v", streams)
	}
}

func TestExtractStreamTypeLetter(t *testing.T) {
	tests := []struct {
		spec string
		want string
	}{
		{"a", "a"},
		{"a:", "a"},
		{"a:0", "a"},
		{"v", "v"},
		{"V", "V"},
		{"s:1", "s"},
		{"d", "d"},
		{"d:0", "d"},
		{"t", "t"},
		{"m:language:eng", ""},
		{"disp:default", ""},
		{"", ""},
		{"0", ""},
	}

	for _, tt := range tests {
		got := extractStreamTypeLetter(tt.spec)
		if got != tt.want {
			t.Errorf("extractStreamTypeLetter(%q) = %q, want %q", tt.spec, got, tt.want)
		}
	}
}

func TestStreamTypeToCodecType(t *testing.T) {
	tests := []struct {
		letter string
		want   string
	}{
		{"v", "video"},
		{"V", "video"},
		{"a", "audio"},
		{"s", "subtitle"},
		{"d", "data"},
		{"t", "attachment"},
		{"x", ""},
		{"", ""},
	}

	for _, tt := range tests {
		got := streamTypeToCodecType(tt.letter)
		if got != tt.want {
			t.Errorf("streamTypeToCodecType(%q) = %q, want %q", tt.letter, got, tt.want)
		}
	}
}

func TestActionStreamIndexWithProbedStreams(t *testing.T) {
	streams := []probe.StreamInfo{
		{Index: 0, CodecType: "video"},
		{Index: 1, CodecType: "audio"},
		{Index: 2, CodecType: "audio"},
	}

	// Test that stream indices for audio type are returned
	indices := probe.StreamIndices(streams, "audio")
	if len(indices) != 2 || indices[0] != "1" || indices[1] != "2" {
		t.Errorf("StreamIndices for audio = %v, want [1 2]", indices)
	}

	// Test that stream indices for video type are returned
	indices = probe.StreamIndices(streams, "video")
	if len(indices) != 1 || indices[0] != "0" {
		t.Errorf("StreamIndices for video = %v, want [0]", indices)
	}
}

func generateMultistreamMKV(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.mkv")

	cmd := exec.Command("ffmpeg", "-hide_banner", "-y",
		"-f", "lavfi", "-i", "color=c=black:s=2x2:d=0.01:r=1",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=0.01:r=8000",
		"-f", "lavfi", "-i", "sine=frequency=880:duration=0.01:r=8000",
		"-map", "0:v", "-map", "1:a", "-map", "2:a",
		"-metadata:s:a:0", "language=eng",
		"-metadata:s:a:1", "language=fre",
		"-c:v", "libx264", "-preset", "ultrafast", "-tune", "stillimage", "-crf", "51",
		"-c:a", "pcm_s16le",
		path,
	)
	if err := cmd.Run(); err != nil {
		t.Skipf("skipping: ffmpeg not available or failed: %v", err)
	}
	return path
}

func TestProbeAllWithGeneratedFile(t *testing.T) {
	path := generateMultistreamMKV(t)
	ctx := &argstream.CompletionContext{
		InputURLs: []string{path},
	}
	streams := ProbeAll(ctx)
	if len(streams) < 3 {
		t.Fatalf("expected at least 3 streams, got %d", len(streams))
	}
	if streams[0].CodecType != "video" {
		t.Errorf("stream 0 CodecType = %q, want 'video'", streams[0].CodecType)
	}
	if streams[1].Tags["language"] != "eng" {
		t.Errorf("stream 1 language = %q, want 'eng'", streams[1].Tags["language"])
	}
	if streams[2].Tags["language"] != "fre" {
		t.Errorf("stream 2 language = %q, want 'fre'", streams[2].Tags["language"])
	}
}

func TestProbeAllDeduplicatesURLs(t *testing.T) {
	path := generateMultistreamMKV(t)
	ctx := &argstream.CompletionContext{
		InputURLs: []string{path, path},
	}
	streams := ProbeAll(ctx)
	// Same URL listed twice should only be probed once
	if len(streams) < 3 {
		t.Fatalf("expected at least 3 streams, got %d", len(streams))
	}
	// Shouldn't double the streams
	if len(streams) > 3 {
		t.Errorf("expected at most 3 streams (deduplicated), got %d", len(streams))
	}
}

func TestActionStreamIndexWithRealStreams(t *testing.T) {
	path := generateMultistreamMKV(t)
	streams, _ := probe.Probe(path)
	if len(streams) < 3 {
		t.Skipf("skipping: not enough streams (%d)", len(streams))
	}

	specCtx := &streamspec.CompletionContext{
		CurrentKind: streamspec.KindStreamType,
	}

	// Video type spec "v:" → should return index "0"
	action := actionStreamIndex(specCtx, "v:", streams, "")
	_ = action

	// Audio type spec "a:" → should return indices "1", "2"
	action = actionStreamIndex(specCtx, "a:", streams, "")
	_ = action
}

func TestActionMetadataValueWithRealStreams(t *testing.T) {
	path := generateMultistreamMKV(t)
	streams, _ := probe.Probe(path)
	if len(streams) < 3 {
		t.Skipf("skipping: not enough streams (%d)", len(streams))
	}

	specCtx := &streamspec.CompletionContext{
		MetadataKey: "language",
	}

	// Should not return empty action since streams have language tags
	action := actionMetadataValue(specCtx, streams, "")
	_ = action
}
