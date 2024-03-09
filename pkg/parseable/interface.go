package parseable

// Environment variables
const (
	ParseableURL      = "PARSEABLE_URL"
	ParseableUsername = "PARSEABLE_USERNAME"
	ParseablePassword = "PARSEABLE_PASSWORD"
)

type Parseable interface {
	QueryStream() (string, error)
	InsertLogs() (int, error)
	CreateStream() (int, error)
}

type Query interface {
	QueryStream() (string, error)
}

type Stream interface {
	InsertLogs() (int, error)
	CreateStream() (int, error)
}
