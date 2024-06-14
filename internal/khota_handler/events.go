package khota_handler

import (
	"github.com/parseablehq/parseable-sdk-go/parseable"
	klog "k8s.io/klog/v2"
)

type eventStreams string

const (
	customersEventStream    eventStreams = "customers"
	dataplanesEventStream   eventStreams = "dataplanes"
	tenantsEventStream      eventStreams = "tenants"
	tenantsInfraEventStream eventStreams = "tenantsinfra"
	applicationsEventStream eventStreams = "applications"
)

type eventType string

const (
	customerCreateSuccessEvent eventType = `
	[
		{
			"message": "customer create successfully",
			"type": "normal",
			"reason": "CustomerCreateSuccess",
			"object": "baaz/customers"
		}
	]`
	customerCreateFailEvent eventType = `[
		{
			"message": "customer creation failed",
			"type": "error",
			"reason": "CustomerCreateFail",
			"object": "baaz/customers"
		}
	]`
	dataplaneInitiationSuccessEvent eventType = `
	[
		{
			"message": "dataplane creation initiated",
			"type": "normal",
			"reason": "DataplaneCreationInitiated",
			"object": "baaz/dataplanes"
		}
	]`
	dataplaneInitiationFailEvent eventType = `[
		{
			"message": "dataplane creation failed",
			"type": "error",
			"reason": "DataplaneCreationFailed",
			"object": "baaz/dataplanes"
		}
	]`
	dataplaneTerminationEvent eventType = `[
		{
			"message": "dataplane termination initiated",
			"type": "error",
			"reason": "DataplaneTerminationInitiated",
			"object": "baaz/dataplanes"
		}
	]`
	tenantsInfraInitiationSuccessEvent eventType = `
	[
		{
			"message": "tenants infra creation initiated",
			"type": "normal",
			"reason": "TenantsInfraCreationInitiated",
			"object": "baaz/tenantsinfra"
		}
	]`
	tenantsInfraInitiationDeletionEvent eventType = `
	[
		{
			"message": "tenants infra deletion initiated",
			"type": "normal",
			"reason": "TenantsInfraDeletionInitiated",
			"object": "baaz/tenantsinfra"
		}
	]`
	tenantsInfraInitiationFailEvent eventType = `[
		{
			"message": "tenants infra creation failed",
			"type": "error",
			"reason": "TenantsInfraCreationFailed",
			"object": "baaz/tenantsinfra"
		}
	]`
	tenantsCreationSuccessEvent eventType = `
	[
		{
			"message": "tenants creation success",
			"type": "normal",
			"reason": "TenantsCreationSuccess",
			"object": "baaz/tenants"
		}
	]`
	tenantsCreationFailEvent eventType = `[
		{
			"message": "tenants creation failed",
			"type": "error",
			"reason": "TenantsCreationFailed",
			"object": "baaz/tenants"
		}
	]`
	ApplicationCreationFailEvent eventType = `[
		{
			"message": "application creation failed",
			"type": "error",
			"reason": "ApplicationCreationFailed",
			"object": "baaz/applications"
		}
	]`
	ApplicationCreationSuccessEvent eventType = `[
		{
			"message": "application creation success",
			"type": "error",
			"reason": "ApplicationCreationSuccess",
			"object": "baaz/applications"
		}
	]`
)

func sendEventParseable(stream eventStreams, eventType eventType, labels, tags map[string]string) {

	switch stream {
	case customersEventStream:
		stream := parseable.NewStreamBuilder(
			string(customersEventStream),
			[]byte(eventType),
			labels,
			tags,
		)
		_, err := stream.InsertLogs()
		if err != nil {
			klog.Error(err)
		}
	case dataplanesEventStream:
		stream := parseable.NewStreamBuilder(
			string(dataplanesEventStream),
			[]byte(eventType),
			labels,
			tags,
		)
		_, err := stream.InsertLogs()
		if err != nil {
			klog.Error(err)
		}
	case tenantsInfraEventStream:
		stream := parseable.NewStreamBuilder(
			string(tenantsInfraEventStream),
			[]byte(eventType),
			labels,
			tags,
		)
		_, err := stream.InsertLogs()
		if err != nil {
			klog.Error(err)
		}
	case tenantsEventStream:
		stream := parseable.NewStreamBuilder(
			string(tenantsEventStream),
			[]byte(eventType),
			labels,
			tags,
		)
		_, err := stream.InsertLogs()
		if err != nil {
			klog.Error(err)
		}
	case applicationsEventStream:
		stream := parseable.NewStreamBuilder(
			string(applicationsEventStream),
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
