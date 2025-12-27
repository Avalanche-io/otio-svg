// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package svg

import (
	"fmt"
	"io"
	"strings"
)

// SVGBuilder helps construct SVG documents.
type SVGBuilder struct {
	w      io.Writer
	indent int
}

// NewSVGBuilder creates a new SVG builder.
func NewSVGBuilder(w io.Writer) *SVGBuilder {
	return &SVGBuilder{w: w, indent: 0}
}

// WriteHeader writes the SVG header with dimensions.
func (b *SVGBuilder) WriteHeader(width, height int) error {
	_, err := fmt.Fprintf(b.w, `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">
`, width, height, width, height)
	b.indent = 1
	return err
}

// WriteFooter writes the closing SVG tag.
func (b *SVGBuilder) WriteFooter() error {
	_, err := fmt.Fprintf(b.w, "</svg>\n")
	return err
}

// StartGroup starts a group element.
func (b *SVGBuilder) StartGroup(id, class string) error {
	attrs := ""
	if id != "" {
		attrs += fmt.Sprintf(` id="%s"`, escapeAttr(id))
	}
	if class != "" {
		attrs += fmt.Sprintf(` class="%s"`, escapeAttr(class))
	}
	_, err := fmt.Fprintf(b.w, "%s<g%s>\n", indent(b.indent), attrs)
	b.indent++
	return err
}

// EndGroup ends a group element.
func (b *SVGBuilder) EndGroup() error {
	b.indent--
	_, err := fmt.Fprintf(b.w, "%s</g>\n", indent(b.indent))
	return err
}

// WriteRect writes a rectangle element.
func (b *SVGBuilder) WriteRect(x, y, width, height float64, fill, stroke string, id, class, text string) error {
	attrs := fmt.Sprintf(`x="%.2f" y="%.2f" width="%.2f" height="%.2f"`, x, y, width, height)
	if fill != "" {
		attrs += fmt.Sprintf(` fill="%s"`, fill)
	}
	if stroke != "" {
		attrs += fmt.Sprintf(` stroke="%s"`, stroke)
	}
	if id != "" {
		attrs += fmt.Sprintf(` id="%s"`, escapeAttr(id))
	}
	if class != "" {
		attrs += fmt.Sprintf(` class="%s"`, escapeAttr(class))
	}

	if text == "" {
		_, err := fmt.Fprintf(b.w, "%s<rect %s />\n", indent(b.indent), attrs)
		return err
	}

	// Write rect with nested text
	if _, err := fmt.Fprintf(b.w, "%s<rect %s />\n", indent(b.indent), attrs); err != nil {
		return err
	}
	// Add text label centered
	textX := x + width/2
	textY := y + height/2
	return b.WriteText(textX, textY, text, "middle", "", "clip-label")
}

// WritePath writes a path element.
func (b *SVGBuilder) WritePath(d string, fill, stroke string, strokeWidth float64, class string) error {
	attrs := fmt.Sprintf(`d="%s"`, d)
	if fill != "" {
		attrs += fmt.Sprintf(` fill="%s"`, fill)
	}
	if stroke != "" {
		attrs += fmt.Sprintf(` stroke="%s"`, stroke)
	}
	if strokeWidth > 0 {
		attrs += fmt.Sprintf(` stroke-width="%.2f"`, strokeWidth)
	}
	if class != "" {
		attrs += fmt.Sprintf(` class="%s"`, escapeAttr(class))
	}
	_, err := fmt.Fprintf(b.w, "%s<path %s />\n", indent(b.indent), attrs)
	return err
}

// WriteText writes a text element.
func (b *SVGBuilder) WriteText(x, y float64, text, anchor, id, class string) error {
	attrs := fmt.Sprintf(`x="%.2f" y="%.2f"`, x, y)
	if anchor != "" {
		attrs += fmt.Sprintf(` text-anchor="%s"`, anchor)
	}
	if id != "" {
		attrs += fmt.Sprintf(` id="%s"`, escapeAttr(id))
	}
	if class != "" {
		attrs += fmt.Sprintf(` class="%s"`, escapeAttr(class))
	}
	attrs += ` dominant-baseline="middle"`
	_, err := fmt.Fprintf(b.w, "%s<text %s>%s</text>\n", indent(b.indent), attrs, escapeText(text))
	return err
}

// WriteLine writes a line element.
func (b *SVGBuilder) WriteLine(x1, y1, x2, y2 float64, stroke string, strokeWidth float64, class string) error {
	attrs := fmt.Sprintf(`x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f"`, x1, y1, x2, y2)
	if stroke != "" {
		attrs += fmt.Sprintf(` stroke="%s"`, stroke)
	}
	if strokeWidth > 0 {
		attrs += fmt.Sprintf(` stroke-width="%.2f"`, strokeWidth)
	}
	if class != "" {
		attrs += fmt.Sprintf(` class="%s"`, escapeAttr(class))
	}
	_, err := fmt.Fprintf(b.w, "%s<line %s />\n", indent(b.indent), attrs)
	return err
}

// WriteStyle writes a style element with CSS.
func (b *SVGBuilder) WriteStyle(css string) error {
	_, err := fmt.Fprintf(b.w, "%s<style>\n%s\n%s</style>\n", indent(b.indent), css, indent(b.indent))
	return err
}

// indent creates an indentation string.
func indent(level int) string {
	return strings.Repeat("  ", level)
}

// escapeAttr escapes attribute values.
func escapeAttr(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// escapeText escapes text content.
func escapeText(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
