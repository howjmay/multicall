package jsonrpc

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type JSONRPCRequest struct {
	Version string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      string      `json:"id"`
}

type JSONRPCResponse struct {
	Version string          `json:"jsonrpc"`
	ID      string          `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Raw     json.RawMessage `json:"-"`
}

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

// DecodeError : error returned when failing to decode a response
type DecodeError struct {
	Raw []byte
	Err error
}

// Error message string
func (d DecodeError) Error() string {
	return fmt.Sprintf("failed to decode %s: %s", d.Raw, d.Err)
}

func (msg *JSONRPCResponse) IsNotification() bool {
	return msg.ID == "" && msg.Method != ""
}

func (msg *JSONRPCResponse) IsResponse() bool {
	return msg.HasValidID() && msg.Method == "" && len(msg.Params) == 0
}

func (msg *JSONRPCResponse) HasValidID() bool {
	return len(msg.ID) > 0 && msg.ID[0] != '{' && msg.ID[0] != '['
}

// ValidID decodes the id if it is valid
func (msg *JSONRPCResponse) ValidID() (string, error) {
	if !msg.HasValidID() {
		return "", fmt.Errorf("message does not have a valid id")
	}
	return string(msg.ID), nil
}

// UINTResult decodes the the result as uint
func (msg *JSONRPCResponse) UINTResult() (uint64, error) {
	var s string
	err := json.Unmarshal(msg.Result, &s)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(s, 0, 64)
}

func (msg *JSONRPCResponse) String() string {
	b, _ := json.Marshal(msg)
	return string(b)
}

type JSONRPCNotification struct {
	ID     string          `json:"subscription"`
	Result json.RawMessage `json:"result"`
}

// ValidID decodes the id if it s valid
func (msg *JSONRPCNotification) ValidID() (string, error) {
	if len(msg.ID) == 0 {
		return "", fmt.Errorf("no ID found")
	}
	return msg.ID, nil
}
