package http_client

func NewSearchBundleRestClient(clientOptions ClientOptions) *Client {
	odmurl := GetBundleUrl(clientOptions, BUNDLE_SEARCHING)
	clientOptions.Url = odmurl
	clientOptions.Method = POST
	return NewClient(clientOptions)
}
func NewGetBundleRestClient(clientOptions ClientOptions) *Client {
	odmurl := GetBundleUrl(clientOptions, BUNDLE_GET)
	clientOptions.Url = odmurl
	clientOptions.Method = GET
	return NewClient(clientOptions)
}
func NewDownloadFileBundleRestClient(clientOptions ClientOptions) *Client {
	clientOptions.Method = GET
	client := NewClient(clientOptions)
	client.OnlyByteRsp()
	return client
}
func NewSearchDevOperations(clientOptions ClientOptions) *Client {
	odmUrl := GetFullUrlFromDevOperations(clientOptions)
	clientOptions.Url = odmUrl
	clientOptions.Method = POST
	return NewClient(clientOptions)
}
func NewSetJobInInProgress(clientOptions ClientOptions) *Client {
	odmUrl := GetManageJobInProgress(clientOptions)
	odmUrl = odmUrl + "?state=IN_PROGRESS"
	clientOptions.Url = odmUrl
	clientOptions.Method = PUT
	return NewClient(clientOptions)
}
func NewCollectRestClient(clientOptions ClientOptions) *Client {
	odmUrl := GetFullUrlFromDeviceId(clientOptions, COLLECTION)
	clientOptions.Url = odmUrl
	clientOptions.Method = POST
	return NewClient(clientOptions)
}
func NewSearchEntitiesRestClient(clientOptions ClientOptions) *Client {
	odmUrl := GetSearchUrl(clientOptions)
	clientOptions.Url = odmUrl
	clientOptions.Method = POST
	return NewClient(clientOptions)
}
func NewOperationRestClient(clientOptions ClientOptions) *Client {
	odmUrl := GetFullUrlFromDeviceId(clientOptions, ASYNC_OP_RSP)
	clientOptions.Url = odmUrl
	clientOptions.Method = POST
	return NewClient(clientOptions)
}
func NewDeviceRestClient(clientOptions ClientOptions) *Client {
	odmUrl := GetFullUrlFromProvision(clientOptions, PROVISION_NEW_DEVICE)
	clientOptions.Url = odmUrl
	clientOptions.Method = POST
	return NewClient(clientOptions)
}
func NewGetDeviceRestClient(clientOptions ClientOptions) *Client {
	odmUrl := GetFullUrlFromProvision(clientOptions, PROVISION_GET_DEVICE)
	clientOptions.Url = odmUrl
	clientOptions.Method = GET
	return NewClient(clientOptions)
}
func NewGetEntitiesRestClient(clientOptions ClientOptions) *Client {
	odmUrl := GetFullUrlFromProvision(clientOptions, PROVISION_GET_ENTITIE)
	clientOptions.Url = odmUrl
	clientOptions.Method = GET
	return NewClient(clientOptions)
}
func NewGetApiKeyRestClient(clientOptions ClientOptions) *Client {
	odmUrl := GetFullUrlFromUser(clientOptions, PROVISION_GET_USERS)
	clientOptions.Url = odmUrl
	clientOptions.Method = GET
	return NewClient(clientOptions)
}

// making health-check with the users
func HealCheck(clientOptions ClientOptions) *Client {
	clientOptions.Protocol = "https"
	odmUrl := GetFullUrlFromUser(clientOptions, PROVISION_GET_USERS)
	clientOptions.Url = odmUrl
	clientOptions.Method = GET
	return NewClient(clientOptions)
}
func GenericRestClient(clientOptions ClientOptions) *Client {
	client := NewClient(clientOptions)
	if clientOptions.Pass != "" {
		client.WithApiPass(clientOptions.Pass)
	}
	return client
}
