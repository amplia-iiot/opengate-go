package connectorog

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/amplia-iiot/opengate-go/http_client"
	"github.com/amplia-iiot/opengate-go/odm_model"
	"github.com/amplia-iiot/opengate-go/logger"
)

// para distinguir entre un valor complejo que dentro lleva un array con un valor complejo normal
// Ejemplo de complejo:
// {"id":"device.software","datapoints":[{"at":1697630789000,"value":[{"name":"28","type":"FIRMWARE","version":"28"}]}]}
// Ejemplo de normal:
// {"id":"nodoMsRaw","datapoints":[{"value":{"concentratorId":"866207055045441","mType":13,"mTypeRepeated":"13-false","repeated":false,"sizeB":140}}]}
type ComplexValueType string

var (
	ARRAY  ComplexValueType = "array"
	SIMPLE ComplexValueType = "simple"
)

type CollectInfo struct {
	ModelName,
	FieldName,
	FieldValue string
	Ts           int64
	complexValue *ComplexValue
}

type ComplexValue struct {
	DataType  ComplexValueType
	SubValues []Values
}
type Values struct {
	SubRelations []SubCollectInfo
}

// para las subrelaciones. Por tanto valores complejos
type SubCollectInfo struct {
	FieldValue string
	FieldName  string
}
type GenerateCollect interface {
	Fill() []odm_model.CollectIot
}
type ParseCollect interface {
	CollectToRestData(collect []odm_model.CollectIot) []string
}
type Normalizer struct {
	ClientOptions       http_client.ClientOptions //clientOptions para el cliente rest que crea el metodo SendCollectIoT()
	CollectionGenerator GenerateCollect
	ParseCollect        ParseCollect
	ManageError         *bool //si true, la propia libreria gestiona los errores al recolectar con su sistema de reintentos, sino devuelve el error
	errorInCollect      error
	CustomRestClient    *http_client.Client //cliente que sustituye el de por defecto que se crea con el ClientOptions
}

func (c *CollectInfo) WithComplexValue(v Values) {
	c.complexValue = &ComplexValue{
		DataType:  SIMPLE,
		SubValues: []Values{v},
	}
}
func (c *CollectInfo) WithComplexValueArr(v []Values) {
	c.complexValue = &ComplexValue{
		DataType:  ARRAY,
		SubValues: v,
	}
}
func NewValues(sub []SubCollectInfo) Values {
	return Values{SubRelations: sub}
}
func NewNormalizer() *Normalizer {
	return &Normalizer{}
}
func NewCollectIoTGrouped(collected []CollectInfo, device string, path []string, byAlias bool) (collects []odm_model.CollectIot) {
	var grouped map[string][]CollectInfo = make(map[string][]CollectInfo)
	for _, coll := range collected {
		grouped[coll.FieldName] = append(grouped[coll.FieldName], coll)
	}
	var dataStreams []odm_model.CollectDatastream = []odm_model.CollectDatastream{}
	for _, collArr := range grouped {
		dataPoints := []odm_model.Datapoint{}
		var ogDataStream string
		for _, collInfo := range collArr {
			if byAlias {
				if dataStream := GetSimpleDSByAlias(collInfo); dataStream != nil {
					ogDataStream = dataStream.Id
					dataPoints = append(dataPoints, dataStream.Datapoints...)
				}
			} else {
				if dataStream := GetSimpleDS(collInfo); dataStream != nil {
					ogDataStream = dataStream.Id
					dataPoints = append(dataPoints, dataStream.Datapoints...)
				}
			}
		}
		if ogDataStream != "" {
			dataStream := &odm_model.CollectDatastream{
				Id:         ogDataStream,
				Datapoints: dataPoints,
			}
			dataStreams = append(dataStreams, *dataStream)
		}
	}
	if len(dataStreams) != 0 {
		collects = []odm_model.CollectIot{{
			Version:     odm_model.Version,
			Device:      device,
			Path:        path,
			Datastreams: dataStreams,
		}}
	}
	return
}
func NewCollectIoTSimple(collected []CollectInfo, device string, path []string, byAlias bool) (collects []odm_model.CollectIot) {
	var dataStreams_cnc []odm_model.CollectDatastream
	for _, c := range collected {
		if byAlias {
			if dataStream := GetSimpleDSByAlias(c); dataStream != nil {
				dataStreams_cnc = append(dataStreams_cnc, *dataStream)
			}
		} else {
			if dataStream := GetSimpleDS(c); dataStream != nil {
				dataStreams_cnc = append(dataStreams_cnc, *dataStream)
			}
		}
	}
	if len(dataStreams_cnc) != 0 {
		collects = []odm_model.CollectIot{{
			Version:     odm_model.Version,
			Device:      device,
			Path:        path,
			Datastreams: dataStreams_cnc,
		}}
	}
	return
}

func GetSimpleDSByAlias(c CollectInfo) *odm_model.CollectDatastream {
	var dataStream *odm_model.CollectDatastream = &odm_model.CollectDatastream{}
	if relation := GetRelationByAlias(c.FieldName, c.ModelName); relation != nil {
		var dataPoint odm_model.Datapoint
		if c.Ts != 0 {
			dataPoint.At = c.Ts
		}
		if c.complexValue != nil {
			dpValue := getComplexValueForDataPoint(c, relation)
			if dpValue == nil {
				return nil
			}
			dataPoint.Value = dpValue
		} else if ogValue := relation.GetValueInRelation(c.FieldValue); ogValue != nil {
			dataPoint.Value = ogValue
		} else {
			return nil
		}
		dataStream.Datapoints = []odm_model.Datapoint{dataPoint}
		dataStream.Id = relation.OgDataStream
		return dataStream
	}
	return nil
}
func GetSimpleDS(c CollectInfo) *odm_model.CollectDatastream {
	var dataStream *odm_model.CollectDatastream = &odm_model.CollectDatastream{}
	if relation := GetRelation(c.FieldName, c.ModelName); relation != nil {
		var dataPoint odm_model.Datapoint
		if c.Ts != 0 {
			dataPoint.At = c.Ts
		}
		if c.complexValue != nil {
			dpValue := getComplexValueForDataPoint(c, relation)
			if dpValue == nil {
				return nil
			}
			dataPoint.Value = dpValue
		} else if ogValue := relation.GetValueInRelation(c.FieldValue); ogValue != nil {
			dataPoint.Value = ogValue
		} else {
			return nil
		}
		dataStream.Datapoints = []odm_model.Datapoint{dataPoint}
		dataStream.Id = relation.OgDataStream
		return dataStream
	}
	return nil
}

func getComplexValueForDataPoint(c CollectInfo, relation *Relation) interface{} {
	if c.complexValue.DataType == ARRAY {
		var finalValueDataPointArr []interface{}
		for _, subValue := range c.complexValue.SubValues {
			var complexValue map[string]interface{} = make(map[string]interface{})
			for _, sub := range subValue.SubRelations {
				relation.BuildMapJsonElementInSub(sub.FieldName, sub.FieldValue, complexValue)
			}
			if len(complexValue) == 0 {
				continue
			}
			finalValueDataPointArr = append(finalValueDataPointArr, complexValue)
		}
		if len(finalValueDataPointArr) == 0 {
			return nil
		}
		return finalValueDataPointArr
	} else {
		var complexValue map[string]interface{} = make(map[string]interface{})
		if len(c.complexValue.SubValues) == 1 {
			for _, sub := range c.complexValue.SubValues[0].SubRelations {
				relation.BuildMapJsonElementInSub(sub.FieldName, sub.FieldValue, complexValue)
			}
		}
		if len(complexValue) == 0 {
			return nil
		}
		return complexValue
	}
}
func (n *Normalizer) GetErrorInCollect() error {
	return n.errorInCollect
}
func (n *Normalizer) WithClientOptions(clientOptions http_client.ClientOptions) {
	n.ClientOptions = clientOptions
}
func (n *Normalizer) WithCollGenerator(g GenerateCollect) {
	n.CollectionGenerator = g
}
func (n *Normalizer) WithParseCollect(p ParseCollect) {
	n.ParseCollect = p
}
func (n *Normalizer) WithManageError(manage bool) {
	n.ManageError = &manage
}
func (n *Normalizer) SendCollectIoT() error {
	if n.ManageError != nil && !*n.ManageError { //no queremos reintentos ya que se ha configurado que la libreria no maneje los errores de http
		n.ClientOptions.MaxRetries = 0
	}
	var restClient *http_client.Client
	if n.CustomRestClient != nil {
		restClient = n.CustomRestClient
	} else {
		restClient = http_client.NewCollectRestClient(n.ClientOptions)
	}
	if n.CollectionGenerator == nil {
		n.errorInCollect = errors.New("it's necessary implement a CollectionGenerator")
		logger.Error(n.errorInCollect.Error())
		return n.errorInCollect
	}
	collectIoT := n.CollectionGenerator.Fill()
	var odmDataStrArr []string
	if n.ParseCollect != nil { //por si hay alguna particularidad a la hora de organizar los []collectIoT
		odmDataStrArr = n.ParseCollect.CollectToRestData(collectIoT)
	} else {
		for _, collection := range collectIoT {
			odmData, err := json.Marshal(collection)
			if err != nil {
				n.errorInCollect = errors.New("error marshalling data: " + err.Error())
				logger.Error(n.errorInCollect.Error())
				return n.errorInCollect
			}
			odmDataStrArr = append(odmDataStrArr, string(odmData))
		}
	}
	for index, odmData := range odmDataStrArr {
		logger.Debug(fmt.Sprintf("[%v]: %v", restClient.RequestId, string(odmData)))
		_, err := restClient.Do(string(odmData))
		if err != nil {
			errDetails := fmt.Sprintf("[%v] error performing: [%v]%v, error: %v", restClient.RequestId, restClient.Method, restClient.Url, err)
			if n.ManageError != nil && !*n.ManageError {
				return fmt.Errorf("[err %v]: %v", index, errDetails)
			}
			if n.errorInCollect == nil {
				n.errorInCollect = fmt.Errorf("[err %v]: %v", index, errDetails)
			} else {
				n.errorInCollect = fmt.Errorf("%v | [err %v]: %v", n.errorInCollect.Error(), index, errDetails)
			}
			logger.Error(errDetails)
			logger.Error(fmt.Sprintf("[%v] error collecting: %v", restClient.RequestId, string(odmData)))
		}
	}
	return nil
}
