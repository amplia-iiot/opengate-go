package http_client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/amplia-iiot/opengate-go/logger"
)

const DEFAULT_TIMEOUT_SECONDS = 3 * time.Second
const DEFAULT_TIME_BETWEEN_RETRIES = 5 * time.Second

type RestMethod string

const (
	POST   RestMethod = "POST"
	GET    RestMethod = "GET"
	PUT    RestMethod = "PUT"
	DELETE RestMethod = "DELETE"
	PATCH  RestMethod = "PATCH"
)

type OdmMethod string

// south
const (
	COLLECTION   OdmMethod = "/south/v80/devices/{device_id}/collect/iot"
	ASYNC_OP_RSP OdmMethod = "/south/v80/devices/{device_id}/operation/response"
)

// north
const (
	PROVISION_NEW_DEVICE           OdmMethod = "/north/v80/provision/organizations/{organization}/devices"
	SEARCH_ENTITIES                OdmMethod = "/north/v80/search/entities"
	SEARCH_POST_DEVICES_OPERATIONS OdmMethod = "/north/v80/search/entities/devices/operations"
	PROVISION_GET_DEVICE           OdmMethod = "/north/v80/provision/organizations/{organization}/devices/{identifier}"
	PROVISION_GET_ENTITIE          OdmMethod = "/north/v80/provision/organizations/{organization}/entities/{identifier}"
	PROVISION_GET_USERS            OdmMethod = "/north/v80/provision/users/{user}"
	PROVISION_DEL_USERS            OdmMethod = "/north/v80/provision/users/{user}"
	PROVISION_PUT_USERS            OdmMethod = "/north/v80/provision/users/{user}"
	PROVISION_POST_USERS           OdmMethod = "/north/v80/provision/users"
	SEARCH_POST_USERS              OdmMethod = "/north/v80/search/users"
	SEARCH_POST_USERS_SUMMARY      OdmMethod = "/north/v80/search/users/summary"
	MANAGE_JOB                     OdmMethod = "/north/v80/operation/jobs/{jobid}/operations/{operationid}"
	BUNDLE_SEARCHING               OdmMethod = "/north/v80/search/bundles"
	BUNDLE_GET                     OdmMethod = "/north/v80/provision/bundles/{bundlename}/versions/{bundleversion}/deploymentElements"
)

type ClientOptions struct {
	OGRestOptions
	RestOptions
}
type RestOptions struct {
	Url     string     //
	Method  RestMethod //POST, DELETE, GET ....
	Headers http.Header
}
type OGRestOptions struct {
	Protocol,
	Host,
	Port,
	User,
	Pass,
	ApiKey,
	ApiPass,
	DeviceId,
	Organization,
	JobId,
	BundleName,
	BundleVersion,
	BundleFileUrl,
	OperationId string
	RemovePrefixNorthSouth bool //si este campo esta a true, los metodos que crean la url de conexion eliminan el /north y /south del path de la uris
	// Method    RestMethod //POST, DELETE, GET ....
	RequestId string //para representacion de los logs
	TransPort http.RoundTripper
	ClientTimes
}
type Client struct {
	StatusCode  int  //
	onlyByteRsp bool //
	stopRetries bool
	BytesRsp    []byte //
	ClientOptions
}
type ClientTimes struct {
	TimeBetweenRetries uint //tiempo en segundos
	MaxRetries         uint
	TimeOutInCalls     time.Duration
}

func GetSearchUrl(clientOptions ClientOptions) string {
	url := clientOptions.Protocol + "://" + clientOptions.Host + ":" + clientOptions.Port + string(SEARCH_ENTITIES) + "?defaultSorted=false"
	return clientOptions.reMakeUrl(url)
}

func GetBundleUrl(clientOptions ClientOptions, odmMethod OdmMethod) string {
	odmMethodReplaced := strings.ReplaceAll(string(odmMethod), "{bundlename}", clientOptions.BundleName)
	odmMethodReplaced = strings.ReplaceAll(odmMethodReplaced, "{bundleversion}", clientOptions.BundleVersion)
	url := clientOptions.Protocol + "://" + clientOptions.Host + ":" + clientOptions.Port + odmMethodReplaced
	return clientOptions.reMakeUrl(url)
}
func GetFullUrlFromProvision(clientOptions ClientOptions, odmMethod OdmMethod) string {
	odmMethodReplaced := strings.ReplaceAll(string(odmMethod), "{organization}", clientOptions.Organization)
	if clientOptions.DeviceId != "" {
		odmMethodReplaced = strings.ReplaceAll(odmMethodReplaced, "{identifier}", string(clientOptions.DeviceId))
	}
	url := clientOptions.Protocol + "://" + clientOptions.Host + ":" + clientOptions.Port + odmMethodReplaced
	return clientOptions.reMakeUrl(url)
}
func GetManageJobInProgress(clientOptions ClientOptions) string {
	odmMethodReplaced := strings.ReplaceAll(strings.ReplaceAll(string(MANAGE_JOB), "{jobid}", clientOptions.JobId), "{operationid}", clientOptions.OperationId)
	url := clientOptions.Protocol + "://" + clientOptions.Host + ":" + clientOptions.Port + odmMethodReplaced
	return clientOptions.reMakeUrl(url)
}
func GetFullUrlFromDevOperations(clientOptions ClientOptions) string {
	return clientOptions.reMakeUrl(clientOptions.Protocol + "://" + clientOptions.Host + ":" + clientOptions.Port + string(SEARCH_POST_DEVICES_OPERATIONS))
}
func GetFullUrlFromDeviceId(clientOptions ClientOptions, odmMethod OdmMethod) string {
	odmMethodReplaced := strings.ReplaceAll(string(odmMethod), "{device_id}", clientOptions.DeviceId)
	url := clientOptions.Protocol + "://" + clientOptions.Host + ":" + clientOptions.Port + odmMethodReplaced
	return clientOptions.reMakeUrl(url)
}
func GetFullUrlFromUser(clientOptions ClientOptions, odmMethod OdmMethod) string {
	odmMethodReplaced := strings.ReplaceAll(string(odmMethod), "{user}", clientOptions.User)
	url := clientOptions.Protocol + "://" + clientOptions.Host + odmMethodReplaced
	return clientOptions.reMakeUrl(url)
}

func (o OGRestOptions) reMakeUrl(url string) (newUrl string) {
	newUrl = url
	if o.RemovePrefixNorthSouth {
		re := regexp.MustCompile(`/north|/south`)
		newUrl = re.ReplaceAllString(url, "")
	}
	return
}
func NewClient(clientOpts ClientOptions) *Client {
	c := &Client{
		ClientOptions: clientOpts,
	}
	if len(c.Headers) == 0 {
		c.Headers = make(http.Header)
	}
	if c.TimeOutInCalls == 0 {
		c.WithTimeout(DEFAULT_TIMEOUT_SECONDS)
	}
	return c
}

func (c *Client) WithHeaders(headers http.Header) {
	c.Headers = headers
}
func (c *Client) WithTimeout(t time.Duration) {
	c.TimeOutInCalls = t
}
func (c *Client) WithTimes(ct ClientTimes) {
	c.ClientTimes = ct
}
func (c *Client) OnlyByteRsp() {
	c.onlyByteRsp = true
}
func (c *Client) WithRequestId(requestId string) {
	c.RequestId = requestId
}
func (c *Client) WithApiPass(apiPass string) {
	c.ApiPass = apiPass
}

// evita que la peticion en curso se vuelva a reintentar
func (c *Client) StopRetries() {
	c.stopRetries = true
}
func (c *Client) Do(payload string) (rsp string, err error) {
	retryCount := 0
	for {
		if c.stopRetries {
			err = fmt.Errorf("[%v][%v]:%v retries stopped", c.Method, c.RequestId, c.Url)
			return
		}
		rsp, err = c.doInternal(payload)
		if err == nil {
			if retryCount != 0 {
				logger.Info(fmt.Sprintf("re-call with success [%v][%v]:%v. Retry [%v/%v]", c.Method, c.RequestId, c.Url, retryCount, c.MaxRetries))
			}
			return
		} else {
			logger.Warn(fmt.Sprintf("error asking: [%v][%v]:%v  --> %v. Retry [%v/%v]", c.Method, c.RequestId, c.Url, err, retryCount, c.MaxRetries))
			if retryCount == int(c.MaxRetries) {
				logger.Error(fmt.Sprintf("error nretries exceeded: [%v][%v]:%v  --> %v. Retry [%v/%v]", c.Method, c.RequestId, c.Url, err, retryCount, c.MaxRetries))
				return
			}
			timeBetweenRetries := DEFAULT_TIME_BETWEEN_RETRIES
			if c.TimeBetweenRetries != 0 {
				timeBetweenRetries = time.Duration(c.TimeBetweenRetries) * time.Second
			}
			time.Sleep(timeBetweenRetries)
			retryCount++
		}
	}
}

//	func (c *Client) DoSslInsecure(payload string) (string, error) {
//		return c.doInternal(payload, true)
//	}
func (c *Client) doInternal(payload string) (string, error) {
	timeout := c.TimeOutInCalls
	client := http.Client{
		Timeout:   timeout,
		Transport: c.TransPort,
	}
	logger.Debug(fmt.Sprintf("url to ask: [%v][%v]:%v", c.Method, c.RequestId, c.Url))
	request, err := http.NewRequest(string(c.Method), c.Url, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return "", err
	}
	request.Header = c.Headers
	request.Header.Set("Content-type", "application/json")
	if c.ApiKey != "" {
		request.Header.Set("X-ApiKey", c.ApiKey)
	}
	if c.ApiPass != "" {
		request.Header.Set("X-ApiPass", c.ApiPass)
	}
	// TODO: ver que hacer con esta traza, ya que junto con el resto que hay en el codigo hace que sea redundate cuando se hace un ataque rest
	// incluso en modo debug
	// logger.Debug(fmt.Sprintf("url to ask: [%v]:%v, with header: %v and body: %v", request.Method, request.URL, request.Header, request.Body))
	resp, err := client.Do(request)
	if err != nil {
		return "", err
	}
	c.StatusCode = resp.StatusCode
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error reading response body: ", err)
	}
	c.BytesRsp = body
	var bodyStr string
	if !isStatusSuccess(resp.StatusCode) {
		return "", fmt.Errorf("httpResponse is not success. Code: %v, and message: %v", resp.StatusCode, string(body))
	}
	if !c.onlyByteRsp {
		bodyStr = string(body)
	}
	return bodyStr, nil
}
func isStatusSuccess(statusCode int) bool { return statusCode >= 200 && statusCode < 300 }
