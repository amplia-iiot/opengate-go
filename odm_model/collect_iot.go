package odm_model

import (
	"encoding/json"
)

type CollectIot struct {
	Version     string              `json:"version,omitempty"`
	Device      string              `json:"device,omitempty"`
	Path        []string            `json:"path,omitempty"`
	Datastreams []CollectDatastream `json:"datastreams,omitempty"`
}
type CollectDatastream struct {
	Id         string      `json:"id,omitempty"`
	Feed       string      `json:"feed,omitempty"`
	Datapoints []Datapoint `json:"datapoints,omitempty"`
}
type Datapoint struct {
	At    int64       `json:"at,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

func ToString(odm interface{}) (string, error) {
	rsp, err := json.Marshal(odm)
	if err != nil {
		return "", err
	}
	return string(rsp), nil
}
