package connectorog

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/amplia-iiot/opengate-go/logger"
	utils "github.com/amplia-iiot/opengate-go/utils"
)

const datamodelsPath = "resources/ds_models/"

var matcher []ModelOG
var crudMatcher CrudMatcher

var filesembed embed.FS

func ReadAllModels(f embed.FS) error {
	setEmbedFiles(f)
	return readAllModels()
}
//alternativa a ReadAllModels. Aqui tu le pasas todos los modelos ya cargados
func SetModels(models []ModelOG) {
	matcher = models
}

// func ReadAllModels(models map[string]interface{})
func setEmbedFiles(f embed.FS) {
	filesembed = f
}
func readAllModels() error {
	filesInfo, err := ReadAll(filesembed, datamodelsPath)
	if err != nil {
		return err
	}
	matcher = make([]ModelOG, 0)
	for _, fileInfo := range filesInfo {
		if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".json") {
			var newModelOG ModelOG
			jsonFile, err := OpenFile(fileInfo.Name(), datamodelsPath, filesembed)
			if err != nil {
				return err
			}
			newModelOG.ModelName = strings.TrimSuffix(fileInfo.Name(), ".json")
			defer jsonFile.Close()
			var match []*Relation = []*Relation{}
			file, err := io.ReadAll(jsonFile)
			if err != nil {
				logger.Error(fmt.Sprintf("error reading json file content %v: %v", fileInfo.Name(), err))
				return err
			}
			err1 := json.Unmarshal([]byte(file), &match)
			if err1 != nil {
				logger.Error(fmt.Sprintf("error unmarshaling json file %v: %v", fileInfo.Name(), err1))
				return err1
			}
			newModelOG.Relations = match
			matcher = append(matcher, newModelOG)
			logger.Info("OG " + newModelOG.ModelName + " catalog loaded")
		}
	}
	return nil
}

type CrudMatcher struct {
}

func GetCrudMatcher() *CrudMatcher {
	return &crudMatcher
}

type Relation struct {
	Field        string         `json:"field,omitempty"`
	OgDataStream string         `json:"ogDataStream,omitempty"`
	Alias        string         `json:"alias,omitempty"`
	DataType     string         `json:"dataType,omitempty"`
	SubRelations []*SubRelation `json:"ogNamesArr,omitempty"`
	Enums        []*Enums       `json:"enums,omitempty"`
	Factor       string         `json:"factor,omitempty"`
}
type Enums struct {
	CollectValue string `json:"collectValue,omitempty"`
	OGValue      string `json:"ogValue,omitempty"`
}
type SubRelation struct {
	Field        string `json:"field,omitempty"`
	OgDataStream string `json:"ogDataStream,omitempty"`
	DataType     string `json:"dataType,omitempty"`
	Factor       string `json:"factor,omitempty"`
}

type ModelOG struct {
	ModelName string
	Relations []*Relation
}

// create a complete model. If the model exists, will be created.
// It updates the datamodel's folder and reaload the info on memory, so it must be
// called after a CRUD over datamodels
func (c *CrudMatcher) CreateModel(modelName string, relations []*Relation) error {
	content, err := json.Marshal(relations)
	if err != nil {
		return err
	}
	fileName := modelName + ".json"
	err1 := WriteInFolder(content, datamodelsPath, fileName)
	if err1 != nil {
		return err1
	}
	readAllModels()
	return nil
}
func (c *CrudMatcher) UpdateModel(modelName string, modifiedRelations []*Relation) error {
	allPresentRelations := GetRelations(modelName)
	if allPresentRelations == nil {
		logger.Info(fmt.Sprintf("%v model does not exist, creating....", modelName))
		return c.CreateModel(modelName, modifiedRelations)
	}
	for _, modifiedRe := range modifiedRelations {
		presentR := GetRelation(modifiedRe.Field, modelName)
		if presentR == nil { //son nuevas relaciones
			allPresentRelations = append(allPresentRelations, presentR)
		} else {
			replaceRelation(modifiedRe, modelName)
		}
	}
	return c.CreateModel(modelName, allPresentRelations)
}
func (c *CrudMatcher) GetAllModels() []ModelOG {
	return matcher
}
func (c *CrudMatcher) DeleteModel(modelName string) {
	for i, model := range matcher {
		if model.ModelName == modelName {
			matcher = append(matcher[:i], matcher[i+1:]...)
		}
	}
	allPresentRelations := GetRelations(modelName)
	c.CreateModel(modelName, allPresentRelations)
}
func (c *CrudMatcher) ExistModel(modelName string) (bool, ModelOG) {
	for _, model := range matcher {
		if model.ModelName == modelName {
			return true, model
		}
	}
	return false, ModelOG{}
}
func replaceRelation(newRelation *Relation, modelName string) {
	if allPresentRelations := GetRelations(modelName); allPresentRelations != nil {
		for pos, presentRelation := range allPresentRelations {
			if presentRelation.Field == newRelation.Field {
				allPresentRelations[pos] = newRelation
				return
			}
		}
	}
}
func GetRelations(modelName string) []*Relation {
	for _, model := range matcher {
		if model.ModelName == modelName {
			return model.Relations
		}
	}
	return nil
}

// if only there is one model it is not necessary pass the name
func GetRelation(fieldName, modelName string) *Relation {
	for _, model := range matcher {
		if modelName == model.ModelName || (len(matcher) == 1) {
			for _, e := range model.Relations {
				if fieldName == e.Field {
					return e
				}
			}
		}
	}
	return nil
}
func GetRelationByAlias(alias, modelName string) *Relation {
	for _, model := range matcher {
		if modelName == model.ModelName || (len(matcher) == 1) {
			for _, e := range model.Relations {
				if alias == e.Alias {
					return e
				}
			}
		}
	}
	return nil
}

func (r *Relation) BuildMapJsonElementInSub(deviceFieldName, deviceValue string, jsonMap map[string]interface{}) {
	var subRelation *SubRelation
	for _, sub := range r.SubRelations {
		if sub.Field == deviceFieldName {
			subRelation = sub
			break
		}
	}
	if subRelation == nil {
		return
	}
	switch subRelation.DataType {
	case "boolean":
		if boolValue, err := strconv.ParseBool(deviceValue); err == nil {
			jsonMap[subRelation.OgDataStream] = boolValue
		}
	case "integer": //viene un entero en un string y hay que pasarlo a entero. Ej: "12" --> 12
		if i, err := strconv.ParseInt(deviceValue, 10, 64); err == nil {
			jsonMap[subRelation.OgDataStream] = i
		}
	case "number": //viene un entero en un string y hay que pasarlo a float. Ej: "12" --> 12
		if inputNumberFloat, err := strconv.ParseFloat(deviceValue, 64); err == nil {
			if r.Factor == "" {
				r.Factor = "1"
			}
			if factor, _ := strconv.ParseFloat(r.Factor, 64); factor != 0 {
				jsonMap[subRelation.OgDataStream] = utils.FloatToFixed(inputNumberFloat*factor, 2)
			}
		}
	case "hexstring-littleendian-int":
		if i, err := strconv.ParseInt(utils.ReverseIn2Bytes(deviceValue), 16, 64); err == nil {
			jsonMap[subRelation.OgDataStream] = i
		}
	case "hexstring-littleendian-intstring":
		if i, err := strconv.ParseInt(utils.ReverseIn2Bytes(deviceValue), 16, 64); err == nil {
			jsonMap[subRelation.OgDataStream] = fmt.Sprintf("%v", i)
		}
	case "hexstring-int": //viene un hex en string y hay que pasarlo a int. Ej: "A" --> 10
		if i, err := strconv.ParseInt(deviceValue, 16, 64); err == nil {
			jsonMap[subRelation.OgDataStream] = i
		}
	case "a2-8-string-float": //viene en hex en formato complemento A2 (8 bits) y hay que pasarlo a float
		if number, err := strconv.ParseInt(deviceValue, 16, 0); err == nil {
			value := utils.ReverseBits(fmt.Sprintf("%08b", number))
			if r.Factor == "" {
				r.Factor = "1"
			}
			if factor, _ := strconv.ParseFloat(subRelation.Factor, 64); factor != 0 {
				jsonMap[subRelation.OgDataStream] = utils.FloatToFixed(float64(value)*factor, 2)
			}
		}
	case "a2-8-string-int": //viene en hex en formato complemento A2 (8 bits) y hay que pasarlo a entero
		if number, err := strconv.ParseInt(deviceValue, 16, 0); err == nil {
			jsonMap[subRelation.OgDataStream] = utils.ReverseBits(fmt.Sprintf("%08b", number))
		}
	case "a2-8-string-intstring": //como el anterior pero devolviendolo en formato string
		if number, err := strconv.ParseInt(deviceValue, 16, 0); err == nil {
			jsonMap[subRelation.OgDataStream] = fmt.Sprintf("%v", utils.ReverseBits(fmt.Sprintf("%08b", number)))
		}
	case "hexstring-intstring": //viene en string en formato hex y hay que pasarlo a string en formato int. Ej: "B002" --> "45058"
		if numberInt, err := strconv.ParseInt(deviceValue, 16, 64); err == nil {
			jsonMap[subRelation.OgDataStream] = fmt.Sprintf("%v", numberInt)
		}
	case "timestamp": //viene fecha en formato epoch en un string y se pasa a string en formato ISO8601
		if i, err := strconv.ParseInt(deviceValue, 10, 64); err == nil {
			jsonMap[subRelation.OgDataStream] = utils.ParseEpoch2IsoMs(i)
		}
	case "string": //se deja tal cual
		jsonMap[subRelation.OgDataStream] = deviceValue
	default:
		jsonMap[subRelation.OgDataStream] = deviceValue
	}
}
func (r *Relation) GetValueInRelation(inputValue string) interface{} {
	if r.Enums != nil {
		return r.getEnumValue(r.Enums, inputValue)
	}
	return r.getConvertedValue(inputValue)
}
func (r *Relation) getConvertedValue(inputValue string) interface{} {
	switch r.DataType {
	case "boolean":
		if boolValue, err := strconv.ParseBool(inputValue); err == nil {
			return boolValue
		}
	case "integer": //viene un entero en un string y hay que pasarlo a entero. Ej: "12" --> 12
		if i, err := strconv.ParseInt(inputValue, 10, 64); err == nil {
			return i
		}
	case "number": //viene un entero en un string y hay que pasarlo a float. Ej: "12" --> 12
		if inputNumberFloat, err := strconv.ParseFloat(inputValue, 64); err == nil {
			if r.Factor == "" {
				r.Factor = "1"
			}
			if factor, _ := strconv.ParseFloat(r.Factor, 64); factor != 0 {
				return utils.FloatToFixed(inputNumberFloat*factor, 2)
			}
		}
		// viene un hex en string, el cual esta el little endian y hay que pasarlo a entero.
		// Por tanto: "3FB60100" --> "0001B63F" --> 112191
	case "hexstring-littleendian-int":
		if i, err := strconv.ParseInt(utils.ReverseIn2Bytes(inputValue), 16, 64); err == nil {
			return i
		}
		// viene un hex en string, el cual esta el little endian y hay que pasarlo a entero en formato string.
		// Por tanto: "3FB60100" --> "0001B63F" --> "112191"
	case "hexstring-littleendian-intstring":
		if i, err := strconv.ParseInt(utils.ReverseIn2Bytes(inputValue), 16, 64); err == nil {
			return fmt.Sprintf("%v", i)
		}
	case "hexstring-int": //viene un hex en string y hay que pasarlo a int. Ej: "A" --> 10
		if i, err := strconv.ParseInt(inputValue, 16, 64); err == nil {
			return i
		}
	case "a2-8-string-float": //viene en hex en formato complemento A2 (8 bits) y hay que pasarlo a float
		if number, err := strconv.ParseInt(inputValue, 16, 0); err == nil {
			value := utils.ReverseBits(fmt.Sprintf("%08b", number))
			if r.Factor == "" {
				r.Factor = "1"
			}
			if factor, _ := strconv.ParseFloat(r.Factor, 64); factor != 0 {
				return utils.FloatToFixed(float64(value)*factor, 2)
			}
		}
	case "a2-8-string-int": //viene en hex en formato complemento A2 (8 bits) y hay que pasarlo a entero
		if number, err := strconv.ParseInt(inputValue, 16, 0); err == nil {
			return utils.ReverseBits(fmt.Sprintf("%08b", number))
		}
	case "a2-8-string-intstring": //como el anterior pero devolviendolo en formato string
		if number, err := strconv.ParseInt(inputValue, 16, 0); err == nil {
			return fmt.Sprintf("%v", utils.ReverseBits(fmt.Sprintf("%08b", number)))
		}
	case "hexstring-intstring": //viene en string en formato hex y hay que pasarlo a string en formato int. Ej: "B002" --> "45058"
		if numberInt, err := strconv.ParseInt(inputValue, 16, 64); err == nil {
			return fmt.Sprintf("%v", numberInt)
		}
	case "timestamp": //viene fecha en formato epoch (s) en un string y se pasa a string en formato ISO8601
		if i, err := strconv.ParseInt(inputValue, 10, 64); err == nil {
			return utils.ParseEpoch2Iso(i)
		}
	case "string": //se deja tal cual
		return inputValue
	default:
		return inputValue
	}
	return inputValue
}
func (r *Relation) getEnumValue(enums []*Enums, value2match string) interface{} {
	for _, enum := range enums {
		if enum.CollectValue == value2match {
			return r.getConvertedValue(enum.OGValue)
		}
	}
	// mejor esto que devolver un nil, de cara a construir el json, al igual que en el resto de funciones
	return value2match
}
