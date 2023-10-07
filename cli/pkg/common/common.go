package common

import "os"

type ServerResp struct {
	Msg        string      `json:"Msg"`
	Status     string      `json:"Status"`
	StatusCode int         `json:"StatusCode"`
	Err        interface{} `json:"Err"`
}

func GetBzUrl() string {
	return os.Getenv("BAAZ_URL")
}
