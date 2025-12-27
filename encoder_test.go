// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package svg

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mrjoshuak/gotio/opentimelineio"
	"github.com/mrjoshuak/gotio/opentime"
)

func TestNewEncoder(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}

	if enc.width != DefaultWidth {
		t.Errorf("Expected width %d, got %d", DefaultWidth, enc.width)
	}

	if enc.height != DefaultHeight {
		t.Errorf("Expected height %d, got %d", DefaultHeight, enc.height)
	}
}

func TestSetSize(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	enc.SetSize(800, 400)

	if enc.width != 800 {
		t.Errorf("Expected width 800, got %d", enc.width)
	}

	if enc.height != 400 {
		t.Errorf("Expected height 400, got %d", enc.height)
	}
}

func TestEncodeSimpleTimeline(t *testing.T) {
	// Create a simple timeline with one track and one clip
	timeline := opentimelineio.NewTimeline("Test Timeline", nil, nil)

	track := opentimelineio.NewTrack("Video Track", nil, opentimelineio.TrackKindVideo, nil, nil)

	// Create a clip with duration
	sourceRange := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(240, 24), // 10 seconds at 24fps
	)
	clip := opentimelineio.NewClip(
		"Test Clip",
		nil,
		&sourceRange,
		nil,
		nil,
		nil,
		"",
		nil,
	)

	if err := track.AppendChild(clip); err != nil {
		t.Fatalf("Failed to append clip: %v", err)
	}

	if err := timeline.Tracks().AppendChild(track); err != nil {
		t.Fatalf("Failed to append track: %v", err)
	}

	// Encode to SVG
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if err := enc.Encode(timeline); err != nil {
		t.Fatalf("Failed to encode timeline: %v", err)
	}

	svg := buf.String()

	// Verify SVG structure
	if !strings.Contains(svg, `<?xml version="1.0" encoding="UTF-8"?>`) {
		t.Error("SVG missing XML declaration")
	}

	if !strings.Contains(svg, `<svg xmlns="http://www.w3.org/2000/svg"`) {
		t.Error("SVG missing opening tag")
	}

	if !strings.Contains(svg, `</svg>`) {
		t.Error("SVG missing closing tag")
	}

	// Verify timeline elements
	if !strings.Contains(svg, `class="track"`) {
		t.Error("SVG missing track")
	}

	if !strings.Contains(svg, `class="clip"`) {
		t.Error("SVG missing clip")
	}

	if !strings.Contains(svg, `class="ruler"`) {
		t.Error("SVG missing time ruler")
	}
}

func TestEncodeMultipleTracksAndClips(t *testing.T) {
	timeline := opentimelineio.NewTimeline("Multi-Track Timeline", nil, nil)

	// Create video track
	videoTrack := opentimelineio.NewTrack("Video 1", nil, opentimelineio.TrackKindVideo, nil, nil)

	sr1 := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(120, 24), // 5 seconds
	)
	clip1 := opentimelineio.NewClip(
		"Clip 1",
		nil,
		&sr1,
		nil,
		nil,
		nil,
		"",
		nil,
	)

	sr2 := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(120, 24), // 5 seconds
	)
	clip2 := opentimelineio.NewClip(
		"Clip 2",
		nil,
		&sr2,
		nil,
		nil,
		nil,
		"",
		nil,
	)

	if err := videoTrack.AppendChild(clip1); err != nil {
		t.Fatalf("Failed to append clip1: %v", err)
	}

	if err := videoTrack.AppendChild(clip2); err != nil {
		t.Fatalf("Failed to append clip2: %v", err)
	}

	// Create audio track
	audioTrack := opentimelineio.NewTrack("Audio 1", nil, opentimelineio.TrackKindAudio, nil, nil)

	srAudio := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 48000),
		opentime.NewRationalTime(480000, 48000), // 10 seconds
	)
	audioClip := opentimelineio.NewClip(
		"Audio Clip",
		nil,
		&srAudio,
		nil,
		nil,
		nil,
		"",
		nil,
	)

	if err := audioTrack.AppendChild(audioClip); err != nil {
		t.Fatalf("Failed to append audio clip: %v", err)
	}

	if err := timeline.Tracks().AppendChild(videoTrack); err != nil {
		t.Fatalf("Failed to append video track: %v", err)
	}

	if err := timeline.Tracks().AppendChild(audioTrack); err != nil {
		t.Fatalf("Failed to append audio track: %v", err)
	}

	// Encode to SVG
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if err := enc.Encode(timeline); err != nil {
		t.Fatalf("Failed to encode timeline: %v", err)
	}

	svg := buf.String()

	// Verify multiple clips
	clipCount := strings.Count(svg, `class="clip"`)
	if clipCount != 3 {
		t.Errorf("Expected 3 clips, found %d", clipCount)
	}

	// Verify track labels
	if !strings.Contains(svg, "Video 1") {
		t.Error("SVG missing video track label")
	}

	if !strings.Contains(svg, "Audio 1") {
		t.Error("SVG missing audio track label")
	}
}

func TestEncodeWithGaps(t *testing.T) {
	timeline := opentimelineio.NewTimeline("Timeline with Gaps", nil, nil)

	track := opentimelineio.NewTrack("Video Track", nil, opentimelineio.TrackKindVideo, nil, nil)

	srGap1 := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(48, 24), // 2 seconds
	)
	clip1 := opentimelineio.NewClip(
		"Clip 1",
		nil,
		&srGap1,
		nil,
		nil,
		nil,
		"",
		nil,
	)

	gap := opentimelineio.NewGapWithDuration(opentime.NewRationalTime(24, 24)) // 1 second gap

	srGap2 := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(48, 24), // 2 seconds
	)
	clip2 := opentimelineio.NewClip(
		"Clip 2",
		nil,
		&srGap2,
		nil,
		nil,
		nil,
		"",
		nil,
	)

	if err := track.AppendChild(clip1); err != nil {
		t.Fatalf("Failed to append clip1: %v", err)
	}

	if err := track.AppendChild(gap); err != nil {
		t.Fatalf("Failed to append gap: %v", err)
	}

	if err := track.AppendChild(clip2); err != nil {
		t.Fatalf("Failed to append clip2: %v", err)
	}

	if err := timeline.Tracks().AppendChild(track); err != nil {
		t.Fatalf("Failed to append track: %v", err)
	}

	// Encode to SVG
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if err := enc.Encode(timeline); err != nil {
		t.Fatalf("Failed to encode timeline: %v", err)
	}

	svg := buf.String()

	// Verify gap is rendered
	if !strings.Contains(svg, `class="gap"`) {
		t.Error("SVG missing gap")
	}

	// Verify both clips
	clipCount := strings.Count(svg, `class="clip"`)
	if clipCount != 2 {
		t.Errorf("Expected 2 clips, found %d", clipCount)
	}
}

func TestEncodeWithTransitions(t *testing.T) {
	timeline := opentimelineio.NewTimeline("Timeline with Transitions", nil, nil)

	track := opentimelineio.NewTrack("Video Track", nil, opentimelineio.TrackKindVideo, nil, nil)

	srTrans1 := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(96, 24), // 4 seconds
	)
	clip1 := opentimelineio.NewClip(
		"Clip 1",
		nil,
		&srTrans1,
		nil,
		nil,
		nil,
		"",
		nil,
	)

	transition := opentimelineio.NewTransition(
		"Dissolve",
		opentimelineio.TransitionTypeSMPTEDissolve,
		opentime.NewRationalTime(12, 24), // 0.5 second in
		opentime.NewRationalTime(12, 24), // 0.5 second out
		nil,
	)

	srTrans2 := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(96, 24), // 4 seconds
	)
	clip2 := opentimelineio.NewClip(
		"Clip 2",
		nil,
		&srTrans2,
		nil,
		nil,
		nil,
		"",
		nil,
	)

	if err := track.AppendChild(clip1); err != nil {
		t.Fatalf("Failed to append clip1: %v", err)
	}

	if err := track.AppendChild(transition); err != nil {
		t.Fatalf("Failed to append transition: %v", err)
	}

	if err := track.AppendChild(clip2); err != nil {
		t.Fatalf("Failed to append clip2: %v", err)
	}

	if err := timeline.Tracks().AppendChild(track); err != nil {
		t.Fatalf("Failed to append track: %v", err)
	}

	// Encode to SVG
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if err := enc.Encode(timeline); err != nil {
		t.Fatalf("Failed to encode timeline: %v", err)
	}

	svg := buf.String()

	// Verify transition is rendered
	if !strings.Contains(svg, `class="transition"`) {
		t.Error("SVG missing transition")
	}

	// Verify both clips
	clipCount := strings.Count(svg, `class="clip"`)
	if clipCount != 2 {
		t.Errorf("Expected 2 clips, found %d", clipCount)
	}
}

func TestEncodeNilTimeline(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	err := enc.Encode(nil)
	if err == nil {
		t.Error("Expected error for nil timeline")
	}

	if !strings.Contains(err.Error(), "timeline is nil") {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestEncodeEmptyTimeline(t *testing.T) {
	timeline := opentimelineio.NewTimeline("Empty Timeline", nil, nil)

	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	err := enc.Encode(timeline)
	if err == nil {
		t.Error("Expected error for empty timeline")
	}
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		seconds  float64
		expected string
	}{
		{0.5, "0.5s"},
		{1.0, "1.0s"},
		{30.5, "30.5s"},
		{60.0, "1:00"},
		{90.0, "1:30"},
		{150.0, "2:30"},
		{3600.0, "1:00:00"},
		{3661.0, "1:01:01"},
		{7265.0, "2:01:05"},
	}

	for _, tt := range tests {
		result := formatTime(tt.seconds)
		if result != tt.expected {
			t.Errorf("formatTime(%.1f) = %s, want %s", tt.seconds, result, tt.expected)
		}
	}
}

func TestSanitizeID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "unnamed"},
		{"valid_id", "valid_id"},
		{"valid-id", "valid-id"},
		{"ValidID123", "ValidID123"},
		{"invalid id!", "invalid_id_"},
		{"123invalid", "id_123invalid"},
		{"my clip name", "my_clip_name"},
		{"clip@#$%", "clip____"},
	}

	for _, tt := range tests {
		result := sanitizeID(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeID(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestCalculateTimeInterval(t *testing.T) {
	tests := []struct {
		duration float64
		wantMin  float64
		wantMax  float64
	}{
		{10.0, 0.5, 2.0},
		{60.0, 2.0, 10.0},
		{300.0, 10.0, 60.0},
		{3600.0, 120.0, 600.0},
	}

	for _, tt := range tests {
		result := calculateTimeInterval(tt.duration)
		if result < tt.wantMin || result > tt.wantMax {
			t.Errorf("calculateTimeInterval(%.1f) = %.1f, want between %.1f and %.1f",
				tt.duration, result, tt.wantMin, tt.wantMax)
		}
	}
}

func TestCustomSize(t *testing.T) {
	timeline := opentimelineio.NewTimeline("Test Timeline", nil, nil)

	track := opentimelineio.NewTrack("Video Track", nil, opentimelineio.TrackKindVideo, nil, nil)

	srCustom := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(240, 24),
	)
	clip := opentimelineio.NewClip(
		"Test Clip",
		nil,
		&srCustom,
		nil,
		nil,
		nil,
		"",
		nil,
	)

	if err := track.AppendChild(clip); err != nil {
		t.Fatalf("Failed to append clip: %v", err)
	}

	if err := timeline.Tracks().AppendChild(track); err != nil {
		t.Fatalf("Failed to append track: %v", err)
	}

	// Encode with custom size
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	enc.SetSize(800, 400)

	if err := enc.Encode(timeline); err != nil {
		t.Fatalf("Failed to encode timeline: %v", err)
	}

	svg := buf.String()

	// Verify custom dimensions
	if !strings.Contains(svg, `width="800"`) {
		t.Error("SVG missing custom width")
	}

	if !strings.Contains(svg, `height="400"`) {
		t.Error("SVG missing custom height")
	}
}
