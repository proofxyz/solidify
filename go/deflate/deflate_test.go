package deflate

import (
	"bytes"
	"math/rand"
	"testing"
)

func randomBytes(n uint) []byte {
	r := make([]byte, n)
	rand.Read(r)
	return r
}

func TestDeInflate(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "Keyboard entropy",
			data: []byte("qwertyuioqwertypasdqwertyfghjklqwerty"),
		},
		{
			name: "Random bytes",
			data: randomBytes(10897),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := Deflate(bytes.NewReader(tt.data))
			if err != nil {
				t.Errorf("Deflate([data]) error %v", err)
			}

			d, err := Inflate(c)
			if err != nil {
				t.Errorf("Inflate(Deflate([data])) error %v", err)
			}

			if !bytes.Equal(d, tt.data) {
				t.Errorf("Inflate(Deflate([data])) got %+x want %+x", d, tt.data)
			}
		})
	}
}
