package hivego

import (
	"bytes"
	"encoding/hex"
	"errors"
	"testing"
)

func TestHashTxForSig(t *testing.T) {
	chainID, _ := hex.DecodeString(HiveMainnetChainID)
	got := hashTxForSig([]byte{189, 140, 95, 226, 111, 69, 241, 121, 168, 87, 1, 0, 5, 120, 101, 114, 111, 99, 5, 120, 101, 114, 111, 99, 6, 112, 105, 115, 116, 111, 110, 16, 39, 0}, chainID)
	expected := []byte{14, 93, 189, 151, 159, 79, 35, 135, 197, 213, 161, 182, 73, 239, 6, 88, 150, 48, 250, 247, 192, 101, 222, 160, 218, 142, 89, 43, 3, 218, 6, 188}

	if !bytes.Equal(got, expected) {
		t.Error("Expected", expected, "got", got)
	}
}

func TestHashTx(t *testing.T) {
	got := hashTx([]byte{189, 140, 95, 226, 111, 69, 241, 121, 168, 87, 1, 0, 5, 120, 101, 114, 111, 99, 5, 120, 101, 114, 111, 99, 6, 112, 105, 115, 116, 111, 110, 16, 39, 0})
	expected := []byte{18, 22, 77, 206, 229, 24, 103, 76, 88, 110, 106, 97, 208, 134, 35, 196, 73, 128, 227, 38, 201, 129, 108, 80, 35, 200, 184, 136, 15, 114, 61, 107}

	if !bytes.Equal(got, expected) {
		t.Error("Expected", expected, "got", got)
	}
}

func TestSignDigest(t *testing.T) {
	key, err := KeyPairFromWif("5JuMt237G3m3BaT7zH4YdoycUtbw4AEPy6DLdCrKAnFGAtXyQ1W")
	if err != nil {
		t.Fatal(err)
	}
	got := SignDigest([]byte{18, 22, 77, 206, 229, 24, 103, 76, 88, 110, 106, 97, 208, 134, 35, 196, 73, 128, 227, 38, 201, 129, 108, 80, 35, 200, 184, 136, 15, 114, 61, 107}, key)
	expected := []byte{31, 135, 178, 255, 150, 145, 101, 147, 159, 43, 208, 208, 212, 141, 167, 79, 124, 202, 216, 106, 84, 92, 126, 124, 217, 124, 39, 47, 105, 76, 20, 113, 229, 2, 5, 110, 116, 181, 214, 32, 111, 134, 253, 231, 100, 158, 241, 102, 238, 32, 213, 22, 155, 115, 226, 172, 85, 229, 183, 99, 0, 155, 252, 229, 186}
	if !bytes.Equal(got, expected) {
		t.Error("Expected", expected, "got", got)
	}
}

func TestGphBase58CheckDecode(t *testing.T) {
	got1, got2, _ := GphBase58CheckDecode("5JuMt237G3m3BaT7zH4YdoycUtbw4AEPy6DLdCrKAnFGAtXyQ1W")
	expected1 := []byte{143, 55, 174, 90, 120, 223, 222, 54, 91, 147, 72, 37, 164, 39, 94, 43, 230, 160, 223, 142, 67, 73, 158, 81, 48, 197, 148, 24, 63, 220, 121, 208}
	expected2 := [1]byte{128}
	if !bytes.Equal(got1, expected1) || got2 != expected2 {
		t.Error("Expected", expected1, "and", expected2, "got", got1, "and", got2)
	}
}

func TestGphBase58Encode(t *testing.T) {
	// Round-trip: decode then re-encode should produce the original string
	input := "5JuMt237G3m3BaT7zH4YdoycUtbw4AEPy6DLdCrKAnFGAtXyQ1W"
	payload, version, err := GphBase58CheckDecode(input)
	if err != nil {
		t.Fatal("GphBase58CheckDecode failed:", err)
	}
	got := GphBase58Encode(payload, version)
	if got != input {
		t.Errorf("Expected %s, got %s", input, got)
	}
}

func TestGphBase58CheckDecodeTooShort(t *testing.T) {
	_, _, err := GphBase58CheckDecode("abc")
	if !errors.Is(err, ErrInvalidFormat) {
		t.Fatalf("expected ErrInvalidFormat, got %v", err)
	}
}

func TestGphBase58CheckDecodeBadChecksum(t *testing.T) {
	// Take a valid WIF and corrupt the last byte.
	payload, version, _ := GphBase58CheckDecode("5JuMt237G3m3BaT7zH4YdoycUtbw4AEPy6DLdCrKAnFGAtXyQ1W")
	// Re-encode with a corrupted payload byte.
	payload[0] ^= 0xFF
	corrupted := GphBase58Encode(payload, version)
	// Corrupt the checksum by flipping a byte at the end.
	// Easier: decode the base58 bytes and flip checksum manually.
	_ = corrupted // we'll just encode a known bad string
	_, _, err := GphBase58CheckDecode("5JuMt237G3m3BaT7zH4YdoycUtbw4AEPy6DLdCrKAnFGAtXyQ1X")
	if !errors.Is(err, ErrChecksumMismatch) {
		t.Fatalf("expected ErrChecksumMismatch, got %v", err)
	}
}
