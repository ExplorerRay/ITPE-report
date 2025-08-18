package plot

import (
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

// styleManager ensures that a given data series (identified by a string key)
// always receives the same line and glyph style across all plots.
type styleManager struct {
	lineStyles  map[string]draw.LineStyle
	glyphStyles map[string]draw.GlyphStyle
	colorIndex  int
	dashIndex   int
	shapeIndex  int
}

// newStyleManager initializes a new style manager.
func newStyleManager() *styleManager {
	return &styleManager{
		lineStyles:  make(map[string]draw.LineStyle),
		glyphStyles: make(map[string]draw.GlyphStyle),
	}
}

// getStyle retrieves the consistent style for a given key. If the key is new,
// it assigns and stores a new style from the default palettes.
// The style palettes (color, dash, shape) are cycled in a nested fashion:
// color cycles first, then dash, then shape.
func (sm *styleManager) getStyle(key string) (draw.LineStyle, draw.GlyphStyle) {
	// If a style for this key already exists, return it.
	if lineStyle, ok := sm.lineStyles[key]; ok {
		return lineStyle, sm.glyphStyles[key]
	}

	// Ensure defaults exist to avoid panics.
	if len(plotutil.DefaultDashes) == 0 {
		plotutil.DefaultDashes = [][]vg.Length{{}} // fallback: solid line
	}
	if len(plotutil.DefaultGlyphShapes) == 0 {
		plotutil.DefaultGlyphShapes = []draw.GlyphDrawer{draw.CircleGlyph{}}
	}

	// Assign new styles from palettes.
	lineStyle := draw.LineStyle{
		Color:  plotutil.DefaultColors[sm.colorIndex],
		Dashes: plotutil.DefaultDashes[sm.dashIndex],
		Width:  vg.Points(1.5),
	}

	glyphStyle := draw.GlyphStyle{
		Color:  lineStyle.Color, // Match line color
		Radius: vg.Points(3),
		Shape:  plotutil.DefaultGlyphShapes[sm.shapeIndex],
	}

	// Store them.
	sm.lineStyles[key] = lineStyle
	sm.glyphStyles[key] = glyphStyle

	// Cycle palettes for next key.
	sm.colorIndex++
	if sm.colorIndex >= len(plotutil.DefaultColors) {
		sm.colorIndex = 0
		sm.dashIndex++
		if sm.dashIndex >= len(plotutil.DefaultDashes) {
			sm.dashIndex = 0
			sm.shapeIndex++
			if sm.shapeIndex >= len(plotutil.DefaultGlyphShapes) {
				sm.shapeIndex = 0
			}
		}
	}

	return lineStyle, glyphStyle
}
