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

// Package blinkygo provides utilities to control a
// BlinkyTape LED strip.
package blinkygo

import (
	"bytes"
	"log"
	"sync"
	"time"

	"github.com/ivahaev/timer"
	"github.com/tarm/serial"
)

const (
	// ControlHeader is the byte sent to the led strip to render a new state.
	ControlHeader byte = 0xFF

	// AnimationDefaultDelay is the default delay to wait between two frames
	// of an pattern.
	AnimationDefaultDelay time.Duration = 75 * time.Millisecond
)

// Status constants.
const (
	// StatusStopped means no animation is running.
	StatusStopped AnimationStatus = iota
	// StatusRunning means an animations is being played.
	StatusRunning
	// StatusPaused means a running animation is paused.
	StatusPaused
)

// AnimationStatus represents the animation loop status
// of a BlinkyTape instance.
type AnimationStatus int

// A BlinkyTape represents a BlinkyTape LED strip.
// All operations that modify the state of the strip are buffered.
type BlinkyTape struct {
	serial               *serial.Port
	currState, nextState []Pixel
	buffer               bytes.Buffer
	stop, pause, resume  chan struct{}
	position             uint
	PixelCount           uint
	status               AnimationStatus
	mutex                sync.Mutex
}

// NewBlinkyTape creates a new BlinkyTape instance.
// The led strip is created with all pixels set to black.
func NewBlinkyTape(portName string, count uint) (*BlinkyTape, error) {
	if count == 0 {
		return nil, ErrNoPixels
	}

	config := serial.Config{
		Name:        portName,
		Baud:        115200,
		ReadTimeout: time.Millisecond * 500,
	}
	port, err := serial.OpenPort(&config)
	if err != nil {
		return nil, err
	}

	blinky := BlinkyTape{
		serial:     port,
		currState:  make([]Pixel, count),
		nextState:  make([]Pixel, count),
		pause:      make(chan struct{}),
		resume:     make(chan struct{}),
		stop:       make(chan struct{}),
		position:   0,
		PixelCount: count,
		status:     StatusStopped,
	}

	// send the control header after initializtion to stop any pattern
	// playing on the LED strip before issuing future commands.
	if err := blinky.sendBytes([]byte{ControlHeader}); err != nil {
		return nil, err
	}
	return &blinky, nil
}

// Close closes the serial port.
func (bt *BlinkyTape) Close() error {
	bt.Stop()
	return bt.serial.Close()
}

// Render sends all accumulated pixel data followed by a control byte
// to the LED strip to render a new state. It also reset the internal
// buffer and reset the next position to 0.
func (bt *BlinkyTape) Render() error {
	if bt.IsRunning() {
		return ErrBusyPlaying
	}
	return bt.render()
}

func (bt *BlinkyTape) render() error {
	if bt.buffer.Len() == 0 {
		return ErrEmptyBuffer
	}
	if _, errW := bt.buffer.Write([]byte{ControlHeader}); errW != nil {
		return ErrWriteCtrlHeader
	}
	if err := bt.sendBytes(bt.buffer.Bytes()); err != nil {
		return err
	}
	bt.clear()
	bt.currState = bt.nextState

	return nil
}

// Reset discards any changes made to the LED strip's state.
func (bt *BlinkyTape) Reset() error {
	if bt.IsRunning() {
		return ErrBusyPlaying
	}
	bt.clear()
	bt.nextState = bt.currState

	return nil
}

func (bt *BlinkyTape) clear() {
	bt.position = 0
	bt.buffer.Reset()
}

// SwitchOff switches off the LED strip.
// This actually set the color to black for all pixels and calls Render().
func (bt *BlinkyTape) SwitchOff() error {
	if err := bt.SetColor(NewRGBColor(0, 0, 0)); err != nil {
		return err
	}
	return bt.Render()
}

// Play plays an Animation with the LED strip.
// If an animation is already being played, it is stopped in
// favor of the new one. The animation loop can be paused, resumed
// or stopped at any moment, regardless its status.
// A negative number of repetitions will start an infinite loop.
func (bt *BlinkyTape) Play(a *Animation, cfg *AnimationConfig) {
	var (
		repeat int
		delay  time.Duration
	)

	if cfg == nil {
		repeat = a.Repeat
		if a.Speed != 0 {
			delay = time.Second / time.Duration(a.Speed)
		} else {
			delay = AnimationDefaultDelay
		}
	} else {
		repeat = cfg.Repeat
		delay = cfg.Delay
	}

	// avoid entering the loop if there is no repetitions to process
	if repeat != 0 {
		bt.Stop()
		go bt.animation(a.Pattern, repeat, delay)
	}
}

// Status returns the animation status of the LED strip.
func (bt *BlinkyTape) Status() AnimationStatus {
	return bt.status
}

func (bt *BlinkyTape) updateStatus(as AnimationStatus) {
	bt.mutex.Lock()
	bt.status = as
	bt.mutex.Unlock()
}

// IsRunning returns whether or not an animation is running.
func (bt *BlinkyTape) IsRunning() bool {
	return bt.status == StatusRunning
}

// Stop stops the animation being played on the LED strip. A stop can
// occur at any moment between the render of two frames regardless the
// delay, or during a pause.
// If there is no animation being played or paused, do nothing.
func (bt *BlinkyTape) Stop() {
	bt.mutex.Lock()
	if bt.status == StatusRunning || bt.status == StatusPaused {
		bt.stop <- struct{}{}
	}
	bt.mutex.Unlock()
}

// Pause pauses the animation being played on the LED strip.
// If there is no animation being played, do nothing.
func (bt *BlinkyTape) Pause() {
	bt.mutex.Lock()
	if bt.status == StatusRunning {
		bt.pause <- struct{}{}
	}
	bt.mutex.Unlock()
}

// Resume resumes a previous animation that was paused.
// If the animation was paused during the delay between the render of
// two frames, the remaining of the delay will be respected.
// If there is no animation to resume, do nothing.
func (bt *BlinkyTape) Resume() {
	bt.mutex.Lock()
	if bt.status == StatusPaused {
		bt.resume <- struct{}{}
	}
	bt.mutex.Unlock()
}

func (bt *BlinkyTape) animation(p Pattern, repeat int, delay time.Duration) {
	bt.updateStatus(StatusRunning)

	// if the number of repetitions is less than zero, launch
	// an infinite loop that can be broken by calling Stop()
	if repeat < 0 {
		for {
			if !bt.playPattern(p, delay) {
				break
			}
		}
	} else {
		for i := 0; i < repeat; i++ {
			if !bt.playPattern(p, delay) {
				break
			}
		}
	}
	bt.updateStatus(StatusStopped)
}

func (bt *BlinkyTape) playPattern(p Pattern, delay time.Duration) bool {
	bt.clear()

	for _, frame := range p {
		bt.setPixels(frame)

		if err := bt.render(); err != nil {
			log.Fatalf("render error: %s\n", err)
		}

		timer := timer.NewTimer(delay)
		timer.Start()

		select {
		case <-bt.stop:
			return false
		case <-bt.pause:
			if paused := timer.Pause(); paused != false {
				bt.updateStatus(StatusPaused)
				select {
				case <-bt.stop:
					return false
				case <-bt.resume:
					if started := timer.Start(); started != false {
						bt.updateStatus(StatusRunning)
						select {
						case <-bt.stop:
							return false
						case <-timer.C:
							continue
						}
					}
				}
			}
		case <-timer.C:
			continue
		}
	}
	return true
}

// SetColor sets all pixels to the same color.
func (bt *BlinkyTape) SetColor(c Color) error {
	if bt.IsRunning() {
		return ErrBusyPlaying
	}
	bt.clear()

	pixel := Pixel{Color: c}
	for i := 0; i < int(bt.PixelCount); i++ {
		if err := bt.setNextPixel(pixel); err != nil {
			return err
		}
	}
	return nil
}

// SetPixels sets pixels from a list.
func (bt *BlinkyTape) SetPixels(p []Pixel) error {
	if bt.IsRunning() {
		return ErrBusyPlaying
	}
	return bt.setPixels(p)
}

func (bt *BlinkyTape) setPixels(pixels []Pixel) error {
	for c, p := range pixels {
		if c > (int(bt.PixelCount) - 1) {
			break
		}
		if err := bt.setNextPixel(p); err != nil {
			return err
		}
	}
	return nil
}

// SetNextPixel sets a pixel at the next position.
func (bt *BlinkyTape) SetNextPixel(p Pixel) error {
	if bt.IsRunning() {
		return ErrBusyPlaying
	}
	return bt.setNextPixel(p)
}

func (bt *BlinkyTape) setNextPixel(p Pixel) error {
	if bt.position > (bt.PixelCount - 1) {
		return RangeError{
			Position: bt.position,
			MaxRange: bt.PixelCount - 1,
		}
	}
	if _, err := bt.buffer.Write(p.clampedRGBTriplet()); err != nil {
		return PixelError{
			Pixel:    p,
			Position: bt.position,
			err:      err,
		}
	}
	bt.nextState[bt.position] = p
	bt.position++

	return nil
}

// SetPixelAt sets a pixel at the specified position.
// The operation has to rewrite the whole buffer.
func (bt *BlinkyTape) SetPixelAt(p *Pixel, position uint) error {
	if bt.IsRunning() {
		return ErrBusyPlaying
	}
	if position > bt.PixelCount {
		return ErrOutOfRange
	}

	bt.nextState[position] = *p
	bt.buffer.Reset()

	for _, p := range bt.nextState {
		if _, err := bt.buffer.Write(p.clampedRGBTriplet()); err != nil {
			return err
		}
	}
	return nil
}

func (bt *BlinkyTape) sendBytes(data []byte) error {
	if err := bt.serial.Flush(); err != nil {
		return err
	}
	if _, err := bt.serial.Write(data); err != nil {
		return err
	}
	return nil
}
