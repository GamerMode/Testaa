package Requests

import (
	"client/ClientErrors"
	"encoding/json"
)

type ResponeType int

const (
	ErrorRespone ResponeType = 999
	ValidRespone ResponeType = 200
)

type ResponeInfo struct {
	Type    ResponeType `json:"Type"`
	Respone string      `json:"Data"`
}

// Encode raw slice of bytes to ResponeInfo struct
func GetResponseInfo(data []byte) (ResponeInfo, error) {
	var response_info ResponeInfo
	err := json.Unmarshal(data, &response_info)
	if err != nil {
		return ResponeInfo{}, &ClientErrors.JsonDecodeError{Err: err}
	}
	return response_info, nil
}
