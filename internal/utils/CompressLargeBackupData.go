package utils

import (
	"bytes"
	"compress/gzip"
)

// When the backup data is too large(larger than gzipThreshold), we will need to compress it before sending it to the client.
// CompressLargeBackupData() takes a []byte of data and compresses it using gzip, returning a bytes.Buffer containing the compressed data.
func CompressLargeBackupData(data []byte) (bytes.Buffer, error) {
	// The type of data is []byte, so we can use bytes.Buffer to write the compressed data to a buffer and then send it as a stream

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		gz.Close()
		return bytes.Buffer{}, err
	}
	gz.Close()
	return buf, nil
}
