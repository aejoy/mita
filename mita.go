package mita

import (
	"image"
	_ "image/png"
	_ "image/jpeg"
	"unsafe"
)

/*
#cgo LDFLAGS: -lavif

#include <avif/avif.h>
#include <stdio.h>
#include <string.h>

uint8_t* encode(
	uint8_t *rgb_in,
	int width, int height,
	int maxThreads,
	int alphaPremultiplied, int depth, int chroma,
	int quality, int speed,
	size_t *size
) {
    avifResult result;

    avifImage *image = avifImageCreate(width, height, depth, chroma);
    if (!image) {
        fprintf(stderr, "[avifImageCreate] Out of memory\n");
        goto cleanup;
    }

    avifRGBImage rgb;
    avifRGBImageSetDefaults(&rgb, image);

    rgb.maxThreads = maxThreads;
    rgb.alphaPremultiplied = alphaPremultiplied;

    result = avifRGBImageAllocatePixels(&rgb);
    if (result != AVIF_RESULT_OK) {
        fprintf(stderr, "[avifRGBImageAllocatePixels] Out of memory\n");
        goto cleanup;
    }

    rgb.pixels = rgb_in;

    result = avifImageRGBToYUV(image, &rgb);
    if (result != AVIF_RESULT_OK) {
        fprintf(stderr, "[avifImageRGBToYUV] Out of memory\n");
        goto cleanup;
    }

    avifRWData output = AVIF_DATA_EMPTY;

    avifEncoder *encoder = avifEncoderCreate();
    if (!encoder) {
        fprintf(stderr, "[avifEncoderCreate] Out of memory\n");
        goto cleanup;
    }

    encoder->maxThreads = maxThreads;
    encoder->quality = quality;
    encoder->speed = speed;

    result = avifEncoderAddImage(encoder, image, 1, AVIF_ADD_IMAGE_FLAG_SINGLE);
    if (result != AVIF_RESULT_OK) {
        fprintf(stderr, "Failed to add image to encoder: %s\n", avifResultToString(result));
        goto cleanup;
    }

    result = avifEncoderFinish(encoder, &output);
    if (result != AVIF_RESULT_OK) {
        fprintf(stderr, "Failed to finish encode: %s\n", avifResultToString(result));
        goto cleanup;
    }

	*size = output.size;


	return output.data;
cleanup:
    if (image) {
        avifImageDestroy(image);
    }
    if (encoder) {
        avifEncoderDestroy(encoder);
    }

    avifRWDataFree(&output);
    avifRGBImageFreePixels(&rgb);
}
*/
import "C"

type Options struct {
    MaxThreads int
    AlphaPremultiplied int
    Depth int
    Chroma ChromaType
    Quality int
    Speed int
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
		C.int(opts.AlphaPremultiplied), C.int(opts.Depth), C.int(opts.Chroma),
		C.int(opts.Quality), C.int(opts.Speed),
		&size)

	return C.GoBytes(unsafe.Pointer(data), C.int(size))
}