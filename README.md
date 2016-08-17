# BlinkyGo

[![GoDoc](https://godoc.org/github.com/wI2l/blinkygo?status.svg)](https://godoc.org/github.com/wI2l/blinkygo)
![License MIT](https://img.shields.io/badge/license-MIT-blue.svg)

A well featured package that allow you control your [**BlinkyTape LED strip**](http://blinkinlabs.com/blinkytape/) from _BlinkyLabs_, using the [Go Programming language](golang.org).


#### Installation

```
go get github.com/wI2L/blinkygo
```

## Overview

### Basic usage

Create a new BlinkyTape instance. You must pass two parameters:
   - the serial port name to open
      eg: *COM3* (Windows), */dev/tty/usbmodem1421* (OSX)
   - the number of pixels on the LED strip

```go
import blinky "github.com/wI2L/blinkygo"

bt, err := blinky.NewBlinkyTape("/dev/tty.usbmodem1421", 60)
if err != nil {
   log.Fatal(err)
}
defer bt.Close()
```

When a new BlinkyTape instance is created, all its pixels are initialised to be black; *RGB(0, 0, 0)*. All operations that modify the state of the LED strip are buffered using an internal bytes buffer. You have to manually "commit" the changes when you want them to take effect.

__Set the next pixel__

```go
// Create a new color
red := blinky.NewRGBColor(255, 0, 0)
// Create the pixel
pixel := blinky.Pixel{Color: red}

for (i := 0; i < bt.PixelCount; i++) {
   err := bt.SetNextPixel(pixel)
}
```

`setNextPixel()` will set a pixel to the next available position. The position is incremented each time this function is called and resets when accumulated data is sent to the LED strip.
Be carefull not to exceed the number of pixels the instance was initialized for when calling this function or it will return a `RangeError`

__Set a pixel at a specified position__

```go

// Set a yellow pixel at the position 1 (second led)
pixel := blinky.Pixel{Color: NewRGBColor(255, 255, 0)}
err := bt.SetPixelAt(pixel, 1)
```

When you set a pixel at a specified position, it has to internally rewrite the whole buffer, which is slower.

__Set a list of pixels__

```go
color := NewRGBColor(0, 255, 0)
pixels := [Pixel{Color: color}]

err := bt.SetPixels(pixels)
```

Set a list of pixels, starting from the beginning. If the slice provided as argument contains more pixels than the BlinkyTape instance was initialized for, the remaining will be genly ignored.

__Set a color__

```go
white := NewRGBColor(255, 255, 255)
err := bt.SetColor(color)
```

It set the same color to all the pixels of the LED strip.

__Render accumulated data__

```go
err := bt.Render()
```

When you want to apply the changes you made, call this function. It will send all accumulated data to the LED strip and reset it state. The strip will immediately reflects the modifications.

__Discard changes__

```go
err := bt.Reset()
```
If you want to discard any previous changes without rendering them, call `Reset()`. It clear the internal buffer and reset the position of the next pixel to 0.

__Switch off LED strip__

You can also switch off the LED strip. It will set black color to all pixels and render the changes. This gives the impression the LED strip extinguished.

```go
err := bt.SwitchOff()

// This is similar to writing:
//    bt.SetColor(NewRGBColor(0, 0, 0))
//    bt.Render()
```

### Colors

There is three different ways to create a `Color` instance.

__RGB triplet__
```go
blue := blinky.NewRGBColor(0, 0, 255)
white :=  blinky.NewRGBColor(255, 255, 255)
red := blinky.NewRGBColor(255, 0, 0)
```

__HTML hex color-string format__
```go
purple, _ := blinky.NewHEXColor("#800080")
orange, _ := blinky.NewHEXColor("FFA500")
pink, _ := blinky.NewHEXColor("#F06")
```

__Named color__
```go
olive, _ := blinky.NewNamedColor("Olive")
violet, _ := blinky.NewNamedColor("Violet")
```
Supported names are from the _colornames_ package, see https://godoc.org/golang.org/x/image/colornames

`NewHEXColor()` and `NewNamedColor()` will return an error if the input format is invalid or the name is unknown.

### Pattern

A `Pattern` is a list of `Frame`, each containing a list of pixels. Patterns can be used to create an `Animation`. You create them manually, or from an external source like an Arduino C header file exported by PatternPaint, or an image.

__Decoding an image__

A pattern can be decoded from an image. You have to indicate how many pixels should be extracted per frame.

Frames will be extracted from axis `y`. If the number of pixels to extract is lower than the image's height, the rest of the pixels in each frame will be black. Instead, if the image's height is greater, then it will be resized to the exact dimension while preserving the original image aspect ratio.

Image types `png`, `jpeg`, `bmp` and `gif` are supported.

```go
pattern, err := blinky.NewPatternFromImage("pattern.png", 60)
```

__Parsing an Arduino header__

_PatternPaint_ can export a pattern drawn with it as an Arduino C Header. You can parse them as well to create a pattern.

```go
pattern, err := blinky.NewPatternFromArduinoExport("pattern.h")
```

### Animation

An `Animation` is the composition of a `Pattern` and a set of parameters to define how it should be played, and how many times.

You can create an animation instance literally, or import it from a file.
```go
anim, err := NewAnimationFromFile("animation.json")
if err != nil {
   // do whatever you want with the animation
}
```

```go
p, _ := blinky.NewPatternFromImage("cylon.png", 60)
anim := blinky.Animation{
   Name:    "clyon",
   Repeat:  10,
   Speed:   50
   Pattern: p,
}
```
   - `repeat` indicate how many times the pattern must be played. A negative number will run an infinite loop.
   - `speed` is a convenient and simple way to add a delay between each frame. The delay, expressed in milliseconds, is calculated as `1000 / speed`.

__Play an animation__

```go
bt, err := blinky.NewBlinkyTape("/dev/tty.usbmodem1421", 60)
// error handling
bt.Play(&anim, nil)
```

You can also provide a configuration struct to override the animation parameters. It allows you to define a specific delay to use between each frame.

```go
// Configure the animation to play the pattern indefinitely,
// with a delay of 66ms between each frame
config := blinky.AnimationConfig{
   Repeat:  -1,
   Delay:   66 * time.Millisecond,
}
bt.Play(&anim, &config)
```

Notes:

   - You can't change the state of the LED strip nor rendering while an animation is being played. This is only possible while an animation is stopped or paused.

   - If `Play()` is called while an animation is running or paused, it will stop it before launching the new one.

__Controlling an animation__

While an animation is being played on the LED strip, you can control it and retrieve its status.

```go
// Pause a running animation
bt.Pause()
// Resume a paused animation
bt.Resume()
// Stop an animation
// Can be calLED even if the animation is in a paused state
bt.Stop()

// Get the status of the animation loop
status := bt.Status()
switch status {
case blinky.StatusRunning:
   // animation is running
case blinky.StatusPaused:
   // animation is paused
case blinky.StatusStopped:
   // no animation is running
}

// similar to "bt.Status() == StatusRunning"
if bt.IsRunning() {
   fmt.Println("An animation is running on this BlinkyTape")
}
```

__Export to a file__

You can export an animation to a file if you want to reuse it later. The file will use the JSON format to represent its content.

```go
// Export
err := anim.SaveToFile("animation.json")
if err != nil {
   // animation has been saved successfully
}
```

### Examples

You can find pattern examples in the folder `patterns` of this repository.

## Share yours

If you create a nice pattern manually of with *PatternPaint* and want to share it with others, send me a mail with the pattern attached to it, and i will add it to the repository.

## License

Copyright (c) 2016, William Poussier <william.poussier@gmail.com>   
BlinkyGo is released under a MIT style license.
