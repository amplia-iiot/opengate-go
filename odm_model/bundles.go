package odm_model

import (
	"encoding/json"
	"errors"
	"net/http"

	og_http "github.com/amplia-iiot/opengate-go/http_client"
	"github.com/amplia-iiot/opengate-go/logger"
)

type DeploymentElement struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Type        string `json:"type"`
	Path        string `json:"path"`
	Order       int    `json:"order"`
	Size        int    `json:"size"`
	Operation   string `json:"operation"`
	OldName     string `json:"oldName"`
	OldVersion  string `json:"oldVersion"`
	OldPath     string `json:"oldPath"`
	DownloadURL string `json:"downloadUrl"`
}
type BundleRsp struct {
	DeploymentElements []DeploymentElement `json:"deploymentElement"`
}
type BundleRequester struct {
	RequestId string
}

func (b *BundleRequester) GetDeployElements(clientOptions og_http.ClientOptions) (BundleRsp, error) {
	restClient := og_http.NewGetBundleRestClient(clientOptions)
	restClient.WithRequestId(b.RequestId)
	rsp, err := restClient.Do("")
	if err != nil {
		return BundleRsp{}, err
	}
	if restClient.StatusCode == http.StatusNoContent {
		return BundleRsp{}, errors.New("deploy element not found")
	}
	bRsp := BundleRsp{}
	err = json.Unmarshal([]byte(rsp), &bRsp)
	if err != nil {
		logger.Error("JSON RSP: ", rsp)
		logger.Error("error unmarshalling: ", err)
		return BundleRsp{}, err
	}
	return bRsp, nil
}
func (b *BundleRequester) GetFileBytes(clientOptions og_http.ClientOptions) ([]byte, error) {
	downClient := og_http.NewDownloadFileBundleRestClient(clientOptions)
	downClient.WithRequestId(b.RequestId)
	_, err := downClient.Do("")
	if err != nil {
		return nil, err
	}
	return downClient.BytesRsp, nil
}
