package probe

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// generateMultistreamMKV creates a tiny MKV with 1 video + 2 audio streams
// (eng, fre) using ffmpeg from lavfi sources. Returns the file path.
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

// generateMonoWAV creates a minimal WAV file with a single audio stream
// using ffmpeg from a lavfi source. Returns the file path.
func generateMonoWAV(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.wav")

	cmd := exec.Command("ffmpeg", "-hide_banner", "-y",
		"-f", "lavfi", "-i", "sine=frequency=1000:duration=0.01:r=8000",
		"-c:a", "pcm_s16le",
		path,
	)
	if err := cmd.Run(); err != nil {
		t.Skipf("skipping: ffmpeg not available or failed: %v", err)
	}
	return path
}

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

func TestMetadataValuesCaseInsensitive(t *testing.T) {
	streams := []StreamInfo{
		{Index: 0, Tags: map[string]string{"ARTIST": "Bach", "Title": "Fugue"}},
	}

	tests := []struct {
		key  string
		want []string
	}{
		{"artist", []string{"Bach"}},
		{"ARTIST", []string{"Bach"}},
		{"Artist", []string{"Bach"}},
		{"title", []string{"Fugue"}},
		{"TITLE", []string{"Fugue"}},
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

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("skipping: cannot determine home directory: %v", err)
	}

	tests := []struct {
		input string
		want  string
	}{
		{"~/slop/example.mkv", filepath.Join(home, "slop/example.mkv")},
		{"~/", home + "/"},
		{"~", "~"},
		{"~user/foo", "~user/foo"},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		{"", ""},
	}

	for _, tt := range tests {
		got := expandPath(tt.input)
		if got != tt.want {
			t.Errorf("expandPath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestProbeWithTildePath(t *testing.T) {
	path := generateMonoWAV(t)

	// Construct a ~/... path by replacing the home dir prefix
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("skipping: cannot determine home directory: %v", err)
	}

	relPath := path
	if strings.HasPrefix(path, home+"/") {
		relPath = "~" + path[len(home):]
	} else {
		// If temp dir isn't under home, symlink it there
		symlinkDir := filepath.Join(home, ".carapace-ffmpeg-test-"+t.Name())
		if err := os.Symlink(filepath.Dir(path), symlinkDir); err != nil {
			t.Skipf("skipping: cannot create symlink in home dir: %v", err)
		}
		defer os.Remove(symlinkDir)
		relPath = filepath.Join("~", ".carapace-ffmpeg-test-"+t.Name(), filepath.Base(path))
	}

	streams, err := Probe(relPath)
	if err != nil {
		t.Fatalf("Probe() error = %v", err)
	}
	if len(streams) == 0 {
		t.Fatal("Probe() returned no streams for tilde-expanded path")
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

// generateTaggedMKV creates an MKV with format-level metadata (title, artist, etc.)
// and stream-level metadata (language) to test tag merging. Returns the file path.
func generateTaggedMKV(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "tagged.mkv")

	cmd := exec.Command("ffmpeg", "-hide_banner", "-y",
		"-f", "lavfi", "-i", "sine=frequency=1000:duration=0.01:r=8000",
		"-metadata", "title=Test Song",
		"-metadata", "artist=Test Artist",
		"-metadata", "album=Test Album",
		"-metadata", "genre=Synthwave",
		"-metadata", "track=1",
		"-metadata", "date=2024",
		"-metadata:s:a:0", "language=eng",
		"-c:a", "pcm_s16le",
		path,
	)
	if err := cmd.Run(); err != nil {
		t.Skipf("skipping: ffmpeg not available or failed: %v", err)
	}
	return path
}

func TestProbeMonoWAV(t *testing.T) {
	path := generateMonoWAV(t)

	streams, err := Probe(path)
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

func TestProbeMultistreamMKV(t *testing.T) {
	path := generateMultistreamMKV(t)

	streams, err := Probe(path)
	if err != nil {
		t.Fatalf("Probe() error = %v", err)
	}
	if len(streams) < 3 {
		t.Fatalf("expected at least 3 streams, got %d", len(streams))
	}

	// Verify video stream
	if streams[0].CodecType != "video" {
		t.Errorf("stream 0 CodecType = %q, want 'video'", streams[0].CodecType)
	}

	// Verify audio streams with language tags
	for i := 1; i <= 2; i++ {
		if streams[i].CodecType != "audio" {
			t.Errorf("stream %d CodecType = %q, want 'audio'", i, streams[i].CodecType)
		}
	}
	if streams[1].Tags["language"] != "eng" {
		t.Errorf("stream 1 language = %q, want 'eng'", streams[1].Tags["language"])
	}
	if streams[2].Tags["language"] != "fre" {
		t.Errorf("stream 2 language = %q, want 'fre'", streams[2].Tags["language"])
	}

	// Verify stream indices
	audioIndices := StreamIndices(streams, "audio")
	if len(audioIndices) != 2 {
		t.Errorf("StreamIndices(audio) = %v, want 2 entries", audioIndices)
	}

	// Verify metadata values
	langs := MetadataValues(streams, "language")
	if len(langs) != 2 {
		t.Errorf("MetadataValues(language) = %v, want 2 entries", langs)
	}

	// Verify dispositions
	disp := ActiveDispositions(streams)
	if len(disp) == 0 {
		t.Error("ActiveDispositions returned empty, expected at least 'default'")
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

func TestProbeTaggedMKV(t *testing.T) {
	path := generateTaggedMKV(t)

	streams, err := Probe(path)
	if err != nil {
		t.Fatalf("Probe() error = %v", err)
	}
	if len(streams) == 0 {
		t.Fatal("Probe() returned no streams")
	}

	s := streams[0]

	// Format-level tags should be merged into stream tags
	// Note: MKV muxer writes some keys in uppercase (e.g. ARTIST)
	if s.Tags["title"] != "Test Song" {
		t.Errorf("title = %q, want 'Test Song'", s.Tags["title"])
	}
	// Case-insensitive matching: ARTIST in format, "artist" in lookup
	artist := MetadataValues(streams, "artist")
	if len(artist) != 1 || artist[0] != "Test Artist" {
		t.Errorf("MetadataValues(artist) = %v, want [Test Artist]", artist)
	}
	artistUpper := MetadataValues(streams, "ARTIST")
	if len(artistUpper) != 1 || artistUpper[0] != "Test Artist" {
		t.Errorf("MetadataValues(ARTIST) = %v, want [Test Artist]", artistUpper)
	}

	// Stream-level tag should be present (language was set on stream)
	if s.Tags["language"] != "eng" {
		t.Errorf("language = %q, want 'eng'", s.Tags["language"])
	}

	// Stream-level tag should take precedence over format-level
	// (if both had the same key, stream wins)
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

func TestStreamIDs(t *testing.T) {
	streams := []StreamInfo{
		{Index: 0, ID: "512", CodecType: "video"},
		{Index: 1, ID: "513", CodecType: "audio"},
		{Index: 2, ID: "", CodecType: "audio"},
	}

	ids := StreamIDs(streams)
	seen := map[string]bool{}
	for _, id := range ids {
		seen[id] = true
	}
	if !seen["512"] {
		t.Errorf("StreamIDs missing decimal '512' in %v", ids)
	}
	if !seen["0x200"] {
		t.Errorf("StreamIDs missing hex '0x200' for decimal 512 in %v", ids)
	}
	if !seen["513"] {
		t.Errorf("StreamIDs missing decimal '513' in %v", ids)
	}
	if !seen["0x201"] {
		t.Errorf("StreamIDs missing hex '0x201' for decimal 513 in %v", ids)
	}
	if seen[""] {
		t.Error("StreamIDs should not include empty ID")
	}
}

func TestStreamIDsEmpty(t *testing.T) {
	if got := StreamIDs(nil); len(got) != 0 {
		t.Errorf("StreamIDs(nil) = %v, want empty", got)
	}
	if got := StreamIDs([]StreamInfo{{Index: 0, ID: ""}}); len(got) != 0 {
		t.Errorf("StreamIDs with empty ID = %v, want empty", got)
	}
}

func TestStreamIDsDedup(t *testing.T) {
	streams := []StreamInfo{
		{Index: 0, ID: "100"},
		{Index: 1, ID: "100"},
	}
	ids := StreamIDs(streams)
	count := 0
	for _, id := range ids {
		if id == "100" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("StreamIDs has %d occurrences of '100', want 1: %v", count, ids)
	}
}

func TestStreamIDsHexInput(t *testing.T) {
	streams := []StreamInfo{
		{Index: 0, ID: "0x1F3"},
	}
	ids := StreamIDs(streams)
	if len(ids) != 1 || ids[0] != "0x1F3" {
		t.Errorf("StreamIDs with hex ID = %v, want [0x1F3]", ids)
	}
}

func TestFormatIDAsHex(t *testing.T) {
	tests := []struct {
		id   string
		want string
	}{
		{"512", "0x200"},
		{"0", "0x0"},
		{"1", "0x1"},
		{"0x200", ""},
		{"0X1F3", ""},
		{"notanumber", ""},
		{"", ""},
	}
	for _, tt := range tests {
		got := formatIDAsHex(tt.id)
		if got != tt.want {
			t.Errorf("formatIDAsHex(%q) = %q, want %q", tt.id, got, tt.want)
		}
	}
}
func generateMpegTS(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.ts")

	cmd := exec.Command("ffmpeg", "-hide_banner", "-y",
		"-f", "lavfi", "-i", "color=c=black:s=2x2:d=0.01:r=1",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=0.01:r=8000",
		"-map", "0:v", "-map", "1:a",
		"-c:v", "libx264", "-preset", "ultrafast", "-crf", "51",
		"-c:a", "mp2",
		path,
	)
	if err := cmd.Run(); err != nil {
		t.Skipf("skipping: ffmpeg not available or failed: %v", err)
	}
	return path
}

func TestProbeMPEGTS(t *testing.T) {
	path := generateMpegTS(t)

	streams, err := Probe(path)
	if err != nil {
		t.Fatalf("Probe() error = %v", err)
	}
	if len(streams) < 2 {
		t.Fatalf("expected at least 2 streams, got %d", len(streams))
	}

	// MPEG-TS should have stream IDs (PIDs)
	ids := StreamIDs(streams)
	if len(ids) == 0 {
		t.Error("expected stream IDs in MPEG-TS, got none")
	}

	// Verify that hex-format IDs are present (0x100, 0x101, etc.)
	foundHex := false
	for _, id := range ids {
		if id == "0x100" || id == "0x101" {
			foundHex = true
		}
	}
	if !foundHex {
		t.Errorf("expected hex stream IDs 0x100 or 0x101 in MPEG-TS, got %v", ids)
	}
}
