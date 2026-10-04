package main

import (
	"bytes"
	"encoding/binary"
)

func audioMIME(format string) string {
	switch format {
	case "mp3":
		return "audio/mpeg"
	case "wav":
		return "audio/wav"
	case "ogg":
		return "audio/ogg"
	}
	return ""
}

// Structural validation proves complete bounded frames/chunks, not audible
// quality or host playback. Never accept a MIME label alone.
func validAudioBytes(data []byte, format string) bool {
	if len(data) == 0 || len(data) > maxAudioOutput {
		return false
	}
	switch format {
	case "wav":
		return validWAV(data)
	case "mp3":
		return validMP3(data)
	case "ogg":
		return validOggAudio(data)
	}
	return false
}

func validWAV(data []byte) bool {
	if len(data) < 44 || string(data[:4]) != "RIFF" || string(data[8:12]) != "WAVE" || uint64(binary.LittleEndian.Uint32(data[4:8]))+8 != uint64(len(data)) {
		return false
	}
	formatSeen, payload := false, false
	blockAlign := uint16(0)
	for offset := 12; offset < len(data); {
		if len(data)-offset < 8 {
			return false
		}
		size := uint64(binary.LittleEndian.Uint32(data[offset+4 : offset+8]))
		end := uint64(offset) + 8 + size
		if end > uint64(len(data)) {
			return false
		}
		chunk := data[offset+8 : int(end)]
		switch string(data[offset : offset+4]) {
		case "fmt ":
			if formatSeen || len(chunk) < 16 {
				return false
			}
			code := binary.LittleEndian.Uint16(chunk)
			channels := binary.LittleEndian.Uint16(chunk[2:])
			rate := binary.LittleEndian.Uint32(chunk[4:])
			bits := binary.LittleEndian.Uint16(chunk[14:])
			blockAlign = binary.LittleEndian.Uint16(chunk[12:])
			if (code != 1 && code != 3) || channels < 1 || channels > 8 || rate < 8000 || rate > 192000 || (bits != 8 && bits != 16 && bits != 24 && bits != 32) ||
				(code == 3 && bits != 32) || blockAlign != channels*(bits/8) || binary.LittleEndian.Uint32(chunk[8:]) != rate*uint32(blockAlign) {
				return false
			}
			formatSeen = true
		case "data":
			if !formatSeen || payload || size == 0 || size%uint64(blockAlign) != 0 {
				return false
			}
			payload = true
		}
		offset = int(end) + int(size%2)
		if offset > len(data) {
			return false
		}
	}
	return formatSeen && payload
}

func validMP3(data []byte) bool {
	offset := 0
	if len(data) >= 10 && string(data[:3]) == "ID3" {
		if data[3] < 2 || data[3] > 4 || data[4] == 255 {
			return false
		}
		size := 0
		for _, b := range data[6:10] {
			if b >= 128 {
				return false
			}
			size = (size << 7) | int(b)
		}
		offset = 10 + size
		if data[3] == 4 && data[5]&0x10 != 0 {
			offset += 10
		}
		if offset > len(data) {
			return false
		}
	}
	frames := 0
	for offset < len(data) {
		if len(data)-offset == 128 && string(data[offset:offset+3]) == "TAG" {
			return frames >= 2
		}
		if len(data)-offset < 4 {
			return false
		}
		h := binary.BigEndian.Uint32(data[offset : offset+4])
		version, layer, bitrate, sample := int(h>>19&3), int(h>>17&3), int(h>>12&15), int(h>>10&3)
		if h>>21 != 0x7ff || version == 1 || layer != 1 || bitrate == 0 || bitrate == 15 || sample == 3 {
			return false
		}
		rate := []int{44100, 48000, 32000}[sample]
		kbps := []int{0, 32, 40, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320}[bitrate]
		multiplier := 144
		if version != 3 {
			rate /= 2
			multiplier = 72
			kbps = []int{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160}[bitrate]
		}
		if version == 0 {
			rate /= 2
		}
		size := multiplier*kbps*1000/rate + int(h>>9&1)
		if size < 4 || size > len(data)-offset {
			return false
		}
		offset += size
		frames++
	}
	return frames >= 2
}

func validOggAudio(data []byte) bool {
	offset, sequence, packets := 0, uint32(0), 0
	var serial uint32
	var pending []byte
	codec, ended := "", false
	for offset < len(data) {
		if ended || len(data)-offset < 27 || string(data[offset:offset+4]) != "OggS" || data[offset+4] != 0 {
			return false
		}
		header := data[offset : offset+27]
		count := int(header[26])
		start := offset + 27 + count
		if start > len(data) {
			return false
		}
		size := 0
		for _, n := range data[offset+27 : start] {
			size += int(n)
		}
		end := start + size
		if end > len(data) || header[5]&^byte(7) != 0 {
			return false
		}
		if sequence == 0 {
			serial = binary.LittleEndian.Uint32(header[14:])
			if header[5]&2 == 0 {
				return false
			}
		} else if header[5]&2 != 0 {
			return false
		}
		if binary.LittleEndian.Uint32(header[14:]) != serial || binary.LittleEndian.Uint32(header[18:]) != sequence || (header[5]&1 != 0) != (len(pending) > 0) {
			return false
		}
		crc := uint32(0)
		for index, b := range data[offset:end] {
			if index >= 22 && index < 26 {
				b = 0
			}
			crc ^= uint32(b) << 24
			for bit := 0; bit < 8; bit++ {
				if crc&0x80000000 != 0 {
					crc = (crc << 1) ^ 0x04c11db7
				} else {
					crc <<= 1
				}
			}
		}
		if crc != binary.LittleEndian.Uint32(header[22:]) {
			return false
		}
		position := start
		for _, n := range data[offset+27 : start] {
			pending = append(pending, data[position:position+int(n)]...)
			position += int(n)
			if n < 255 {
				if packets == 0 {
					switch {
					case len(pending) == 19 && bytes.HasPrefix(pending, []byte("OpusHead")) &&
						pending[8] == 1 && pending[9] >= 1 && pending[9] <= 2 && pending[18] == 0:
						codec = "opus"
					case len(pending) == 30 && bytes.HasPrefix(pending, []byte{1, 'v', 'o', 'r', 'b', 'i', 's'}) &&
						binary.LittleEndian.Uint32(pending[7:]) == 0 && pending[11] >= 1 &&
						binary.LittleEndian.Uint32(pending[12:]) > 0 && pending[29] == 1:
						small, large := pending[28]&15, pending[28]>>4
						if small < 6 || large > 13 || small > large {
							return false
						}
						codec = "vorbis"
					default:
						return false
					}
				} else if packets == 1 {
					prefix := []byte("OpusTags")
					if codec == "vorbis" {
						prefix = []byte{3, 'v', 'o', 'r', 'b', 'i', 's'}
					}
					if !validOggComment(pending, prefix, codec == "vorbis") {
						return false
					}
				} else if packets == 2 && codec == "vorbis" {
					if len(pending) < 8 || !bytes.HasPrefix(pending, []byte{5, 'v', 'o', 'r', 'b', 'i', 's'}) {
						return false
					}
				} else if len(pending) == 0 {
					return false
				}
				packets++
				pending = nil
			}
		}
		ended = header[5]&4 != 0
		offset = end
		sequence++
	}
	return ended && len(pending) == 0 && ((codec == "opus" && packets >= 3) || (codec == "vorbis" && packets >= 4))
}

func validOggComment(data, prefix []byte, framing bool) bool {
	if !bytes.HasPrefix(data, prefix) || len(data) < len(prefix)+8 {
		return false
	}
	offset := uint64(len(prefix))
	vendor := uint64(binary.LittleEndian.Uint32(data[offset:]))
	offset += 4 + vendor
	if offset+4 > uint64(len(data)) {
		return false
	}
	count := binary.LittleEndian.Uint32(data[offset:])
	offset += 4
	for i := uint32(0); i < count; i++ {
		if offset+4 > uint64(len(data)) {
			return false
		}
		size := uint64(binary.LittleEndian.Uint32(data[offset:]))
		offset += 4 + size
		if offset > uint64(len(data)) {
			return false
		}
	}
	return !framing || (offset+1 == uint64(len(data)) && data[offset] == 1)
}
