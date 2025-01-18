package mita

import (
	"os"
	"image"
	_ "image/png"
	"bytes"
	"testing"
)

func TestEncode(t *testing.T) {
	t.Run("toAVIF", func (t *testing.T) {
		src, err := os.ReadFile("./testdata/src/pikachu.png")
		if err != nil {
			panic(err)
		}

		img, _, err := image.Decode(bytes.NewReader(src))
		if err != nil {
			panic(err)
		}

		dst, err := os.Create("./testdata/dst/pikachu.avif")
		if err != nil {
			panic(err)
		}
		defer dst.Close()

		if _, err := dst.Write(Encode(img, Options{
			Quality: 75,
			Speed: 10,
		})); err != nil {
			panic(err)
		}
	})
}