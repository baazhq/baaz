package parseable

import (
	"fmt"
	"net/http"
)

const (
	streamPath = "/api/v1/logstream/"
)

func (s *Stream) makeStreamPath() string {
	return s.Url + streamPath + s.StreamName
}

func (s *Stream) CreateStream() (int, error) {
	pbClient := NewHTTPClient(
		&http.Client{},
		&Auth{
			BasicAuth: BasicAuth{
				UserName: s.UserName,
				Password: s.Password,
			},
		},
		"PUT",
		s.makeStreamPath(),
		s.Data,
		s.Metadata,
		s.Tags,
	)
	resp, err := pbClient.Do()
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("resp error [%s], status code [%d]", resp.ResponseBody, resp.StatusCode)
	}

	return resp.StatusCode, nil
}

func (s *Stream) Insertlogs() (int, error) {
	pbClient := NewHTTPClient(
		&http.Client{},
		&Auth{
			BasicAuth: BasicAuth{
				UserName: s.UserName,
				Password: s.Password,
			},
		},
		"POST",
		s.makeStreamPath(),
		s.Data,
		s.Metadata,
		s.Tags,
	)
	resp, err := pbClient.Do()
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("resp error [%s], status code [%d]", resp.ResponseBody, resp.StatusCode)
	}

	return resp.StatusCode, nil
}
