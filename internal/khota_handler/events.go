package khota_handler

import (
	"github.com/parseablehq/parseable-sdk-go/parseable"
	klog "k8s.io/klog/v2"
)

type eventStreams string

const (
	customerEventStream     eventStreams = "customers"
	dataplanesEventStream   eventStreams = "dataplanes"
	tenatsEventStream       eventStreams = "tenants"
	tenatsInfraEventStream  eventStreams = "tenantsinfra"
	applicationsEventStream eventStreams = "applications"
)

type eventType string

const (
	customerCreateSuccess eventType = `
	[
		{
			"message": "customer create successfully",
			"type": "normal",
			"reason": "CustomerCreateSuccess",
			"object": "baaz/customer"
		}
	]`
	customerCreateFail eventType = `[
		{
			"message": "customer creation failed",
			"type": "error",
			"reason": "CustomerCreateFail",
			"object": "baaz/customer"
		}
	]`
	dataplaneInitiationSuccess eventType = `
	[
		{
			"message": "dataplane creation initiated",
			"type": "normal",
			"reason": "DataplaneCreationInitiated",
			"object": "baaz/dataplane"
		}
	]`
	dataplaneInitiationFail eventType = `[
		{
			"message": "dataplane creation failed",
			"type": "error",
			"reason": "DataplaneCreationFailed",
			"object": "baaz/dataplane"
		}
	]`
	tenantsInfraInitiationSuccess eventType = `
	[
		{
			"message": "tenants infra creation initiated",
			"type": "normal",
			"reason": "TenantsInfraCreationInitiated",
			"object": "baaz/tenantsinfra"
		}
	]`
	tenantsInfraInitiationFail eventType = `[
		{
			"message": "tenants infra creation failed",
			"type": "error",
			"reason": "TenantsInfraCreationFailed",
			"object": "baaz/tenantsinfra"
		}
	]`
	tenantsCreationSuccess eventType = `
	[
		{
			"message": "tenants creation success",
			"type": "normal",
			"reason": "TenantsCreationSuccess",
			"object": "baaz/tenants"
		}
	]`
	tenantsCreationFail eventType = `[
		{
			"message": "tenants creation failed",
			"type": "error",
			"reason": "TenantsCreationFailed",
			"object": "baaz/tenants"
		}
	]`
)

func sendEventParseable(stream eventStreams, eventType eventType, labels, tags map[string]string) {

	switch stream {
	case customerEventStream:
		stream := parseable.NewStreamBuilder(
			string(customerEventStream),
			[]byte(eventType),
			labels,
			tags,
		)
		_, err := stream.InsertLogs()
		if err != nil {
			klog.Error(err)
		}
	}

}
