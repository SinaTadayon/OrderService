package frame

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFrameHeader(t *testing.T) {
	frame := Factory().
		SetDefaultHeader(HeaderOrderId, 1234456).
		SetBody(11111111).Build()

	frame2 := Factory().
		SetDefaultHeader(HeaderItemId, 9999999).
		SetBody(222222222).Build()

	frame.Header().CopyFrom(frame2.Header())
	require.True(t, frame.Header().KeyExists(string(HeaderItemId)))
	require.Equal(t, 9999999, frame.Header().Value(string(HeaderItemId)))
	require.Equal(t, 11111111, frame.Body().Content())
}
