package common

import "os"

type ServerResp struct {
	
}

func GetBzUrl() string {
	return os.Getenv("BAAZ_URL")
}

type CustomError string

const (
	InvalidConfig CustomError = "Invalid Config"
)
