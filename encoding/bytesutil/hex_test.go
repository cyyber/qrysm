package bytesutil_test

import (
	"testing"

	"github.com/theQRL/qrysm/encoding/bytesutil"
	"github.com/theQRL/qrysm/testing/assert"
)

func TestIsHex(t *testing.T) {
	tests := []struct {
		a []byte
		b bool
	}{
		{nil, false},
		{[]byte(""), false},
		{[]byte("0x"), false},
		{[]byte("0x0"), true},
		{[]byte("foo"), false},
		{[]byte("1234567890abcDEF"), false},
		{[]byte("XYZ4567890abcDEF1234567890abcDEF1234567890abcDEF1234567890abcDEF"), false},
		{[]byte("0x1234567890abcDEF1234567890abcDEF1234567890abcDEF1234567890abcDEF"), true},
		{[]byte("1234567890abcDEF1234567890abcDEF1234567890abcDEF1234567890abcDEF"), false},
	}
	for _, tt := range tests {
		isHex := bytesutil.IsHex(tt.a)
		assert.Equal(t, tt.b, isHex)
	}
}
