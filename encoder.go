// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

// Package svg provides a write-only adapter for visualizing OpenTimelineIO timelines as SVG graphics.
//
// The encoder renders timelines with horizontal track lanes, showing clips as colored rectangles,
// transitions as diagonal lines, and gaps with dashed borders. Each track type (video/audio) uses
// different colors, and a time ruler is included at the top of the visualization.
//
// Basic usage:
//
//	encoder := svg.NewEncoder(outputWriter)
//	encoder.SetSize(1200, 600)  // optional
//	err := encoder.Encode(timeline)
package svg

import (
	"fmt"
	"io"
	"math"

	"github.com/mrjoshuak/gotio/opentimelineio"
	"github.com/mrjoshuak/gotio/opentime"
)

// Default dimensions and styling constants.
const (
	DefaultWidth       = 1200
	DefaultHeight      = 600
	TrackHeight        = 80
	MarginTop          = 60
	MarginBottom       = 40
	MarginLeft         = 100
	MarginRight        = 40
	RulerHeight        = 40
	TransitionWidth    = 20
	MinClipWidth       = 5
	FontSize           = 12
	SmallFontSize      = 10
)

// Color scheme.
const (
	VideoTrackColor      = "#4A90E2"
	AudioTrackColor      = "#50C878"
	GapColor             = "#E0E0E0"
	TransitionColor      = "#FFB84D"
	BackgroundColor      = "#FFFFFF"
	GridColor            = "#CCCCCC"
	TextColor            = "#333333"
	RulerTextColor       = "#666666"
	TrackLabelBg         = "#F5F5F5"
)

// Encoder encodes OTIO timelines as SVG.
type Encoder struct {
	w      io.Writer
	width  int
	height int
}

// NewEncoder creates a new SVG encoder.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w:      w,
		width:  DefaultWidth,
		height: DefaultHeight,
	}
}

// SetSize sets the SVG canvas size.
func (e *Encoder) SetSize(width, height int) {
	e.width = width
	e.height = height
}

// Encode encodes a timeline to SVG.
func (e *Encoder) Encode(t *opentimelineio.Timeline) error {
	if t == nil {
		return fmt.Errorf("timeline is nil")
	}

	builder := NewSVGBuilder(e.w)

	// Write SVG header
	if err := builder.WriteHeader(e.width, e.height); err != nil {
		return err
	}

	// Write CSS styles
	if err := e.writeStyles(builder); err != nil {
		return err
	}

	// Get timeline duration
	duration, err := t.Duration()
	if err != nil {
		return fmt.Errorf("failed to get timeline duration: %w", err)
	}

	if duration.Value() <= 0 {
		return fmt.Errorf("timeline has no duration")
	}

	// Calculate content area
	contentWidth := float64(e.width - MarginLeft - MarginRight)
	contentHeight := float64(e.height - MarginTop - MarginBottom)

	// Get all tracks
	tracks := t.Tracks()
	if tracks == nil {
		return fmt.Errorf("timeline has no tracks")
	}

	allTracks := tracks.Children()
	numTracks := len(allTracks)
	if numTracks == 0 {
		return fmt.Errorf("timeline has no tracks")
	}

	// Calculate scale: pixels per second
	durationSeconds := duration.ToSeconds()
	timeScale := contentWidth / durationSeconds

	// Draw time ruler at top
	if err := e.drawTimeRuler(builder, duration, timeScale); err != nil {
		return err
	}

	// Draw each track
	trackHeight := TrackHeight
	if numTracks > 0 {
		availableHeight := contentHeight - RulerHeight
		trackHeight = int(availableHeight / float64(numTracks))
		if trackHeight > TrackHeight {
			trackHeight = TrackHeight
		}
		if trackHeight < 40 {
			trackHeight = 40
		}
	}

	for i, child := range allTracks {
		track, ok := child.(*opentimelineio.Track)
		if !ok {
			continue
		}

		yOffset := MarginTop + RulerHeight + float64(i*trackHeight)
		if err := e.drawTrack(builder, track, yOffset, float64(trackHeight), timeScale); err != nil {
			return err
		}
	}

	// Write SVG footer
	if err := builder.WriteFooter(); err != nil {
		return err
	}

	return nil
}

// writeStyles writes CSS styles for the SVG.
func (e *Encoder) writeStyles(builder *SVGBuilder) error {
	css := `
    .track-label {
      font-family: Arial, sans-serif;
      font-size: 12px;
      fill: #333;
      font-weight: bold;
    }
    .clip-label {
      font-family: Arial, sans-serif;
      font-size: 10px;
      fill: white;
      pointer-events: none;
    }
    .ruler-text {
      font-family: Arial, sans-serif;
      font-size: 10px;
      fill: #666;
    }
    .clip {
      stroke: #333;
      stroke-width: 1;
    }
    .gap {
      stroke: #999;
      stroke-width: 1;
      stroke-dasharray: 2,2;
    }
    .transition {
      stroke: #333;
      stroke-width: 2;
      fill: none;
    }
  `
	return builder.WriteStyle(css)
}

// drawTimeRuler draws the time ruler at the top.
func (e *Encoder) drawTimeRuler(builder *SVGBuilder, duration opentime.RationalTime, timeScale float64) error {
	if err := builder.StartGroup("time-ruler", "ruler"); err != nil {
		return err
	}

	// Draw ruler background
	rulerY := float64(MarginTop)
	rulerWidth := float64(e.width - MarginLeft - MarginRight)
	if err := builder.WriteRect(float64(MarginLeft), rulerY, rulerWidth, RulerHeight, TrackLabelBg, GridColor, "", "ruler-bg", ""); err != nil {
		return err
	}

	// Draw time markers
	durationSeconds := duration.ToSeconds()

	// Calculate appropriate interval
	interval := calculateTimeInterval(durationSeconds)

	time := 0.0
	for time <= durationSeconds {
		x := float64(MarginLeft) + time*timeScale

		// Draw tick mark
		if err := builder.WriteLine(x, rulerY, x, rulerY+RulerHeight, GridColor, 1, "tick"); err != nil {
			return err
		}

		// Draw time label
		timeLabel := formatTime(time)
		if err := builder.WriteText(x, rulerY+RulerHeight/2, timeLabel, "middle", "", "ruler-text"); err != nil {
			return err
		}

		time += interval
	}

	return builder.EndGroup()
}

// drawTrack draws a single track.
func (e *Encoder) drawTrack(builder *SVGBuilder, track *opentimelineio.Track, yOffset, height, timeScale float64) error {
	trackID := fmt.Sprintf("track-%s", sanitizeID(track.Name()))
	if err := builder.StartGroup(trackID, "track"); err != nil {
		return err
	}

	// Draw track background
	trackColor := VideoTrackColor
	if track.Kind() == opentimelineio.TrackKindAudio {
		trackColor = AudioTrackColor
	}

	// Track background with slight transparency
	bgColor := trackColor + "33" // Add alpha
	if err := builder.WriteRect(float64(MarginLeft), yOffset, float64(e.width-MarginLeft-MarginRight), height, bgColor, GridColor, "", "track-bg", ""); err != nil {
		return err
	}

	// Draw track label
	labelText := track.Name()
	if labelText == "" {
		labelText = fmt.Sprintf("%s Track", track.Kind())
	}
	if err := builder.WriteText(float64(MarginLeft-10), yOffset+height/2, labelText, "end", "", "track-label"); err != nil {
		return err
	}

	// Draw items in the track
	currentTime := 0.0
	for _, child := range track.Children() {
		dur, err := child.Duration()
		if err != nil {
			continue
		}

		durSeconds := dur.ToSeconds()

		switch item := child.(type) {
		case *opentimelineio.Clip:
			x := float64(MarginLeft) + currentTime*timeScale
			width := math.Max(durSeconds*timeScale, MinClipWidth)
			if err := e.drawClip(builder, item, x, yOffset, width, height, trackColor); err != nil {
				return err
			}
			if child.Visible() {
				currentTime += durSeconds
			}

		case *opentimelineio.Gap:
			x := float64(MarginLeft) + currentTime*timeScale
			width := math.Max(durSeconds*timeScale, MinClipWidth)
			if err := e.drawGap(builder, item, x, yOffset, width, height); err != nil {
				return err
			}
			if child.Visible() {
				currentTime += durSeconds
			}

		case *opentimelineio.Transition:
			x := float64(MarginLeft) + currentTime*timeScale
			width := math.Max(durSeconds*timeScale, MinClipWidth)
			if err := e.drawTransition(builder, item, x, yOffset, width, height); err != nil {
				return err
			}
			// Transitions don't advance time (they overlap)
		}
	}

	return builder.EndGroup()
}

// drawClip draws a clip.
func (e *Encoder) drawClip(builder *SVGBuilder, clip *opentimelineio.Clip, x, y, width, height float64, trackColor string) error {
	clipID := fmt.Sprintf("clip-%s", sanitizeID(clip.Name()))

	// Adjust clip rectangle to have some padding
	padding := 2.0
	clipY := y + padding
	clipHeight := height - 2*padding

	// Draw clip rectangle
	if err := builder.WriteRect(x, clipY, width, clipHeight, trackColor, "#333", clipID, "clip", ""); err != nil {
		return err
	}

	// Draw clip name if there's room
	if width > 30 {
		clipName := clip.Name()
		if clipName == "" {
			clipName = "Clip"
		}
		textX := x + width/2
		textY := y + height/2
		if err := builder.WriteText(textX, textY, clipName, "middle", "", "clip-label"); err != nil {
			return err
		}
	}

	return nil
}

// drawGap draws a gap.
func (e *Encoder) drawGap(builder *SVGBuilder, gap *opentimelineio.Gap, x, y, width, height float64) error {
	gapID := fmt.Sprintf("gap-%p", gap)

	padding := 2.0
	gapY := y + padding
	gapHeight := height - 2*padding

	// Draw gap rectangle with dashed border
	return builder.WriteRect(x, gapY, width, gapHeight, GapColor, "#999", gapID, "gap", "")
}

// drawTransition draws a transition as a diagonal line.
func (e *Encoder) drawTransition(builder *SVGBuilder, transition *opentimelineio.Transition, x, y, width, height float64) error {
	padding := 2.0
	transY := y + padding
	transHeight := height - 2*padding

	// Draw diagonal line from bottom-left to top-right
	x1 := x
	y1 := transY + transHeight
	x2 := x + width
	y2 := transY

	// Draw the transition path
	path := fmt.Sprintf("M %.2f %.2f L %.2f %.2f", x1, y1, x2, y2)
	return builder.WritePath(path, "none", TransitionColor, 3, "transition")
}

// calculateTimeInterval calculates an appropriate time interval for ruler marks.
func calculateTimeInterval(durationSeconds float64) float64 {
	intervals := []float64{0.1, 0.5, 1, 2, 5, 10, 15, 30, 60, 120, 300, 600, 1800, 3600}

	// Aim for about 10-20 marks
	targetMarks := 12.0
	idealInterval := durationSeconds / targetMarks

	// Find closest interval
	bestInterval := intervals[0]
	for _, interval := range intervals {
		if interval >= idealInterval {
			bestInterval = interval
			break
		}
		bestInterval = interval
	}

	return bestInterval
}

// formatTime formats seconds as a time string.
func formatTime(seconds float64) string {
	if seconds < 60 {
		return fmt.Sprintf("%.1fs", seconds)
	}

	minutes := int(seconds / 60)
	secs := int(seconds) % 60

	if minutes < 60 {
		return fmt.Sprintf("%d:%02d", minutes, secs)
	}

	hours := minutes / 60
	minutes = minutes % 60
	return fmt.Sprintf("%d:%02d:%02d", hours, minutes, secs)
}

// sanitizeID sanitizes a string for use as an XML ID.
func sanitizeID(s string) string {
	if s == "" {
		return "unnamed"
	}

	// Replace non-alphanumeric characters with underscores
	result := ""
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			result += string(r)
		} else {
			result += "_"
		}
	}

	// Ensure it starts with a letter
	if len(result) > 0 {
		first := result[0]
		if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z')) {
			result = "id_" + result
		}
	}

	return result
}
