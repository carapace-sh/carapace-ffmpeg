package ffmpeg

import (
	"testing"

	"github.com/carapace-sh/carapace"
	"github.com/carapace-sh/carapace/pkg/sandbox"
)

func TestBsfStreamType(t *testing.T) {
	tests := []struct {
		name     string
		audio    bool
		video    bool
		subtitle bool
	}{
		{"aac_adtstoasc", true, false, false},
		{"eac3_core", true, false, false},
		{"opus_metadata", true, false, false},
		{"truehd_core", true, false, false},
		{"ahx_to_mp2", true, false, false},
		{"pcm_rechunk", true, false, false},
		{"dca_core", true, false, false},
		{"h264_metadata", false, true, false},
		{"h264_mp4toannexb", false, true, false},
		{"h264_redundant_pps", false, true, false},
		{"hevc_metadata", false, true, false},
		{"hevc_mp4toannexb", false, true, false},
		{"vvc_metadata", false, true, false},
		{"vvc_mp4toannexb", false, true, false},
		{"av1_frame_merge", false, true, false},
		{"av1_frame_split", false, true, false},
		{"av1_metadata", false, true, false},
		{"vp9_metadata", false, true, false},
		{"vp9_superframe", false, true, false},
		{"vp9_superframe_split", false, true, false},
		{"mpeg2_metadata", false, true, false},
		{"mpeg4_unpack_bframes", false, true, false},
		{"prores_metadata", false, true, false},
		{"mjpeg2jpeg", false, true, false},
		{"evc_frame_merge", false, true, false},
		{"apv_metadata", false, true, false},
		{"dovi_rpu", false, true, false},
		{"lcevc_metadata", false, true, false},
		{"dv_error_marker", false, true, false},
		{"hapqa_extract", false, true, false},
		{"imxdump", false, true, false},
		{"pgs_frame_merge", false, false, true},
		{"mov2textsub", false, false, true},
		{"text2movsub", false, false, true},
		{"null", true, true, true},
		{"noise", true, true, true},
		{"chomp", true, true, true},
		{"dump_extra", true, true, true},
		{"extract_extradata", true, true, true},
		{"filter_units", true, true, true},
		{"remove_extra", true, true, true},
		{"setts", true, true, true},
		{"showinfo", true, true, true},
		{"trace_headers", true, true, true},
	}

	for _, tt := range tests {
		audio, video, subtitle := bsfStreamType(tt.name)
		if audio != tt.audio || video != tt.video || subtitle != tt.subtitle {
			t.Errorf("bsfStreamType(%q) = (%v, %v, %v), want (%v, %v, %v)",
				tt.name, audio, video, subtitle, tt.audio, tt.video, tt.subtitle)
		}
	}
}

func TestActionDispositionsDefault(t *testing.T) {
	sandbox.Action(t, func() carapace.Action {
		return ActionDispositions(DispositionOpts{}.Default())
	})(func(s *sandbox.Sandbox) {
		s.Run("").Expect(carapace.ActionValues(
			"default", "dub", "original", "comment", "lyrics", "karaoke",
			"forced", "hearing_impaired", "visual_impaired", "clean_effects",
			"attached_pic", "timed_thumbnails", "non_diegetic", "captions",
			"descriptions", "metadata", "dependent", "still_image", "multilayer",
		).Tag("dispositions"))
	})
}

func TestActionDispositionsAudioOnly(t *testing.T) {
	sandbox.Action(t, func() carapace.Action {
		return ActionDispositions(DispositionOpts{Audio: true})
	})(func(s *sandbox.Sandbox) {
		s.Run("").Expect(carapace.ActionValues(
			"default", "dub", "original", "comment", "lyrics", "karaoke",
			"forced", "hearing_impaired", "visual_impaired", "clean_effects",
			"non_diegetic", "captions", "descriptions", "metadata", "dependent",
		).Tag("dispositions"))
	})
}

func TestActionDispositionsVideoOnly(t *testing.T) {
	sandbox.Action(t, func() carapace.Action {
		return ActionDispositions(DispositionOpts{Video: true})
	})(func(s *sandbox.Sandbox) {
		s.Run("").Expect(carapace.ActionValues(
			"default", "dub", "original", "comment", "lyrics",
			"forced", "attached_pic", "timed_thumbnails", "non_diegetic",
			"captions", "descriptions", "metadata", "dependent",
			"still_image", "multilayer",
		).Tag("dispositions"))
	})
}

func TestActionDispositionsSubtitleOnly(t *testing.T) {
	sandbox.Action(t, func() carapace.Action {
		return ActionDispositions(DispositionOpts{Subtitle: true})
	})(func(s *sandbox.Sandbox) {
		s.Run("").Expect(carapace.ActionValues(
			"default", "dub", "original", "comment", "lyrics",
			"forced", "non_diegetic", "captions", "descriptions",
			"metadata", "dependent",
		).Tag("dispositions"))
	})
}

// Filter tests use ExpectNot since filters produce large dynamic result sets
// and Expect requires exact match.

func TestActionFiltersAudioOnlyExcludesVideoFilters(t *testing.T) {
	sandbox.Action(t, func() carapace.Action {
		return ActionFilters(FilterOpts{Audio: true, Video: false})
	})(func(s *sandbox.Sandbox) {
		s.Run("").ExpectNot(carapace.ActionValues("scale"))
		s.Run("").ExpectNot(carapace.ActionValues("color"))
	})
}

func TestActionFiltersAudioOnlyIncludesAudioFilters(t *testing.T) {
	sandbox.Action(t, func() carapace.Action {
		return ActionFilters(FilterOpts{Audio: true, Video: false})
	})(func(s *sandbox.Sandbox) {
		s.Run("").ExpectNot(carapace.ActionValues())
	})
}

func TestActionFiltersVideoOnlyExcludesAudioFilters(t *testing.T) {
	sandbox.Action(t, func() carapace.Action {
		return ActionFilters(FilterOpts{Audio: false, Video: true})
	})(func(s *sandbox.Sandbox) {
		s.Run("").ExpectNot(carapace.ActionValues("acompressor"))
		s.Run("").ExpectNot(carapace.ActionValues("aecho"))
	})
}

func TestActionFiltersVideoOnlyIncludesVideoFilters(t *testing.T) {
	sandbox.Action(t, func() carapace.Action {
		return ActionFilters(FilterOpts{Audio: false, Video: true})
	})(func(s *sandbox.Sandbox) {
		s.Run("").ExpectNot(carapace.ActionValues())
	})
}

// BSF tests use ExpectNot for the same reason as filters — large dynamic result sets.

func TestActionBitstreamFiltersAudioOnlyExcludesVideoBSFs(t *testing.T) {
	sandbox.Action(t, func() carapace.Action {
		return ActionBitstreamFilters(BsfOpts{Audio: true})
	})(func(s *sandbox.Sandbox) {
		s.Run("").ExpectNot(carapace.ActionValues("h264_mp4toannexb"))
		s.Run("").ExpectNot(carapace.ActionValues("hevc_metadata"))
		s.Run("").ExpectNot(carapace.ActionValues("vp9_superframe"))
	})
}

func TestActionBitstreamFiltersVideoOnlyExcludesAudioBSFs(t *testing.T) {
	sandbox.Action(t, func() carapace.Action {
		return ActionBitstreamFilters(BsfOpts{Video: true})
	})(func(s *sandbox.Sandbox) {
		s.Run("").ExpectNot(carapace.ActionValues("aac_adtstoasc"))
		s.Run("").ExpectNot(carapace.ActionValues("eac3_core"))
		s.Run("").ExpectNot(carapace.ActionValues("opus_metadata"))
	})
}

func TestActionBitstreamFiltersSubtitleOnlyExcludesVideoAndAudioBSFs(t *testing.T) {
	sandbox.Action(t, func() carapace.Action {
		return ActionBitstreamFilters(BsfOpts{Subtitle: true})
	})(func(s *sandbox.Sandbox) {
		s.Run("").ExpectNot(carapace.ActionValues("h264_mp4toannexb"))
		s.Run("").ExpectNot(carapace.ActionValues("aac_adtstoasc"))
	})
}
