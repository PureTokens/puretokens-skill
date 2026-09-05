package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
)

// Check a supported binary format, not merely the server's Content-Type.
func validMediaPrefix(data []byte, contentType string) bool {
	switch contentType {
	case "image/png":
		return len(data) >= 24 && bytes.Equal(data[:8], []byte{137, 80, 78, 71, 13, 10, 26, 10}) && string(data[12:16]) == "IHDR" && binary.BigEndian.Uint32(data[16:20]) > 0 && binary.BigEndian.Uint32(data[20:24]) > 0
	case "image/jpeg":
		return len(data) >= 4 && data[0] == 0xff && data[1] == 0xd8 && data[2] == 0xff
	case "image/gif":
		return len(data) >= 13 && (string(data[:6]) == "GIF87a" || string(data[:6]) == "GIF89a")
	case "image/webp":
		return len(data) >= 20 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WEBP"
	case "image/avif":
		return len(data) >= 16 && string(data[4:8]) == "ftyp" && (bytes.Contains(data[8:], []byte("avif")) || bytes.Contains(data[8:], []byte("avis")))
	case "video/mp4":
		return len(data) >= 24 && string(data[4:8]) == "ftyp"
	case "video/webm":
		return len(data) >= 16 && bytes.Equal(data[:4], []byte{0x1a, 0x45, 0xdf, 0xa3}) && bytes.Contains(data, []byte("webm"))
	}
	return false
}

func validMediaFile(path, contentType string) bool {
	if _, err := os.Lstat(path + ".incomplete"); !os.IsNotExist(err) {
		return false
	}
	before, err := os.Lstat(path)
	if err != nil || !before.Mode().IsRegular() {
		return false
	}
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil || !stat.Mode().IsRegular() || !os.SameFile(before, stat) {
		return false
	}
	prefix := make([]byte, 512)
	n, _ := f.Read(prefix)
	if !validMediaPrefix(prefix[:n], contentType) {
		return false
	}
	// Cheap container termination checks avoid accepting common truncated files,
	// without decoding a large bitmap/video into RAM.
	if contentType == "image/png" {
		if stat.Size() < 45 {
			return false
		}
		tail := make([]byte, 12)
		_, err = f.ReadAt(tail, stat.Size()-12)
		return err == nil && bytes.Equal(tail, []byte{0, 0, 0, 0, 73, 69, 78, 68, 174, 66, 96, 130})
	}
	if contentType == "image/jpeg" {
		tail := make([]byte, 2)
		_, err = f.ReadAt(tail, stat.Size()-2)
		return err == nil && bytes.Equal(tail, []byte{255, 217})
	}
	if contentType == "image/webp" {
		return stat.Size() >= int64(binary.LittleEndian.Uint32(prefix[4:8]))+8
	}
	if contentType == "video/mp4" || contentType == "image/avif" {
		var offset int64
		hasMetadata, hasPayload := false, false
		for offset < stat.Size() {
			var header [16]byte
			if _, err = f.ReadAt(header[:8], offset); err != nil {
				return false
			}
			boxType := string(header[4:8])
			if boxType == "moov" || boxType == "moof" || boxType == "meta" {
				hasMetadata = true
			}
			if boxType == "mdat" {
				hasPayload = true
			}
			size := int64(binary.BigEndian.Uint32(header[:4]))
			minSize := int64(8)
			if size == 1 {
				if _, err = f.ReadAt(header[8:], offset+8); err != nil {
					return false
				}
				raw := binary.BigEndian.Uint64(header[8:])
				if raw > uint64(stat.Size()) {
					return false
				}
				size = int64(raw)
				minSize = 16
			}
			if size == 0 {
				return stat.Size()-offset > 8 && hasMetadata && (hasPayload || contentType == "image/avif")
			}
			if size < minSize || size > stat.Size()-offset {
				return false
			}
			offset += size
		}
		return hasMetadata && (hasPayload || contentType == "image/avif")
	}
	_, err = f.Seek(0, io.SeekStart)
	return err == nil
}

// Filesystems such as exFAT cannot hard-link. The transient marker prevents
// reuse until copying has completed, including after an interrupted process.
func copyDownloadExclusive(source, destination string) (err error) {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	marker := destination + ".incomplete"
	if err := os.Mkdir(marker, 0700); err != nil {
		return err
	}
	removeMarker := true
	defer func() {
		if !removeMarker {
			return
		}
		if removeErr := os.Remove(marker); err == nil {
			err = removeErr
		}
	}()
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if removeErr := os.Remove(destination); removeErr != nil && !os.IsNotExist(removeErr) {
				removeMarker = false
			}
		}
	}()
	_, copyErr := io.Copy(output, input)
	closeErr := output.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}
