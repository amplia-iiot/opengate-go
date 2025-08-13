package odm_model

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amplia-iiot/opengate-go/http_client"
	"github.com/amplia-iiot/opengate-go/logger"
)

type FilterType string
type OperatorType string

// eq: Equals.
// neq: Not equals.
// like: Regex pattern like.
// gt: Greater than.
// lt: Lower than.
// gte: Greater than or equals.
// lte: Lower than or equals.
// in[]: Included in a concrete group.
// nin[]: Not included in a concrete group.
// exists: Exists. within
const (
	IN FilterType = "in"
	EQ FilterType = "eq"
)
const (
	OR  OperatorType = "or"
	AND OperatorType = "and"
)

type SearchReq struct {
	Filter interface{} `json:"filter,omitempty"`
	Limit  *Limit      `json:"limit,omitempty"`
	Select []Select    `json:"select,omitempty"`
}

type Limit struct {
	Size  int `json:"size,omitempty"`
	Start int `json:"start,omitempty"`
}
type Fields struct {
	Field string `json:"field,omitempty"`
	Alias string `json:"alias,omitempty"`
}
type Select struct {
	Name   string   `json:"name,omitempty"`
	Fields []Fields `json:"fields,omitempty"`
}

type LaunchEntityRsp struct {
	StatusCode       int
	EntitieRsp       map[string]interface{}
	FullRspRaw       []byte
	OpSearchResponse SearchEntitiesRsp
}

func CreateFilterType(filterType FilterType, keyName string, value interface{}) map[string]interface{} {
	jsonElement := make(map[string]interface{})
	jsonElement[keyName] = value
	jsonFilter := make(map[string]interface{})
	jsonFilter[string(filterType)] = jsonElement
	return jsonFilter
}
func CreateLogicalOperator(operatorType OperatorType, elementInOperators []map[string]interface{}) map[string]interface{} {
	jsonElement := make(map[string]interface{})
	jsonElement[string(operatorType)] = elementInOperators
	return jsonElement
}
func CreateNewSelect(name string) Select {
	return Select{
		Fields: []Fields{{
			Field: FIELD_VALUE,
			Alias: FIELD_ALIAS_IDENTIFICATOR,
		}},
		Name: name,
	}
}

type EntitiesSearcher struct {
	RequestId string
	Filter    interface{} `json:"filter,omitempty"`
	Selects   []Select    `json:"select,omitempty"`
	Limit     *Limit      `json:"limit,omitempty"`
}

func (s *EntitiesSearcher) WithRequestId(id string) {
	s.RequestId = id
}
func (s *EntitiesSearcher) LaunchNewSearchDevOperations(clientOptions http_client.ClientOptions) (leRsp LaunchEntityRsp, err error) {
	ogRestClient := http_client.NewSearchDevOperations(clientOptions)
	searchReq := SearchReq{
		Filter: s.Filter,
		Select: s.Selects,
		Limit:  s.Limit,
	}
	reqByte, err := json.Marshal(searchReq)
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Debug(fmt.Sprintf("[%v]: searching req: %v", ogRestClient.RequestId, string(reqByte)))
	rsp, err := ogRestClient.Do(string(reqByte))
	if err != nil {
		logger.Error("error in searching: ", err)
		return
	}
	leRsp.FullRspRaw = []byte(rsp)
	logger.Debug(fmt.Sprintf("[%v]: searching rsp: %v", ogRestClient.RequestId, rsp))
	leRsp.StatusCode = ogRestClient.StatusCode

	if ogRestClient.StatusCode == http.StatusOK {
		opRsp := SearchEntitiesRsp{}
		err = json.Unmarshal([]byte(rsp), &opRsp)
		if err != nil {
			logger.Error("Req: ", string(reqByte))
			logger.Error("JSON RSP: ", rsp)
			logger.Error("error unmarshalling: ", err)
			return
		}
		leRsp.OpSearchResponse = opRsp
		return
	}
	err = fmt.Errorf("status code rsp: %v", ogRestClient.StatusCode)
	return
}
func (s *EntitiesSearcher) LaunchNewSearchEntity(clientOpts http_client.ClientOptions) (leRsp LaunchEntityRsp, err error) {
	ogResClient := http_client.NewSearchEntitiesRestClient(clientOpts)
	searchReq := SearchReq{
		Filter: s.Filter,
		Select: s.Selects,
		Limit:  s.Limit,
	}
	reqByte, err := json.Marshal(searchReq)
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Debug(fmt.Sprintf("[%v]: searching req: %v", ogResClient.RequestId, string(reqByte)))
	rsp, err := ogResClient.Do(string(reqByte))
	leRsp.StatusCode = ogResClient.StatusCode
	if err != nil {
		logger.Error("error in searching: ", err)
		return
	}
	leRsp.FullRspRaw = []byte(rsp)
	logger.Debug(fmt.Sprintf("[%v]: searching rsp: %v", ogResClient.RequestId, rsp))

	if ogResClient.StatusCode == http.StatusOK {
		var dataRsp = map[string]interface{}{}
		err = json.Unmarshal([]byte(rsp), &dataRsp)
		if err != nil {
			logger.Error("Req: ", string(reqByte))
			logger.Error("JSON RSP: ", rsp)
			logger.Error("error unmarshalling: ", err)
			return
		}
		if eIface, found := dataRsp["entities"]; found {
			if entitiesI, ok := eIface.([]interface{}); ok {
				if len(entitiesI) != 0 {
					if entitieRsp, ok2 := entitiesI[0].(map[string]interface{}); ok2 {
						leRsp.EntitieRsp = entitieRsp
						return
					}
				}
			}
		}
	}
	err = fmt.Errorf("status code rsp: %v", ogResClient.StatusCode)
	return
}
