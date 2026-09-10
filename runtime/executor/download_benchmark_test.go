package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
)

// Structurally valid container for I/O measurement, not a playable-video
// fixture or evidence of real-host attachment delivery.
func structuralVideoBytes(size int) []byte {
	data := make([]byte, size)
	binary.BigEndian.PutUint32(data[:4], 24)
	copy(data[4:8], "ftyp")
	copy(data[8:12], "isom")
	binary.BigEndian.PutUint32(data[24:28], 8)
	copy(data[28:32], "moov")
	binary.BigEndian.PutUint32(data[32:36], uint32(size-32))
	copy(data[36:40], "mdat")
	for i := 40; i < size; i += 4096 {
		data[i] = byte(i / 4096)
	}
	return data
}

func localDownloadService(data []byte) service {
	return service{
		baseURL: "https://fixture.invalid",
		client: &http.Client{Transport: pollingFixtureTransport(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode:    200,
				Header:        http.Header{"Content-Type": []string{"video/mp4"}},
				ContentLength: int64(len(data)),
				Body:          io.NopCloser(bytes.NewReader(data)),
			}, nil
		})},
		downloadProofs: map[string]downloadProof{},
	}
}

// Explicit only: go test -run '^$' -bench BenchmarkDownloadPipeline -benchtime=3x
// Large media allocations and disk measurement are not part of npm run check.
func BenchmarkDownloadPipeline(b *testing.B) {
	for _, mib := range []int{100, 512} {
		b.Run(fmt.Sprintf("%dMiB", mib), func(b *testing.B) {
			svc := localDownloadService(structuralVideoBytes(mib << 20))
			dir := b.TempDir()
			b.SetBytes(int64(mib << 20))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				path, _, _, _, _, err := svc.download(context.Background(), contentPath("video", "bench", i), "video", dir)
				if err != nil {
					b.Fatal(err)
				}
				b.StopTimer()
				if err := os.Remove(path); err != nil {
					b.Fatal(err)
				}
				delete(svc.downloadProofs, path)
				b.StartTimer()
			}
		})
	}
}

func TestStreamedDownloadRejectsSameSizeTimestampReplacement(t *testing.T) {
	svc := localDownloadService(structuralVideoBytes(1 << 20))
	dir, route := t.TempDir(), contentPath("video", "tamper", 0)
	path, _, _, _, _, err := svc.download(context.Background(), route, "video", dir)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(path, os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.WriteAt([]byte{99}, 500)
	closeErr := f.Close()
	if err != nil || closeErr != nil {
		t.Fatalf("fixture modification: %v %v", err, closeErr)
	}
	if err := os.Chtimes(path, info.ModTime(), info.ModTime()); err != nil {
		t.Fatal(err)
	}
	if !validMediaFile(path, "video/mp4") || matchesDownloadProof(path, svc.downloadProofs[path]) {
		t.Fatal("same-size, same-time content replacement was not detected")
	}
	if _, _, _, _, _, err := svc.download(context.Background(), route, "video", dir); err == nil {
		t.Fatal("modified output was reused")
	}
}
