package events

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/parseablehq/parseable-sdk-go/parseable"
)

type PbEvent struct {
	Message    string `json:"message"`
	Object     string `json:"object"`
	PMetadata  string `json:"p_metadata"`
	PTags      string `json:"p_tags"`
	PTimestamp string `json:"p_timestamp"`
	Reason     string `json:"reason"`
	Type       string `json:"type"`
}

type QueryBuilder struct {
	Query     string    `json:"query"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

func GetEvents(enitityName, durationStr string) error {

	duration, err := time.ParseDuration("-" + durationStr) // Convert duration string to negative duration (to go back in time)
	if err != nil {
		return err
	}

	currentTime := time.Now() // Current time
	startTime := currentTime.Add(duration)

	newQuery := QueryBuilder{
		Query:     fmt.Sprintf("select * from %s", enitityName),
		StartTime: startTime,
		EndTime:   currentTime,
	}

	body, err := json.Marshal(newQuery)
	if err != nil {
		return err
	}

	queryBuilder := parseable.NewQueryBuilder(string(body))

	resp, err := queryBuilder.QueryStream()
	if err != nil {
		return err
	}

	var customers []PbEvent
	err = json.Unmarshal([]byte(resp), &customers)
	if err != nil {
		return err
	}

	// Create new table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Message", "Object", "Metadata", "Tags", "Timestamp", "Reason", "Type"})

	// Add data to table
	for _, c := range customers {
		timestamp, err := time.Parse("2006-01-02T15:04:05.999", c.PTimestamp)
		if err != nil {
			return err
		}

		// Replace ^ delimiter with , in Metadata
		metadata := strings.ReplaceAll(c.PMetadata, "^", ", ")

		// Replace ^ delimiter with , in Tags
		tags := strings.ReplaceAll(c.PTags, "^", ", ")

		table.Append([]string{
			c.Message,
			c.Object,
			metadata,
			tags,
			timestamp.Format("2006-01-02 15:04:05"),
			c.Reason,
			c.Type,
		})
	}

	// Render table
	table.Render()

	return nil
}
