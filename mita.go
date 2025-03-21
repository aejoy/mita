package mita

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"unsafe"
)

// #cgo LDFLAGS: -lavif
// #include "encode.h"
// #include "decode.h"
import "C"

type Options struct {
	MaxThreads         int
	AlphaPremultiplied int
	Depth              int
	Chroma             ChromaType
	Quality            int
	Speed              int
}

type ChromaType int

const (
	ChromaYUV444 ChromaType = C.AVIF_PIXEL_FORMAT_YUV444
	ChromaYUV422 ChromaType = C.AVIF_PIXEL_FORMAT_YUV422
	ChromaYUV420 ChromaType = C.AVIF_PIXEL_FORMAT_YUV420
)

func Encode(img image.Image, opts Options) []byte {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	pixels := make([]byte, width*height*4)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			i := (y*width + x) * 4
			pixels[i+0] = byte(r >> 8)
			pixels[i+1] = byte(g >> 8)
			pixels[i+2] = byte(b >> 8)
			pixels[i+3] = byte(a >> 8)
		}
	}

	var size C.size_t

	if opts.Depth < 8 {
		opts.Depth = 8
	}

	if opts.Chroma == 0 {
		opts.Chroma = ChromaYUV444
	}

	data := C.encode((*C.uint8_t)(unsafe.Pointer(&pixels[0])),
		C.int(width), C.int(height),
		C.int(opts.MaxThreads),
		C.int(opts.AlphaPremultiplied), C.int(opts.Depth), C.avifPixelFormat(opts.Chroma),
		C.int(opts.Quality), C.int(opts.Speed),
		&size)

	return C.GoBytes(unsafe.Pointer(data), C.int(size))
}

func Decode(r io.Reader) (image.Image, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, errors.New("failed to read AVIF data")
	}

	cfg, err := DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	r = bytes.NewReader(data)

	var width, height, depth, count C.uint32_t
	var delay [8]byte

	height = C.uint32_t(cfg.Height)
	width = C.uint32_t(cfg.Width)

	rgbDepth := 8
	if depth > 8 {
		rgbDepth = 16
	}

	bytesPerChannel := 1
	if rgbDepth == 16 {
		bytesPerChannel = 2
	}

	bytesPerPixel := 4 * bytesPerChannel
	bufSize := int(width) * int(height) * bytesPerPixel
	out := make([]byte, bufSize)

	result := C.decode(
		(*C.uint8_t)(unsafe.Pointer(&data[0])), C.int(len(data)),
		0, 0,
		&width, &height, &depth, 1, &count,
		(*C.uint8_t)(unsafe.Pointer(&delay[0])),
		(*C.uint8_t)(unsafe.Pointer(&out[0])),
	)
	if result == 0 {
		return nil, errors.New("failed to decode AVIF")
	}

	var img image.Image

	if rgbDepth == 8 {
		rgba := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
		copy(rgba.Pix, out)
		img = rgba
	} else {
		rgba64 := image.NewRGBA64(image.Rect(0, 0, int(width), int(height)))
		copy(rgba64.Pix, out)
		img = rgba64
	}

	return img, nil
}

func DecodeConfig(r io.Reader) (image.Config, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return image.Config{}, err
	}

	var width, height, depth, count C.uint32_t

	result := C.decode(
		(*C.uint8_t)(unsafe.Pointer(&data[0])), C.int(len(data)),
		1, 0,
		&width, &height, &depth, 1, &count,
		(*C.uint8_t)(unsafe.Pointer(nil)), (*C.uint8_t)(unsafe.Pointer(nil)),
	)
	if result == 0 {
		return image.Config{}, errors.New("failed to read AVIF config")
	}

	var cm color.Model
	if depth > 8 {
		cm = color.RGBA64Model
	} else {
		cm = color.RGBAModel
	}

	return image.Config{
		Width:      int(width),
		Height:     int(height),
		ColorModel: cm,
	}, nil
}
