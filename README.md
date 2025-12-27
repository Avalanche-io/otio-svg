# otio-svg - OpenTimelineIO SVG Adapter for Go

A write-only adapter that visualizes OpenTimelineIO timelines as SVG graphics.

## Features

- Visualizes timelines as SVG with horizontal track lanes
- Shows clips as colored rectangles with names
- Displays transitions as diagonal lines between clips
- Renders gaps with dashed borders
- Includes time ruler with appropriate time markers
- Different colors for video vs audio tracks
- Customizable canvas dimensions

## Installation

```bash
go get github.com/mrjoshuak/otio-svg
```

## Usage

```go
package main

import (
    "os"

    svg "github.com/mrjoshuak/otio-svg"
    "github.com/mrjoshuak/gotio/opentimelineio"
    "github.com/mrjoshuak/gotio/opentime"
)

func main() {
    // Create a timeline
    timeline := opentimelineio.NewTimeline("My Timeline", nil, nil)

    // Create a video track
    videoTrack := opentimelineio.NewTrack("Video 1", nil, opentimelineio.TrackKindVideo, nil, nil)

    // Add a clip
    sourceRange := opentime.NewTimeRange(
        opentime.NewRationalTime(0, 24),
        opentime.NewRationalTime(240, 24), // 10 seconds at 24fps
    )
    clip := opentimelineio.NewClip("My Clip", nil, &sourceRange, nil, nil, nil, "", nil)
    videoTrack.AppendChild(clip)

    // Add track to timeline
    timeline.Tracks().AppendChild(videoTrack)

    // Encode to SVG
    file, _ := os.Create("timeline.svg")
    defer file.Close()

    encoder := svg.NewEncoder(file)
    encoder.SetSize(1200, 600) // Optional: set custom dimensions
    encoder.Encode(timeline)
}
```

## API

### NewEncoder

```go
func NewEncoder(w io.Writer) *Encoder
```

Creates a new SVG encoder that writes to the given writer.

### SetSize

```go
func (e *Encoder) SetSize(width, height int)
```

Sets the canvas dimensions. Default is 1200x600.

### Encode

```go
func (e *Encoder) Encode(t *opentimelineio.Timeline) error
```

Encodes a timeline to SVG format.

## Visual Elements

### Tracks
- **Video tracks**: Blue background (#4A90E2)
- **Audio tracks**: Green background (#50C878)
- Each track is rendered as a horizontal lane with its name on the left

### Clips
- Rendered as filled rectangles
- Display clip name if there's sufficient width
- Positioned sequentially along the track timeline

### Gaps
- Rendered with light gray fill (#E0E0E0)
- Dashed border to distinguish from clips

### Transitions
- Rendered as diagonal lines in orange (#FFB84D)
- Connect between adjacent clips

### Time Ruler
- Displayed at the top of the visualization
- Shows time markers with appropriate intervals
- Formats time as seconds, minutes:seconds, or hours:minutes:seconds

## Limitations

This is a write-only adapter. It does not support:
- Reading SVG files back to OTIO
- Nested compositions
- Effects visualization
- Marker visualization (future enhancement)

## Development

```bash
# Run tests
go test -v

# Build
go build ./...
```

## License

Apache-2.0 - See LICENSE file for details

## Contributing

Contributions welcome! This adapter is part of the gotio project.

## Reference

Based on the [Python SVG adapter](https://github.com/OpenTimelineIO/otio-svg-adapter) for OpenTimelineIO.
