package layout

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// PlaceOverlay positions a dialog overlay on top of background content.
func PlaceOverlay(bgWidth, bgHeight int, fg string, bg string) string {
	fgLines := strings.Split(fg, "\n")
	bgLines := strings.Split(bg, "\n")

	fgW := lipgloss.Width(fg)
	fgH := len(fgLines)

	// Center the overlay
	startX := (bgWidth - fgW) / 2
	startY := (bgHeight - fgH) / 2

	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}

	// Pad background to full size
	for len(bgLines) < bgHeight {
		bgLines = append(bgLines, strings.Repeat(" ", bgWidth))
	}

	// Overlay foreground onto background
	for i, fgLine := range fgLines {
		y := startY + i
		if y >= len(bgLines) {
			break
		}

		bgLine := bgLines[y]
		// Pad background line if needed
		for len(bgLine) < bgWidth {
			bgLine += " "
		}

		// Replace portion of background with foreground
		before := ""
		if startX > 0 && startX < len(bgLine) {
			before = bgLine[:startX]
		}
		after := ""
		endX := startX + lipgloss.Width(fgLine)
		if endX < len(bgLine) {
			after = bgLine[endX:]
		}

		bgLines[y] = before + fgLine + after
	}

	return strings.Join(bgLines[:bgHeight], "\n")
}
