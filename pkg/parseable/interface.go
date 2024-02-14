package parseable

import "os"

const (
	parseable_url      = "PARSEABLE_URL"
	parseable_username = "PARSEABLE_USERNAME"
	parseable_password = "PARSEABLE_PASSWORD"
)

type Stream struct {
	StreamName string
	UserName   string
	Password   string
	Url        string
	Data       []byte
	Metadata   map[string]string
	Tags       map[string]string
}

type Streams interface {
	CreateStream() (int, error)
	Insertlogs() (int, error)
}

func NewStream(
	streamName string,
	data []byte,
	metadata, tags map[string]string) Streams {
	return &Stream{
		StreamName: streamName,
		UserName:   os.Getenv(parseable_username),
		Password:   os.Getenv(parseable_password),
		Url:        os.Getenv(parseable_url),
		Data:       data,
		Metadata:   metadata,
		Tags:       tags,
	}
}
