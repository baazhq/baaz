package parseable

import (
	"bytes"
	"io"
	"net/http"
	"os"
)

// HTTP interface defines the Do method
type HTTP interface {
	do() (*response, error)
}

// pbClient represents an HTTP client
type pbClient struct {
	httpClient *http.Client
	basicAuth  *basicAuth
	method     string
	url        string
	body       []byte
	metadata   map[string]string
	tags       map[string]string
}

// newHTTPClient creates a new HTTP client instance
func newHTTPClient(
	client *http.Client,
	method string,
	url string,
	body []byte,
	metadata, tags map[string]string,
) HTTP {
	return &pbClient{
		httpClient: client,
		basicAuth:  newBasicAuth(),
		method:     method,
		url:        url,
		body:       body,
		metadata:   metadata,
		tags:       tags,
	}
}

// basicAuth represents basic authentication credentials
type basicAuth struct {
	userName string
	password string
}

// newBasicAuth creates a new basicAuth instance with provided username and password
func newBasicAuth() *basicAuth {
	return &basicAuth{
		userName: os.Getenv(ParseableUsername),
		password: os.Getenv(ParseablePassword),
	}
}

// response represents the response from HTTP requests
type response struct {
	responseBody string
	statusCode   int
}

// do performs the HTTP request
func (c *pbClient) do() (*response, error) {
	req, err := http.NewRequest(c.method, c.url, bytes.NewBuffer(c.body))
	if err != nil {
		return nil, err
	}

	if c.basicAuth.userName != "" && c.basicAuth.password != "" {
		req.SetBasicAuth(c.basicAuth.userName, c.basicAuth.password)
	}

	for key, value := range c.metadata {
		req.Header.Add("X-P-META-"+key, value)
	}

	for key, value := range c.tags {
		req.Header.Add("X-P-TAG-"+key, value)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &response{
		responseBody: string(responseBody),
		statusCode:   resp.StatusCode,
	}, nil
}
