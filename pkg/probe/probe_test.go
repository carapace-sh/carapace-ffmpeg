package probe

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMetadataValues(t *testing.T) {
	streams := []StreamInfo{
		{Index: 0, CodecType: "video", Tags: map[string]string{"language": "eng", "title": "Main"}},
		{Index: 1, CodecType: "audio", Tags: map[string]string{"language": "eng"}},
		{Index: 2, CodecType: "audio", Tags: map[string]string{"language": "fre"}},
		{Index: 3, CodecType: "subtitle", Tags: map[string]string{"language": "fre"}},
	}

	tests := []struct {
		key  string
		want []string
	}{
		{"language", []string{"eng", "fre"}},
		{"title", []string{"Main"}},
		{"missing", nil},
	}

	for _, tt := range tests {
		got := MetadataValues(streams, tt.key)
		if len(got) != len(tt.want) {
			t.Errorf("MetadataValues(%q) = %v, want %v", tt.key, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("MetadataValues(%q)[%d] = %q, want %q", tt.key, i, got[i], tt.want[i])
			}
		}
	}
}

func TestMetadataValuesEmpty(t *testing.T) {
	if got := MetadataValues(nil, "language"); len(got) != 0 {
		t.Errorf("MetadataValues(nil) = %v, want empty", got)
	}
	if got := MetadataValues([]StreamInfo{{}}, "language"); len(got) != 0 {
		t.Errorf("MetadataValues with no tags = %v, want empty", got)
	}
}

func TestStreamIndices(t *testing.T) {
	streams := []StreamInfo{
		{Index: 0, CodecType: "video"},
		{Index: 1, CodecType: "audio"},
		{Index: 2, CodecType: "audio"},
		{Index: 3, CodecType: "subtitle"},
	}

	tests := []struct {
		codecType string
		want      []string
	}{
		{"video", []string{"0"}},
		{"audio", []string{"1", "2"}},
		{"subtitle", []string{"3"}},
		{"", []string{"0", "1", "2", "3"}},
		{"data", nil},
	}

	for _, tt := range tests {
		got := StreamIndices(streams, tt.codecType)
		if len(got) != len(tt.want) {
			t.Errorf("StreamIndices(%q) = %v, want %v", tt.codecType, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("StreamIndices(%q)[%d] = %q, want %q", tt.codecType, i, got[i], tt.want[i])
			}
		}
	}
}

func TestStreamIndicesEmpty(t *testing.T) {
	if got := StreamIndices(nil, "audio"); len(got) != 0 {
		t.Errorf("StreamIndices(nil) = %v, want empty", got)
	}
}

func TestActiveDispositions(t *testing.T) {
	streams := []StreamInfo{
		{Disposition: map[string]int{"default": 1, "forced": 0, "dub": 0}},
		{Disposition: map[string]int{"default": 0, "forced": 1, "comment": 1}},
	}
	got := ActiveDispositions(streams)

	seen := map[string]bool{}
	for _, d := range got {
		seen[d] = true
	}
	for _, want := range []string{"default", "forced", "comment"} {
		if !seen[want] {
			t.Errorf("ActiveDispositions missing %q in %v", want, got)
		}
	}
	if seen["dub"] {
		t.Errorf("ActiveDispositions should not include 'dub'")
	}
}

func TestActiveDispositionsEmpty(t *testing.T) {
	if got := ActiveDispositions(nil); len(got) != 0 {
		t.Errorf("ActiveDispositions(nil) = %v, want empty", got)
	}
	if got := ActiveDispositions([]StreamInfo{{Disposition: map[string]int{"default": 0}}}); len(got) != 0 {
		t.Errorf("ActiveDispositions with all-zero = %v, want empty", got)
	}
}

func TestIsLocalFile(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"input.mp4", true},
		{"/home/user/video.mkv", true},
		{"./relative.mp4", true},
		{"http://example.com/video.mp4", false},
		{"https://example.com/video.mp4", false},
		{"rtmp://server/live", false},
		{"ftp://server/video.mp4", false},
		{"pipe:", false},
		{"lavfi:color=red", false},
		{"concat:1.txt", false},
		{"-", false},
		{"/dev/stdin", false},
		{"", false},
		{"udp://239.255.0.1:1234", false},
		{"rtp://239.255.0.1:1234", false},
	}

	for _, tt := range tests {
		got := isLocalFile(tt.url)
		if got != tt.want {
			t.Errorf("isLocalFile(%q) = %v, want %v", tt.url, got, tt.want)
		}
	}
}

func TestProbeWithRealFile(t *testing.T) {
	wavPath := filepath.Join("/", "usr", "share", "filezilla", "resources", "finished.wav")
	if _, err := os.Stat(wavPath); err != nil {
		t.Skip("skipping: test wav file not found")
	}

	streams, err := Probe(wavPath)
	if err != nil {
		t.Fatalf("Probe() error = %v", err)
	}
	if len(streams) == 0 {
		t.Fatal("Probe() returned no streams")
	}
	if streams[0].CodecType != "audio" {
		t.Errorf("first stream CodecType = %q, want 'audio'", streams[0].CodecType)
	}
	if streams[0].CodecName == "" {
		t.Error("first stream CodecName is empty")
	}
}

func TestProbeNonLocalFile(t *testing.T) {
	streams, err := Probe("http://example.com/video.mp4")
	if err != nil {
		t.Errorf("Probe() unexpected error = %v", err)
	}
	if streams != nil {
		t.Errorf("Probe() for remote URL = %v, want nil", streams)
	}
}

func TestProbeNonexistentFile(t *testing.T) {
	streams, err := Probe("/nonexistent/path/video.mp4")
	if err != nil {
		t.Errorf("Probe() unexpected error = %v", err)
	}
	if streams != nil {
		t.Errorf("Probe() for nonexistent file = %v, want nil", streams)
	}
}
