package libxcursor

// https://gitlab.freedesktop.org/xorg/lib/libxcursor/-/blob/master/include/X11/Xcursor/Xcursor.h.in
/*
 * Each cursor image occupies a separate image chunk.
 * The length of the image header follows the chunk header
 * so that future versions can extend the header without
 * breaking older applications
 *
 *  Image:
 *	ChunkHeader	header	chunk header
 *	CARD32		width	actual width
 *	CARD32		height	actual height
 *	CARD32		xhot	hot spot x
 *	CARD32		yhot	hot spot y
 *	CARD32		delay	animation delay
 *	LISTofCARD32	pixels	ARGB pixels
 */
type XcursorImage struct {
	Width  uint32
	Height uint32
	Size   uint32
	XHot   uint32
	YHot   uint32
	Delay  uint32
	Pixels []uint32 // ARGB (premultiplied)
}

type XcursorImages struct {
	NImage int            // number of images
	Images []XcursorImage // array of XcursorImage pointers
	Name   string         // name used to load images
}
