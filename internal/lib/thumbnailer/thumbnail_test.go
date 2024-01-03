package thumbnailer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestThumbnail(t *testing.T) {
	dims := Dims{Width: 512, Height: 512}

	thumb, err := Thumbnail(dims, "gopher.jpg")
	assert.Nil(t, err)
	bounds := thumb.Bounds()
	assert.Equal(t, 512, bounds.Dx())
	assert.Equal(t, 512, bounds.Dy())
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