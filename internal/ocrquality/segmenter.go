package ocrquality

import "image"

// SegmentRequest carries everything a figure segmenter may need to locate the
// figure content on a page image. Image is the decoded page; ImagePath is the
// source file for segmenters that run out of process and want to open the
// image themselves.
type SegmentRequest struct {
	Image       image.Image
	ImagePath   string
	PageNumber  int
	FigureIndex int
	Description string
	Margin      int
}

// FigureSegmenter computes the crop rectangle for one figure on a page image.
// The returned method string is recorded in the figure sidecar for
// provenance (e.g. "ink-band-v1").
type FigureSegmenter interface {
	SegmentFigure(req SegmentRequest) (image.Rectangle, string, error)
}

// InkBandSegmenter is the built-in pixel-heuristic segmenter: it unions
// horizontal ink bands, drops header/footer furniture zones, and pads the
// result by the requested margin.
type InkBandSegmenter struct{}

func (InkBandSegmenter) SegmentFigure(req SegmentRequest) (image.Rectangle, string, error) {
	_, rect := cropNonWhite(req.Image, req.Margin)
	return rect, "ink-band-v1", nil
}
