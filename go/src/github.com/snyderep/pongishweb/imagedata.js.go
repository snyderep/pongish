// +build js

package main

import (
	"image/color"

	"github.com/gopherjs/gopherjs/js"
	"honnef.co/go/js/dom"
)

// copied from https://github.com/Archs/js/blob/master/canvas/canvas.go

type imageData struct {
	// ref: https://github.com/gopherjs/gopherjs/blob/master/js/js.go
	*js.Object
	// ImageData.data Read only
	// Is a Uint8ClampedArray representing a one-dimensional array containing the data in the RGBA order, with integer values between 0 and 255 (included).
	Data *js.Object `js:"data"`
	// ImageData.height Read only
	// Is an unsigned long representing the actual height, in pixels, of the ImageData.
	Height int `js:"height"`
	// ImageData.width Read only
	// Is an unsigned long representing the actual width, in pixels, of the ImageData.
	Width int `js:"width"`
}

func (i *imageData) at(x, y int) *color.RGBA {
	idx := 4 * (y*i.Width + x)
	rgba := &color.RGBA{}
	rgba.R = uint8(i.Data.Index(idx).Int())
	rgba.G = uint8(i.Data.Index(idx + 1).Int())
	rgba.B = uint8(i.Data.Index(idx + 2).Int())
	rgba.A = uint8(i.Data.Index(idx + 3).Int())
	//println("at:", x, y, rgba)
	return rgba
}

func (i *imageData) anyBlue() bool {
	for x := 0; x < i.Width; x++ {
		for y := 0; y < i.Height; y++ {
			rgb := i.at(x, y)
			if rgb.B > 0 {
				return true
			}
		}
	}
	return false
}

// The CanvasRenderingContext2D.getImageData() method of the Canvas 2D API returns an ImageData object
// representing the underlying pixel data for the area of the canvas
// denoted by the rectangle which starts at (sx, sy) and has an sw width and sh height.
// x The x coordinate of the upper left corner of the rectangle from which the ImageData will be extracted.
// y The y coordinate of the upper left corner of the rectangle from which the ImageData will be extracted.
// width The width of the rectangle from which the ImageData will be extracted.
// height The height of the rectangle from which the ImageData will be extracted.
func getImageData(ctx *dom.CanvasRenderingContext2D, x, y, width, heigth int) *imageData {
	o := ctx.Call("getImageData", x, y, width, heigth)
	return &imageData{Object: o}
}
