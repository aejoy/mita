#include <avif/avif.h>
#include <stdio.h>
#include <string.h>

int decode(
    uint8_t *avif_in,
    int avif_in_size,
    int config_only,
    int decode_all,
    uint32_t *width, uint32_t *height,
    uint32_t *depth,
	int maxThreads,
    uint32_t *count,
    uint8_t *delay,
    uint8_t *out
) {
    avifDecoder *decoder = avifDecoderCreate();
    decoder->ignoreExif = 1;
    decoder->ignoreXMP = 1;
    decoder->maxThreads = maxThreads;
    decoder->strictFlags = 0;

    avifResult result = avifDecoderSetIOMemory(decoder, avif_in, avif_in_size);
    if(result != AVIF_RESULT_OK) {
        avifDecoderDestroy(decoder);
        return 0;
    }

    result = avifDecoderParse(decoder);
    if(result != AVIF_RESULT_OK) {
        avifDecoderDestroy(decoder);
        return 0;
    }

    *width = (uint32_t)decoder->image->width;
    *height = (uint32_t)decoder->image->height;
    *depth = (uint32_t)decoder->image->depth;
    *count = (uint32_t)decoder->imageCount;

    if(config_only) {
        avifDecoderDestroy(decoder);
        return 1;
    }

    avifRGBImage rgb;
    avifRGBImageSetDefaults(&rgb, decoder->image);

    rgb.maxThreads = maxThreads;
    rgb.alphaPremultiplied = 1;

    if(decoder->image->depth > 8) {
        rgb.depth = 16;
    }

    if(decoder->imageCount > 1 && decode_all) {
        rgb.chromaUpsampling = AVIF_CHROMA_UPSAMPLING_BEST_QUALITY;
    }

    while(avifDecoderNextImage(decoder) == AVIF_RESULT_OK) {
        result = avifRGBImageAllocatePixels(&rgb);
        if(result != AVIF_RESULT_OK) {
            avifDecoderDestroy(decoder);
            return 0;
        }

        result = avifImageYUVToRGB(decoder->image, &rgb);
        if(result != AVIF_RESULT_OK) {
            avifRGBImageFreePixels(&rgb);
            avifDecoderDestroy(decoder);
            return 0;
        }

        int buf_size = rgb.rowBytes * rgb.height;
        memcpy(out + buf_size*decoder->imageIndex, rgb.pixels, buf_size);

        memcpy(delay + sizeof(double)*decoder->imageIndex, &decoder->imageTiming.duration, sizeof(double));

        avifRGBImageFreePixels(&rgb);

        if(!decode_all) {
            avifDecoderDestroy(decoder);
            return 1;
        }
    }

    avifDecoderDestroy(decoder);
    return 1;
}