package parseable

import (
	"bytes"
	"io"
	"net/http"
)

// HTTP interface
type HTTP interface {
	Do() (*Response, error)
}

// HTTP client
type PbClient struct {
	HTTPClient *http.Client
	Auth       *Auth
	Method     string
	Url        string
	Body       []byte
	Metadata   map[string]string
	Tags       map[string]string
}

func NewHTTPClient(
	client *http.Client,
	auth *Auth,
	method string,
	url string,
	body []byte,
	metadata map[string]string,
	tags map[string]string,
) HTTP {
	newClient := &PbClient{
		HTTPClient: client,
		Auth:       auth,
		Method:     method,
		Url:        url,
		Body:       body,
		Metadata:   metadata,
		Tags:       tags,
	}

	return newClient
}

// Auth mechanisms supported by control plane to authenticate with parseable
type Auth struct {
	BasicAuth BasicAuth
}

// BasicAuth
type BasicAuth struct {
	UserName string
	Password string
}

// Response passed to controller
type Response struct {
	ResponseBody string
	StatusCode   int
}

// Do method to be used schema and tenant controller.
func (c *PbClient) Do() (*Response, error) {

	req, err := http.NewRequest(c.Method, c.Url, bytes.NewBuffer(c.Body))
	if err != nil {
		return nil, err
	}

	if c.Auth.BasicAuth != (BasicAuth{}) {
		req.SetBasicAuth(c.Auth.BasicAuth.UserName, c.Auth.BasicAuth.Password)
	}

	for key, value := range c.Metadata {
		req.Header.Add("X-P-META-"+key, value)
	}

	for key, value := range c.Tags {
		req.Header.Add("X-P-TAG-"+key, value)
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &Response{ResponseBody: string(responseBody), StatusCode: resp.StatusCode}, nil
}
