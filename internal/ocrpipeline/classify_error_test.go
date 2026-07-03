package ocrpipeline

import (
	"errors"
	"testing"

	"github.com/go-go-golems/scraper/pkg/workflow"
	"github.com/stretchr/testify/require"
)

type fakeHintedError struct{ retryable bool }

func (f fakeHintedError) Error() string       { return "plugin call failed" }
func (f fakeHintedError) RetryableHint() bool { return f.retryable }

// A plugin's explicit retryability verdict must beat the string-based
// classification in both directions.
func TestClassifyStructuredPageErrorHonorsRetryHint(t *testing.T) {
	var werr *workflow.Error

	err := classifyStructuredPageError(fakeHintedError{retryable: true})
	require.True(t, errors.As(err, &werr))
	require.True(t, werr.Retryable)

	err = classifyStructuredPageError(fakeHintedError{retryable: false})
	require.True(t, errors.As(err, &werr))
	require.False(t, werr.Retryable)

	// Unhinted errors keep the existing string classification: a page
	// mismatch is permanent, an unknown error retryable.
	err = classifyStructuredPageError(errors.New("structured OCR page mismatch: input page 005 response page 006"))
	require.True(t, errors.As(err, &werr))
	require.False(t, werr.Retryable)

	err = classifyStructuredPageError(errors.New("some transient provider hiccup"))
	require.True(t, errors.As(err, &werr))
	require.True(t, werr.Retryable)
}
