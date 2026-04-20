package libxcursor

func XcursorImagesCreate(count int) XcursorImages {
	return XcursorImages{
		NImage: count,
		Images: make([]XcursorImage, count),
	}
}
