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
)

var (
	// ErrBusyPlaying is returned when a command that attempt to modify the
	// state of the LED strip is called when an animation is currently running.
	ErrBusyPlaying = errors.New("led strip is busy playing an animation")

	// ErrNoPixels is returned when a null number of pixels is used to create
	// a BlinkyTape instance.
	ErrNoPixels = errors.New("number of pixels cannot be null")

	// ErrEmptyBuffer is returned when an attempt to send accumulated data to the
	// led strip find an empty buffer.
	ErrEmptyBuffer = errors.New("nothing to render, the buffer is empty")

	// ErrWriteCtrlHeader is returned when an error occur while attempting to write
	// the control header at the end of the buffer.
	ErrWriteCtrlHeader = errors.New("couldn't write control header at the end of the buffer")

	// ErrOutOfRange is returned when attempting to set a pixel out of the range of
	// the led strip available pixels.
	ErrOutOfRange = errors.New("attempting to set pixel outside of range")

	// ErrUnknownColorName is returned when a named color is unknown.
	ErrUnknownColorName = errors.New("unknown color name")
)

// PixelError describes an error related to a pixel command.
// It provides a copy of the pixel, and its position inside the
// LED strip's pixels range.
type PixelError struct {
	Pixel    Pixel
	Position uint
	err      error
}

func (e PixelError) Error() string {
	return fmt.Sprintf("pixel error at position #%d - RGB(%v,%v,%v): %s\n",
		e.Position,
		e.Pixel.Color.R, e.Pixel.Color.G, e.Pixel.Color.B,
		e.err,
	)
}

// RangeError describes an error related to a range excess.
// It provides the position that caused the error, and the max
// range of the led strip's pixels.
type RangeError struct {
	Position uint
	MaxRange uint
}

func (e RangeError) Error() string {
	return fmt.Sprintf("range error: trying to set pixel at position %d, allowed range is [0-%d]",
		e.Position,
		e.MaxRange,
	)
}

// InvalidHEXColor describes an error related to an HTML hex-color parsing.
type InvalidHEXColor struct {
	color string
	err   error
}

func (e InvalidHEXColor) Error() string {
	return fmt.Sprintf("%v is not an hexadecimal color: %s", e.color, e.err.Error())
}
