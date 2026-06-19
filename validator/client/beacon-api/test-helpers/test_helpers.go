package test_helpers

import (
	"github.com/theQRL/go-qrl/common/hexutil"
	fieldparams "github.com/theQRL/qrysm/config/fieldparams"
)

func FillByteSlice(sliceLength int, value byte) []byte {
	bytes := make([]byte, sliceLength)

	for index := range bytes {
		bytes[index] = value
	}

	return bytes
}

func FillByteArraySlice(sliceLength int, value []byte) [][]byte {
	bytes := make([][]byte, sliceLength)

	for index := range bytes {
		bytes[index] = value
	}

	return bytes
}

func FillEncodedByteSlice(sliceLength int, value byte) string {
	return hexutil.Encode(FillByteSlice(sliceLength, value))
}

func FillPubkey(value byte) []byte {
	return FillByteSlice(fieldparams.MLDSA87PubkeyLength, value)
}

func FillEncodedPubkey(value byte) string {
	return FillEncodedByteSlice(fieldparams.MLDSA87PubkeyLength, value)
}

func FillSignature(value byte) []byte {
	return FillByteSlice(fieldparams.MLDSA87SignatureLength, value)
}

func FillEncodedSignature(value byte) string {
	return FillEncodedByteSlice(fieldparams.MLDSA87SignatureLength, value)
}

func FillEncodedByteArraySlice(sliceLength int, value string) []string {
	encodedBytes := make([]string, sliceLength)
	for index := range encodedBytes {
		encodedBytes[index] = value
	}
	return encodedBytes
}
