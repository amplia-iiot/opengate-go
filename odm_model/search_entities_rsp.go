package odm_model

type SearchEntitiesRsp struct {
	Page       Page                 `json:"page"`
	Entities   []Entities           `json:"entities,omitempty"`
	Operations []OperationSearching `json:"operations,omitempty"`
}
type Page struct {
	Number int `json:"number"`
}
type Entities struct {
	Provision          *ProvisionEntitie   `json:"provision,omitempty"`
	Device             *DeviceEntitie      `json:"device,omitempty"`
	CollectProtVersion *CollectProtVersion `json:"protVersion,omitempty"`
	// Others
}
type OperationSearching struct {
	JobId        string                 `json:"jobId"`
	EntityId     string                 `json:"entityId"`
	ResourceType string                 `json:"resourceType"`
	OperationId  string                 `json:"operationId"`
	Name         string                 `json:"name"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	Attempts     Attempts               `json:"attempts"`
	Notify       bool                   `json:"notify"`
	Execution    Execution              `json:"execution"`
	User         string                 `json:"user"`
	Status       string                 `json:"status"`
	Steps        []Step                 `json:"steps,omitempty"`
	Date         string                 `json:"date"`
}
type Attempts struct {
	Total   int `json:"total"`
	Current int `json:"current"`
}
type Execution struct {
	ActivatedDate string `json:"activatedDate"`
}
type DeviceEntitie struct {
	Topology *Topology `json:"topology,omitempty"`
	Software *Software `json:"software,omitempty"`
}
type Software struct {
	Current Current `json:"_current"`
}
type Topology struct {
	Path *Path `json:"path,omitempty"`
}
type Path struct {
	Current Current `json:"_current"`
}
type CollectProtVersion struct {
	Current Current `json:"_current"`
}
