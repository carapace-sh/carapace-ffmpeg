package completer

import (
	"os"
	"path/filepath"
	"testing"

	ffmpeg "github.com/carapace-sh/carapace-ffmpeg/pkg/actions/tools/ffmpeg"
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

	indices := probe.StreamIndices(streams, "audio")
	if len(indices) != 2 || indices[0] != "1" || indices[1] != "2" {
		t.Errorf("StreamIndices for audio = %v, want [1 2]", indices)
	}

	indices = probe.StreamIndices(streams, "video")
	if len(indices) != 1 || indices[0] != "0" {
		t.Errorf("StreamIndices for video = %v, want [0]", indices)
	}
}

// testdataPath resolves a path to a file in the project's testdata/ directory.
// Skips the test if the file doesn't exist (run `go generate ./testdata/` to create it).
func testdataPath(t *testing.T, filename string) string {
	t.Helper()
	// Try relative to the test binary location first (project root).
	candidates := []string{
		filepath.Join("testdata", filename),
		filepath.Join("..", "..", "testdata", filename),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			abs, _ := filepath.Abs(p)
			return abs
		}
	}
	t.Skipf("skipping: testdata/%s not found (run `go generate ./testdata/`)", filename)
	return ""
}

func TestProbeAllWithMultistream(t *testing.T) {
	path := testdataPath(t, "multistream.mkv")
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
	path := testdataPath(t, "multistream.mkv")
	ctx := &argstream.CompletionContext{
		InputURLs: []string{path, path},
	}
	streams := ProbeAll(ctx)
	if len(streams) < 3 {
		t.Fatalf("expected at least 3 streams, got %d", len(streams))
	}
	if len(streams) > 3 {
		t.Errorf("expected at most 3 streams (deduplicated), got %d", len(streams))
	}
}

func TestProbeAllWithSubtitles(t *testing.T) {
	path := testdataPath(t, "subtitles.mkv")
	ctx := &argstream.CompletionContext{
		InputURLs: []string{path},
	}
	streams := ProbeAll(ctx)
	if len(streams) < 3 {
		t.Fatalf("expected at least 3 streams, got %d", len(streams))
	}
	found := false
	for _, s := range streams {
		if s.CodecType == "subtitle" {
			found = true
			if s.Tags["language"] != "eng" {
				t.Errorf("subtitle language = %q, want 'eng'", s.Tags["language"])
			}
		}
	}
	if !found {
		t.Error("expected subtitle stream in subtitles.mkv")
	}
}

func TestProbeAllWithAudioOnly(t *testing.T) {
	path := testdataPath(t, "audio_only.wav")
	ctx := &argstream.CompletionContext{
		InputURLs: []string{path},
	}
	streams := ProbeAll(ctx)
	if len(streams) < 1 {
		t.Fatalf("expected at least 1 stream, got %d", len(streams))
	}
	if streams[0].CodecType != "audio" {
		t.Errorf("stream 0 CodecType = %q, want 'audio'", streams[0].CodecType)
	}
}

func TestProbeAllWithAttachment(t *testing.T) {
	path := testdataPath(t, "attachment.mkv")
	ctx := &argstream.CompletionContext{
		InputURLs: []string{path},
	}
	streams := ProbeAll(ctx)
	found := false
	for _, s := range streams {
		if s.CodecType == "attachment" {
			found = true
		}
	}
	if !found {
		t.Error("expected attachment stream in attachment.mkv")
	}
}

func TestProbeAllWith5Point1(t *testing.T) {
	path := testdataPath(t, "surround.mkv")
	ctx := &argstream.CompletionContext{
		InputURLs: []string{path},
	}
	streams := ProbeAll(ctx)
	// Just verify the audio stream exists with codec_type audio.
	found := false
	for _, s := range streams {
		if s.CodecType == "audio" {
			found = true
		}
	}
	if !found {
		t.Error("expected audio stream in surround.mkv")
	}
}

func TestActionStreamIndexWithRealStreams(t *testing.T) {
	path := testdataPath(t, "multistream.mkv")
	streams, _ := probe.Probe(path)
	if len(streams) < 3 {
		t.Skipf("skipping: not enough streams (%d)", len(streams))
	}

	specCtx := &streamspec.CompletionContext{
		CurrentKind: streamspec.KindStreamType,
	}

	action := actionStreamIndex(specCtx, "v:", streams, "")
	_ = action

	action = actionStreamIndex(specCtx, "a:", streams, "")
	_ = action
}

func TestActionMetadataValueWithRealStreams(t *testing.T) {
	path := testdataPath(t, "multistream.mkv")
	streams, _ := probe.Probe(path)
	if len(streams) < 3 {
		t.Skipf("skipping: not enough streams (%d)", len(streams))
	}

	specCtx := &streamspec.CompletionContext{
		MetadataKey: "language",
	}

	action := actionMetadataValue(specCtx, streams, "")
	_ = action
}

func TestFilterOptsFromContext(t *testing.T) {
	tests := []struct {
		name    string
		ctx     *argstream.CompletionContext
		want    ffmpeg.FilterOpts
	}{
		{
			"nil option returns default",
			&argstream.CompletionContext{},
			ffmpeg.FilterOpts{Audio: true, Video: true},
		},
		{
			"no specifier no implicit spec returns default",
			&argstream.CompletionContext{
				CurrentOption: &argstream.OptionContext{Name: "filter_complex", CanonicalName: "filter_complex"},
			},
			ffmpeg.FilterOpts{Audio: true, Video: true},
		},
		{
			"stream specifier a",
			&argstream.CompletionContext{
				CurrentOption: &argstream.OptionContext{Name: "filter", CanonicalName: "filter", StreamSpecifier: "a", AcceptsSpec: true},
			},
			ffmpeg.FilterOpts{Audio: true, Video: false},
		},
		{
			"stream specifier v",
			&argstream.CompletionContext{
				CurrentOption: &argstream.OptionContext{Name: "filter", CanonicalName: "filter", StreamSpecifier: "v", AcceptsSpec: true},
			},
			ffmpeg.FilterOpts{Audio: false, Video: true},
		},
		{
			"stream specifier V",
			&argstream.CompletionContext{
				CurrentOption: &argstream.OptionContext{Name: "filter", CanonicalName: "filter", StreamSpecifier: "V", AcceptsSpec: true},
			},
			ffmpeg.FilterOpts{Audio: false, Video: true},
		},
		{
			"stream specifier s",
			&argstream.CompletionContext{
				CurrentOption: &argstream.OptionContext{Name: "filter", CanonicalName: "filter", StreamSpecifier: "s", AcceptsSpec: true},
			},
			ffmpeg.FilterOpts{Audio: false, Video: false},
		},
		{
			"implicit spec vf",
			&argstream.CompletionContext{
				CurrentOption: &argstream.OptionContext{Name: "vf", CanonicalName: "filter"},
			},
			ffmpeg.FilterOpts{Audio: false, Video: true},
		},
		{
			"implicit spec af",
			&argstream.CompletionContext{
				CurrentOption: &argstream.OptionContext{Name: "af", CanonicalName: "filter"},
			},
			ffmpeg.FilterOpts{Audio: true, Video: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterOptsFromContext(tt.ctx)
			if got != tt.want {
				t.Errorf("FilterOptsFromContext() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
