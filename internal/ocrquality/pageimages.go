package ocrquality

import (
	"path/filepath"
	"regexp"
	"strconv"
)

var pageImageNumberRE = regexp.MustCompile(`(\d+)`)

// indexPageImages maps page numbers to image paths by globbing dir and
// parsing the last run of digits in each filename — the same inference the
// discovery step uses — so any zero-padding width (page_001.png,
// page_0001.png, scan-12.png) resolves correctly.
func indexPageImages(dir string, glob string) map[int]string {
	if glob == "" {
		glob = "page_*.png"
	}
	matches, err := filepath.Glob(filepath.Join(dir, glob))
	if err != nil {
		return nil
	}
	index := make(map[int]string, len(matches))
	for _, match := range matches {
		nums := pageImageNumberRE.FindAllString(filepath.Base(match), -1)
		if len(nums) == 0 {
			continue
		}
		n, err := strconv.Atoi(nums[len(nums)-1])
		if err != nil || n <= 0 {
			continue
		}
		if _, exists := index[n]; !exists {
			index[n] = match
		}
	}
	return index
}
