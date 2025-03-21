#include <avif/avif.h>
#include <stdio.h>
#include <string.h>

uint8_t* encode(
	uint8_t *rgb_in,
	int width, int height,
	int maxThreads,
	int alphaPremultiplied, int depth, avifPixelFormat chroma,
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