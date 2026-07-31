package mozlz4_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"mozsession-merge/internal/mozlz4"
)

func TestRoundTrip(t *testing.T) {
	testCases := [][]byte{
		[]byte(""),
		[]byte("Hello, World! This is a test string for mozLz4 codec round-trip validation."),
		[]byte(`{"windows":[{"tabs":[{"entries":[{"url":"https://example.com","title":"Example"}]}]}]}`),
		bytes.Repeat([]byte("Repeated string data 1234567890\n"), 100),
	}

	for i, original := range testCases {
		encoded, err := mozlz4.Encode(original)
		if err != nil {
			t.Fatalf("case %d: Encode failed: %v", i, err)
		}

		decoded, err := mozlz4.Decode(encoded)
		if err != nil {
			t.Fatalf("case %d: Decode failed: %v", i, err)
		}

		if !bytes.Equal(original, decoded) {
			t.Fatalf("case %d: decoded data does not match original data.\nOriginal len: %d, Decoded len: %d", i, len(original), len(decoded))
		}
	}
}

func TestDecodeErrors(t *testing.T) {
	t.Run("short file", func(t *testing.T) {
		_, err := mozlz4.Decode([]byte("short"))
		if err != mozlz4.ErrFileTooShort {
			t.Fatalf("expected ErrFileTooShort, got %v", err)
		}
	})

	t.Run("invalid header", func(t *testing.T) {
		invalidHeader := make([]byte, 16)
		copy(invalidHeader, []byte("invalid_"))
		_, err := mozlz4.Decode(invalidHeader)
		if err != mozlz4.ErrInvalidHeader {
			t.Fatalf("expected ErrInvalidHeader, got %v", err)
		}
	})
}

func TestRealSamples(t *testing.T) {
	samples := []string{
		"../../testdata/Profile1/sessionstore.jsonlz4",
		"../../testdata/Profile1/zen-sessions.jsonlz4",
	}

	for _, samplePath := range samples {
		absPath, err := filepath.Abs(samplePath)
		if err != nil {
			t.Fatalf("failed to resolve path %s: %v", samplePath, err)
		}

		data, err := os.ReadFile(absPath)
		if err != nil {
			t.Fatalf("failed to read sample file %s: %v", absPath, err)
		}

		decoded, err := mozlz4.Decode(data)
		if err != nil {
			t.Fatalf("failed to decode sample file %s: %v", absPath, err)
		}

		if len(decoded) == 0 {
			t.Fatalf("decoded sample file %s is empty", absPath)
		}

		// Verify round-trip re-encoding of real sample JSON
		reEncoded, err := mozlz4.Encode(decoded)
		if err != nil {
			t.Fatalf("failed to re-encode sample %s: %v", absPath, err)
		}

		reDecoded, err := mozlz4.Decode(reEncoded)
		if err != nil {
			t.Fatalf("failed to decode re-encoded sample %s: %v", absPath, err)
		}

		if !bytes.Equal(decoded, reDecoded) {
			t.Fatalf("re-decoded sample %s does not match original decoded JSON", absPath)
		}

		// Print first 200 chars to test output for visual confirmation
		previewLen := 200
		if len(decoded) < previewLen {
			previewLen = len(decoded)
		}
		t.Logf("Sample %s decoded & round-trip re-encoded successfully (%d bytes). First %d chars:\n%s\n",
			filepath.Base(samplePath), len(decoded), previewLen, string(decoded[:previewLen]))
	}
}
