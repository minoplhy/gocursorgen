package imagedecode

import (
	"image"

	xdraw "golang.org/x/image/draw"
)

func Resize(img image.Image, size uint32, xhot, yhot uint32) (image.Image, uint32, uint32) {
	src := img.Bounds()
	srcW := uint32(src.Dx())
	srcH := uint32(src.Dy())

	if srcW == size {
		return img, xhot, yhot
	}

	dstW := int(size)
	dstH := int(size * srcH / srcW)

	dst := image.NewNRGBA(image.Rect(0, 0, dstW, dstH))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), img, src, xdraw.Over, nil)

	scale := float64(size) / float64(srcW)
	scaledX := uint32(float64(xhot) * scale)
	scaledY := uint32(float64(yhot) * scale)

	return dst, scaledX, scaledY
}
