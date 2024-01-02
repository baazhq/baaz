package common

import (
	"bytes"
	"fmt"
	"os"
)

var (
	BaazPath        = "/api/v1"
	DataplanePath   = "/dataplane"
	CustomerPath    = "/customer"
	TenantPath      = "/tenant"
	TenantInfraPath = "/tenantsinfra"
	TenantSizesPath = "/sizes"
	KubeConfigPath  = "/config"
)

func GetBzUrl() string {
	return os.Getenv("BAAZ_URL")
}

type CustomError string

const (
	InvalidConfig CustomError = "Invalid Config"
)

func CreateKeyValuePairs(m map[string]string) string {
	b := new(bytes.Buffer)
	for key, value := range m {
		fmt.Fprintf(b, "%s: %s\n", key, value)
	}
	return b.String()
}
