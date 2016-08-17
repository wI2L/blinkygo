/*
	The MIT License

   Copyright (c) 2016, William Poussier <william.poussier@gmail.com>

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in
	all copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
	THE SOFTWARE.
*/

package blinkygo

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"golang.org/x/image/colornames"
)

// HTML hecadecimal color-string formats.
const (
	HEX3DigitsForm = "%1x%1x%1x"
	HEX6DigitsForm = "%02x%02x%02x"
)

// Exponents are the factors used to convert a color
// from the screen space to the LED space.
var (
	RedExponent   = 1.8
	GreenExponent = 1.8
	BlueExponent  = 2.1
)

// A Pixel represents a led of the strip.
type Pixel struct {
	Color Color `json:"color"`
}

// A Color represents a color.
type Color struct {
	R byte `json:"r"`
	G byte `json:"g"`
	B byte `json:"b"`
}

// NewRGBColor returns a new Color from its RGB representation.
func NewRGBColor(r, g, b byte) Color {
	r, g, b = brightnessCorrect(r, g, b)
	return Color{R: r, G: g, B: b}
}

// brightnessCorrect performs a rough brightness correction
// on a given RGB color triplet.
func brightnessCorrect(r, g, b byte) (byte, byte, byte) {
	f := func(v byte, exp float64) byte {
		return byte(255 * math.Pow(float64(v)/255.0, exp))
	}
	return f(r, RedExponent), f(g, GreenExponent), f(b, BlueExponent)
}

// NewNamedColor returns a new color from its name.
// Supported names are from the package "colornames",
// see https://godoc.org/golang.org/x/image/colornames
func NewNamedColor(name string) (Color, error) {
	color, ok := colornames.Map[name]
	if !ok {
		return Color{}, ErrUnknownColorName
	}
	return NewRGBColor(color.R, color.G, color.B), nil
}

// NewHEXColor returns a new color parsed from its "html" hex color-string format,
// either in the 3 "#F06" or 6 "#FF0066" digits form. First char '#' is optional.
func NewHEXColor(color string) (Color, error) {
	var format string

	// remove "#" if it present at the beginning of the format
	color = strings.TrimPrefix(color, "#")

	if len(color) == 3 {
		format = HEX3DigitsForm
	} else if len(color) == 6 {
		format = HEX6DigitsForm
	} else {
		return Color{}, InvalidHEXColor{
			color: color,
			err:   errors.New("invalid format"),
		}
	}
	var r, g, b byte

	n, err := fmt.Sscanf(color, format, &r, &g, &b)
	if err != nil || n != 3 {
		return Color{}, InvalidHEXColor{
			color: color,
			err:   err,
		}
	}
	return NewRGBColor(r, g, b), nil
}

// clampedRGBTriplet returns the pixel data triplet in RGB format,
// with its values clamped to 0-254 to avoid confusion with the LED
// strip's control header (0xFF).
func (p Pixel) clampedRGBTriplet() []byte {
	return []byte{
		clamp(p.Color.R),
		clamp(p.Color.G),
		clamp(p.Color.B),
	}
}

func clamp(v byte) byte {
	return byte(math.Min(float64(ControlHeader-1), float64(v)))
}
