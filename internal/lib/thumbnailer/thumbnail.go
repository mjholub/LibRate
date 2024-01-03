package thumbnailer

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	_ "golang.org/x/image/webp"

	xdraw "golang.org/x/image/draw"
)

type Dims struct {
	Width, Height uint
}

// experimental implementation with image/draw, w/o ffmpeg dependency
func Thumbnail(dims Dims, inputFile string) (thumb image.Image, err error) {
	f, err := os.Open(inputFile)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("error decoding image: %v", err)
	}
	resized := image.NewRGBA(image.Rect(0, 0, int(dims.Width), int(dims.Height)))
	xdraw.CatmullRom.Scale(resized, resized.Bounds(), img, img.Bounds(), draw.Over, nil)
	return resized, nil
}
