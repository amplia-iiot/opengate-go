package odm_model

import (
	"encoding/json"
	"errors"

	"github.com/amplia-iiot/opengate-go/http_client"
	"github.com/amplia-iiot/opengate-go/logger"
)

type OperationStatusRsp string

const (
	IN_PROGRESS            OperationStatusRsp = "IN_PROGRESS"
	WAITING_FOR_CONNECTION OperationStatusRsp = "WAITING_FOR_CONNECTION"
)

type OpSender struct {
	Steps           []Step
	OperationStatus AsyncRspType
	RspDescription  string
	OperationName   string
	DeviceId        string
	OperationId     string
}

type OperationRsp struct {
	Version   string    `json:"version,omitempty"`
	Operation Operation `json:"operation,omitempty"`
}
type Operation struct {
	Response Response `json:"response,omitempty"`
}
type Response struct {
	Id string `json:"id,omitempty"`
	// Timestamp         int64  `json:"timestamp,omitempty"`
	DeviceId          string `json:"deviceId,omitempty"`
	Name              string `json:"name,omitempty"`
	ResultCode        string `json:"resultCode,omitempty"`
	ResultDescription string `json:"resultDescription,omitempty"`
	Steps             []Step `json:"steps,omitempty"`
}
type Step struct {
	Name string `json:"name,omitempty"`
	// Timestamp   int64  `json:"timestamp,omitempty"`
	Result      string `json:"result,omitempty"`
	Description string `json:"description,omitempty"`
}

func BuildNewStep(stepName string, stepStatus AsyncRspType, description string) Step {
	return Step{
		Name:        stepName,
		Result:      string(stepStatus),
		Description: description,
	}
}
func (o *OpSender) SendIntermediateStep(clientOptions http_client.ClientOptions) error {
	o.OperationStatus = EMPTY_FOR_STEP
	o.RspDescription = ""
	return o.SendAsyncOdmRspSteps(clientOptions)
}
func (o *OpSender) SendAsyncOdmRspSteps(clientOptions http_client.ClientOptions) error {
	restClient := http_client.NewOperationRestClient(clientOptions)
	if o.OperationId == "" || o.DeviceId == "" || o.OperationName == "" {
		return errors.New("imcomplete info to build operation rsp")
	}
	response := Response{
		Id:                o.OperationId,
		DeviceId:          o.DeviceId,
		Name:              o.OperationName,
		ResultCode:        string(o.OperationStatus),
		ResultDescription: o.RspDescription,
		Steps:             o.Steps,
	}
	operationRsp := OperationRsp{
		Version: Version,
		Operation: Operation{
			Response: response,
		},
	}
	odmAsyncRsp, err := json.Marshal(operationRsp)
	if err != nil {
		return errors.New("Error marshalling data: " + err.Error())
	}
	logger.Debug("JsonRspSteps:")
	logger.Debug(string(odmAsyncRsp))
	_, err = restClient.Do(string(odmAsyncRsp))
	return err
}
