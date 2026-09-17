package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"hash/crc32"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func fixtureJPEG(t *testing.T) []byte {
	t.Helper()
	var out bytes.Buffer
	if err := jpeg.Encode(&out, image.NewRGBA(image.Rect(0, 0, 2, 2)), nil); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func TestImagePayloadIntegrity(t *testing.T) {
	png := fixturePNG(t)
	noPixels := append(append([]byte{}, png[:33]...), png[len(png)-12:]...)
	badCRC := append([]byte{}, png...)
	badCRC[len(badCRC)-13] ^= 1
	large := append([]byte{}, png...)
	binary.BigEndian.PutUint32(large[16:20], 100000)
	binary.BigEndian.PutUint32(large[20:24], 100000)
	binary.BigEndian.PutUint32(large[29:33], crc32.ChecksumIEEE(large[12:29]))
	jpg := fixtureJPEG(t)
	webp, err := base64.StdEncoding.DecodeString("UklGRh4AAABXRUJQVlA4TBEAAAAvAAAAAAfQ//73v/+BiOh/AAA=")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, mediaType string
		body            []byte
		valid           bool
	}{
		{"png", "image/png", png, true},
		{"jpeg", "image/jpeg", jpg, true},
		{"webp-lossless", "image/webp", webp, true},
		{"png-without-idat", "image/png", noPixels, false},
		{"png-corrupt-payload", "image/png", badCRC, false},
		{"png-oversized-dimensions", "image/png", large, false},
		{"jpeg-without-scan", "image/jpeg", []byte{255, 216, 255, 217}, false},
		{"jpeg-truncated-scan", "image/jpeg", append(append([]byte{}, jpg[:len(jpg)-10]...), 255, 217), false},
		{"webp-without-payload", "image/webp", []byte("RIFF\x0c\x00\x00\x00WEBPVP8 \x00\x00\x00\x00"), false},
		{"webp-truncated", "image/webp", webp[:len(webp)-1], false},
		{"webp-trailing-bytes", "image/webp", append(append([]byte{}, webp...), 0), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "result")
			if err := os.WriteFile(file, tc.body, 0600); err != nil {
				t.Fatal(err)
			}
			if got := validMediaFile(file, tc.mediaType); got != tc.valid {
				t.Errorf("validMediaFile = %v, want %v", got, tc.valid)
			}
			_, err := fingerprintDownload(file, tc.mediaType)
			if (err == nil) != tc.valid {
				t.Errorf("fingerprintDownload accepted invalid content: %v", err)
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", tc.mediaType)
				w.Write(tc.body)
			}))
			defer server.Close()
			request := taskRequest{Kind: "image", TaskID: "same-task", TaskStatus: "completed", RequestedCount: 1, OutputDir: t.TempDir()}
			data, _ := json.Marshal(request)
			var out bytes.Buffer
			err = executeExistingTask("content", bytes.NewReader(data), &out, fixtureService(server))
			result := decodeReceipt(t, &out)
			if tc.valid {
				if err != nil || result.NextStep != "deliver" || len(result.DownloadedPaths) != 1 {
					t.Fatalf("valid image failed: %+v %v", result, err)
				}
			} else if err == nil || result.LocalErrorCode != "invalid_media_content" || len(result.DownloadedPaths) != 0 || result.TaskID != "same-task" {
				t.Fatalf("invalid image accepted or task lost: %+v %v", result, err)
			}
		})
	}
}

func BenchmarkPNGIntegrity(b *testing.B) {
	var encoded bytes.Buffer
	// A generated bitmap fixture, never a downloaded user image.
	if err := png.Encode(&encoded, image.NewRGBA(image.Rect(0, 0, 2048, 2048))); err != nil {
		b.Fatal(err)
	}
	file := filepath.Join(b.TempDir(), "result.png")
	if err := os.WriteFile(file, encoded.Bytes(), 0600); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !validMediaFile(file, "image/png") {
			b.Fatal("valid image rejected")
		}
	}
}
