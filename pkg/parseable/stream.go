package parseable

import (
	"fmt"
	"net/http"
	"os"
)

const (
	apiVersion = "/v1"
	streamPath = "/api" + apiVersion + "/logstream/"
)

// StreamBuilder represents a data stream builder
type StreamBuilder struct {
	StreamName string            // Stream Name
	Data       []byte            // Data to be inserted into the stream
	Metadata   map[string]string // Metadata associated with the stream
	Tags       map[string]string // Tags associated with the stream
}

// NewStreamBuilder creates a new stream builder with provided configurations
func NewStreamBuilder(streamName string, data []byte, metadata, tags map[string]string) Stream {
	return &StreamBuilder{
		StreamName: streamName,
		Data:       data,
		Metadata:   metadata,
		Tags:       tags,
	}
}

// makeStreamPath constructs the URL for the stream
func (s *StreamBuilder) makeStreamPath() string {
	return fmt.Sprintf("%s%s%s", os.Getenv(ParseableURL), streamPath, s.StreamName)
}

// CreateStream creates a new stream
func (s *StreamBuilder) CreateStream() (int, error) {
	// Create HTTP client with provided configurations
	pbClient := newHTTPClient(
		&http.Client{},
		http.MethodPut, // Use constants instead of hardcoded strings
		s.makeStreamPath(),
		s.Data,
		s.Metadata,
		s.Tags,
	)
	// Perform HTTP request
	resp, err := pbClient.do()
	if err != nil {
		return 0, err
	}

	// Check response status code
	if resp.statusCode != http.StatusOK {
		return 0, fmt.Errorf("response error [%s], status code [%d]", resp.responseBody, resp.statusCode)
	}

	return resp.statusCode, nil
}

// InsertLogs inserts logs into the stream
func (s *StreamBuilder) InsertLogs() (int, error) {
	// Create HTTP client with provided configurations
	pbClient := newHTTPClient(
		&http.Client{},
		http.MethodPost, // Use constants instead of hardcoded strings
		s.makeStreamPath(),
		s.Data,
		s.Metadata,
		s.Tags,
	)
	// Perform HTTP request
	resp, err := pbClient.do()
	if err != nil {
		return 0, err
	}

	// Check response status code
	if resp.statusCode != http.StatusOK {
		return 0, fmt.Errorf("response error [%s], status code [%d]", resp.responseBody, resp.statusCode)
	}

	return resp.statusCode, nil
}
