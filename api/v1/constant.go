package v1

type ApplicationType string

const (
	Druid      ApplicationType = "druid"
	ClickHouse ApplicationType = "clickhouse"
	Pinot      ApplicationType = "pinot"
)
