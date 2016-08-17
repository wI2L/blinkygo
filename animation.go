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
	"encoding/json"
	"io/ioutil"
	"time"
)

// An Animation is composed of a Pattern to play with a BlinkyTape
// based on a playback speed and an number of repetitions.
type Animation struct {
	Name    string  `json:"name"`
	Repeat  int     `json:"repeat"`
	Speed   uint    `json:"speed"`
	Pattern Pattern `json:"pattern"`
}

// AnimationConfig represents the configuration of an Animation.
type AnimationConfig struct {
	// Repeat indicates how many times the pattern has to be played
	Repeat int
	// Delay is the duration to wait between the rendering of two frames
	Delay time.Duration
}

// NewAnimationFromFile create a new Animation instance from a file.
// The animation file must use JSON as its marshalling format.
func NewAnimationFromFile(path string) (*Animation, error) {
	anim := Animation{}

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(raw, &anim)

	return &anim, nil
}

// SaveToFile marshall an Animation to JSON format and
// write it to a file.
func (a Animation) SaveToFile(path string) error {
	data, err := json.Marshal(a)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}
	return nil
}
