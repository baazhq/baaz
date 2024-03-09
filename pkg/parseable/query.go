package parseable

import (
	"fmt"
	"net/http"
	"os"
)

const (
	queryPath = "/api" + apiVersion + "/query"
)

// QueryBuilder represents a query builder
type QueryBuilder struct {
	Query string // Query string for the query
}

// NewQueryBuilder creates a new QueryBuilder instance with provided configurations
func NewQueryBuilder(query string) *QueryBuilder {
	return &QueryBuilder{
		Query: query,
	}
}

// makeQueryPath constructs the URL for the query
func (s *QueryBuilder) makeQueryPath() string {
	return fmt.Sprintf("%s%s", os.Getenv(ParseableURL), queryPath)
}

// QueryStream queries into the stream
func (s *QueryBuilder) QueryStream() (string, error) {
	pbClient := newHTTPClient(
		&http.Client{},
		http.MethodPost, // Use constants instead of hardcoded strings
		s.makeQueryPath(),
		[]byte(s.Query),
		nil,
		nil,
	)

	resp, err := pbClient.do()
	if err != nil {
		return "", err
	}

	if resp.statusCode != http.StatusOK {
		return "", fmt.Errorf("response error [%s], status code [%d]", resp.responseBody, resp.statusCode)
	}

	return resp.responseBody, nil
}
