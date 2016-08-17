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
	"bufio"
	"fmt"
	"image"
	"image/draw"
	"os"
	"strings"

	// Image decoding
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/nfnt/resize"
	// Image decoding
	_ "golang.org/x/image/bmp"
)

// A Frame represents a list of pixels.
type Frame []Pixel

// A Pattern represents a list of frames.
type Pattern []Frame

// NewPatternFromImage returns a new pattern created from an image.
// Types 'jpeg', 'png', 'gif' and 'bmp' are supported.
func NewPatternFromImage(path string, pixelCount uint) (Pattern, error) {
	if pixelCount == 0 {
		return nil, ErrNoPixels
	}
	source, err := readImage(path)
	if err != nil {
		return nil, err
	}

	var img image.Image
	bounds := source.Bounds()
	if bounds.Dy() > int(pixelCount) {
		// the resize function preserves the aspect ratio
		img = resize.Resize(0, pixelCount, source, resize.Bilinear)
	} else {
		img = source
	}

	// draw a new rgba image from source, so we can directly
	// access to the rgb representation of each pixel in Pix field
	bounds = img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	// loop across the pixels and create frames
	pattern := make(Pattern, bounds.Dx())

	for x := 0; x < bounds.Dx(); x++ {
		f := make(Frame, pixelCount)
		for y := 0; y < bounds.Dy(); y++ {
			r := rgba.Pix[rgba.PixOffset(x, y)]
			g := rgba.Pix[rgba.PixOffset(x, y)+1]
			b := rgba.Pix[rgba.PixOffset(x, y)+2]

			f[y] = Pixel{
				Color: NewRGBColor(brightnessCorrect(r, g, b)),
			}
		}
		pattern[x] = f
	}
	return pattern, nil
}

// readImage open a file and return an image.
func readImage(path string) (image.Image, error) {
	reader, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}
	reader.Close()

	return img, nil
}

// NewPatternFromArduinoExport returns a new pattern created
// from an Arduino C header file exported from PatternPaint.
func NewPatternFromArduinoExport(path string) (Pattern, error) {
	pattern := make(Pattern, 0)

	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}
	fc := -1
	ll := len(lines) - 3

	// loop across the lines and create frames
	// ignore the first line, and the last three
	for i := 1; i < ll; i++ {
		if strings.HasPrefix(lines[i], "//") {
			pattern = append(pattern, make(Frame, 0))
			fc++
		} else {
			var r, g, b byte
			_, err := fmt.Sscanf(lines[i], "%v,%v,%v,", &r, &g, &b)
			if err != nil {
				return nil, err
			}
			pattern[fc] = append(pattern[fc], Pixel{
				Color: NewRGBColor(r, g, b),
			})
		}
	}
	return pattern, nil
}

// readLines reads the content of a file
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	reader, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	var lines []string
	scanner := bufio.NewScanner(reader)

	// read each line and append it in a slice that
	// grows dynamically
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	reader.Close()

	return lines, scanner.Err()
}
