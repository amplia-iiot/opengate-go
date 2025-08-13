package odm_model

import (
	"encoding/json"
	"fmt"
)

type AdministrativeStateType string

const (
	DEVICE_ADMIN_STATE_ACTIVE  AdministrativeStateType = "ACTIVE"
	DEVICE_SPECIFICTYPE_METER  AdministrativeStateType = "METER"
	DEVICE_SPECIFICTYPE_CNC    AdministrativeStateType = "CONCENTRATOR"
	DEVICE_SPECIFICTYPE_COMHUB AdministrativeStateType = "COMHUB"
)

type ProvisionOrgan struct {
	Provision ProvisionEntitie `json:"provision,omitempty"`
}
type ProvisionEntitie struct {
	Administration Administration `json:"administration,omitempty"`
	Device         *Device        `json:"device,omitempty"`
	// Others
}
type Administration struct {
	Channel      Channel      `json:"channel,omitempty"`
	Organization Organization `json:"organization,omitempty"`
	ServiceGroup ServiceGroup `json:"serviceGroup,omitempty"`
	Plan         Plan         `json:"plan,omitempty"`
}
type Device struct {
	Identifier          Identifier          `json:"identifier,omitempty"`
	AdministrativeState AdministrativeState `json:"administrativeState,omitempty"`
	Name                Name                `json:"name,omitempty"`
	SpecificType        *SpecificType       `json:"specificType,omitempty"`
}

type Current struct {
	Value      interface{} `json:"value"`
	Date       string      `json:"date,omitempty"`
	At         string      `json:"at,omitempty"`
	From       string      `json:"from,omitempty"`
	Source     string      `json:"source,omitempty"`
	SourceInfo string      `json:"sourceInfo,omitempty"`
}
type Channel struct {
	Current Current `json:"_current"`
}
type Organization struct {
	Current Current `json:"_current"`
}
type ServiceGroup struct {
	Current Current `json:"_current"`
}
type Plan struct {
	Current Current `json:"_current"`
}
type Identifier struct {
	Current Current `json:"_current"`
}
type AdministrativeState struct {
	Current Current `json:"_current"`
}
type Name struct {
	Current Current `json:"_current"`
}
type SpecificType struct {
	Current Current `json:"_current"`
}

type Provisioner struct {
	ProOptions
}
type ProOptions struct {
	Channel      string
	Organization string
	ServiceGroup string
	Plan         string
	DeviceId     string
}

func (p *Provisioner) NewProvisionMeter() string {
	administration := Administration{
		Channel: Channel{
			Current: Current{
				Value: p.Channel,
			},
		},
		Organization: Organization{
			Current: Current{
				Value: p.Organization,
			},
		},
		ServiceGroup: ServiceGroup{
			Current: Current{
				Value: p.ServiceGroup,
			},
		},
		Plan: Plan{
			Current: Current{
				Value: p.Plan,
			},
		},
	}
	device := &Device{
		Identifier: Identifier{
			Current: Current{
				Value: p.DeviceId,
			},
		},
		AdministrativeState: AdministrativeState{
			Current: Current{
				Value: DEVICE_ADMIN_STATE_ACTIVE,
			},
		},
		Name: Name{
			Current: Current{
				Value: p.DeviceId,
			},
		},
		SpecificType: &SpecificType{
			Current: Current{
				Value: DEVICE_SPECIFICTYPE_METER,
			},
		},
	}
	provision := ProvisionOrgan{
		Provision: ProvisionEntitie{
			Administration: administration,
			Device:         device,
		},
	}
	marshalled, _ := json.Marshal(provision)
	return string(marshalled)
}

func (c Current) GetStrValue() string {
	return fmt.Sprintf("%v", c.Value)
}
func (c Current) GetElementFromObjectArr(elementNumber int, targetName string) (jsonValue string) {
	if result, ok := c.Value.([]interface{}); ok {
		if result2, ok2 := result[elementNumber].(map[string]interface{}); ok2 {
			jsonValue = fmt.Sprintf("%v", result2[targetName])
		}
	}
	return
}
func (c Current) GetArrayStr() (result []string) {
	if resultI, ok := c.Value.([]interface{}); ok {
		for _, element := range resultI {
			if elementString, ok2 := element.(string); ok2 {
				result = append(result, elementString)
			}
		}
	}
	return
}
