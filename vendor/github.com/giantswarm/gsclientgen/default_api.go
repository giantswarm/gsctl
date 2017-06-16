/*
 * Giant Swarm legacy API
 *
 * Caution: This is an incomplete description of some legacy API functions.
 *
 * OpenAPI spec version: legacy
 *
 * Generated by: https://github.com/swagger-api/swagger-codegen.git
 */

package gsclientgen

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

type DefaultApi struct {
	Configuration *Configuration
}

func NewDefaultApi() *DefaultApi {
	configuration := NewConfiguration()
	return &DefaultApi{
		Configuration: configuration,
	}
}

func NewDefaultApiWithBasePath(basePath string) *DefaultApi {
	configuration := NewConfiguration()
	configuration.BasePath = basePath

	return &DefaultApi{
		Configuration: configuration,
	}
}

/**
 * Create cluster
 * This operation is used to create a new Kubernetes cluster for an organization. The desired configuration can be specified using the __cluster definition format__ (see [external documentation](https://github.com/giantswarm/api-spec/blob/ master/details/CLUSTER_DEFINITION.md) for details). The cluster definition format allows to set a number of optional configuration details, like memory size and number of CPU cores. However, one attribute is __mandatory__ upon creation: The &#x60;owner&#x60; attribute must carry the name of the organization the cluster will belong to. Note that the acting user must be a member of that organization in order to create a cluster. It is *recommended* to also specify the &#x60;name&#x60; attribute to give the cluster a friendly name, like e. g. \&quot;Development Cluster\&quot;. Additional definition attributes can be used. Where attributes are ommitted, default configuration values will be applied. For example, if no &#x60;kubernetes_version&#x60; is specified, the latest version tested and provided by Giant Swarm is used. The &#x60;workers&#x60; attribute, if present, must contain an array of node definition objects. The number of objects given determines the number of workers created. For example, requesting three worker nodes with default configuration can be achieved by submitting an array of three empty objects: &#x60;&#x60;&#x60;\&quot;workers\&quot;: [{}, {}, {}]&#x60;&#x60;&#x60;
 *
 * @param authorization Header to pass an authorization token. The value has to be in the form &#x60;giantswarm &lt;token&gt;&#x60;.
 * @param body New cluster definition
 * @param xRequestID A randomly generated key that can be used to track a request throughout services of Giant Swarm
 * @param xGiantSwarmActivity Name of an activity to track, like \&quot;list-clusters\&quot;
 * @param xGiantSwarmCmdLine If activity has been issued by a CLI, this header can contain the command line
 * @return *V4GenericResponse
 */
func (a DefaultApi) AddCluster(authorization string, body V4AddClusterRequest, xRequestID string, xGiantSwarmActivity string, xGiantSwarmCmdLine string) (*V4GenericResponse, *APIResponse, error) {

	var localVarHttpMethod = strings.ToUpper("Post")
	// create path and map variables
	localVarPath := a.Configuration.BasePath + "/v4/clusters/"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := make(map[string]string)
	var localVarPostBody interface{}
	var localVarFileName string
	var localVarFileBytes []byte
	// add default headers if any
	for key := range a.Configuration.DefaultHeader {
		localVarHeaderParams[key] = a.Configuration.DefaultHeader[key]
	}

	// to determine the Content-Type header
	localVarHttpContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHttpContentType := a.Configuration.APIClient.SelectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHttpContentType
	}
	// to determine the Accept header
	localVarHttpHeaderAccepts := []string{
		"application/json",
	}

	// set Accept header
	localVarHttpHeaderAccept := a.Configuration.APIClient.SelectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHttpHeaderAccept
	}
	// header params "Authorization"
	localVarHeaderParams["Authorization"] = a.Configuration.APIClient.ParameterToString(authorization, "")
	// header params "X-Request-ID"
	localVarHeaderParams["X-Request-ID"] = a.Configuration.APIClient.ParameterToString(xRequestID, "")
	// header params "X-Giant-Swarm-Activity"
	localVarHeaderParams["X-Giant-Swarm-Activity"] = a.Configuration.APIClient.ParameterToString(xGiantSwarmActivity, "")
	// header params "X-Giant-Swarm-CmdLine"
	localVarHeaderParams["X-Giant-Swarm-CmdLine"] = a.Configuration.APIClient.ParameterToString(xGiantSwarmCmdLine, "")
	// body params
	localVarPostBody = &body
	var successPayload = new(V4GenericResponse)
	localVarHttpResponse, err := a.Configuration.APIClient.CallAPI(localVarPath, localVarHttpMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)

	var localVarURL, _ = url.Parse(localVarPath)
	localVarURL.RawQuery = localVarQueryParams.Encode()
	var localVarAPIResponse = &APIResponse{Operation: "AddCluster", Method: localVarHttpMethod, RequestURL: localVarURL.String()}
	if localVarHttpResponse != nil {
		localVarAPIResponse.Response = localVarHttpResponse.RawResponse
		localVarAPIResponse.Payload = localVarHttpResponse.Body()
	}

	if err != nil {
		return successPayload, localVarAPIResponse, err
	}
	err = json.Unmarshal(localVarHttpResponse.Body(), &successPayload)
	return successPayload, localVarAPIResponse, err
}

/**
 * Add key-pair for cluster
 *
 * @param authorization Header to pass an authorization token. The value has to be in the form &#x60;giantswarm &lt;token&gt;&#x60;.
 * @param clusterId ID of the cluster to create the key-pair for
 * @param body Description and expiry time for the new key-pair
 * @param xRequestID A randomly generated key that can be used to track a request throughout services of Giant Swarm
 * @param xGiantSwarmActivity Name of an activity to track, like \&quot;list-clusters\&quot;
 * @param xGiantSwarmCmdLine If activity has been issued by a CLI, this header can contain the command line
 * @return *V4AddKeyPairResponse
 */
func (a DefaultApi) AddKeyPair(authorization string, clusterId string, body V4AddKeyPairBody, xRequestID string, xGiantSwarmActivity string, xGiantSwarmCmdLine string) (*V4AddKeyPairResponse, *APIResponse, error) {

	var localVarHttpMethod = strings.ToUpper("Post")
	// create path and map variables
	localVarPath := a.Configuration.BasePath + "/v4/clusters/{cluster_id}/key-pairs/"
	localVarPath = strings.Replace(localVarPath, "{"+"cluster_id"+"}", fmt.Sprintf("%v", clusterId), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := make(map[string]string)
	var localVarPostBody interface{}
	var localVarFileName string
	var localVarFileBytes []byte
	// add default headers if any
	for key := range a.Configuration.DefaultHeader {
		localVarHeaderParams[key] = a.Configuration.DefaultHeader[key]
	}

	// to determine the Content-Type header
	localVarHttpContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHttpContentType := a.Configuration.APIClient.SelectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHttpContentType
	}
	// to determine the Accept header
	localVarHttpHeaderAccepts := []string{
		"application/json",
	}

	// set Accept header
	localVarHttpHeaderAccept := a.Configuration.APIClient.SelectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHttpHeaderAccept
	}
	// header params "Authorization"
	localVarHeaderParams["Authorization"] = a.Configuration.APIClient.ParameterToString(authorization, "")
	// header params "X-Request-ID"
	localVarHeaderParams["X-Request-ID"] = a.Configuration.APIClient.ParameterToString(xRequestID, "")
	// header params "X-Giant-Swarm-Activity"
	localVarHeaderParams["X-Giant-Swarm-Activity"] = a.Configuration.APIClient.ParameterToString(xGiantSwarmActivity, "")
	// header params "X-Giant-Swarm-CmdLine"
	localVarHeaderParams["X-Giant-Swarm-CmdLine"] = a.Configuration.APIClient.ParameterToString(xGiantSwarmCmdLine, "")
	// body params
	localVarPostBody = &body
	var successPayload = new(V4AddKeyPairResponse)
	localVarHttpResponse, err := a.Configuration.APIClient.CallAPI(localVarPath, localVarHttpMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)

	var localVarURL, _ = url.Parse(localVarPath)
	localVarURL.RawQuery = localVarQueryParams.Encode()
	var localVarAPIResponse = &APIResponse{Operation: "AddKeyPair", Method: localVarHttpMethod, RequestURL: localVarURL.String()}
	if localVarHttpResponse != nil {
		localVarAPIResponse.Response = localVarHttpResponse.RawResponse
		localVarAPIResponse.Payload = localVarHttpResponse.Body()
	}

	if err != nil {
		return successPayload, localVarAPIResponse, err
	}
	err = json.Unmarshal(localVarHttpResponse.Body(), &successPayload)
	return successPayload, localVarAPIResponse, err
}

/**
 * Delete cluster
 * This operation allows to delete a cluster.  __Caution:__ Deleting a cluster causes the termination of all workloads running on the cluster. Data stored on the worker nodes will be lost. There is no way to undo this operation.  The response is sent as soon as the request is validated. At that point, workloads might still be running on the cluster and may be accessible for a little wile, until the cluster is actually deleted.
 *
 * @param authorization Header to pass an authorization token. The value has to be in the form &#x60;giantswarm &lt;token&gt;&#x60;.
 * @param clusterId Cluster ID
 * @param xRequestID A randomly generated key that can be used to track a request throughout services of Giant Swarm
 * @param xGiantSwarmActivity Name of an activity to track, like \&quot;list-clusters\&quot;
 * @param xGiantSwarmCmdLine If activity has been issued by a CLI, this header can contain the command line
 * @return *V4GenericResponse
 */
func (a DefaultApi) DeleteCluster(authorization string, clusterId string, xRequestID string, xGiantSwarmActivity string, xGiantSwarmCmdLine string) (*V4GenericResponse, *APIResponse, error) {

	var localVarHttpMethod = strings.ToUpper("Delete")
	// create path and map variables
	localVarPath := a.Configuration.BasePath + "/v4/clusters/{cluster_id}/"
	localVarPath = strings.Replace(localVarPath, "{"+"cluster_id"+"}", fmt.Sprintf("%v", clusterId), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := make(map[string]string)
	var localVarPostBody interface{}
	var localVarFileName string
	var localVarFileBytes []byte
	// add default headers if any
	for key := range a.Configuration.DefaultHeader {
		localVarHeaderParams[key] = a.Configuration.DefaultHeader[key]
	}

	// to determine the Content-Type header
	localVarHttpContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHttpContentType := a.Configuration.APIClient.SelectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHttpContentType
	}
	// to determine the Accept header
	localVarHttpHeaderAccepts := []string{
		"application/json",
	}

	// set Accept header
	localVarHttpHeaderAccept := a.Configuration.APIClient.SelectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHttpHeaderAccept
	}
	// header params "Authorization"
	localVarHeaderParams["Authorization"] = a.Configuration.APIClient.ParameterToString(authorization, "")
	// header params "X-Request-ID"
	localVarHeaderParams["X-Request-ID"] = a.Configuration.APIClient.ParameterToString(xRequestID, "")
	// header params "X-Giant-Swarm-Activity"
	localVarHeaderParams["X-Giant-Swarm-Activity"] = a.Configuration.APIClient.ParameterToString(xGiantSwarmActivity, "")
	// header params "X-Giant-Swarm-CmdLine"
	localVarHeaderParams["X-Giant-Swarm-CmdLine"] = a.Configuration.APIClient.ParameterToString(xGiantSwarmCmdLine, "")
	var successPayload = new(V4GenericResponse)
	localVarHttpResponse, err := a.Configuration.APIClient.CallAPI(localVarPath, localVarHttpMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)

	var localVarURL, _ = url.Parse(localVarPath)
	localVarURL.RawQuery = localVarQueryParams.Encode()
	var localVarAPIResponse = &APIResponse{Operation: "DeleteCluster", Method: localVarHttpMethod, RequestURL: localVarURL.String()}
	if localVarHttpResponse != nil {
		localVarAPIResponse.Response = localVarHttpResponse.RawResponse
		localVarAPIResponse.Payload = localVarHttpResponse.Body()
	}

	if err != nil {
		return successPayload, localVarAPIResponse, err
	}
	err = json.Unmarshal(localVarHttpResponse.Body(), &successPayload)
	return successPayload, localVarAPIResponse, err
}

/**
 * Get cluster details
 *
 * @param authorization Header to pass an authorization token. The value has to be in the form &#x60;giantswarm &lt;token&gt;&#x60;.
 * @param clusterId Cluster ID
 * @param xRequestID A randomly generated key that can be used to track a request throughout services of Giant Swarm
 * @param xGiantSwarmActivity Name of an activity to track, like \&quot;list-clusters\&quot;
 * @param xGiantSwarmCmdLine If activity has been issued by a CLI, this header can contain the command line
 * @return *V4ClusterDetailsModel
 */
func (a DefaultApi) GetCluster(authorization string, clusterId string, xRequestID string, xGiantSwarmActivity string, xGiantSwarmCmdLine string) (*V4ClusterDetailsModel, *APIResponse, error) {

	var localVarHttpMethod = strings.ToUpper("Get")
	// create path and map variables
	localVarPath := a.Configuration.BasePath + "/v4/clusters/{cluster_id}/"
	localVarPath = strings.Replace(localVarPath, "{"+"cluster_id"+"}", fmt.Sprintf("%v", clusterId), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := make(map[string]string)
	var localVarPostBody interface{}
	var localVarFileName string
	var localVarFileBytes []byte
	// add default headers if any
	for key := range a.Configuration.DefaultHeader {
		localVarHeaderParams[key] = a.Configuration.DefaultHeader[key]
	}

	// to determine the Content-Type header
	localVarHttpContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHttpContentType := a.Configuration.APIClient.SelectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHttpContentType
	}
	// to determine the Accept header
	localVarHttpHeaderAccepts := []string{
		"application/json",
	}

	// set Accept header
	localVarHttpHeaderAccept := a.Configuration.APIClient.SelectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHttpHeaderAccept
	}
	// header params "Authorization"
	localVarHeaderParams["Authorization"] = a.Configuration.APIClient.ParameterToString(authorization, "")
	// header params "X-Request-ID"
	localVarHeaderParams["X-Request-ID"] = a.Configuration.APIClient.ParameterToString(xRequestID, "")
	// header params "X-Giant-Swarm-Activity"
	localVarHeaderParams["X-Giant-Swarm-Activity"] = a.Configuration.APIClient.ParameterToString(xGiantSwarmActivity, "")
	// header params "X-Giant-Swarm-CmdLine"
	localVarHeaderParams["X-Giant-Swarm-CmdLine"] = a.Configuration.APIClient.ParameterToString(xGiantSwarmCmdLine, "")
	var successPayload = new(V4ClusterDetailsModel)
	localVarHttpResponse, err := a.Configuration.APIClient.CallAPI(localVarPath, localVarHttpMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)

	var localVarURL, _ = url.Parse(localVarPath)
	localVarURL.RawQuery = localVarQueryParams.Encode()
	var localVarAPIResponse = &APIResponse{Operation: "GetCluster", Method: localVarHttpMethod, RequestURL: localVarURL.String()}
	if localVarHttpResponse != nil {
		localVarAPIResponse.Response = localVarHttpResponse.RawResponse
		localVarAPIResponse.Payload = localVarHttpResponse.Body()
	}

	if err != nil {
		return successPayload, localVarAPIResponse, err
	}
	err = json.Unmarshal(localVarHttpResponse.Body(), &successPayload)
	return successPayload, localVarAPIResponse, err
}

/**
 * Get clusters
 * This operation fetches a list of clusters. The result depends on the permissions of the user. A normal user will get all the clusters the user has access to, via organization membership.  A user with admin permission will receive a list of all existing clusters.  The result array items are sparse representations of the cluster objects.
 *
 * @param authorization Header to pass an authorization token. The value has to be in the form &#x60;giantswarm &lt;token&gt;&#x60;.
 * @param xRequestID A randomly generated key that can be used to track a request throughout services of Giant Swarm
 * @param xGiantSwarmActivity Name of an activity to track, like \&quot;list-clusters\&quot;
 * @param xGiantSwarmCmdLine If activity has been issued by a CLI, this header can contain the command line
 * @return []V4ClusterListItem
 */
func (a DefaultApi) GetClusters(authorization string, xRequestID string, xGiantSwarmActivity string, xGiantSwarmCmdLine string) ([]V4ClusterListItem, *APIResponse, error) {

	var localVarHttpMethod = strings.ToUpper("Get")
	// create path and map variables
	localVarPath := a.Configuration.BasePath + "/v4/clusters/"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := make(map[string]string)
	var localVarPostBody interface{}
	var localVarFileName string
	var localVarFileBytes []byte
	// add default headers if any
	for key := range a.Configuration.DefaultHeader {
		localVarHeaderParams[key] = a.Configuration.DefaultHeader[key]
	}

	// to determine the Content-Type header
	localVarHttpContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHttpContentType := a.Configuration.APIClient.SelectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHttpContentType
	}
	// to determine the Accept header
	localVarHttpHeaderAccepts := []string{
		"application/json",
	}

	// set Accept header
	localVarHttpHeaderAccept := a.Configuration.APIClient.SelectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHttpHeaderAccept
	}
	// header params "Authorization"
	localVarHeaderParams["Authorization"] = a.Configuration.APIClient.ParameterToString(authorization, "")
	// header params "X-Request-ID"
	localVarHeaderParams["X-Request-ID"] = a.Configuration.APIClient.ParameterToString(xRequestID, "")
	// header params "X-Giant-Swarm-Activity"
	localVarHeaderParams["X-Giant-Swarm-Activity"] = a.Configuration.APIClient.ParameterToString(xGiantSwarmActivity, "")
	// header params "X-Giant-Swarm-CmdLine"
	localVarHeaderParams["X-Giant-Swarm-CmdLine"] = a.Configuration.APIClient.ParameterToString(xGiantSwarmCmdLine, "")
	var successPayload = new([]V4ClusterListItem)
	localVarHttpResponse, err := a.Configuration.APIClient.CallAPI(localVarPath, localVarHttpMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)

	var localVarURL, _ = url.Parse(localVarPath)
	localVarURL.RawQuery = localVarQueryParams.Encode()
	var localVarAPIResponse = &APIResponse{Operation: "GetClusters", Method: localVarHttpMethod, RequestURL: localVarURL.String()}
	if localVarHttpResponse != nil {
		localVarAPIResponse.Response = localVarHttpResponse.RawResponse
		localVarAPIResponse.Payload = localVarHttpResponse.Body()
	}

	if err != nil {
		return *successPayload, localVarAPIResponse, err
	}
	err = json.Unmarshal(localVarHttpResponse.Body(), &successPayload)
	return *successPayload, localVarAPIResponse, err
}

/**
 * Get key-pairs for cluster
 *
 * @param authorization Header to pass an authorization token. The value has to be in the form &#x60;giantswarm &lt;token&gt;&#x60;.
 * @param clusterId
 * @param xRequestID A randomly generated key that can be used to track a request throughout services of Giant Swarm
 * @param xGiantSwarmActivity Name of an activity to track, like \&quot;list-clusters\&quot;
 * @param xGiantSwarmCmdLine If activity has been issued by a CLI, this header can contain the command line
 * @return []KeyPairModel
 */
func (a DefaultApi) GetKeyPairs(authorization string, clusterId string, xRequestID string, xGiantSwarmActivity string, xGiantSwarmCmdLine string) ([]KeyPairModel, *APIResponse, error) {

	var localVarHttpMethod = strings.ToUpper("Get")
	// create path and map variables
	localVarPath := a.Configuration.BasePath + "/v4/clusters/{cluster_id}/key-pairs/"
	localVarPath = strings.Replace(localVarPath, "{"+"cluster_id"+"}", fmt.Sprintf("%v", clusterId), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := make(map[string]string)
	var localVarPostBody interface{}
	var localVarFileName string
	var localVarFileBytes []byte
	// add default headers if any
	for key := range a.Configuration.DefaultHeader {
		localVarHeaderParams[key] = a.Configuration.DefaultHeader[key]
	}

	// to determine the Content-Type header
	localVarHttpContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHttpContentType := a.Configuration.APIClient.SelectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHttpContentType
	}
	// to determine the Accept header
	localVarHttpHeaderAccepts := []string{
		"application/json",
	}

	// set Accept header
	localVarHttpHeaderAccept := a.Configuration.APIClient.SelectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHttpHeaderAccept
	}
	// header params "Authorization"
	localVarHeaderParams["Authorization"] = a.Configuration.APIClient.ParameterToString(authorization, "")
	// header params "X-Request-ID"
	localVarHeaderParams["X-Request-ID"] = a.Configuration.APIClient.ParameterToString(xRequestID, "")
	// header params "X-Giant-Swarm-Activity"
	localVarHeaderParams["X-Giant-Swarm-Activity"] = a.Configuration.APIClient.ParameterToString(xGiantSwarmActivity, "")
	// header params "X-Giant-Swarm-CmdLine"
	localVarHeaderParams["X-Giant-Swarm-CmdLine"] = a.Configuration.APIClient.ParameterToString(xGiantSwarmCmdLine, "")
	var successPayload = new([]KeyPairModel)
	localVarHttpResponse, err := a.Configuration.APIClient.CallAPI(localVarPath, localVarHttpMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)

	var localVarURL, _ = url.Parse(localVarPath)
	localVarURL.RawQuery = localVarQueryParams.Encode()
	var localVarAPIResponse = &APIResponse{Operation: "GetKeyPairs", Method: localVarHttpMethod, RequestURL: localVarURL.String()}
	if localVarHttpResponse != nil {
		localVarAPIResponse.Response = localVarHttpResponse.RawResponse
		localVarAPIResponse.Payload = localVarHttpResponse.Body()
	}

	if err != nil {
		return *successPayload, localVarAPIResponse, err
	}
	err = json.Unmarshal(localVarHttpResponse.Body(), &successPayload)
	return *successPayload, localVarAPIResponse, err
}

/**
 * Get organizations for user
 * This operation allows to fetch a list of organizations the user is a member of. In the case of an admin user, the result includes all existing organizations.
 *
 * @param authorization Header to pass an authorization token. The value has to be in the form &#x60;giantswarm &lt;token&gt;&#x60;.
 * @param xRequestID A randomly generated key that can be used to track a request throughout services of Giant Swarm
 * @param xGiantSwarmActivity Name of an activity to track, like \&quot;list-clusters\&quot;
 * @param xGiantSwarmCmdLine If activity has been issued by a CLI, this header can contain the command line
 * @return []V4OrganizationListItem
 */
func (a DefaultApi) GetUserOrganizations(authorization string, xRequestID string, xGiantSwarmActivity string, xGiantSwarmCmdLine string) ([]V4OrganizationListItem, *APIResponse, error) {

	var localVarHttpMethod = strings.ToUpper("Get")
	// create path and map variables
	localVarPath := a.Configuration.BasePath + "/v4/organizations/"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := make(map[string]string)
	var localVarPostBody interface{}
	var localVarFileName string
	var localVarFileBytes []byte
	// add default headers if any
	for key := range a.Configuration.DefaultHeader {
		localVarHeaderParams[key] = a.Configuration.DefaultHeader[key]
	}

	// to determine the Content-Type header
	localVarHttpContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHttpContentType := a.Configuration.APIClient.SelectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHttpContentType
	}
	// to determine the Accept header
	localVarHttpHeaderAccepts := []string{
		"application/json",
	}

	// set Accept header
	localVarHttpHeaderAccept := a.Configuration.APIClient.SelectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHttpHeaderAccept
	}
	// header params "Authorization"
	localVarHeaderParams["Authorization"] = a.Configuration.APIClient.ParameterToString(authorization, "")
	// header params "X-Request-ID"
	localVarHeaderParams["X-Request-ID"] = a.Configuration.APIClient.ParameterToString(xRequestID, "")
	// header params "X-Giant-Swarm-Activity"
	localVarHeaderParams["X-Giant-Swarm-Activity"] = a.Configuration.APIClient.ParameterToString(xGiantSwarmActivity, "")
	// header params "X-Giant-Swarm-CmdLine"
	localVarHeaderParams["X-Giant-Swarm-CmdLine"] = a.Configuration.APIClient.ParameterToString(xGiantSwarmCmdLine, "")
	var successPayload = new([]V4OrganizationListItem)
	localVarHttpResponse, err := a.Configuration.APIClient.CallAPI(localVarPath, localVarHttpMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)

	var localVarURL, _ = url.Parse(localVarPath)
	localVarURL.RawQuery = localVarQueryParams.Encode()
	var localVarAPIResponse = &APIResponse{Operation: "GetUserOrganizations", Method: localVarHttpMethod, RequestURL: localVarURL.String()}
	if localVarHttpResponse != nil {
		localVarAPIResponse.Response = localVarHttpResponse.RawResponse
		localVarAPIResponse.Payload = localVarHttpResponse.Body()
	}

	if err != nil {
		return *successPayload, localVarAPIResponse, err
	}
	err = json.Unmarshal(localVarHttpResponse.Body(), &successPayload)
	return *successPayload, localVarAPIResponse, err
}

/**
 * Log in as a user
 * This method takes email and password of a user and returns a new session token. The token can be found in the &#x60;data.Id&#x60; field of the response object.
 *
 * @param email User email address
 * @param payload base64 encoded password
 * @param xRequestID A randomly generated key that can be used to track a request throughout services of Giant Swarm
 * @param xGiantSwarmActivity Name of an activity to track, like \&quot;list-clusters\&quot;
 * @param xGiantSwarmCmdLine If activity has been issued by a CLI, this header can contain the command line
 * @return *LoginResponseModel
 */
func (a DefaultApi) UserLogin(email string, payload LoginBodyModel, xRequestID string, xGiantSwarmActivity string, xGiantSwarmCmdLine string) (*LoginResponseModel, *APIResponse, error) {

	var localVarHttpMethod = strings.ToUpper("Post")
	// create path and map variables
	localVarPath := a.Configuration.BasePath + "/v1/user/{email}/login"
	localVarPath = strings.Replace(localVarPath, "{"+"email"+"}", fmt.Sprintf("%v", email), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := make(map[string]string)
	var localVarPostBody interface{}
	var localVarFileName string
	var localVarFileBytes []byte
	// add default headers if any
	for key := range a.Configuration.DefaultHeader {
		localVarHeaderParams[key] = a.Configuration.DefaultHeader[key]
	}

	// to determine the Content-Type header
	localVarHttpContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHttpContentType := a.Configuration.APIClient.SelectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHttpContentType
	}
	// to determine the Accept header
	localVarHttpHeaderAccepts := []string{
		"application/json",
	}

	// set Accept header
	localVarHttpHeaderAccept := a.Configuration.APIClient.SelectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHttpHeaderAccept
	}
	// header params "X-Request-ID"
	localVarHeaderParams["X-Request-ID"] = a.Configuration.APIClient.ParameterToString(xRequestID, "")
	// header params "X-Giant-Swarm-Activity"
	localVarHeaderParams["X-Giant-Swarm-Activity"] = a.Configuration.APIClient.ParameterToString(xGiantSwarmActivity, "")
	// header params "X-Giant-Swarm-CmdLine"
	localVarHeaderParams["X-Giant-Swarm-CmdLine"] = a.Configuration.APIClient.ParameterToString(xGiantSwarmCmdLine, "")
	// body params
	localVarPostBody = &payload
	var successPayload = new(LoginResponseModel)
	localVarHttpResponse, err := a.Configuration.APIClient.CallAPI(localVarPath, localVarHttpMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)

	var localVarURL, _ = url.Parse(localVarPath)
	localVarURL.RawQuery = localVarQueryParams.Encode()
	var localVarAPIResponse = &APIResponse{Operation: "UserLogin", Method: localVarHttpMethod, RequestURL: localVarURL.String()}
	if localVarHttpResponse != nil {
		localVarAPIResponse.Response = localVarHttpResponse.RawResponse
		localVarAPIResponse.Payload = localVarHttpResponse.Body()
	}

	if err != nil {
		return successPayload, localVarAPIResponse, err
	}
	err = json.Unmarshal(localVarHttpResponse.Body(), &successPayload)
	return successPayload, localVarAPIResponse, err
}

/**
 * Expire the currently used auth token
 *
 * @param authorization Header to pass an authorization token. The value has to be in the form &#x60;giantswarm &lt;token&gt;&#x60;.
 * @param xRequestID A randomly generated key that can be used to track a request throughout services of Giant Swarm
 * @param xGiantSwarmActivity Name of an activity to track, like \&quot;list-clusters\&quot;
 * @param xGiantSwarmCmdLine If activity has been issued by a CLI, this header can contain the command line
 * @return *GenericResponseModel
 */
func (a DefaultApi) UserLogout(authorization string, xRequestID string, xGiantSwarmActivity string, xGiantSwarmCmdLine string) (*GenericResponseModel, *APIResponse, error) {

	var localVarHttpMethod = strings.ToUpper("Post")
	// create path and map variables
	localVarPath := a.Configuration.BasePath + "/v1/token/logout"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := make(map[string]string)
	var localVarPostBody interface{}
	var localVarFileName string
	var localVarFileBytes []byte
	// add default headers if any
	for key := range a.Configuration.DefaultHeader {
		localVarHeaderParams[key] = a.Configuration.DefaultHeader[key]
	}

	// to determine the Content-Type header
	localVarHttpContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHttpContentType := a.Configuration.APIClient.SelectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHttpContentType
	}
	// to determine the Accept header
	localVarHttpHeaderAccepts := []string{
		"application/json",
		"text/plain",
	}

	// set Accept header
	localVarHttpHeaderAccept := a.Configuration.APIClient.SelectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHttpHeaderAccept
	}
	// header params "Authorization"
	localVarHeaderParams["Authorization"] = a.Configuration.APIClient.ParameterToString(authorization, "")
	// header params "X-Request-ID"
	localVarHeaderParams["X-Request-ID"] = a.Configuration.APIClient.ParameterToString(xRequestID, "")
	// header params "X-Giant-Swarm-Activity"
	localVarHeaderParams["X-Giant-Swarm-Activity"] = a.Configuration.APIClient.ParameterToString(xGiantSwarmActivity, "")
	// header params "X-Giant-Swarm-CmdLine"
	localVarHeaderParams["X-Giant-Swarm-CmdLine"] = a.Configuration.APIClient.ParameterToString(xGiantSwarmCmdLine, "")
	var successPayload = new(GenericResponseModel)
	localVarHttpResponse, err := a.Configuration.APIClient.CallAPI(localVarPath, localVarHttpMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)

	var localVarURL, _ = url.Parse(localVarPath)
	localVarURL.RawQuery = localVarQueryParams.Encode()
	var localVarAPIResponse = &APIResponse{Operation: "UserLogout", Method: localVarHttpMethod, RequestURL: localVarURL.String()}
	if localVarHttpResponse != nil {
		localVarAPIResponse.Response = localVarHttpResponse.RawResponse
		localVarAPIResponse.Payload = localVarHttpResponse.Body()
	}

	if err != nil {
		return successPayload, localVarAPIResponse, err
	}
	err = json.Unmarshal(localVarHttpResponse.Body(), &successPayload)
	return successPayload, localVarAPIResponse, err
}
