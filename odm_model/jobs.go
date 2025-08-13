package odm_model

import (
	og_http "github.com/amplia-iiot/opengate-go/http_client"
)

func SetJobInProgress(clientOptions og_http.ClientOptions) (err error) {
	ogClient := og_http.NewSetJobInInProgress(clientOptions)
	_, err = ogClient.Do("")
	return
}
