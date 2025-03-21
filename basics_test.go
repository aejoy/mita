package mita

import (
	"bytes"
	"image"
	"image/png"
	_ "image/png"
	"os"
	"testing"
)

func TestBasics(t *testing.T) {
	t.Run("encode", func(t *testing.T) {
		src, err := os.ReadFile("./testdata/src/go.png")
		if err != nil {
			t.Error(err)
		}

		img, _, err := image.Decode(bytes.NewReader(src))
		if err != nil {
			t.Error(err)
		}

		dst, err := os.Create("./testdata/dst/go.avif")
		if err != nil {
			t.Error(err)
		}
		defer dst.Close()

		if _, err := dst.Write(Encode(img, Options{
			Quality: 75,
			Speed:   10,
		})); err != nil {
			t.Error(err)
		}
	})

	t.Run("decode", func(t *testing.T) {
		src, err := os.ReadFile("./testdata/dst/go.avif")
		if err != nil {
			t.Error(err)
		}

		img, err := Decode(bytes.NewReader(src))
		if err != nil {
			panic(err)
		}

		dst, err := os.Create("./testdata/dst/go.png")
		if err != nil {
			t.Error(err)
		}

		if err := png.Encode(dst, img); err != nil {
			t.Error(err)
		}
	})
}
