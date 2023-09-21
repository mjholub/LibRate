package thumbnailer

import (
	"bytes"
	"fmt"
	"image"
	"os"

	"github.com/154pinkchairs/imaging"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type Dims struct {
	Width, Height uint
}

func Thumbnail(dims Dims, inputFile string) (thumb image.Image, err error) {
	buf := bytes.NewBuffer(nil)
	err = ffmpeg.Input(inputFile).
		Filter("select", ffmpeg.Args{fmt.Sprintf("eq(n,%d)", 0)}).
		Output("pipe:", ffmpeg.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg"}).
		WithOutput(buf, os.Stdout).
		Run()
	if err != nil {
		return nil, err
	}

	thumb, _, err = image.Decode(buf)
	if err != nil {
		return nil, err
	}
	thumbNRGBA := imaging.Resize(thumb, int(dims.Width), int(dims.Height), imaging.Lanczos)
	centreX := float64(dims.Width) * 0.9
	centreY := float64(dims.Height) * 0.9
	thumb = imaging.CropCenter(thumbNRGBA, int(centreX), int(centreY))

	return thumb, err
}
