package confcore

import (
	"bytes"
	"compress/gzip"
	"io"
)

// confcoreXML returns raw, uncompressed file data.
func confcoreXML() []byte {
	gz, err := gzip.NewReader(bytes.NewBuffer([]byte{
		0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x00, 0xff, 0xbc, 0x55,
		0xc1, 0x8e, 0xda, 0x30, 0x10, 0xbd, 0xef, 0x57, 0xb8, 0xbe, 0x56, 0x21,
		0xa5, 0x5c, 0xf6, 0x90, 0x64, 0x55, 0xad, 0xb4, 0xa8, 0x6a, 0x55, 0x55,
		0x85, 0xb6, 0x47, 0x64, 0xec, 0x81, 0xb8, 0x18, 0x3b, 0xb5, 0x27, 0x0b,
		0xfc, 0x7d, 0x9d, 0x38, 0x40, 0x80, 0xc0, 0x42, 0x57, 0xed, 0x05, 0x64,
		0xcf, 0xcc, 0x9b, 0x99, 0x37, 0xcf, 0x93, 0xe4, 0x61, 0xbd, 0x54, 0xe4,
		0x19, 0xac, 0x93, 0x46, 0xa7, 0xb4, 0xdf, 0x7b, 0x47, 0x09, 0x68, 0x6e,
		0x84, 0xd4, 0xf3, 0x94, 0x7e, 0x1f, 0x3f, 0x45, 0xf7, 0xf4, 0x21, 0xbb,
		0x4b, 0xde, 0x44, 0x11, 0x19, 0x82, 0x06, 0xcb, 0x10, 0x04, 0x59, 0x49,
		0xcc, 0xc9, 0x5c, 0x31, 0x01, 0x64, 0xd0, 0xeb, 0xdf, 0xf7, 0x06, 0x24,
		0x8a, 0xbc, 0x93, 0xd4, 0x08, 0x76, 0xc6, 0x38, 0x64, 0x77, 0x84, 0x24,
		0x16, 0x7e, 0x97, 0xd2, 0x82, 0x23, 0x4a, 0x4e, 0x53, 0x3a, 0xc7, 0xc5,
		0x5b, 0xba, 0x4f, 0x34, 0xf0, 0x89, 0xe2, 0xda, 0xcd, 0x4c, 0x7f, 0x01,
		0x47, 0xc2, 0x15, 0x73, 0x2e, 0xa5, 0x43, 0x5c, 0x7c, 0x96, 0x0e, 0x47,
		0x68, 0x2c, 0x50, 0x22, 0x45, 0x4a, 0x97, 0x46, 0x80, 0xa2, 0x95, 0xab,
		0x77, 0xe6, 0x46, 0x95, 0x4b, 0xed, 0xc2, 0xc9, 0x9f, 0xab, 0xb2, 0xc2,
		0x5d, 0xa4, 0xd9, 0x12, 0xc8, 0x27, 0xd8, 0xd4, 0xa5, 0x34, 0xe6, 0x60,
		0x22, 0xb8, 0x29, 0xc0, 0x57, 0xc0, 0x73, 0x66, 0x99, 0xb5, 0x6c, 0x13,
		0x32, 0x77, 0x02, 0x7c, 0xe4, 0x46, 0x9f, 0x45, 0x18, 0x8a, 0xc5, 0x57,
		0xb9, 0x9e, 0x96, 0xb3, 0x0b, 0x00, 0x5f, 0xaa, 0x9f, 0x57, 0x95, 0x30,
		0x36, 0x46, 0xa1, 0x2c, 0x6e, 0x00, 0x49, 0xe2, 0x16, 0x2f, 0x49, 0x1c,
		0x18, 0xed, 0x26, 0x77, 0xc4, 0xad, 0x51, 0x0a, 0xc4, 0x4f, 0xa9, 0x85,
		0x59, 0x05, 0x86, 0x57, 0x52, 0xcc, 0x01, 0xb7, 0x14, 0x17, 0xd6, 0x14,
		0x60, 0x71, 0x43, 0xaa, 0x62, 0x52, 0xfa, 0x2c, 0x9d, 0x9c, 0x2a, 0xa0,
		0xd9, 0xd8, 0x96, 0x90, 0xc4, 0x5b, 0x6b, 0xb7, 0x33, 0x67, 0x7a, 0x32,
		0x33, 0xbc, 0x74, 0x34, 0x7b, 0x62, 0xca, 0xbd, 0xe8, 0xef, 0x72, 0xe6,
		0xab, 0x98, 0x54, 0x7d, 0xd1, 0x0c, 0x90, 0xe7, 0x20, 0x22, 0xa9, 0x4f,
		0xa2, 0x78, 0x2e, 0x95, 0xd8, 0x91, 0x71, 0xd2, 0xd3, 0xd8, 0x02, 0xfc,
		0x90, 0xd0, 0x74, 0x83, 0xfe, 0x44, 0xb7, 0xce, 0x37, 0xf6, 0xf3, 0x42,
		0x4f, 0xd7, 0x86, 0xe4, 0xcc, 0xdd, 0x1a, 0xd2, 0xa8, 0xbc, 0xfe, 0xbb,
		0x2a, 0x05, 0xf8, 0xc7, 0x67, 0xdd, 0x64, 0xd7, 0x4e, 0x27, 0xdf, 0x97,
		0x22, 0xb9, 0x92, 0x7c, 0xc1, 0x6e, 0x8b, 0x75, 0xc0, 0x2c, 0xcf, 0x27,
		0x41, 0x6d, 0x34, 0x7b, 0x7f, 0x4d, 0x0c, 0x06, 0x39, 0xef, 0x82, 0x06,
		0x9d, 0x41, 0xf5, 0x88, 0x49, 0xbd, 0x40, 0x34, 0x53, 0x51, 0x7d, 0xac,
		0xf2, 0x29, 0x3f, 0x6a, 0xbf, 0x30, 0x5a, 0x03, 0x3d, 0xa7, 0x80, 0xd1,
		0xce, 0xb7, 0x96, 0xc1, 0x3e, 0x34, 0x6e, 0xa5, 0x89, 0x0f, 0xa4, 0x74,
		0x2c, 0xad, 0xcb, 0xf2, 0x7a, 0x0c, 0x0d, 0xd4, 0xe8, 0xa1, 0x99, 0x6a,
		0x59, 0x1c, 0x54, 0xd6, 0x05, 0xd8, 0x0d, 0xfa, 0x08, 0x4a, 0x7d, 0x03,
		0xed, 0x27, 0x01, 0xb6, 0xd9, 0x29, 0x01, 0xd8, 0xdf, 0xd7, 0xb0, 0xf1,
		0x09, 0x06, 0x43, 0xb4, 0x72, 0x5a, 0x22, 0xb8, 0x63, 0x53, 0xdb, 0xd8,
		0xb0, 0x5e, 0x04, 0xcc, 0xac, 0x9f, 0xc4, 0x3b, 0xd3, 0x09, 0x62, 0x7c,
		0x0e, 0xf2, 0x84, 0xa7, 0xc3, 0xb5, 0xf2, 0x4f, 0xc8, 0xac, 0x16, 0xe7,
		0x31, 0x99, 0xc7, 0x52, 0x92, 0xe8, 0xe5, 0x4a, 0xd0, 0x32, 0xed, 0x14,
		0xc3, 0x4a, 0xbc, 0x29, 0xdd, 0x80, 0x7f, 0x66, 0x1f, 0x8a, 0x42, 0x01,
		0x76, 0x29, 0xab, 0x0b, 0xc7, 0x19, 0x8b, 0x8d, 0x1e, 0x27, 0x52, 0xd4,
		0x24, 0x9d, 0x09, 0xfc, 0x8b, 0x69, 0x8e, 0x61, 0x8d, 0xfb, 0x59, 0xda,
		0xe6, 0x16, 0xfd, 0x6d, 0xff, 0xb5, 0x43, 0x5d, 0x32, 0xbb, 0x28, 0x8b,
		0xfa, 0xdd, 0xfd, 0xa7, 0xa1, 0x1e, 0x3a, 0xb4, 0x8c, 0x7b, 0x43, 0x12,
		0xb7, 0x3e, 0xfb, 0x7f, 0x02, 0x00, 0x00, 0xff, 0xff, 0xf4, 0xe0, 0xc0,
		0xbd, 0x4f, 0x08, 0x00, 0x00,
	}))

	if err != nil {
		panic("Decompression failed: " + err.Error())
	}

	var b bytes.Buffer
	io.Copy(&b, gz)
	gz.Close()

	return b.Bytes()
}
