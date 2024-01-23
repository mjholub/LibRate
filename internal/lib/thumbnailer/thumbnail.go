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

// Dims denote the maximum width and height of the thumbnail
// The algorithm will try to resize the image to the maximum possible size that
// preserves the aspect ratio and fits within the given dimensions.
type Dims struct {
	Width  uint `yaml:"width"`
	Height uint `yaml:"height"`
}

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

	originalWidth := float64(img.Bounds().Dx())
	originalHeight := float64(img.Bounds().Dy())
	aspectRatio := originalWidth / originalHeight
	// calculate the maximum possible width and height
	// branching on 1:1 is not worth it as benchmarking shows
	targeth := uint(float64(dims.Width) / aspectRatio)

	resized := image.NewRGBA(image.Rect(0, 0, int(dims.Width), int(targeth)))
	xdraw.CatmullRom.Scale(resized, resized.Bounds(), img, img.Bounds(), draw.Over, nil)
	return resized, nil
}

func ThumbnailPercentage(percentage float64, inputFile string) (thumb image.Image, err error) {
	f, err := os.Open(inputFile)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("error decoding image: %v", err)
	}
	resized := image.NewRGBA(
		image.Rect(
			0, 0, int(
				float64(
					img.Bounds().Dx())*percentage), int(
				float64(
					img.Bounds().Dy())*percentage),
		),
	)
	xdraw.CatmullRom.Scale(resized, resized.Bounds(), img, img.Bounds(), draw.Over, nil)
	return resized, nil
}
