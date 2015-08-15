package handbook

import (
	"bytes"
	"compress/gzip"
	"io"
)

// handbookXML returns raw, uncompressed file data.
func handbookXML() []byte {
	gz, err := gzip.NewReader(bytes.NewBuffer([]byte{
		0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x00, 0xff, 0xd4, 0x58,
		0x4f, 0x53, 0xdb, 0x3e, 0x10, 0xbd, 0xf3, 0x29, 0xf4, 0xd3, 0xf5, 0x37,
		0x21, 0x30, 0xf4, 0xc0, 0xc1, 0x31, 0x33, 0x1c, 0xa0, 0x9d, 0xe9, 0x91,
		0x9e, 0x3d, 0xb2, 0xbd, 0xb1, 0x97, 0xc8, 0x92, 0x2b, 0xad, 0x93, 0xf0,
		0xed, 0xab, 0xc4, 0xa4, 0x24, 0xb5, 0x1c, 0xc7, 0x0a, 0x1d, 0xe8, 0x0d,
		0xec, 0x7d, 0x59, 0xbd, 0xb7, 0x7f, 0xe5, 0xe8, 0x6e, 0x5d, 0x49, 0xb6,
		0x04, 0x63, 0x51, 0xab, 0x19, 0xbf, 0xbe, 0xbc, 0xe2, 0x0c, 0x54, 0xa6,
		0x73, 0x54, 0xc5, 0x8c, 0xff, 0x78, 0x7a, 0x98, 0xdc, 0xf2, 0xbb, 0xf8,
		0x22, 0xfa, 0x6f, 0x32, 0x61, 0x8f, 0xa0, 0xc0, 0x08, 0x82, 0x9c, 0xad,
		0x90, 0x4a, 0x56, 0x48, 0x91, 0x03, 0xbb, 0xb9, 0xbc, 0xbe, 0xbd, 0xbc,
		0x61, 0x93, 0x89, 0x33, 0x42, 0x45, 0x60, 0xe6, 0x22, 0x83, 0xf8, 0x82,
		0xb1, 0xc8, 0xc0, 0xcf, 0x06, 0x0d, 0x58, 0x26, 0x31, 0x9d, 0xf1, 0x82,
		0x16, 0xff, 0xf3, 0x37, 0x47, 0x0e, 0x76, 0xc5, 0xa7, 0x5b, 0x3b, 0x9d,
		0x3e, 0x43, 0x46, 0x2c, 0x93, 0xc2, 0xda, 0x19, 0x7f, 0xa4, 0xc5, 0x83,
		0x11, 0x15, 0x70, 0x86, 0xf9, 0x8c, 0x97, 0x42, 0xe5, 0xa9, 0xd6, 0x0b,
		0xbe, 0xb1, 0x74, 0xb6, 0xb5, 0xd1, 0x35, 0x18, 0x7a, 0x61, 0xca, 0x99,
		0xcc, 0xf8, 0x12, 0x2d, 0xa6, 0x12, 0x78, 0xfc, 0x64, 0x1a, 0x88, 0xa6,
		0xbb, 0xb7, 0x7e, 0xe3, 0x4c, 0xa8, 0x64, 0xae, 0xb3, 0xc6, 0xf2, 0xf8,
		0x41, 0x48, 0x3b, 0x68, 0x9f, 0x6a, 0x93, 0x83, 0x49, 0x56, 0x98, 0x53,
		0xc9, 0xe3, 0x2f, 0x43, 0xe6, 0x52, 0xa4, 0x20, 0x93, 0xb5, 0x90, 0x58,
		0x28, 0x1e, 0x5f, 0x0d, 0x99, 0xdb, 0x52, 0xe4, 0x7a, 0x95, 0xd0, 0x4b,
		0xed, 0x8e, 0xaf, 0x1b, 0xea, 0xd8, 0x67, 0x25, 0xca, 0xbc, 0xfd, 0xdb,
		0x27, 0xd2, 0xbd, 0x5e, 0xb7, 0x12, 0xcd, 0x37, 0x6a, 0x7d, 0x4d, 0xdd,
		0xbf, 0x3b, 0xe3, 0x91, 0x3a, 0x85, 0x68, 0xe5, 0x65, 0x54, 0x8b, 0xcc,
		0x65, 0x8d, 0x47, 0xaa, 0x2e, 0x9f, 0x01, 0x4e, 0x04, 0x6b, 0xba, 0x3f,
		0x60, 0x14, 0xc4, 0x2a, 0x94, 0x99, 0xd7, 0xd9, 0x6b, 0x60, 0x2d, 0x09,
		0xd3, 0x09, 0x56, 0x1f, 0x68, 0x20, 0x85, 0xfa, 0x60, 0xda, 0x20, 0x28,
		0x12, 0xe4, 0x2a, 0x85, 0xc7, 0xae, 0x64, 0x08, 0x33, 0x21, 0x7b, 0xc1,
		0x1d, 0x61, 0xfd, 0xe2, 0x7e, 0xdf, 0xe4, 0x67, 0x2b, 0xaf, 0x68, 0xa8,
		0xd4, 0x86, 0xff, 0x89, 0x09, 0x54, 0xf8, 0x1c, 0x95, 0x7d, 0x58, 0xa5,
		0x13, 0x5b, 0xba, 0xd2, 0x10, 0x52, 0x8e, 0x77, 0x5c, 0x9e, 0x12, 0x26,
		0x1f, 0xb0, 0xb1, 0x90, 0x54, 0xc2, 0x2c, 0x9a, 0x7a, 0xbc, 0xd3, 0xe7,
		0xc6, 0x12, 0xce, 0x5f, 0x78, 0x3c, 0x47, 0xd9, 0x1b, 0x27, 0x1f, 0xd0,
		0xba, 0x7a, 0x91, 0x90, 0x48, 0x54, 0xce, 0xb9, 0xce, 0x07, 0x85, 0x8e,
		0xa6, 0x6d, 0x58, 0x3b, 0xcf, 0x5d, 0xe9, 0x2d, 0xdc, 0x6f, 0x0d, 0x7b,
		0x84, 0x75, 0xed, 0x5a, 0x6a, 0x40, 0x54, 0x36, 0xd4, 0xc6, 0x2b, 0x53,
		0x8b, 0x3c, 0xef, 0xef, 0x09, 0xbd, 0x28, 0x6d, 0xb1, 0x4d, 0xfe, 0x4e,
		0x1b, 0xdd, 0xd7, 0xc2, 0x4b, 0x3a, 0x9a, 0x7a, 0xea, 0x21, 0xa0, 0x46,
		0x72, 0xb0, 0x99, 0xc1, 0xba, 0x3d, 0xc6, 0xe7, 0x2c, 0x94, 0xe0, 0x5c,
		0x3f, 0xa9, 0x97, 0x79, 0x3d, 0xee, 0xd2, 0x67, 0x2c, 0xcd, 0x65, 0x70,
		0xde, 0x7d, 0x48, 0x5d, 0xae, 0x8c, 0x08, 0x70, 0x67, 0x41, 0xba, 0x74,
		0x12, 0x27, 0xe4, 0xc1, 0xfb, 0xd5, 0xf1, 0xd8, 0x33, 0x06, 0x96, 0xf1,
		0xef, 0x82, 0xbc, 0x7e, 0x8f, 0x82, 0xf4, 0xd1, 0xf7, 0x53, 0x0f, 0xa1,
		0x3d, 0x9e, 0xf2, 0xe8, 0xfe, 0xe3, 0xa1, 0xda, 0xa1, 0x39, 0x6a, 0xe1,
		0xc1, 0xaa, 0x20, 0x4d, 0xba, 0x4e, 0x3f, 0xd7, 0xd2, 0xf3, 0x57, 0x16,
		0x91, 0xbd, 0xf5, 0xbe, 0x36, 0xb0, 0x44, 0x58, 0xb5, 0x4f, 0x3e, 0x69,
		0x97, 0x75, 0xcd, 0xa7, 0x40, 0x95, 0xc0, 0x26, 0xed, 0x46, 0x4d, 0xb1,
		0x81, 0x0b, 0xc1, 0x31, 0xe8, 0xc1, 0xe5, 0x00, 0xd5, 0x51, 0xa8, 0x57,
		0x75, 0xbf, 0xf2, 0xdf, 0x2a, 0x51, 0x1c, 0x2a, 0xdf, 0x3e, 0xf1, 0xa1,
		0xcf, 0x50, 0xff, 0xdc, 0x08, 0x1c, 0x89, 0xc2, 0x76, 0x62, 0xb9, 0x2e,
		0x74, 0x54, 0xcd, 0x23, 0xf8, 0x6d, 0x14, 0x83, 0xd1, 0xae, 0x40, 0xcf,
		0x40, 0xa7, 0x9a, 0x48, 0x57, 0xc3, 0x3f, 0xd0, 0x37, 0x1c, 0x7a, 0x76,
		0x9b, 0xed, 0x8b, 0xed, 0x73, 0xb6, 0xc9, 0x97, 0x5d, 0xde, 0x21, 0x41,
		0xe5, 0x0d, 0x6c, 0x54, 0x4b, 0x77, 0x43, 0x2f, 0xb5, 0x74, 0x37, 0x93,
		0xe9, 0xc9, 0x3e, 0xfe, 0xbd, 0xcd, 0xf3, 0x7d, 0x77, 0xc8, 0xf3, 0x47,
		0xd6, 0xa8, 0xc6, 0x1b, 0x30, 0xb3, 0xdc, 0x59, 0x5e, 0x1b, 0x86, 0x4b,
		0xf2, 0xf1, 0x93, 0xae, 0x67, 0xb0, 0x0f, 0x4f, 0xba, 0x43, 0x65, 0x0e,
		0x5e, 0x76, 0xd3, 0x92, 0xf7, 0x7f, 0xd0, 0xd8, 0xdb, 0xbd, 0x09, 0x49,
		0xc2, 0x47, 0x7f, 0xce, 0x18, 0x5e, 0x3c, 0x7b, 0x99, 0xbf, 0xbd, 0x88,
		0xa6, 0x7b, 0xdf, 0xc4, 0x7e, 0x05, 0x00, 0x00, 0xff, 0xff, 0xb2, 0xdd,
		0x6c, 0x7c, 0x6c, 0x13, 0x00, 0x00,
	}))

	if err != nil {
		panic("Decompression failed: " + err.Error())
	}

	var b bytes.Buffer
	io.Copy(&b, gz)
	gz.Close()

	return b.Bytes()
}