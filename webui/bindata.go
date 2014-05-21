package webui

import (
    "bytes"
    "compress/gzip"
    "fmt"
    "io"
)

func bindata_read(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	return buf.Bytes(), nil
}

func content_css_styles_css() ([]byte, error) {
	return bindata_read([]byte{
		0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x00, 0xff, 0x8c, 0x54,
		0x5d, 0x92, 0x9b, 0x30, 0x0c, 0x7e, 0xde, 0x9c, 0xc2, 0x07, 0x58, 0x32,
		0x90, 0x6c, 0xa6, 0x19, 0xf7, 0xa9, 0x77, 0xe8, 0x05, 0x44, 0x30, 0xc4,
		0x53, 0x63, 0x51, 0x23, 0xf2, 0xd3, 0x4e, 0xef, 0x5e, 0x19, 0x1b, 0xc2,
		0x12, 0xd8, 0xc9, 0xa3, 0xe5, 0x4f, 0xdf, 0x8f, 0x90, 0xc9, 0xb1, 0xb8,
		0x8b, 0xbf, 0x9b, 0xb7, 0x1a, 0x5c, 0xa5, 0x6d, 0x42, 0xd8, 0x48, 0x91,
		0xa5, 0xcd, 0xed, 0xfb, 0xe6, 0xdf, 0x66, 0xd3, 0x38, 0xf5, 0xb8, 0x93,
		0xa2, 0x2f, 0xbf, 0x35, 0x50, 0x14, 0xda, 0x56, 0xf1, 0xc8, 0x28, 0x03,
		0xb9, 0x32, 0x1e, 0x57, 0xa2, 0xa5, 0xa4, 0x84, 0x5a, 0x9b, 0xbb, 0x14,
		0x3f, 0x9c, 0x06, 0xc3, 0xf0, 0x13, 0x1a, 0x74, 0x52, 0x54, 0x0e, 0xee,
		0x3d, 0x5a, 0xd7, 0x95, 0xc7, 0xe6, 0xe8, 0x0a, 0xc5, 0xf5, 0xb4, 0x2f,
		0x6e, 0x95, 0x73, 0xe8, 0xfa, 0x3a, 0x9c, 0x7e, 0x55, 0x0e, 0x3b, 0x5b,
		0x48, 0x01, 0xbf, 0x3b, 0x60, 0x6d, 0x6d, 0x55, 0x00, 0x8d, 0x42, 0xd1,
		0x6c, 0x8e, 0x44, 0x58, 0x4b, 0x71, 0x88, 0x46, 0xb6, 0x27, 0x36, 0xa0,
		0x2c, 0x79, 0xc8, 0x55, 0x17, 0x74, 0x96, 0xe2, 0x98, 0x06, 0xd3, 0xb1,
		0xc3, 0xa8, 0x92, 0x98, 0xb7, 0x23, 0x7c, 0xd4, 0x9c, 0xae, 0xce, 0x63,
		0xd1, 0xb3, 0x34, 0x60, 0x83, 0x4c, 0x83, 0xad, 0x26, 0x8d, 0x9c, 0xdc,
		0x29, 0x03, 0xa4, 0x2f, 0xec, 0x63, 0x62, 0x30, 0x89, 0xd1, 0xee, 0xca,
		0x18, 0xbc, 0x46, 0x8b, 0x2c, 0x90, 0x2c, 0x10, 0x40, 0xde, 0xa2, 0xe9,
		0xc8, 0x13, 0x04, 0x0f, 0xc9, 0x2e, 0x9b, 0x8d, 0x33, 0x9e, 0xa3, 0xf1,
		0xec, 0x38, 0x8c, 0x77, 0xdb, 0x1b, 0x7c, 0x89, 0xf4, 0xf8, 0x32, 0xa7,
		0xc1, 0x6a, 0x35, 0xe1, 0xbc, 0x1b, 0x2f, 0xca, 0x95, 0x9c, 0x50, 0x8a,
		0xf6, 0xe4, 0xd0, 0x98, 0x79, 0xd2, 0x77, 0x31, 0xb5, 0xc8, 0xa7, 0x48,
		0x3e, 0x7c, 0xe2, 0xac, 0xb9, 0x09, 0x36, 0xaa, 0x0b, 0x61, 0x3c, 0x2c,
		0x6c, 0x42, 0xbc, 0x4d, 0x1c, 0x14, 0xba, 0x6b, 0xa5, 0xd8, 0x0d, 0xce,
		0x4a, 0x6d, 0x88, 0xeb, 0xfe, 0x53, 0xb2, 0xd8, 0xd3, 0x5a, 0xfd, 0x84,
		0x33, 0xd6, 0xf0, 0x3e, 0xae, 0xd7, 0x68, 0x76, 0xdf, 0x7b, 0x7d, 0xd6,
		0xfc, 0x42, 0x2e, 0x30, 0xb7, 0xfa, 0x8f, 0xe2, 0x86, 0xac, 0xa1, 0x4f,
		0x06, 0x0c, 0xaf, 0xdd, 0xc2, 0xae, 0x65, 0x87, 0x99, 0xd1, 0xbc, 0xe3,
		0x0b, 0xdb, 0x7a, 0x28, 0xa9, 0x1b, 0x25, 0xc0, 0x21, 0xfd, 0x3c, 0x7d,
		0xd4, 0x05, 0xe0, 0x18, 0x28, 0xca, 0xee, 0xbc, 0xec, 0xf2, 0xcb, 0x19,
		0xa3, 0x1d, 0x26, 0xd1, 0xc6, 0x04, 0x1f, 0x83, 0x0f, 0x8b, 0xd7, 0x15,
		0xee, 0x74, 0x9d, 0x7b, 0x6d, 0xfa, 0xa4, 0x6b, 0x35, 0x9d, 0xfd, 0xb0,
		0x39, 0xdf, 0x1e, 0xdb, 0x08, 0xb6, 0xfa, 0x02, 0xb2, 0xfa, 0x36, 0x43,
		0x5f, 0x67, 0x35, 0x4d, 0xa6, 0x1a, 0xd6, 0x76, 0xfc, 0xdf, 0x6c, 0x5b,
		0x75, 0x59, 0xa0, 0xde, 0xa7, 0xb3, 0x99, 0xfb, 0x41, 0x83, 0x53, 0xb0,
		0x20, 0x7f, 0x56, 0xe1, 0x39, 0x67, 0xbb, 0xc9, 0xae, 0x83, 0x9f, 0xe3,
		0x93, 0xec, 0xfe, 0xf8, 0x31, 0x75, 0xdc, 0xff, 0xfa, 0x0e, 0xb1, 0xeb,
		0x7f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x9b, 0x14, 0xd4, 0x89, 0x17, 0x05,
		0x00, 0x00,
		},
		"content/css/styles.css",
	)
}

func content_img_loading_gif() ([]byte, error) {
	return bindata_read([]byte{
		0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x00, 0xff, 0xc4, 0x94,
		0x69, 0x50, 0x13, 0x09, 0xde, 0xc6, 0x9b, 0xf4, 0x91, 0xa3, 0x5b, 0x5f,
		0xed, 0x0e, 0x4e, 0x20, 0x30, 0x06, 0x25, 0x22, 0x8c, 0x03, 0x41, 0x85,
		0xe1, 0x54, 0x21, 0x40, 0x80, 0x80, 0x1c, 0x02, 0x82, 0xe2, 0x41, 0x8c,
		0x0a, 0xde, 0x72, 0x08, 0x04, 0x81, 0x84, 0x10, 0x12, 0x90, 0x23, 0x84,
		0x70, 0x99, 0xa0, 0x24, 0x90, 0x18, 0x22, 0x23, 0x22, 0x1e, 0xa0, 0xa3,
		0x83, 0x20, 0x03, 0xea, 0x60, 0xe1, 0x78, 0xeb, 0x2b, 0x0b, 0x38, 0xa3,
		0x82, 0x78, 0x97, 0xb3, 0xce, 0xec, 0xd4, 0xba, 0x30, 0xbb, 0x5b, 0x35,
		0x1f, 0xf8, 0xb0, 0x55, 0x5b, 0x35, 0xdb, 0x55, 0xfd, 0xb1, 0xab, 0xeb,
		0xf7, 0xfb, 0x3f, 0xcf, 0xc3, 0x0b, 0x09, 0xf2, 0xf0, 0x4c, 0x62, 0x01,
		0x2c, 0xe0, 0x3d, 0x00, 0x7c, 0xfa, 0xf4, 0x49, 0x22, 0x91, 0x3c, 0x79,
		0xf2, 0xa4, 0xa7, 0xa7, 0xe7, 0xd1, 0xa3, 0x47, 0x83, 0x83, 0x83, 0x5a,
		0xad, 0xb6, 0xb5, 0xb5, 0x75, 0x72, 0x72, 0xf2, 0xdd, 0xbb, 0x77, 0xc3,
		0xc3, 0xc3, 0x4a, 0xa5, 0x52, 0x2a, 0x95, 0x4a, 0xfe, 0xfd, 0xd8, 0xfd,
		0x02, 0x51, 0x69, 0x00, 0x00, 0xd8, 0xfd, 0xdd, 0x9a, 0x9b, 0xba, 0x35,
		0x29, 0x7d, 0xab, 0x90, 0x95, 0x99, 0x92, 0x9e, 0xcc, 0x4a, 0xda, 0x91,
		0x94, 0xb5, 0x6b, 0x6f, 0x92, 0xd0, 0x39, 0x65, 0xcf, 0xb6, 0xbd, 0x80,
		0xdd, 0x27, 0x74, 0x75, 0x60, 0xcc, 0x1a, 0xae, 0x5f, 0x64, 0xe0, 0x52,
		0x67, 0x0e, 0x68, 0x31, 0xf5, 0xc1, 0x92, 0xa9, 0x17, 0x98, 0xfe, 0x29,
		0x00, 0x3d, 0x9b, 0x33, 0x10, 0x92, 0x92, 0x64, 0x98, 0x18, 0x7c, 0x26,
		0xe0, 0xcb, 0xec, 0x1b, 0x82, 0x30, 0xcb, 0xa6, 0x68, 0x92, 0x9f, 0x3a,
		0xe6, 0xd4, 0x12, 0xb2, 0x61, 0xe9, 0x9a, 0x31, 0x27, 0x8e, 0x9b, 0x8b,
		0xcb, 0xd5, 0xdd, 0xfa, 0x7d, 0x76, 0xa2, 0xaa, 0xcb, 0xdf, 0x73, 0xac,
		0x48, 0x73, 0xbc, 0xed, 0xfb, 0x38, 0x5c, 0x0b, 0xad, 0x33, 0x29, 0xc4,
		0xc9, 0xee, 0xb7, 0x60, 0xee, 0x62, 0xbf, 0x55, 0x73, 0xf7, 0x5a, 0x40,
		0x52, 0xbb, 0x65, 0x9e, 0x31, 0x6e, 0xe7, 0x13, 0x6f, 0x7a, 0x38, 0x9a,
		0x66, 0xc9, 0x1b, 0x37, 0x5f, 0x7e, 0x75, 0xfa, 0x94, 0x90, 0x4e, 0xce,
		0xcc, 0xe2, 0xf9, 0x26, 0x40, 0xb4, 0xed, 0x74, 0x48, 0x9a, 0x19, 0x2c,
		0x28, 0xb2, 0x3d, 0x80, 0xf8, 0xf9, 0xaa, 0x38, 0xd4, 0xb8, 0x44, 0x6d,
		0xa2, 0xc2, 0x9b, 0x4a, 0x82, 0xc1, 0x66, 0x83, 0xf6, 0xa8, 0xb7, 0xc1,
		0xac, 0x09, 0x36, 0x95, 0xe8, 0x8f, 0xb7, 0x36, 0x74, 0xd8, 0xd6, 0x71,
		0xce, 0x9e, 0x49, 0x67, 0x57, 0x06, 0xa7, 0x15, 0x1f, 0x4a, 0x4d, 0xf0,
		0x09, 0xce, 0x77, 0x96, 0x10, 0x7d, 0xc5, 0xe7, 0x6a, 0xa8, 0x97, 0x54,
		0x82, 0x6f, 0x89, 0x75, 0xd6, 0x82, 0xfe, 0x88, 0xc0, 0xed, 0x5e, 0x50,
		0x5d, 0x3e, 0x2f, 0x61, 0x83, 0x73, 0xb7, 0xef, 0x43, 0xbf, 0xbf, 0x44,
		0x44, 0x69, 0xd3, 0x8a, 0x9e, 0x02, 0x13, 0x89, 0x82, 0xde, 0x64, 0x67,
		0xcf, 0x49, 0xdf, 0xb4, 0xed, 0x1f, 0xf2, 0x8f, 0x6c, 0x29, 0x7d, 0xbf,
		0x75, 0xcc, 0x16, 0xbf, 0xef, 0x84, 0x57, 0xdc, 0x2d, 0xda, 0x06, 0x7c,
		0x15, 0x30, 0x25, 0xe9, 0x5f, 0xb2, 0xfe, 0xc0, 0x3e, 0x31, 0xcd, 0x9e,
		0xd0, 0x3c, 0x31, 0x78, 0x2c, 0x21, 0xc8, 0x8d, 0xde, 0x10, 0xa4, 0xbc,
		0xad, 0x8f, 0x46, 0x7a, 0xd5, 0x31, 0x02, 0x06, 0xef, 0x7c, 0xa8, 0xec,
		0x42, 0xf7, 0x64, 0x6b, 0x98, 0xb9, 0x41, 0x38, 0x6a, 0xfd, 0x66, 0x91,
		0x2c, 0x61, 0x77, 0x16, 0xf3, 0x79, 0xb9, 0xba, 0x17, 0xb3, 0x84, 0x58,
		0x89, 0x18, 0x9d, 0xbe, 0xaa, 0xce, 0x0a, 0x17, 0x4f, 0x5c, 0xb6, 0x3c,
		0xcc, 0xa2, 0x40, 0x27, 0x21, 0x57, 0xc6, 0x6f, 0x8b, 0xa2, 0xa5, 0xed,
		0x2a, 0x22, 0x18, 0xa7, 0x2d, 0xb7, 0x70, 0x8d, 0x3a, 0xfd, 0xe8, 0x60,
		0x1c, 0x62, 0xc5, 0x26, 0x89, 0xa0, 0x03, 0x74, 0x28, 0x7f, 0xe1, 0x0e,
		0x8e, 0x9c, 0x56, 0x4e, 0xdf, 0x55, 0xca, 0xdb, 0x9e, 0xc3, 0xc9, 0x03,
		0x8b, 0x7d, 0x24, 0x3e, 0xd4, 0xc2, 0xf5, 0xfa, 0xf5, 0xf2, 0xe4, 0xca,
		0x2c, 0x83, 0x8e, 0xd7, 0x02, 0xd1, 0xb2, 0x5b, 0xf4, 0x2d, 0xc7, 0xda,
		0xe4, 0x4d, 0x67, 0x6b, 0xd6, 0x1f, 0xe1, 0xd4, 0x9e, 0x67, 0x7e, 0xc3,
		0x3b, 0x5d, 0xc9, 0x34, 0x19, 0x23, 0x8b, 0x44, 0x5a, 0xbc, 0x3f, 0xf9,
		0xea, 0x25, 0xdd, 0x95, 0xba, 0x9d, 0xa2, 0x1e, 0x3c, 0xc5, 0xba, 0xf6,
		0x2a, 0x25, 0xbb, 0xbc, 0xf1, 0xe6, 0xb5, 0xe4, 0x83, 0xa2, 0x3b, 0xc9,
		0x2a, 0xde, 0xe3, 0xc2, 0xb1, 0xb8, 0x3e, 0x7d, 0x9b, 0x64, 0x1c, 0x78,
		0xd9, 0xf5, 0x22, 0x31, 0x99, 0x78, 0xb6, 0xbe, 0xad, 0xfc, 0x35, 0xd0,
		0xdf, 0xac, 0x28, 0xf9, 0xb0, 0xa8, 0xfa, 0x8b, 0xca, 0xc6, 0xb5, 0x47,
		0xbe, 0x3a, 0xe7, 0x31, 0x62, 0x35, 0x60, 0x47, 0x9a, 0x09, 0xff, 0xd5,
		0x14, 0xbe, 0xa3, 0xeb, 0x14, 0xfe, 0x33, 0x57, 0x9e, 0x9b, 0x70, 0x83,
		0x6c, 0x71, 0xb5, 0x2e, 0xba, 0xff, 0x94, 0x3a, 0x66, 0x69, 0xa9, 0x7a,
		0x47, 0x58, 0x61, 0x4e, 0xef, 0x2d, 0x1f, 0x16, 0xb1, 0x7e, 0xc8, 0x6d,
		0x24, 0x9c, 0x7e, 0x69, 0x84, 0xc6, 0x19, 0xb2, 0x72, 0x94, 0x59, 0xd1,
		0x58, 0x61, 0x9d, 0xf3, 0x9a, 0x43, 0x48, 0x45, 0x6f, 0x76, 0xcb, 0x06,
		0x03, 0x4d, 0xfd, 0x9b, 0xe9, 0xa7, 0x21, 0x78, 0x1f, 0x6e, 0x88, 0xa5,
		0xa1, 0x12, 0x68, 0xe3, 0x36, 0x3a, 0x9b, 0x52, 0xb0, 0xb1, 0x1d, 0x1e,
		0xf8, 0xd1, 0x96, 0x84, 0x64, 0x60, 0xde, 0x04, 0x49, 0x24, 0x22, 0xe5,
		0x72, 0x0a, 0xe3, 0x99, 0x54, 0x5a, 0xf1, 0x1a, 0x0e, 0xb8, 0x35, 0xd3,
		0x36, 0x1b, 0xda, 0xe9, 0x43, 0xd6, 0xb2, 0xa9, 0x47, 0x23, 0x9a, 0x9a,
		0x8a, 0xab, 0xa8, 0x95, 0x46, 0x51, 0xb3, 0x39, 0x47, 0xf5, 0x75, 0xab,
		0xf9, 0x60, 0x8b, 0xb9, 0xf9, 0x6c, 0x93, 0x9e, 0xdd, 0x65, 0x73, 0xf8,
		0x02, 0xfd, 0x62, 0x50, 0x5b, 0x4a, 0xa9, 0x6b, 0x8f, 0x9a, 0xa3, 0x28,
		0x16, 0xcb, 0x8d, 0xd7, 0xbf, 0x49, 0x38, 0xad, 0x21, 0x0e, 0xf7, 0x7a,
		0x5c, 0xc9, 0x2a, 0xf9, 0x56, 0x55, 0x4a, 0x15, 0x1d, 0xbd, 0xba, 0xaa,
		0x18, 0x2f, 0xbe, 0xe7, 0xf3, 0xf8, 0x38, 0x30, 0x06, 0x1c, 0xee, 0xbf,
		0x3c, 0xfc, 0xe4, 0x39, 0xd0, 0x76, 0x89, 0x39, 0xe9, 0xf1, 0xf6, 0x59,
		0x82, 0x4f, 0x5b, 0xe9, 0x04, 0x7e, 0xd5, 0x40, 0xf1, 0x78, 0x2f, 0x3f,
		0x96, 0x28, 0x8f, 0xb4, 0xb4, 0x3f, 0x73, 0xc1, 0x60, 0xa7, 0xba, 0x3c,
		0x6b, 0xee, 0x8c, 0x09, 0x98, 0x4e, 0x7f, 0x0b, 0x43, 0x3f, 0x31, 0xb8,
		0xdd, 0x2a, 0x30, 0x76, 0x4e, 0x43, 0x0d, 0xeb, 0x4e, 0x63, 0xb4, 0x3c,
		0x49, 0x1d, 0xe3, 0x3f, 0xe7, 0x41, 0xf3, 0x6c, 0x62, 0xf4, 0xd2, 0xbe,
		0x15, 0x0e, 0x67, 0xab, 0x9b, 0xb7, 0x52, 0x3e, 0xd8, 0x2b, 0xea, 0x17,
		0x28, 0x40, 0x7a, 0xa2, 0x7c, 0x61, 0xa0, 0x2b, 0xd7, 0x62, 0x8f, 0xf8,
		0x4a, 0x39, 0x5d, 0x7a, 0x35, 0xef, 0xde, 0x1d, 0x20, 0xd4, 0x6a, 0xc9,
		0xbc, 0xff, 0x5f, 0xe2, 0x97, 0x54, 0x4d, 0x31, 0x5b, 0x37, 0xc4, 0x66,
		0x6e, 0x5c, 0x6e, 0x11, 0xc2, 0x8e, 0x7c, 0x3c, 0x4c, 0x99, 0x9d, 0x81,
		0x46, 0xcd, 0x5b, 0xb6, 0x8c, 0x94, 0xc3, 0x91, 0x20, 0x29, 0xae, 0x31,
		0x32, 0x1e, 0xbc, 0x3d, 0xd3, 0x26, 0x3b, 0x97, 0x0d, 0x2a, 0xd8, 0x0e,
		0xd1, 0xda, 0x06, 0x02, 0x6b, 0xd4, 0xa1, 0x30, 0xb5, 0x5c, 0x56, 0xea,
		0xab, 0x33, 0x63, 0xc6, 0x63, 0x0d, 0xc7, 0xb1, 0x26, 0xc3, 0xb1, 0x65,
		0x47, 0xce, 0xf2, 0x6a, 0xd8, 0x5d, 0xea, 0x7d, 0xe7, 0x2f, 0x72, 0x8c,
		0xb4, 0x7a, 0xa2, 0x23, 0x34, 0x7a, 0xaf, 0xcc, 0x8d, 0x18, 0x28, 0xe2,
		0x84, 0x5c, 0x31, 0xee, 0xce, 0xde, 0xa9, 0x80, 0x7a, 0xb3, 0x06, 0x09,
		0x4a, 0x8c, 0x62, 0x53, 0xee, 0xc0, 0xe2, 0x0a, 0x40, 0x76, 0x6f, 0xe3,
		0xd8, 0x58, 0x36, 0xbf, 0xa1, 0x3c, 0xed, 0x0c, 0x3e, 0xa9, 0x1d, 0x5f,
		0x91, 0xb2, 0xc9, 0xc4, 0x2e, 0xbf, 0xf3, 0x6a, 0xcf, 0xaf, 0xbe, 0x0f,
		0xf0, 0xb1, 0x87, 0xd7, 0xf6, 0xdf, 0x5f, 0xbe, 0x75, 0xf0, 0xcb, 0xf0,
		0x9e, 0xe6, 0x99, 0xd3, 0xff, 0xfb, 0xf9, 0x9d, 0x4c, 0x13, 0x83, 0x0e,
		0x8e, 0x81, 0x06, 0xe1, 0x86, 0x6a, 0x25, 0xc1, 0x30, 0xf6, 0x46, 0x46,
		0x53, 0xfd, 0x28, 0x6a, 0x2f, 0xbb, 0xaf, 0xff, 0x26, 0x4a, 0x55, 0x15,
		0x54, 0x69, 0x05, 0x99, 0x55, 0xb3, 0xd8, 0xee, 0xe2, 0x05, 0xbc, 0xe6,
		0xc5, 0x42, 0xfb, 0xf5, 0x0b, 0xe4, 0xaa, 0x8f, 0xa1, 0xed, 0x9f, 0x77,
		0x07, 0x25, 0x7f, 0xad, 0x44, 0xc0, 0x24, 0x9c, 0x19, 0x05, 0x45, 0x96,
		0x6d, 0xde, 0xc7, 0xb6, 0xe9, 0x29, 0x72, 0xf1, 0x0f, 0x8a, 0x4c, 0xdc,
		0x12, 0x3b, 0x6b, 0x6c, 0xd6, 0xca, 0x98, 0xa1, 0x74, 0xd2, 0x5a, 0x97,
		0x7d, 0x08, 0x99, 0xb7, 0x28, 0x22, 0x3d, 0x00, 0x4e, 0xda, 0x64, 0x93,
		0x06, 0x65, 0xdb, 0xee, 0x0a, 0x70, 0x08, 0xaf, 0xab, 0x27, 0x50, 0x6d,
		0xc3, 0x7e, 0x6a, 0x49, 0xfa, 0x96, 0xba, 0x06, 0x03, 0xaa, 0xd3, 0xd7,
		0x1b, 0x51, 0x41, 0xa3, 0x7e, 0xe9, 0xe1, 0xf6, 0x00, 0xb0, 0x4a, 0x19,
		0x0e, 0x62, 0x98, 0x97, 0x6d, 0xe5, 0x19, 0xeb, 0xae, 0x2e, 0xb4, 0x42,
		0xa8, 0x53, 0xb0, 0x49, 0x48, 0x37, 0x86, 0xe4, 0xeb, 0x97, 0xb3, 0xd3,
		0x1c, 0xd1, 0xee, 0x0b, 0x95, 0x19, 0x52, 0x48, 0x1a, 0xb8, 0x8c, 0xde,
		0x89, 0x35, 0x01, 0x09, 0xb1, 0x62, 0xab, 0x43, 0x40, 0xfa, 0xc3, 0x9a,
		0xa7, 0x4f, 0x2b, 0xbf, 0x0b, 0x28, 0xb1, 0x7f, 0x1a, 0x9a, 0x12, 0xfe,
		0xd2, 0x67, 0x24, 0x61, 0xc4, 0xbb, 0xe4, 0xfe, 0x49, 0x42, 0x5c, 0x27,
		0xf4, 0xa9, 0x62, 0x45, 0x90, 0x0a, 0xd2, 0xe3, 0xdf, 0xf9, 0xec, 0x77,
		0xf8, 0x62, 0xae, 0xff, 0xec, 0x99, 0xcf, 0xff, 0x7e, 0x7a, 0x00, 0xa2,
		0x5b, 0xa6, 0x06, 0x60, 0x41, 0x28, 0x9b, 0xd5, 0x50, 0xa3, 0xa4, 0xeb,
		0x8c, 0x81, 0x51, 0xeb, 0xcc, 0xca, 0x13, 0xf8, 0x0f, 0xf8, 0xeb, 0x30,
		0xa1, 0xaa, 0xe4, 0xaf, 0x5f, 0x26, 0x9c, 0x30, 0xd2, 0x3e, 0x24, 0x17,
		0xd4, 0xec, 0x7c, 0x1e, 0xa5, 0x3b, 0x28, 0x2e, 0xc4, 0xdd, 0x46, 0x33,
		0xf0, 0x4b, 0xd8, 0xf2, 0x84, 0x8f, 0x21, 0x74, 0xa9, 0x85, 0xe5, 0xf1,
		0x46, 0x12, 0xb4, 0x7a, 0x96, 0x60, 0x2d, 0x2a, 0x2f, 0x13, 0x67, 0x54,
		0xca, 0xdd, 0x34, 0xb4, 0x57, 0x97, 0x53, 0x0b, 0x7e, 0xa6, 0xb6, 0xb2,
		0xb3, 0x05, 0xcc, 0xcf, 0x56, 0x6c, 0x64, 0x93, 0xf6, 0xba, 0xa4, 0x1e,
		0xb0, 0xdd, 0xe5, 0xe9, 0x10, 0xa6, 0xaa, 0x22, 0x90, 0x9a, 0x5a, 0xb2,
		0xf7, 0xfc, 0x22, 0xb9, 0x67, 0xed, 0x51, 0x44, 0xa3, 0xad, 0x6a, 0x44,
		0xea, 0x65, 0x4d, 0x6a, 0xb3, 0x27, 0x28, 0x66, 0x97, 0xd9, 0xc0, 0x28,
		0x0a, 0x1e, 0x4a, 0xf1, 0xec, 0xe8, 0x40, 0x1a, 0x70, 0x13, 0x8b, 0x4b,
		0x21, 0x77, 0xa1, 0xe4, 0x05, 0x45, 0xfe, 0xb6, 0x20, 0xe6, 0x0f, 0x21,
		0x5d, 0x67, 0x02, 0xd2, 0x88, 0x0a, 0x3a, 0x19, 0xc3, 0x10, 0x3a, 0x78,
		0x63, 0x63, 0x3c, 0xb4, 0x2d, 0xd7, 0x3e, 0x92, 0x7a, 0x0f, 0xcb, 0x09,
		0x1b, 0x5e, 0x19, 0x09, 0x5c, 0xc3, 0x50, 0xb5, 0x26, 0x62, 0x39, 0x80,
		0xde, 0x1b, 0xea, 0x0e, 0xbc, 0xf5, 0x1e, 0xfa, 0x49, 0xa5, 0x89, 0x7b,
		0x61, 0xcf, 0xe8, 0x7c, 0x7a, 0x6d, 0xae, 0x85, 0x10, 0x20, 0x85, 0x0e,
		0x2d, 0x74, 0xc4, 0x65, 0xff, 0xb7, 0x2f, 0x6e, 0x6e, 0x4f, 0xb7, 0xfd,
		0x8c, 0x4b, 0xf0, 0xe2, 0x0f, 0x1e, 0x64, 0x53, 0x51, 0x60, 0x59, 0xea,
		0xa2, 0x7b, 0xd6, 0xad, 0x2e, 0x76, 0x8a, 0x74, 0xb2, 0x38, 0x3e, 0xea,
		0x7d, 0xc7, 0x3e, 0x32, 0xc7, 0x49, 0xbd, 0xba, 0xa0, 0xfb, 0x65, 0x62,
		0xe0, 0x90, 0xf2, 0x9d, 0x1d, 0x25, 0xdf, 0x35, 0x36, 0x62, 0xe9, 0xe3,
		0x00, 0x6a, 0xfd, 0xa6, 0x54, 0xf7, 0x90, 0x9f, 0x05, 0x34, 0x5d, 0xf3,
		0x4f, 0xb7, 0xaf, 0x04, 0xd7, 0x1e, 0xf6, 0xa0, 0xfa, 0xd7, 0x7a, 0x53,
		0x4f, 0xb5, 0x2e, 0x70, 0xa0, 0x98, 0xb5, 0xeb, 0x1e, 0x3d, 0x4c, 0xbf,
		0xa8, 0x12, 0xd8, 0x88, 0xf9, 0x0b, 0xb9, 0x0e, 0x7c, 0x45, 0x31, 0x01,
		0x96, 0x96, 0xc1, 0x99, 0xb9, 0x2b, 0x15, 0x65, 0x6a, 0xb0, 0x52, 0xc5,
		0xaf, 0x06, 0x2b, 0x52, 0x44, 0xee, 0x25, 0x8d, 0x5e, 0x60, 0x56, 0x01,
		0x1f, 0x42, 0x10, 0x38, 0xdf, 0x8f, 0x6b, 0x32, 0x91, 0x8f, 0x2e, 0xca,
		0xf5, 0xcc, 0x27, 0x9f, 0x40, 0xc8, 0xe9, 0xa2, 0x44, 0x5b, 0x18, 0x4d,
		0x84, 0xce, 0x9a, 0xcc, 0x01, 0x6e, 0x0c, 0xef, 0x2c, 0x10, 0x45, 0xbb,
		0x08, 0xd8, 0xd4, 0x1e, 0x25, 0x11, 0x2e, 0xc6, 0xee, 0x02, 0xd4, 0xeb,
		0x68, 0x3b, 0x7f, 0x2d, 0x7e, 0x17, 0xc3, 0xaf, 0xa1, 0x48, 0x49, 0xb8,
		0xff, 0x08, 0x81, 0x5c, 0x07, 0x15, 0x21, 0xa7, 0xc7, 0x71, 0xe8, 0xa1,
		0x22, 0xdc, 0xf7, 0x2d, 0x4e, 0xbe, 0xde, 0xcc, 0x4c, 0x23, 0x3e, 0x8e,
		0x75, 0x5e, 0xf4, 0x08, 0x48, 0xb2, 0x90, 0xc2, 0xba, 0x50, 0xfa, 0xe6,
		0x55, 0x61, 0x3c, 0x68, 0x26, 0xfe, 0xd7, 0x7f, 0x06, 0x3f, 0x4f, 0xc4,
		0x98, 0x02, 0x56, 0x54, 0x78, 0xe4, 0x8a, 0x8a, 0x55, 0x65, 0xea, 0xbc,
		0x12, 0x0d, 0x3b, 0xbe, 0x90, 0x4f, 0x02, 0x41, 0x4f, 0x5b, 0xa9, 0x97,
		0xde, 0x24, 0xe1, 0xe4, 0x36, 0xb3, 0xa9, 0xb0, 0x09, 0x84, 0xa7, 0xc9,
		0x21, 0xc4, 0x9f, 0xd2, 0xde, 0xbc, 0x8b, 0xb1, 0x05, 0x03, 0xad, 0x41,
		0x04, 0x21, 0x87, 0xe9, 0x05, 0x51, 0x12, 0x04, 0xc3, 0xfa, 0xd1, 0x9b,
		0x00, 0xb5, 0x0f, 0x51, 0x2a, 0xe0, 0x1b, 0x98, 0xd5, 0x4d, 0x14, 0x87,
		0xfb, 0xfa, 0x8b, 0xd1, 0x1b, 0x3d, 0xc0, 0x63, 0x82, 0xdc, 0xd7, 0xc1,
		0x07, 0x6f, 0x0c, 0xe3, 0xcf, 0xf1, 0xce, 0x3e, 0xc5, 0x8f, 0xd8, 0xe4,
		0xb3, 0x1f, 0xc6, 0xfb, 0xb2, 0x6c, 0x6e, 0x20, 0x8b, 0x07, 0x34, 0x91,
		0x23, 0x87, 0x67, 0xc7, 0x5b, 0x62, 0x4b, 0x28, 0x03, 0xe3, 0x77, 0x0b,
		0x0e, 0xba, 0x2c, 0x98, 0xb1, 0x0b, 0x6f, 0xfe, 0x47, 0x0e, 0x64, 0x5e,
		0x14, 0x72, 0xa1, 0x9f, 0xcb, 0x54, 0xf7, 0x41, 0x6e, 0x2e, 0x83, 0x09,
		0xa1, 0x46, 0x63, 0x51, 0x46, 0x22, 0x81, 0xb4, 0x61, 0x48, 0x07, 0x93,
		0x04, 0x26, 0xe7, 0x0a, 0x0c, 0x46, 0x14, 0x0e, 0xe8, 0xb1, 0x86, 0x40,
		0xb0, 0x77, 0xc3, 0x7c, 0x5f, 0xd8, 0x68, 0x22, 0xc8, 0x28, 0x6a, 0x42,
		0xa6, 0x32, 0x30, 0x00, 0x4a, 0xe2, 0x1c, 0x6f, 0xa1, 0x56, 0x77, 0x11,
		0x9c, 0xf6, 0xb0, 0x04, 0xb9, 0x35, 0x34, 0x46, 0xc0, 0x03, 0x9e, 0x5c,
		0xf0, 0xd6, 0x13, 0xfc, 0x45, 0xc0, 0x53, 0xfe, 0x73, 0x54, 0x09, 0x7c,
		0xc0, 0x5f, 0x7e, 0xc3, 0x37, 0xca, 0xf4, 0x19, 0x65, 0xd4, 0x73, 0x78,
		0xbf, 0x93, 0x0d, 0x87, 0xa3, 0x5c, 0xf3, 0x39, 0x78, 0x2d, 0x7d, 0xe6,
		0x4d, 0xfc, 0x53, 0x8a, 0x30, 0x25, 0x01, 0x46, 0x31, 0x65, 0x65, 0xb1,
		0xc8, 0x8d, 0x56, 0x59, 0x8b, 0x55, 0xa9, 0xcb, 0x2a, 0xea, 0x4a, 0x8e,
		0x72, 0x65, 0x5e, 0x54, 0x70, 0x5a, 0x04, 0x8a, 0xc2, 0xbf, 0x8b, 0x40,
		0xcc, 0xe6, 0x69, 0x11, 0xe4, 0x0e, 0x94, 0x7c, 0x8e, 0x19, 0x38, 0xe5,
		0xa1, 0xc5, 0x8c, 0xb8, 0x03, 0x94, 0x3e, 0xeb, 0xf5, 0x74, 0xca, 0x7c,
		0x5f, 0xc8, 0x6c, 0xe8, 0x44, 0x10, 0x03, 0x78, 0x0f, 0x58, 0x9b, 0xb5,
		0xc8, 0xf1, 0x36, 0x62, 0x75, 0xef, 0x87, 0x10, 0x4f, 0x05, 0xf9, 0x76,
		0x2b, 0x30, 0xba, 0x6a, 0x2f, 0x17, 0xbe, 0x4d, 0x26, 0x26, 0x62, 0x56,
		0xf2, 0x87, 0xbf, 0x1b, 0xbf, 0x4f, 0x08, 0x1f, 0x31, 0x5f, 0xd3, 0xdf,
		0x6d, 0xf8, 0x78, 0xd9, 0xff, 0x2b, 0x5d, 0x37, 0x50, 0xd0, 0xe8, 0x4e,
		0x61, 0x0c, 0x1d, 0x0f, 0x98, 0xb1, 0x08, 0xcf, 0xff, 0x2b, 0x7e, 0x0a,
		0xf6, 0x9d, 0xd8, 0xf2, 0x77, 0xfe, 0x39, 0x1c, 0xca, 0xb2, 0x79, 0xb0,
		0xb9, 0x43, 0x9d, 0x7c, 0xa7, 0x93, 0x90, 0x6a, 0x92, 0xba, 0x6b, 0x77,
		0xb0, 0x62, 0x71, 0x18, 0xcd, 0xe1, 0x86, 0xec, 0xb4, 0xd9, 0x2d, 0xe0,
		0x2f, 0x94, 0x6f, 0xe1, 0x57, 0xaa, 0x70, 0xb1, 0x07, 0xb5, 0x30, 0x27,
		0x49, 0x55, 0xed, 0x51, 0xbf, 0x47, 0x53, 0xcd, 0xa8, 0xab, 0xaf, 0x32,
		0xc8, 0x0b, 0xe2, 0x94, 0x5e, 0x0a, 0x17, 0x18, 0x41, 0xf6, 0x7b, 0x85,
		0x4b, 0x98, 0x34, 0x72, 0x5b, 0x1b, 0x37, 0x43, 0x9c, 0x07, 0x9c, 0xeb,
		0x24, 0x17, 0x5b, 0x07, 0x26, 0x77, 0xec, 0x3c, 0xd1, 0x46, 0x76, 0x07,
		0xa8, 0xe0, 0x45, 0x60, 0xda, 0xc2, 0xa6, 0xb6, 0x76, 0x29, 0x08, 0xba,
		0x47, 0xfa, 0xb8, 0xc4, 0x52, 0x1e, 0x80, 0x09, 0xc1, 0x41, 0xaa, 0x51,
		0xcf, 0x95, 0x78, 0x78, 0x0f, 0x9b, 0x36, 0xca, 0xbb, 0xfe, 0xfd, 0x38,
		0x77, 0x74, 0xe7, 0x7b, 0xe0, 0x80, 0x89, 0xf9, 0xe0, 0xed, 0x0a, 0xe7,
		0x0d, 0xbf, 0xba, 0xd3, 0x5b, 0xde, 0xd8, 0xc9, 0x1d, 0x67, 0x7f, 0x46,
		0x8a, 0x53, 0xfc, 0x27, 0x0d, 0x88, 0xc5, 0xfe, 0x09, 0x4f, 0xed, 0x65,
		0xa8, 0x43, 0xbb, 0x77, 0x7c, 0x5c, 0xbd, 0x34, 0x8d, 0xef, 0x5e, 0x41,
		0x6b, 0xd7, 0x60, 0xc2, 0x43, 0x21, 0x73, 0xc4, 0x03, 0x96, 0x8e, 0xa8,
		0x15, 0x6d, 0x4e, 0xd8, 0xfd, 0xe0, 0x93, 0x6b, 0x75, 0x3c, 0xcc, 0x9d,
		0x4a, 0xba, 0xf6, 0x19, 0x3f, 0x9e, 0x61, 0xc4, 0xa5, 0x93, 0x8d, 0xed,
		0x2a, 0x67, 0x77, 0xbd, 0xf0, 0xa4, 0x89, 0x99, 0x47, 0x42, 0x44, 0x55,
		0xc9, 0x0f, 0x7a, 0xe8, 0xe4, 0x8c, 0x0c, 0x72, 0xaa, 0x8b, 0x3b, 0x2b,
		0xde, 0x0d, 0x91, 0x86, 0x78, 0x6f, 0x2a, 0xb2, 0x49, 0x47, 0xe1, 0x88,
		0x85, 0x81, 0xd4, 0x9c, 0x08, 0x8d, 0x46, 0xe1, 0x19, 0x47, 0x46, 0xf4,
		0x4d, 0xda, 0x23, 0x9e, 0x4d, 0xa6, 0xe5, 0x0d, 0x47, 0x18, 0x90, 0xae,
		0x45, 0xdb, 0xae, 0xa9, 0xf5, 0x8e, 0xab, 0xe1, 0xba, 0xd0, 0x40, 0x50,
		0xdc, 0xe5, 0xce, 0x60, 0x52, 0x2e, 0xf4, 0xd6, 0xe5, 0x39, 0x5d, 0xc4,
		0x1d, 0x7a, 0xe1, 0x12, 0xeb, 0xf0, 0xbd, 0xf3, 0x3b, 0xf1, 0xf3, 0x17,
		0x60, 0xf1, 0x9e, 0x1b, 0xfd, 0x82, 0xf9, 0xf1, 0x3d, 0xa0, 0x17, 0x9e,
		0xb7, 0xe1, 0x8b, 0x38, 0xa7, 0xc1, 0xdd, 0x96, 0x3f, 0x41, 0x56, 0xfc,
		0x4d, 0x23, 0x11, 0x37, 0x53, 0xf9, 0xc0, 0xdb, 0x8e, 0xd5, 0xab, 0x9c,
		0x81, 0x09, 0xcd, 0x4d, 0xe1, 0x2f, 0x40, 0xde, 0x8d, 0xfd, 0x43, 0xc6,
		0x1f, 0xe7, 0x13, 0x12, 0x1b, 0x3c, 0x40, 0x4f, 0xf6, 0xbf, 0x7b, 0x08,
		0xa5, 0xb8, 0x4d, 0x35, 0xc0, 0xfb, 0x1f, 0x01, 0x00, 0x00, 0xff, 0xff,
		0xc0, 0x05, 0xf9, 0x4b, 0x7f, 0x0c, 0x00, 0x00,
		},
		"content/img/loading.gif",
	)
}

func content_index_html() ([]byte, error) {
	return bindata_read([]byte{
		0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x00, 0xff, 0xbc, 0x56,
		0x5d, 0x6f, 0x9b, 0x30, 0x14, 0x7d, 0xee, 0x7e, 0x85, 0xe7, 0x87, 0xbd,
		0x81, 0x9b, 0x6e, 0x9d, 0xaa, 0x0e, 0x78, 0xd9, 0xc7, 0xcb, 0xa4, 0x6d,
		0x5a, 0xf7, 0xb2, 0x47, 0x07, 0x5f, 0x62, 0xa7, 0x60, 0x33, 0xdb, 0x24,
		0xe1, 0xdf, 0xcf, 0xe0, 0x90, 0x00, 0x4d, 0x1a, 0xd4, 0x44, 0x95, 0x22,
		0xf9, 0xeb, 0x5c, 0x9f, 0x73, 0x4f, 0xae, 0x2e, 0x8e, 0xde, 0x7e, 0xf9,
		0xf9, 0xf9, 0xcf, 0xdf, 0x5f, 0x5f, 0x11, 0xb7, 0x45, 0x9e, 0xbc, 0x89,
		0xfc, 0x70, 0x15, 0x71, 0xa0, 0xcc, 0x8d, 0x57, 0x51, 0x2e, 0xe4, 0x23,
		0xd2, 0x90, 0xc7, 0xd8, 0xd8, 0x3a, 0x07, 0xc3, 0x01, 0x2c, 0x46, 0x5c,
		0x43, 0x16, 0x63, 0x92, 0x1a, 0x43, 0xfc, 0x76, 0xe8, 0xa6, 0xb8, 0x0d,
		0x30, 0xa9, 0x16, 0xa5, 0x45, 0x46, 0xa7, 0x31, 0xe6, 0xd6, 0x96, 0xe6,
		0x9e, 0x90, 0x54, 0x31, 0x08, 0x97, 0xff, 0x2a, 0xd0, 0x75, 0x98, 0xaa,
		0x82, 0xf8, 0x69, 0x30, 0x0b, 0x67, 0xee, 0x17, 0x16, 0x42, 0x86, 0x4b,
		0x83, 0x91, 0xad, 0x4b, 0x88, 0xb1, 0x85, 0x8d, 0x25, 0x4b, 0xba, 0xa2,
		0xfe, 0x22, 0x9c, 0x44, 0xc4, 0xcf, 0x1a, 0x59, 0x64, 0xab, 0x2b, 0x9a,
		0x2b, 0x56, 0xb7, 0x74, 0x4c, 0xac, 0x50, 0x9a, 0x53, 0x63, 0x62, 0x9c,
		0x2a, 0x69, 0x41, 0xda, 0x56, 0xc6, 0xe0, 0xa0, 0xa4, 0x12, 0x72, 0xbf,
		0x3d, 0xd8, 0xcf, 0x21, 0xb3, 0x41, 0xff, 0x70, 0x70, 0x9a, 0x89, 0xdc,
		0x82, 0x0e, 0x9c, 0x01, 0xd0, 0x1d, 0x0f, 0xa3, 0xe9, 0xbc, 0x09, 0x8c,
		0xda, 0x31, 0xf9, 0xa6, 0x55, 0x71, 0x1f, 0x11, 0xbf, 0x88, 0x88, 0xc3,
		0xed, 0x62, 0x84, 0x2c, 0x2b, 0x8b, 0x04, 0x8b, 0x31, 0xa3, 0x16, 0x82,
		0xcc, 0x21, 0xfb, 0xc9, 0xe2, 0x11, 0x61, 0x93, 0x86, 0x56, 0x39, 0xb2,
		0xa2, 0x80, 0x6e, 0x81, 0xd1, 0x8a, 0xe6, 0x95, 0x0b, 0xc0, 0xa8, 0xa0,
		0x9b, 0x1c, 0xe4, 0xc2, 0xf2, 0x18, 0xcf, 0xae, 0xf1, 0x01, 0x92, 0x36,
		0xee, 0x72, 0x24, 0x77, 0x7b, 0x8e, 0x79, 0x65, 0xad, 0x92, 0xdd, 0x4d,
		0x52, 0xad, 0x03, 0xbf, 0x83, 0x93, 0x1f, 0x6a, 0x1d, 0x11, 0xbf, 0xe8,
		0x9c, 0xec, 0x59, 0xf0, 0x72, 0x57, 0x7f, 0x53, 0xb9, 0x80, 0x93, 0xb6,
		0xea, 0x06, 0x35, 0x25, 0xdb, 0x16, 0xf8, 0x24, 0xdd, 0x60, 0x36, 0x74,
		0xf5, 0xe3, 0x21, 0x57, 0x7d, 0x68, 0x25, 0x85, 0x0d, 0x76, 0xa5, 0xaa,
		0x29, 0x13, 0x0a, 0x23, 0x49, 0x0b, 0xe8, 0x03, 0xcc, 0x8e, 0x7e, 0xbf,
		0xb7, 0x23, 0x33, 0xfb, 0xcb, 0xdb, 0xa4, 0x50, 0xa6, 0xf4, 0xe8, 0xf6,
		0xc4, 0x74, 0x09, 0x3f, 0x2b, 0xa3, 0x78, 0x91, 0x8c, 0x94, 0x43, 0xfa,
		0x08, 0xac, 0x93, 0x53, 0x9c, 0x92, 0xe3, 0x00, 0xc5, 0x24, 0x39, 0xfc,
		0x1c, 0x57, 0xf8, 0x29, 0x19, 0x0e, 0xc0, 0x27, 0xc9, 0x60, 0xe7, 0xc8,
		0x60, 0xa7, 0x64, 0x38, 0x00, 0x1b, 0xca, 0xb8, 0x4c, 0x9d, 0x3f, 0xc0,
		0x0a, 0xb4, 0xb0, 0xf5, 0xc9, 0x52, 0x77, 0xbd, 0x32, 0x30, 0xb0, 0x9a,
		0x52, 0xec, 0x0e, 0xf6, 0xa4, 0xd4, 0xaf, 0x07, 0x95, 0xfe, 0x1e, 0x27,
		0xef, 0xe4, 0xdc, 0x94, 0x9f, 0x02, 0x3f, 0x1c, 0xa0, 0xa3, 0x9b, 0x73,
		0xe8, 0x6e, 0x6e, 0x6f, 0xc7, 0x84, 0x93, 0x5c, 0xf3, 0x9d, 0xc4, 0x1c,
		0xeb, 0x3c, 0x03, 0x94, 0x23, 0xa6, 0x3a, 0xe5, 0xbb, 0x56, 0xf4, 0xd0,
		0x2e, 0x8f, 0x77, 0xa3, 0xfe, 0xb4, 0xc7, 0xac, 0xc5, 0x82, 0x5f, 0xec,
		0x63, 0xf0, 0xa0, 0x2a, 0x9d, 0x82, 0x39, 0xf2, 0x6f, 0x36, 0x2e, 0x52,
		0x0d, 0xb4, 0x75, 0xd8, 0x78, 0xe8, 0x31, 0x57, 0xb7, 0xcb, 0x2e, 0x64,
		0xd4, 0xfe, 0x6f, 0x3e, 0x34, 0x9f, 0xc7, 0xee, 0xf0, 0xa2, 0x15, 0xf9,
		0x1d, 0xea, 0xb5, 0xd2, 0x6c, 0x52, 0x0e, 0x8f, 0x5b, 0xec, 0xeb, 0x25,
		0xf1, 0x3a, 0x05, 0xd2, 0x9b, 0xf5, 0x7d, 0x52, 0x0b, 0x3c, 0x3c, 0xde,
		0x4e, 0xc6, 0xaf, 0x1f, 0xb2, 0x34, 0x84, 0x96, 0xe5, 0x33, 0x6f, 0x1b,
		0xd7, 0x8d, 0xa9, 0x36, 0x60, 0x63, 0x5c, 0xd9, 0x2c, 0xb8, 0x1b, 0xbd,
		0x75, 0xfc, 0x1b, 0xc7, 0xbd, 0x79, 0x9a, 0x37, 0xd9, 0xff, 0x00, 0x00,
		0x00, 0xff, 0xff, 0x12, 0xe4, 0x58, 0x9f, 0xaa, 0x09, 0x00, 0x00,
		},
		"content/index.html",
	)
}

func content_js_app_js() ([]byte, error) {
	return bindata_read([]byte{
		0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x00, 0xff, 0x9c, 0x56,
		0xdf, 0x6f, 0xdb, 0xb6, 0x13, 0x7f, 0x56, 0xff, 0x0a, 0x7e, 0xf3, 0x0d,
		0x2a, 0x6a, 0xae, 0xe5, 0x24, 0xd8, 0x93, 0x9b, 0x34, 0x0f, 0xdb, 0x82,
		0x0e, 0x6b, 0x86, 0x62, 0x09, 0x06, 0x0c, 0x69, 0x06, 0xb0, 0x12, 0x25,
		0x13, 0x91, 0x48, 0x83, 0xa2, 0xec, 0xb8, 0xad, 0xff, 0xf7, 0xdd, 0xf1,
		0x87, 0x24, 0xcb, 0x4e, 0x9a, 0x2d, 0x0f, 0xce, 0xf1, 0xee, 0x78, 0xfc,
		0xdc, 0x7d, 0xee, 0x28, 0xd2, 0xa2, 0x95, 0x99, 0x11, 0x4a, 0x12, 0x9a,
		0x7c, 0x7d, 0x15, 0x1d, 0xd3, 0x5c, 0x65, 0x6d, 0xcd, 0xa5, 0x49, 0x52,
		0xcd, 0x59, 0xbe, 0x19, 0xda, 0x09, 0x38, 0x44, 0x0d, 0x37, 0x57, 0x5a,
		0xd5, 0xbf, 0xab, 0x35, 0x4d, 0xde, 0xba, 0xf5, 0x07, 0x55, 0xde, 0x88,
		0x2f, 0x1c, 0xd7, 0xa0, 0x38, 0xa6, 0x71, 0xda, 0x70, 0xa6, 0xb3, 0xc5,
		0xf4, 0x73, 0x6b, 0x8c, 0x92, 0x71, 0x92, 0x66, 0x95, 0xc8, 0x1e, 0xf6,
		0x42, 0x45, 0x2b, 0xa6, 0x49, 0x01, 0xc1, 0xc8, 0x05, 0x29, 0x5d, 0x58,
		0x9a, 0xbc, 0x41, 0x43, 0xa4, 0x99, 0x2c, 0xb9, 0x53, 0xff, 0x81, 0x62,
		0xd0, 0xd7, 0x42, 0xde, 0xf0, 0x95, 0x33, 0x80, 0x40, 0x8d, 0x6e, 0x79,
		0x30, 0xb1, 0xc7, 0x1d, 0x53, 0xc1, 0xaa, 0x26, 0xd8, 0x1a, 0xd5, 0xea,
		0x8c, 0x37, 0xde, 0xe8, 0x16, 0x21, 0xe4, 0x03, 0xdf, 0xac, 0x95, 0xce,
		0xbd, 0xf1, 0x37, 0xbf, 0xf2, 0xc9, 0x44, 0xa2, 0x20, 0xd4, 0x42, 0x7c,
		0xfd, 0x9a, 0x38, 0x50, 0x20, 0x78, 0x14, 0x28, 0xd9, 0x43, 0x7d, 0x3a,
		0x36, 0xf5, 0x4a, 0x95, 0x90, 0xb0, 0x5d, 0x46, 0xbc, 0x5e, 0x9a, 0x0d,
		0x0d, 0x2b, 0xb6, 0x5c, 0x72, 0x99, 0xd3, 0xf8, 0x5c, 0xd4, 0x25, 0x69,
		0x74, 0x76, 0x71, 0x34, 0x03, 0x69, 0x56, 0x29, 0x96, 0x0b, 0x59, 0xa6,
		0xa5, 0x28, 0x8e, 0x48, 0x56, 0xb1, 0xa6, 0xb9, 0x38, 0xf2, 0xba, 0xa3,
		0x77, 0xb1, 0x87, 0x11, 0x1d, 0xa7, 0x00, 0x8e, 0xc6, 0x18, 0xfd, 0x8d,
		0x3f, 0x2d, 0x42, 0x5c, 0x73, 0x5b, 0x40, 0x97, 0x89, 0xab, 0xda, 0xdc,
		0xe1, 0xf4, 0x2a, 0x07, 0x75, 0xee, 0x21, 0x07, 0xa5, 0x45, 0x3d, 0xf7,
		0xe8, 0xbd, 0xd2, 0xd7, 0x68, 0x4e, 0xbc, 0xe0, 0xd5, 0xa1, 0x3c, 0x73,
		0x12, 0x24, 0xab, 0xdf, 0x26, 0x69, 0xae, 0x24, 0x1f, 0x70, 0x9a, 0x33,
		0xc3, 0x42, 0x21, 0x2c, 0xb3, 0x80, 0x15, 0x6a, 0xda, 0xd7, 0xc4, 0x57,
		0xc3, 0xc7, 0x8d, 0x84, 0xcf, 0x2c, 0x2a, 0x94, 0x26, 0x54, 0x80, 0xeb,
		0xc9, 0x5b, 0x22, 0xc8, 0x39, 0xc1, 0x40, 0x69, 0xc5, 0x65, 0x69, 0x16,
		0xa0, 0x98, 0x4c, 0xba, 0xa0, 0x11, 0xc4, 0x49, 0x7d, 0x15, 0xbd, 0x06,
		0x4b, 0x7e, 0xbe, 0xd4, 0xfc, 0xdd, 0xf9, 0x0c, 0x7f, 0xe1, 0x14, 0xc3,
		0x1f, 0x4d, 0x67, 0xc5, 0xd8, 0x35, 0xc3, 0x06, 0xfd, 0x45, 0x1a, 0xbd,
		0xb1, 0x18, 0xef, 0xc4, 0x7d, 0x12, 0xec, 0x41, 0xb0, 0x9d, 0x8c, 0x59,
		0x85, 0xdc, 0x0a, 0x26, 0xaa, 0x2e, 0x37, 0xaa, 0x79, 0xd3, 0x56, 0xa6,
		0xc3, 0xb1, 0x97, 0xd2, 0x08, 0xd4, 0xf3, 0x98, 0x5c, 0xb0, 0xb4, 0x31,
		0xcc, 0xb4, 0x0d, 0x99, 0x90, 0x78, 0x4e, 0x62, 0xf8, 0xb7, 0xa3, 0xbe,
		0x85, 0x0d, 0xaf, 0x76, 0x20, 0x7a, 0x84, 0x5b, 0xf7, 0x1f, 0x71, 0x6e,
		0xfb, 0x71, 0x93, 0x6a, 0xfd, 0xd4, 0xac, 0x85, 0x51, 0x1b, 0x8f, 0xad,
		0x0d, 0xe4, 0x42, 0x1c, 0xd3, 0xb5, 0x90, 0xb9, 0x5a, 0xe3, 0xc4, 0x37,
		0x38, 0xc7, 0xbb, 0x57, 0xc2, 0x68, 0xc2, 0xfd, 0xa6, 0xce, 0x67, 0x68,
		0xb5, 0x47, 0x1d, 0xa0, 0xde, 0x52, 0x0e, 0xd2, 0x47, 0x96, 0x63, 0x5b,
		0x83, 0x69, 0xc9, 0x74, 0xc3, 0x7f, 0x95, 0x86, 0x22, 0xa5, 0x59, 0xd3,
		0xd0, 0x78, 0xe9, 0x6c, 0x71, 0xe2, 0xbc, 0x3f, 0xab, 0x7c, 0x73, 0xcd,
		0x74, 0x29, 0xe4, 0xd0, 0x1b, 0x22, 0xa2, 0x01, 0xb3, 0xc4, 0x3d, 0xb5,
		0x75, 0x98, 0x1a, 0xb5, 0x0c, 0xdb, 0x5c, 0x26, 0xef, 0xb9, 0x28, 0x17,
		0xc6, 0x22, 0x08, 0xa9, 0x09, 0x29, 0xb9, 0x76, 0x7a, 0x3f, 0xda, 0xdd,
		0xc9, 0x0b, 0xab, 0x85, 0xb1, 0xa2, 0x3b, 0xbb, 0xa7, 0xe4, 0xec, 0x87,
		0x01, 0x0a, 0x5c, 0xf6, 0x29, 0x24, 0x48, 0xdc, 0xf2, 0x31, 0xb6, 0x05,
		0x19, 0x96, 0x63, 0xd4, 0x70, 0x78, 0x91, 0x76, 0x55, 0x31, 0x78, 0xc7,
		0x48, 0xbe, 0x26, 0x3f, 0x33, 0xc3, 0xe9, 0x35, 0x33, 0x8b, 0xb4, 0xa8,
		0x94, 0xd2, 0xe8, 0x95, 0xde, 0x36, 0x64, 0x46, 0x4e, 0x4f, 0xec, 0x5f,
		0x92, 0x20, 0x40, 0xcd, 0x4d, 0xab, 0x25, 0xb9, 0x31, 0x1a, 0xaf, 0x87,
		0xa5, 0x56, 0x46, 0x99, 0xcd, 0x92, 0xa7, 0x99, 0x92, 0x19, 0x73, 0xfd,
		0x64, 0x9a, 0xd4, 0xa8, 0x0f, 0x2a, 0x63, 0x15, 0xc7, 0x98, 0xce, 0xd5,
		0x8f, 0x58, 0x4c, 0xe2, 0x37, 0x23, 0x9f, 0x5b, 0x51, 0x8f, 0x7d, 0xee,
		0x9c, 0x13, 0x22, 0xb8, 0xd1, 0x99, 0x53, 0xde, 0x93, 0x81, 0xd2, 0xdf,
		0x0f, 0xd8, 0xa5, 0x9d, 0xf2, 0xba, 0x29, 0x5f, 0xb9, 0x96, 0xdc, 0x8e,
		0x5a, 0xa1, 0xeb, 0xb1, 0x2e, 0x69, 0xe8, 0xcd, 0x61, 0xd6, 0x5d, 0xd3,
		0xfe, 0x1f, 0x66, 0x91, 0x4f, 0xf1, 0xf2, 0x02, 0x36, 0x57, 0xac, 0xa2,
		0xcf, 0xe5, 0x09, 0x41, 0xf0, 0xfe, 0xbb, 0x6a, 0xab, 0xea, 0x2f, 0xf8,
		0xb4, 0x04, 0xf4, 0x53, 0x07, 0xe9, 0x0b, 0xd7, 0xea, 0xa3, 0xe6, 0x85,
		0x78, 0xa4, 0xde, 0xf1, 0x5a, 0x49, 0xb3, 0xa0, 0x48, 0xd2, 0xe9, 0xf7,
		0x5c, 0x1d, 0x2a, 0x9c, 0xb2, 0xa4, 0xc7, 0x66, 0xa0, 0x50, 0x2f, 0xc5,
		0xb6, 0x1f, 0xf2, 0x3d, 0x5c, 0xa0, 0xf0, 0x01, 0xf1, 0x27, 0xcf, 0x9f,
		0x04, 0x29, 0x64, 0x6b, 0xf8, 0x0b, 0x1c, 0x6f, 0x38, 0x1c, 0x87, 0x9f,
		0xa4, 0x01, 0xca, 0xae, 0xe8, 0x43, 0x77, 0x3f, 0xeb, 0xbe, 0x71, 0x28,
		0x95, 0x70, 0x9b, 0x9e, 0x9e, 0x24, 0xe4, 0x92, 0xc4, 0x27, 0x31, 0x01,
		0x06, 0x63, 0x2c, 0x89, 0xb4, 0xf3, 0x3f, 0xe2, 0xae, 0xfb, 0xfe, 0x76,
		0xc4, 0x21, 0x3f, 0x57, 0xb6, 0x9b, 0x81, 0xbf, 0xd9, 0xdf, 0x9f, 0xf2,
		0xaf, 0x3f, 0x6e, 0xa7, 0xf0, 0x7b, 0xe6, 0x7f, 0x8f, 0x67, 0xae, 0xbd,
		0xa0, 0x54, 0xbb, 0x6e, 0x67, 0xdb, 0xf9, 0xe0, 0xd7, 0xbb, 0x61, 0x34,
		0x77, 0x25, 0x0c, 0x89, 0xef, 0x22, 0x78, 0xd3, 0xa0, 0xee, 0xdd, 0xae,
		0x3f, 0x59, 0x05, 0x56, 0x94, 0x2c, 0x15, 0xfd, 0x1e, 0x67, 0x40, 0x69,
		0x60, 0x50, 0x0f, 0xb6, 0x3c, 0xf8, 0xd9, 0xfe, 0x5f, 0x9f, 0x01, 0xdc,
		0xc0, 0x8d, 0xa1, 0x3e, 0x5a, 0xe2, 0xcb, 0x64, 0x43, 0xc2, 0x38, 0xff,
		0x84, 0xdf, 0x5c, 0x1a, 0x73, 0xad, 0x95, 0x8e, 0xdd, 0xe5, 0x48, 0x38,
		0xbc, 0x1d, 0x06, 0x5e, 0x9a, 0xd7, 0x6a, 0xc5, 0xf7, 0x1d, 0xe1, 0x38,
		0x84, 0x00, 0x8f, 0x10, 0x57, 0xd3, 0x70, 0x72, 0x5f, 0x14, 0x77, 0xb2,
		0x87, 0x1b, 0x4e, 0xb6, 0x98, 0xbf, 0x7b, 0xb2, 0xf5, 0xfa, 0x17, 0x27,
		0x07, 0xda, 0xd5, 0x03, 0x32, 0xee, 0x93, 0x0d, 0x9d, 0x8a, 0xb7, 0x01,
		0x09, 0x30, 0xa0, 0x15, 0x56, 0x4a, 0xe4, 0xf0, 0xc1, 0xdd, 0xef, 0x02,
		0xff, 0xdc, 0xea, 0xda, 0xa0, 0xe8, 0xb9, 0xbd, 0x9b, 0x4c, 0xef, 0x2f,
		0x3f, 0xe5, 0x13, 0x4f, 0x69, 0x78, 0xa3, 0x21, 0x71, 0x56, 0xf6, 0xa4,
		0x59, 0xd9, 0x91, 0x63, 0xc5, 0x01, 0x3b, 0xad, 0xaa, 0x9d, 0xa1, 0xdb,
		0x33, 0x6d, 0xa5, 0x30, 0xd3, 0x66, 0x9e, 0x2d, 0x78, 0xf6, 0xc0, 0x73,
		0x3f, 0x6f, 0x09, 0xf9, 0xf6, 0x8d, 0xb8, 0x0f, 0xdf, 0xd8, 0xb5, 0x3e,
		0xe0, 0x7a, 0xd8, 0x73, 0xf1, 0xf2, 0xa0, 0xf9, 0xd8, 0xb5, 0x6f, 0xa2,
		0x62, 0x40, 0x63, 0xc8, 0x2c, 0xf0, 0xe8, 0xd2, 0x3b, 0x48, 0x64, 0x60,
		0xa3, 0xab, 0xf3, 0x0e, 0xb5, 0x6e, 0xe3, 0x53, 0xdc, 0x0e, 0xd9, 0xec,
		0x8a, 0x39, 0x21, 0xae, 0x78, 0x07, 0x28, 0xc3, 0xd7, 0x2e, 0xbc, 0xef,
		0x0e, 0x72, 0xd6, 0xd3, 0xd5, 0xd8, 0xc7, 0xb1, 0x73, 0xbc, 0xb4, 0x05,
		0x00, 0x71, 0x0a, 0xda, 0x18, 0xfb, 0xc1, 0xae, 0xd9, 0xa3, 0x5b, 0x07,
		0x7f, 0xc7, 0x15, 0x08, 0xcf, 0x15, 0xc5, 0xf9, 0x25, 0xdd, 0x3b, 0x63,
		0xf5, 0x5f, 0x0a, 0x82, 0xdb, 0x5e, 0x52, 0x0e, 0x77, 0xd8, 0xa1, 0x1a,
		0x84, 0x47, 0xbd, 0x8d, 0xe7, 0xbd, 0x31, 0x29, 0xff, 0x9a, 0xed, 0x89,
		0xdd, 0xdb, 0xda, 0x3f, 0xf9, 0xc7, 0x7b, 0xc3, 0x8b, 0x77, 0x67, 0xf3,
		0x36, 0x01, 0xe1, 0x9f, 0x00, 0x00, 0x00, 0xff, 0xff, 0xfa, 0xcb, 0xf2,
		0xb7, 0x37, 0x0d, 0x00, 0x00,
		},
		"content/js/app.js",
	)
}


// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	if f, ok := _bindata[name]; ok {
		return f()
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string] func() ([]byte, error) {
	"content/css/styles.css": content_css_styles_css,
	"content/img/loading.gif": content_img_loading_gif,
	"content/index.html": content_index_html,
	"content/js/app.js": content_js_app_js,

}