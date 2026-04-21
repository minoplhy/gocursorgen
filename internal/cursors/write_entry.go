package cursors

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	imagedecode "gocursorgen/internal/image_decode"
	libxcursor "gocursorgen/internal/libXCursor"
)

// WriteEntry converts a theme.CursorEntry directly into a X11 cursor file.
func (entry *CursorEntry) WriteEntry(filename string, prefix string) error {
	list, count, err := entry.entryToFileList(prefix)
	if err != nil {
		return err
	}
	if !entry.Options.RetainFrames {
		defer cleanupFileList(list)
	}

	var prefixPtr *string
	if prefix != "" {
		prefixPtr = &prefix
	}

	re, err := list.CreateCursors(count, filename, prefixPtr)
	if err != nil {
		return fmt.Errorf("xcursorgen: failed writing cursor %q to %q", entry.Name, filename)
	}

	var fp *os.File
	if filename != "-" {
		var err error
		fp, err = os.Create(filename)
		if err != nil {
			return fmt.Errorf("xcursorgen: cannot open output file %s: %v\n", filename, err)
		}
		defer fp.Close()
	} else {
		fp = os.Stdout
	}

	if !libxcursor.XcursorFileSaveImages(fp, re) {
		return fmt.Errorf("xcursorgen: cannot write cursor file %s: %v\n", filename, err)
	}

	return nil
}

func (entry *CursorEntry) entryToFileList(prefix string) (CursorEntities, int, error) {
	files, err := entry.ResolveFiles(prefix)
	if err != nil {
		return nil, 0, err
	}
	if len(files) == 0 {
		return nil, 0, fmt.Errorf("xcursorgen: cursor %q has no files", entry.Name)
	}

	type decodedFile struct {
		path   string
		frames []image.Image
	}
	decoded := make([]decodedFile, len(files))
	totalFrames := 0
	for i, path := range files {
		frames, err := imagedecode.Decode(path)
		if err != nil {
			return nil, 0, fmt.Errorf("xcursorgen: cursor %q file %q: decode error: %w", entry.Name, path, err)
		}
		decoded[i] = decodedFile{path: path, frames: frames}
		totalFrames += len(frames)
	}

	// Flatten all frames across all files
	type rawFrame struct {
		img      image.Image
		path     string
		fileIdx  int
		frameIdx int
	}
	var allFrames []rawFrame
	for i, df := range decoded {
		for frameIdx, frame := range df.frames {
			allFrames = append(allFrames, rawFrame{
				img:      frame,
				path:     df.path,
				fileIdx:  i,
				frameIdx: frameIdx,
			})
		}
	}
	animated := len(allFrames) > 1

	sizes := entry.Sizes
	if len(sizes) == 0 {
		sizes = []uint32{uint32(allFrames[0].img.Bounds().Dx())}
	}

	var Cursors []CursorsEntity
	count := 0

	// Outer loop: size (largest first - already sorted desc by parse time)
	// Inner loop: frame
	// This groups all animation frames for a given size contiguously in the TOC
	for _, size := range sizes {
		for _, rf := range allFrames {
			hs, err := entry.resolveHotSpot(rf.fileIdx, rf.path)
			if err != nil {
				return nil, 0, fmt.Errorf("xcursorgen: cursor %q file %q: %w", entry.Name, rf.path, err)
			}

			resized, scaledX, scaledY := imagedecode.Resize(rf.img, size, hs.X, hs.Y)

			framePath, err := writeFrame(resized, rf.path, rf.frameIdx, size, entry.Options)
			if err != nil {
				return nil, 0, fmt.Errorf("xcursorgen: cursor %q frame %d size %d: %w",
					entry.Name, rf.frameIdx, size, err)
			}

			Cursors = append(Cursors, CursorsEntity{
				Size:    size,
				XHot:    scaledX,
				YHot:    scaledY,
				PNGFile: framePath,
				Delay:   delayForEntry(animated),
			})
			count++
		}
	}

	return Cursors, count, nil
}

// writeTempFrame encodes a single frame as a PNG into a temp file.
// Caller is responsible for cleanup via cleanupFileList.
func writeFrame(img image.Image, sourcePath string, frameIdx int, size uint32, opts Options) (string, error) {
	var path string

	if opts.RetainFrames {
		srcDir := filepath.Join(opts.ThemeDir, "src")
		if err := os.MkdirAll(srcDir, 0755); err != nil {
			return "", fmt.Errorf("cannot create src dir: %w", err)
		}
		ext := filepath.Ext(sourcePath)
		base := strings.TrimSuffix(filepath.Base(sourcePath), ext)
		path = filepath.Join(srcDir, fmt.Sprintf("%s_frame%d_%d.png", base, frameIdx, size))
	} else {
		f, err := os.CreateTemp("", "xcursorgen-frame-*.png")
		if err != nil {
			return "", err
		}
		f.Close()
		path = f.Name()
	}

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		if !opts.RetainFrames {
			os.Remove(path)
		}
		return "", err
	}
	return path, nil
}

// cleanupFileList removes any temp files written by writeTempFrame.
// Call this after WriteCursors returns.
func cleanupFileList(list []CursorsEntity) {
	for _, k := range list {
		if strings.HasPrefix(filepath.Base(k.PNGFile), "xcursorgen-frame-") {
			os.Remove(k.PNGFile)
		}
	}
}

// imageSize returns the nominal cursor size from the image width.
// Uses DecodeConfig to avoid loading full pixel data.
func imageSize(path string) (uint32, error) {
	cfg, err := imagedecode.DecodeConfig(path)
	if err != nil {
		return 0, err
	}
	if cfg.Width <= 0 {
		return 0, fmt.Errorf("image %q has zero width", path)
	}
	return uint32(cfg.Width), nil
}

// delayForEntry returns the default animation delay when the cursor is animated,
// 0 for static cursors.
func delayForEntry(animated bool) uint32 {
	if animated {
		return 50
	}
	return 0
}
