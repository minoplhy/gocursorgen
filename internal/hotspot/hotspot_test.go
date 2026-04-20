package hotspot

import (
	"image"
	"image/color"
	"testing"
)

func TestFindHotSpot_FirstOpaquePixel(t *testing.T) {
	tests := []struct {
		name  string
		setup func() image.Image
		wantX uint32
		wantY uint32
	}{
		{
			name: "top-left corner is opaque",
			setup: func() image.Image {
				img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
				img.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
				return img
			},
			wantX: 0, wantY: 0,
		},
		{
			name: "opaque pixel on second row third column",
			setup: func() image.Image {
				img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
				img.SetNRGBA(2, 1, color.NRGBA{R: 255, A: 255})
				return img
			},
			wantX: 2, wantY: 1,
		},
		{
			name: "row beats column priority",
			setup: func() image.Image {
				img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
				img.SetNRGBA(3, 2, color.NRGBA{R: 255, A: 255})
				img.SetNRGBA(0, 1, color.NRGBA{G: 255, A: 255}) // wins
				return img
			},
			wantX: 0, wantY: 1,
		},
		{
			name: "fully transparent falls back to origin",
			setup: func() image.Image {
				return image.NewNRGBA(image.Rect(0, 0, 4, 4))
			},
			wantX: 0, wantY: 0,
		},
		{
			name: "non-zero image origin is normalised",
			setup: func() image.Image {
				img := image.NewNRGBA(image.Rect(10, 10, 14, 14))
				img.SetNRGBA(12, 11, color.NRGBA{B: 255, A: 255})
				return img
			},
			wantX: 2, wantY: 1,
		},
		{
			name: "semi-transparent pixel counts as hit",
			setup: func() image.Image {
				img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
				img.SetNRGBA(1, 2, color.NRGBA{R: 128, A: 1})
				return img
			},
			wantX: 1, wantY: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := findHotSpotInImage(tc.setup())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.X != tc.wantX || got.Y != tc.wantY {
				t.Errorf("got (%d,%d), want (%d,%d)", got.X, got.Y, tc.wantX, tc.wantY)
			}
		})
	}
}

func TestFindHotSpot_EmptyImage(t *testing.T) {
	_, err := findHotSpotInImage(image.NewNRGBA(image.Rect(0, 0, 0, 0)))
	if err == nil {
		t.Fatal("expected error for empty image, got nil")
	}
}
