package v1

type ApplicationType string

const (
	Druid      ApplicationType = "Druid"
	ClickHouse ApplicationType = "ClickHouse"
	Pinot      ApplicationType = "Pinot"
)
