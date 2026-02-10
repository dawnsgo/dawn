package packet

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestIntCodecRoundTrip(t *testing.T) {
	tests := []struct {
		name      string
		size      int
		byteOrder binary.ByteOrder
		value     int32
	}{
		{"1byte-BE-positive", 1, binary.BigEndian, 42},
		{"1byte-BE-negative", 1, binary.BigEndian, -10},
		{"1byte-LE-zero", 1, binary.LittleEndian, 0},
		{"2byte-BE-positive", 2, binary.BigEndian, 12345},
		{"2byte-LE-negative", 2, binary.LittleEndian, -1000},
		{"4byte-BE-large", 4, binary.BigEndian, 1234567890},
		{"4byte-LE-negative", 4, binary.LittleEndian, -999999},
		{"4byte-BE-zero", 4, binary.BigEndian, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			codec := newIntCodec(tc.size, tc.byteOrder)

			// Test Size
			if codec.Size() != tc.size {
				t.Fatalf("expected Size() = %d, got %d", tc.size, codec.Size())
			}

			// Test EncodeToBuf + Decode round trip
			buf := &bytes.Buffer{}
			if err := codec.EncodeToBuf(buf, tc.value); err != nil {
				t.Fatalf("EncodeToBuf failed: %v", err)
			}

			if buf.Len() != tc.size {
				t.Fatalf("expected %d bytes, got %d", tc.size, buf.Len())
			}

			reader := bytes.NewReader(buf.Bytes())
			decoded, err := codec.Decode(reader)
			if err != nil {
				t.Fatalf("Decode failed: %v", err)
			}

			// For 1-byte and 2-byte, values are truncated
			expected := tc.value
			switch tc.size {
			case 1:
				expected = int32(int8(tc.value))
			case 2:
				expected = int32(int16(tc.value))
			}

			if decoded != expected {
				t.Fatalf("expected %d, got %d", expected, decoded)
			}
		})
	}
}

func TestIntCodecZeroSize(t *testing.T) {
	codec := newIntCodec(0, binary.BigEndian)

	if codec.Size() != 0 {
		t.Fatalf("expected Size() = 0, got %d", codec.Size())
	}

	// EncodeToBuf should be a no-op
	buf := &bytes.Buffer{}
	if err := codec.EncodeToBuf(buf, 42); err != nil {
		t.Fatalf("EncodeToBuf failed: %v", err)
	}
	if buf.Len() != 0 {
		t.Fatalf("expected 0 bytes, got %d", buf.Len())
	}

	// Decode should return 0
	reader := bytes.NewReader(nil)
	val, err := codec.Decode(reader)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if val != 0 {
		t.Fatalf("expected 0, got %d", val)
	}
}

func TestIntCodecMultipleValues(t *testing.T) {
	codec := newIntCodec(4, binary.BigEndian)

	// Encode multiple values
	buf := &bytes.Buffer{}
	values := []int32{100, 200, 300, -400}
	for _, v := range values {
		if err := codec.EncodeToBuf(buf, v); err != nil {
			t.Fatalf("EncodeToBuf failed: %v", err)
		}
	}

	// Decode all values
	reader := bytes.NewReader(buf.Bytes())
	for i, expected := range values {
		decoded, err := codec.Decode(reader)
		if err != nil {
			t.Fatalf("Decode[%d] failed: %v", i, err)
		}
		if decoded != expected {
			t.Fatalf("value[%d]: expected %d, got %d", i, expected, decoded)
		}
	}
}

func BenchmarkIntCodecEncode4Byte(b *testing.B) {
	codec := newIntCodec(4, binary.BigEndian)
	buf := &bytes.Buffer{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		codec.EncodeToBuf(buf, int32(i))
	}
}

func BenchmarkIntCodecDecode4Byte(b *testing.B) {
	codec := newIntCodec(4, binary.BigEndian)
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, 42)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(data)
		codec.Decode(reader)
	}
}
