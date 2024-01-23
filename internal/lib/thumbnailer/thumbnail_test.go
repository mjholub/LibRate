package thumbnailer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestThumbnail(t *testing.T) {
	dims := Dims{Width: 512, Height: 512}

	// source image 1024x1024
	thumb, err := Thumbnail(dims, "gopher.jpg")
	assert.Nil(t, err)
	bounds := thumb.Bounds()
	assert.Equal(t, 512, bounds.Dx())
	assert.Equal(t, 512, bounds.Dy())

	// source image 1600x900-
	t16x9, err := Thumbnail(dims, "paprikas.png")
	assert.Nil(t, err)
	b := t16x9.Bounds()
	assert.Equal(t, 512, b.Dx())
	assert.Equal(t, 288, b.Dy())
}

func BenchmarkThumbnail(b *testing.B) {
	dims := Dims{Width: 512, Height: 512}

	for i := 0; i < b.N; i++ {
		_, err := Thumbnail(dims, "gopher.jpg")
		if err != nil {
			b.Fatal(err)
		}
	}
}
