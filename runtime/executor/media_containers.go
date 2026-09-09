package main

import (
	"bufio"
	"bytes"
	"compress/lzw"
	"encoding/binary"
	"errors"
	"io"
)

// Validate all GIF blocks and each frame's bounded LZW payload without retaining
// decoded bitmaps. A complete trailer alone is not evidence of complete frames.
func validGIF(input io.Reader) bool {
	r := bufio.NewReader(io.LimitReader(input, maxImageBytes+1))
	var header [13]byte
	if _, err := io.ReadFull(r, header[:]); err != nil || !validMediaPrefix(header[:], "image/gif") {
		return false
	}
	width, height := int(binary.LittleEndian.Uint16(header[6:8])), int(binary.LittleEndian.Uint16(header[8:10]))
	if width == 0 || height == 0 {
		return false
	}
	skip := func(n int64) bool { copied, err := io.CopyN(io.Discard, r, n); return err == nil && copied == n }
	if header[10]&0x80 != 0 && !skip(int64(3<<(1+(header[10]&7)))) {
		return false
	}
	subBlocks := func() ([]byte, bool) {
		var body bytes.Buffer
		for {
			size, err := r.ReadByte()
			if err != nil {
				return nil, false
			}
			if size == 0 {
				return body.Bytes(), true
			}
			if body.Len()+int(size) > int(maxImageBytes) {
				return nil, false
			}
			if _, err := io.CopyN(&body, r, int64(size)); err != nil {
				return nil, false
			}
		}
	}
	frames, pixels := 0, int64(0)
	for {
		tag, err := r.ReadByte()
		if err != nil {
			return false
		}
		switch tag {
		case 0x3b:
			_, err := r.ReadByte()
			return frames > 0 && err == io.EOF
		case 0x21:
			if _, err := r.ReadByte(); err != nil {
				return false
			}
			if _, ok := subBlocks(); !ok {
				return false
			}
		case 0x2c:
			var frame [9]byte
			if _, err := io.ReadFull(r, frame[:]); err != nil {
				return false
			}
			x, y := int(binary.LittleEndian.Uint16(frame[:2])), int(binary.LittleEndian.Uint16(frame[2:4]))
			w, h := int(binary.LittleEndian.Uint16(frame[4:6])), int(binary.LittleEndian.Uint16(frame[6:8]))
			count := int64(w) * int64(h)
			pixels += count
			// Bound decompression work as well as compressed file size.
			if count == 0 || x+w > width || y+h > height || pixels > 256<<20 {
				return false
			}
			if frame[8]&0x80 != 0 && !skip(int64(3<<(1+(frame[8]&7)))) {
				return false
			}
			codeSize, err := r.ReadByte()
			if err != nil || codeSize < 2 || codeSize > 8 {
				return false
			}
			compressed, ok := subBlocks()
			if !ok {
				return false
			}
			decoder := lzw.NewReader(bytes.NewReader(compressed), lzw.LSB, int(codeSize))
			decoded, err := io.Copy(io.Discard, io.LimitReader(decoder, count+1))
			decoder.Close()
			if err != nil || decoded != count {
				return false
			}
			frames++
		default:
			return false
		}
	}
}

type ebmlElement struct {
	id, value  uint64
	start, end int64
	unknown    bool
}

func ebmlVint(r io.ReaderAt, offset, end int64, id bool) (uint64, int64, bool, error) {
	var data [8]byte
	if offset >= end {
		return 0, offset, false, io.ErrUnexpectedEOF
	}
	if _, err := r.ReadAt(data[:1], offset); err != nil {
		return 0, offset, false, err
	}
	mask, length := byte(0x80), 1
	for mask != 0 && data[0]&mask == 0 {
		mask >>= 1
		length++
	}
	if mask == 0 || (id && length > 4) || offset+int64(length) > end {
		return 0, offset, false, errors.New("invalid EBML integer")
	}
	if _, err := r.ReadAt(data[:length], offset); err != nil {
		return 0, offset, false, err
	}
	value := uint64(data[0] & (mask - 1))
	unknown := value == uint64(mask-1)
	if id {
		value = uint64(data[0])
	}
	for _, b := range data[1:length] {
		value = value<<8 | uint64(b)
		unknown = unknown && b == 255
	}
	return value, offset + int64(length), !id && unknown, nil
}

func readEBML(r io.ReaderAt, offset, end int64) (ebmlElement, bool) {
	id, at, _, err := ebmlVint(r, offset, end, true)
	if err != nil {
		return ebmlElement{}, false
	}
	size, at, unknown, err := ebmlVint(r, at, end, false)
	if err != nil || (!unknown && size > uint64(end-at)) {
		return ebmlElement{}, false
	}
	element := ebmlElement{id: id, start: at, end: end, unknown: unknown}
	if !unknown {
		element.end = at + int64(size)
	}
	if size > 0 && size <= 8 && !unknown {
		var data [8]byte
		if _, err := r.ReadAt(data[:size], at); err != nil {
			return ebmlElement{}, false
		}
		for _, b := range data[:size] {
			element.value = element.value<<8 | uint64(b)
		}
	}
	return element, true
}

// WebM container checks: exact DocType, bounded element sizes, a declared video
// track, and complete blocks referencing it. This is not a codec/playback test.
// Unknown-length Segment/Cluster containers are supported for streamed WebM.
func validWebM(r io.ReaderAt, end int64) bool {
	header, ok := readEBML(r, 0, end)
	if !ok || header.id != 0x1a45dfa3 || header.unknown || header.end > 4096 {
		return false
	}
	docType := false
	for at := header.start; at < header.end; {
		el, ok := readEBML(r, at, header.end)
		if !ok || el.unknown {
			return false
		}
		if el.id == 0x4282 {
			var name [4]byte
			_, err := r.ReadAt(name[:], el.start)
			docType = err == nil && el.end-el.start == 4 && string(name[:]) == "webm"
		}
		at = el.end
	}
	segment, ok := readEBML(r, header.end, end)
	if !docType || !ok || segment.id != 0x18538067 || segment.end != end {
		return false
	}
	videoTracks := map[uint64]bool{}
	blockTracks := map[uint64]bool{}
	budget := 1000000
	for at := segment.start; at < segment.end; {
		el, ok := readEBML(r, at, segment.end)
		budget--
		if !ok || budget <= 0 || (el.unknown && el.id != 0x1f43b675) {
			return false
		}
		switch el.id {
		case 0x1654ae6b: // Tracks
			for pos := el.start; pos < el.end; {
				entry, ok := readEBML(r, pos, el.end)
				budget--
				if !ok || entry.unknown || budget <= 0 {
					return false
				}
				if entry.id == 0xae {
					var number, kind uint64
					codec := false
					for fieldAt := entry.start; fieldAt < entry.end; {
						field, ok := readEBML(r, fieldAt, entry.end)
						budget--
						if !ok || field.unknown || budget <= 0 {
							return false
						}
						switch field.id {
						case 0xd7:
							number = field.value
						case 0x83:
							kind = field.value
						case 0x86:
							var prefix [2]byte
							_, err := r.ReadAt(prefix[:], field.start)
							codec = err == nil && field.end-field.start > 2 && string(prefix[:]) == "V_"
						}
						fieldAt = field.end
					}
					if number > 0 && kind == 1 && codec {
						videoTracks[number] = true
					}
				}
				pos = entry.end
			}
		case 0x1f43b675: // Cluster
			stop, valid := webMCluster(r, el, blockTracks, &budget)
			if !valid {
				return false
			}
			el.end = stop
		}
		at = el.end
	}
	for track := range blockTracks {
		if videoTracks[track] {
			return true
		}
	}
	return false
}

func webMCluster(r io.ReaderAt, cluster ebmlElement, tracks map[uint64]bool, budget *int) (int64, bool) {
	block := func(el ebmlElement) bool {
		track, at, unknown, err := ebmlVint(r, el.start, el.end, false)
		if err != nil || unknown || track == 0 || el.end-at <= 3 {
			return false
		}
		tracks[track] = true
		return true
	}
	for at := cluster.start; at < cluster.end; {
		el, ok := readEBML(r, at, cluster.end)
		*budget--
		if !ok || *budget <= 0 {
			return at, false
		}
		switch el.id {
		case 0x1f43b675, 0x1c53bb6b, 0x1654ae6b, 0x1549a966, 0x114d9b74, 0x1254c367:
			return at, cluster.unknown
		}
		if el.unknown {
			return at, false
		}
		if el.id == 0xa3 && !block(el) {
			return at, false
		}
		if el.id == 0xa0 {
			for pos := el.start; pos < el.end; {
				child, ok := readEBML(r, pos, el.end)
				*budget--
				if !ok || child.unknown || *budget <= 0 || (child.id == 0xa1 && !block(child)) {
					return pos, false
				}
				pos = child.end
			}
		}
		at = el.end
	}
	return cluster.end, true
}
