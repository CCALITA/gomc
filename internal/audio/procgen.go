package audio

import (
	"encoding/binary"
	"math"
	"math/rand/v2"
)

// GenerateNoiseBurst produces a white-noise burst suitable for block break/place
// sounds. The returned slice contains a valid WAV file (16-bit mono LE PCM).
func GenerateNoiseBurst(duration float64, sampleRate int, freq float64) []byte {
	numSamples := int(duration * float64(sampleRate))
	pcm := make([]int16, numSamples)

	// Amplitude envelope: quick attack, exponential decay.
	for i := range pcm {
		t := float64(i) / float64(sampleRate)
		envelope := math.Exp(-t * freq)
		noise := rand.Float64()*2.0 - 1.0 // [-1, 1]
		pcm[i] = int16(noise * envelope * math.MaxInt16)
	}

	return encodePCMToWAV(pcm, sampleRate)
}

// GenerateSineBeep produces a pure sine-wave tone suitable for UI feedback
// sounds (level-up, menu click). The returned slice contains a valid WAV file.
func GenerateSineBeep(duration float64, sampleRate int, freq float64) []byte {
	numSamples := int(duration * float64(sampleRate))
	pcm := make([]int16, numSamples)

	for i := range pcm {
		t := float64(i) / float64(sampleRate)
		// Fade-in / fade-out envelope to avoid clicks.
		envelope := sineEnvelope(float64(i), float64(numSamples))
		sample := math.Sin(2.0 * math.Pi * freq * t) * envelope
		pcm[i] = int16(sample * math.MaxInt16)
	}

	return encodePCMToWAV(pcm, sampleRate)
}

// GenerateChirp produces a frequency sweep from startFreq to endFreq, useful
// for hurt, eat, and splash sounds. The returned slice contains a valid WAV file.
func GenerateChirp(duration float64, sampleRate int, startFreq, endFreq float64) []byte {
	numSamples := int(duration * float64(sampleRate))
	pcm := make([]int16, numSamples)

	for i := range pcm {
		t := float64(i) / float64(sampleRate)
		progress := float64(i) / float64(numSamples)
		// Linear interpolation of frequency.
		currentFreq := startFreq + (endFreq-startFreq)*progress
		envelope := sineEnvelope(float64(i), float64(numSamples))
		sample := math.Sin(2.0 * math.Pi * currentFreq * t) * envelope
		pcm[i] = int16(sample * math.MaxInt16)
	}

	return encodePCMToWAV(pcm, sampleRate)
}

// sineEnvelope returns a smooth fade-in/fade-out value for the given position
// within a buffer of total samples. The shape is a half-sine so the amplitude
// ramps from 0 -> 1 -> 0 without clicks.
func sineEnvelope(pos, total float64) float64 {
	if total <= 0 {
		return 0
	}
	return math.Sin(math.Pi * pos / total)
}

// encodePCMToWAV wraps raw 16-bit mono PCM samples in a valid WAV container.
func encodePCMToWAV(pcm []int16, sampleRate int) []byte {
	const (
		bitsPerSample  = 16
		numChannels    = 1
		bytesPerSample = bitsPerSample / 8
	)

	dataSize := len(pcm) * bytesPerSample
	// WAV header is 44 bytes.
	buf := make([]byte, 44+dataSize)

	// RIFF header
	copy(buf[0:4], "RIFF")
	binary.LittleEndian.PutUint32(buf[4:8], uint32(36+dataSize))
	copy(buf[8:12], "WAVE")

	// fmt sub-chunk
	copy(buf[12:16], "fmt ")
	binary.LittleEndian.PutUint32(buf[16:20], 16) // sub-chunk size
	binary.LittleEndian.PutUint16(buf[20:22], 1)  // PCM format
	binary.LittleEndian.PutUint16(buf[22:24], numChannels)
	binary.LittleEndian.PutUint32(buf[24:28], uint32(sampleRate))
	binary.LittleEndian.PutUint32(buf[28:32], uint32(sampleRate*numChannels*bytesPerSample)) // byte rate
	binary.LittleEndian.PutUint16(buf[32:34], numChannels*bytesPerSample)                    // block align
	binary.LittleEndian.PutUint16(buf[34:36], bitsPerSample)

	// data sub-chunk
	copy(buf[36:40], "data")
	binary.LittleEndian.PutUint32(buf[40:44], uint32(dataSize))

	// Write PCM samples.
	offset := 44
	for _, s := range pcm {
		binary.LittleEndian.PutUint16(buf[offset:offset+2], uint16(s))
		offset += 2
	}

	return buf
}
