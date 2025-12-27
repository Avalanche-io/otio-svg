// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package svg_test

import (
	"os"

	svg "github.com/mrjoshuak/otio-svg"
	"github.com/Avalanche-io/gotio/opentimelineio"
	"github.com/Avalanche-io/gotio/opentime"
)

func ExampleEncoder() {
	// Create a timeline
	timeline := opentimelineio.NewTimeline("Example Timeline", nil, nil)

	// Create a video track
	videoTrack := opentimelineio.NewTrack("Video 1", nil, opentimelineio.TrackKindVideo, nil, nil)

	// Add some clips
	sr1 := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(120, 24), // 5 seconds
	)
	clip1 := opentimelineio.NewClip("Opening Shot", nil, &sr1, nil, nil, nil, "", nil)
	videoTrack.AppendChild(clip1)

	// Add a transition
	transition := opentimelineio.NewTransition(
		"Dissolve",
		opentimelineio.TransitionTypeSMPTEDissolve,
		opentime.NewRationalTime(12, 24),
		opentime.NewRationalTime(12, 24),
		nil,
	)
	videoTrack.AppendChild(transition)

	// Add another clip
	sr2 := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 24),
		opentime.NewRationalTime(144, 24), // 6 seconds
	)
	clip2 := opentimelineio.NewClip("Main Scene", nil, &sr2, nil, nil, nil, "", nil)
	videoTrack.AppendChild(clip2)

	// Add to timeline
	timeline.Tracks().AppendChild(videoTrack)

	// Create audio track
	audioTrack := opentimelineio.NewTrack("Audio 1", nil, opentimelineio.TrackKindAudio, nil, nil)
	srAudio := opentime.NewTimeRange(
		opentime.NewRationalTime(0, 48000),
		opentime.NewRationalTime(528000, 48000), // 11 seconds
	)
	audioClip := opentimelineio.NewClip("Background Music", nil, &srAudio, nil, nil, nil, "", nil)
	audioTrack.AppendChild(audioClip)
	timeline.Tracks().AppendChild(audioTrack)

	// Encode to SVG
	encoder := svg.NewEncoder(os.Stdout)
	encoder.SetSize(1200, 400)
	encoder.Encode(timeline)

	// Output would be SVG XML
}
