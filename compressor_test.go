package gocql_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/golang/snappy"

	"github.com/gocql/gocql"
)

type frameExample struct {
	Name     string
	Frame    []byte
	FilePath string
}

var frameExamples = struct {
	Requests  []frameExample
	Responses []frameExample
}{
	Requests: []frameExample{
		{
			Name:     "Small query request",
			FilePath: "testdata/frames/small_query_request.bin",
		},
		{
			Name:     "Medium query request",
			FilePath: "testdata/frames/medium_query_request.bin",
		},
		{
			Name:     "Big query request",
			FilePath: "testdata/frames/big_query_request.bin",
		},
		{
			Name:     "Prepare statement request",
			FilePath: "testdata/frames/prepare_statement_request.bin",
		},
	},
	Responses: []frameExample{
		{
			Name:     "Small query response",
			FilePath: "testdata/frames/small_query_response.bin",
		},
		{
			Name:     "Medium query response",
			FilePath: "testdata/frames/medium_query_response.bin",
		},
		{
			Name:     "Big query response",
			FilePath: "testdata/frames/big_query_response.bin",
		},
		{
			Name:     "Prepare statement response",
			FilePath: "testdata/frames/prepare_statement_response.bin",
		},
	},
}

func TestSnappyCompressor(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		c := gocql.SnappyCompressor{}
		if c.Name() != "snappy" {
			t.Fatalf("expected name to be 'snappy', got %v", c.Name())
		}

		str := "My Test String"
		//Test Encoding
		expected := snappy.Encode(nil, []byte(str))
		if res, err := c.Encode([]byte(str)); err != nil {
			t.Fatalf("failed to encode '%v' with error %v", str, err)
		} else if bytes.Compare(expected, res) != 0 {
			t.Fatal("failed to match the expected encoded value with the result encoded value.")
		}

		val, err := c.Encode([]byte(str))
		if err != nil {
			t.Fatalf("failed to encode '%v' with error '%v'", str, err)
		}

		//Test Decoding
		if expected, err := snappy.Decode(nil, val); err != nil {
			t.Fatalf("failed to decode '%v' with error %v", val, err)
		} else if res, err := c.Decode(val); err != nil {
			t.Fatalf("failed to decode '%v' with error %v", val, err)
		} else if bytes.Compare(expected, res) != 0 {
			t.Fatal("failed to match the expected decoded value with the result decoded value.")
		}
	})

	t.Run("frame-examples", func(t *testing.T) {
		c := gocql.SnappyCompressor{}

		t.Run("Encode", func(t *testing.T) {
			for _, frame := range frameExamples.Requests {
				t.Run(frame.Name, func(t *testing.T) {
					encoded, err := c.Encode(frame.Frame)
					if err != nil {
						t.Fatalf("failed to encode frame %s", frame.Name)
					}
					decoded, err := c.Decode(encoded)
					if err != nil {
						t.Fatalf("failed to decode frame %s", frame.Name)
					}

					if bytes.Compare(decoded, frame.Frame) != 0 {
						t.Fatalf("failed to match the decoded value with the original value")
					}
					t.Logf("Compression rate %f", float64(len(encoded))/float64(len(frame.Frame)))
				})
			}
		})

		t.Run("Decode", func(t *testing.T) {
			for _, frame := range frameExamples.Responses {
				t.Run(frame.Name, func(t *testing.T) {
					decoded, err := c.Decode(frame.Frame)
					if err != nil {
						t.Fatalf("failed to decode frame %s", frame.Name)
					}

					if len(decoded) == 0 {
						t.Fatalf("frame was decoded to empty slice")
					}
				})
			}
		})
	})
}

func BenchmarkSnappyCompressor(b *testing.B) {
	c := gocql.SnappyCompressor{}
	b.Run("Decode", func(b *testing.B) {
		for _, frame := range frameExamples.Responses {
			b.Run(frame.Name, func(b *testing.B) {
				for x := 0; x < b.N; x++ {
					_, _ = c.Decode(frame.Frame)
				}
			})
		}
	})

	b.Run("Encode", func(b *testing.B) {
		for _, frame := range frameExamples.Requests {
			b.Run(frame.Name, func(b *testing.B) {
				for x := 0; x < b.N; x++ {
					_, _ = c.Encode(frame.Frame)
				}
			})
		}
	})
}

func init() {
	var err error
	for id, def := range frameExamples.Requests {
		frameExamples.Requests[id].Frame, err = os.ReadFile(def.FilePath)
		if err != nil {
			panic("can't read file " + def.FilePath)
		}
	}
	for id, def := range frameExamples.Responses {
		frameExamples.Responses[id].Frame, err = os.ReadFile(def.FilePath)
		if err != nil {
			panic("can't read file " + def.FilePath)
		}
	}
}
