package cades

import (
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/exp/slog"
)

type CallbackData struct {
	Id     uint32 `json:"callback_id"`
	Object string `json:"object"`
	Type   string `json:"type"`
	Value  string `json:"value"`
}

type CadesRequestData struct {
	CallbackId  uint32       `json:"callback_id,omitempty"`
	RequestId   uint32       `json:"requestid,omitempty"`
	ObjId       uint32       `json:"objid,omitempty"`
	Object      string       `json:"object,omitempty"`
	Destination string       `json:"destination,omitempty"`
	Type        string       `json:"type,omitempty"`
	Value       string       `json:"value,omitempty"`
	Method      string       `json:"method,omitempty"`
	GetProperty string       `json:"get_property,omitempty"`
	SetProperty string       `json:"set_property,omitempty"`
	Url         string       `json:"url,omitempty"`
	Params      []CadesParam `json:"params,omitempty"`
}

type ReturnValue struct {
	Type       string      `json:"type"`
	Value      interface{} `json:"value"`
	Methods    []string    `json:"methods,omitempty"`
	Properties []string    `json:"properties,omitempty"`
}

type CadesResponseData struct {
	RequestId   uint32      `json:"requestid,omitempty"`
	ReturnValue ReturnValue `json:"retval,omitempty"`
	Type        string      `json:"type,omitempty"`
	Value       interface{} `json:"value,omitempty"`
	Message     string      `json:"message,omitempty"`
}

func CadesDataFromAnswer(answer *CadesResponseBody) (*CadesResponseData, error) {
	var data CadesResponseData
	if err := json.Unmarshal(*answer.Data, &data); err != nil {
		return &CadesResponseData{}, err
	}

	if value, ok := data.ReturnValue.Value.(float64); ok {
		if intValue := int32(value); value != float64(intValue) {
			data.ReturnValue.Value = intValue
		}
	}

	return &data, nil
}

type CadesRequestBody struct {
	Tabid string            `json:"tabid"`
	Data  *CadesRequestData `json:"data"`
}

type CadesResponseBody struct {
	Tabid string           `json:"tabid"`
	Data  *json.RawMessage `json:"data"`
}

type CadesParam struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

type Cades struct {
	Id        string
	RequestId uint32
	ObjId     uint32
	Process   *CadesProcess
}

type CadesObject struct {
	Cades *Cades
	ObjId uint32
}

func NewCades() (*Cades, error) {
	process, err := NewNMCadesProcess()
	if err != nil {
		return &Cades{}, err
	}

	cades := Cades{
		Id:        "CadesAgent",
		RequestId: 0,
		ObjId:     0,
		Process:   process,
	}

	body := &CadesRequestBody{
		Tabid: cades.Id,
		Data: &CadesRequestData{
			RequestId:   cades.RequestId,
			Destination: "nmcades",
			Type:        "init",
			Url:         "http://localhost:42217/cades",
		},
	}

	message, err := json.Marshal(body)
	if err != nil {
		return &Cades{}, err
	}

	answer, err := cades.sendRequestToProcess(message)
	if err != nil {
		return &cades, err
	}

	data, err := CadesDataFromAnswer(answer)
	if err != nil {
		return &cades, err

	} else if strings.ToLower(data.Type) == "error" {
		return &cades, fmt.Errorf("[nmcades] %s", data.Message)
	}

	return &cades, nil
}

func (cades *Cades) Close() {
	cades.Process.Cmd.Process.Kill()
}

func (cades *Cades) handlerCallback(answer *CadesResponseBody) (*CadesResponseBody, error) {
	slog.Debug("[Cades.HandlerCallback -> receive callback]")
	var callback CallbackData
	err := json.Unmarshal(*answer.Data, &callback)
	if err != nil {
		slog.Debug(fmt.Sprintf("[Cades.HandlerCallback] Fail to parse json: %s", err))
		return answer, err
	}

	body := &CadesRequestBody{
		Tabid: cades.Id,
		Data: &CadesRequestData{
			ObjId:       cades.ObjId,
			Destination: "nmcades",
			Type:        "result",
		},
	}

	body.Data.CallbackId = callback.Id
	if callback.Type == "callback" {
		switch callback.Value {
		case "result = cadesplugin.EnableInternalCSP":
			body.Data.Value = "result = cadesplugin.EnableInternalCSP"
			body.Data.Params = []CadesParam{
				{Type: "boolean", Value: false},
			}

		case "result = window.document.URL":
			body.Data.Value = "result = window.document.URL"
			body.Data.Params = []CadesParam{
				{Type: "string", Value: "http://localhost:42217/cades"},
			}
		}

	} else if callback.Type == "approved_site" {
		body.Data.Type = "approved_site"
		body.Data.Value = "is_approved_site: http://localhost:42217/cades"
		body.Data.Params = []CadesParam{
			{Type: "boolean", Value: true},
		}
	} else {
		return answer, ErrUnknownCallback
	}

	message, err := json.Marshal(body)
	if err != nil {
		return answer, err
	}

	return cades.sendRequestToProcess(message)
}

func (cades *Cades) SendRequest(request *CadesRequestBody) (*CadesResponseData, error) {
	message, err := json.Marshal(request)
	if err != nil {
		return &CadesResponseData{}, err
	}

	answer, err := cades.sendRequestToProcess(message)
	if err != nil {
		return &CadesResponseData{}, err
	}

	data, err := CadesDataFromAnswer(answer)
	if err != nil {
		return &CadesResponseData{}, err

	} else if strings.ToLower(data.Type) == "error" {
		return &CadesResponseData{}, fmt.Errorf("[nmcades] %s", data.Message)
	}

	return data, nil
}

func (cades *Cades) sendRequestToProcess(request []byte) (*CadesResponseBody, error) {

	slog.Debug(fmt.Sprintf("[Cades.send] Send message: %s", string(request)))
	PostMessage(*cades.Process.Stdin, request)

	message := GetMessage(*cades.Process.Stdout)
	slog.Debug(fmt.Sprintf("[Cades.send] Receive message: %s", message))

	var answer CadesResponseBody
	if err := json.Unmarshal([]byte(message), &answer); err != nil {
		slog.Debug(fmt.Sprintf("[Cades.send] Fail to parse json: %s", err))
		return &answer, err
	}

	if strings.Contains(message, "callback_id") {
		return cades.handlerCallback(&answer)
	}

	cades.RequestId++
	return &answer, nil
}
