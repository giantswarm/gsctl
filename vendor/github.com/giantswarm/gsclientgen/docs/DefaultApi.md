# \DefaultApi

All URIs are relative to *https://api.giantswarm.io*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AddCluster**](DefaultApi.md#AddCluster) | **Post** /v4/clusters/ | Create cluster
[**AddKeyPair**](DefaultApi.md#AddKeyPair) | **Post** /v4/clusters/{cluster_id}/key-pairs/ | Add key-pair for cluster
[**DeleteCluster**](DefaultApi.md#DeleteCluster) | **Delete** /v4/clusters/{cluster_id}/ | Delete cluster
[**GetCluster**](DefaultApi.md#GetCluster) | **Get** /v4/clusters/{cluster_id}/ | Get cluster details
[**GetClusters**](DefaultApi.md#GetClusters) | **Get** /v4/clusters/ | Get clusters
[**GetKeyPairs**](DefaultApi.md#GetKeyPairs) | **Get** /v4/clusters/{cluster_id}/key-pairs/ | Get key-pairs for cluster
[**GetReleases**](DefaultApi.md#GetReleases) | **Get** /v4/releases/ | Get releases
[**GetUserOrganizations**](DefaultApi.md#GetUserOrganizations) | **Get** /v4/organizations/ | Get organizations for user
[**ModifyCluster**](DefaultApi.md#ModifyCluster) | **Patch** /v4/clusters/{cluster_id}/ | Modify cluster
[**UserLogin**](DefaultApi.md#UserLogin) | **Post** /v1/user/{email}/login | Log in as a user
[**UserLogout**](DefaultApi.md#UserLogout) | **Post** /v1/token/logout | Expire the currently used auth token


# **AddCluster**
> V4GenericResponse AddCluster($authorization, $body, $xRequestID, $xGiantSwarmActivity, $xGiantSwarmCmdLine)

Create cluster

This operation is used to create a new Kubernetes cluster for an organization. The desired configuration can be specified using the __cluster definition format__ (see [external documentation](https://github.com/giantswarm/api-spec/blob/ master/details/CLUSTER_DEFINITION.md) for details). The cluster definition format allows to set a number of optional configuration details, like memory size and number of CPU cores. However, one attribute is __mandatory__ upon creation: The `owner` attribute must carry the name of the organization the cluster will belong to. Note that the acting user must be a member of that organization in order to create a cluster. It is *recommended* to also specify the `name` attribute to give the cluster a friendly name, like e. g. \"Development Cluster\". Additional definition attributes can be used. Where attributes are ommitted, default configuration values will be applied. For example, if no `kubernetes_version` is specified, the latest version tested and provided by Giant Swarm is used. The `workers` attribute, if present, must contain an array of node definition objects. The number of objects given determines the number of workers created. For example, requesting three worker nodes with default configuration can be achieved by submitting an array of three empty objects: ```\"workers\": [{}, {}, {}]```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **authorization** | **string**| Header to pass an authorization token. The value has to be in the form &#x60;giantswarm &lt;token&gt;&#x60;.  | 
 **body** | [**V4AddClusterRequest**](V4AddClusterRequest.md)| New cluster definition | 
 **xRequestID** | **string**| A randomly generated key that can be used to track a request throughout services of Giant Swarm  | [optional] 
 **xGiantSwarmActivity** | **string**| Name of an activity to track, like \&quot;list-clusters\&quot; | [optional] 
 **xGiantSwarmCmdLine** | **string**| If activity has been issued by a CLI, this header can contain the command line  | [optional] 

### Return type

[**V4GenericResponse**](V4GenericResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **AddKeyPair**
> V4AddKeyPairResponse AddKeyPair($authorization, $clusterId, $body, $xRequestID, $xGiantSwarmActivity, $xGiantSwarmCmdLine)

Add key-pair for cluster


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **authorization** | **string**| Header to pass an authorization token. The value has to be in the form &#x60;giantswarm &lt;token&gt;&#x60;.  | 
 **clusterId** | **string**| ID of the cluster to create the key-pair for | 
 **body** | [**V4AddKeyPairBody**](V4AddKeyPairBody.md)| Description and expiry time for the new key-pair | 
 **xRequestID** | **string**| A randomly generated key that can be used to track a request throughout services of Giant Swarm  | [optional] 
 **xGiantSwarmActivity** | **string**| Name of an activity to track, like \&quot;list-clusters\&quot; | [optional] 
 **xGiantSwarmCmdLine** | **string**| If activity has been issued by a CLI, this header can contain the command line  | [optional] 

### Return type

[**V4AddKeyPairResponse**](V4AddKeyPairResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DeleteCluster**
> V4GenericResponse DeleteCluster($authorization, $clusterId, $xRequestID, $xGiantSwarmActivity, $xGiantSwarmCmdLine)

Delete cluster

This operation allows to delete a cluster.  __Caution:__ Deleting a cluster causes the termination of all workloads running on the cluster. Data stored on the worker nodes will be lost. There is no way to undo this operation.  The response is sent as soon as the request is validated. At that point, workloads might still be running on the cluster and may be accessible for a little wile, until the cluster is actually deleted. 


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **authorization** | **string**| Header to pass an authorization token. The value has to be in the form &#x60;giantswarm &lt;token&gt;&#x60;.  | 
 **clusterId** | **string**| Cluster ID | 
 **xRequestID** | **string**| A randomly generated key that can be used to track a request throughout services of Giant Swarm  | [optional] 
 **xGiantSwarmActivity** | **string**| Name of an activity to track, like \&quot;list-clusters\&quot; | [optional] 
 **xGiantSwarmCmdLine** | **string**| If activity has been issued by a CLI, this header can contain the command line  | [optional] 

### Return type

[**V4GenericResponse**](V4GenericResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetCluster**
> V4ClusterDetailsModel GetCluster($authorization, $clusterId, $xRequestID, $xGiantSwarmActivity, $xGiantSwarmCmdLine)

Get cluster details


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **authorization** | **string**| Header to pass an authorization token. The value has to be in the form &#x60;giantswarm &lt;token&gt;&#x60;.  | 
 **clusterId** | **string**| Cluster ID | 
 **xRequestID** | **string**| A randomly generated key that can be used to track a request throughout services of Giant Swarm  | [optional] 
 **xGiantSwarmActivity** | **string**| Name of an activity to track, like \&quot;list-clusters\&quot; | [optional] 
 **xGiantSwarmCmdLine** | **string**| If activity has been issued by a CLI, this header can contain the command line  | [optional] 

### Return type

[**V4ClusterDetailsModel**](V4ClusterDetailsModel.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetClusters**
> []V4ClusterListItem GetClusters($authorization, $xRequestID, $xGiantSwarmActivity, $xGiantSwarmCmdLine)

Get clusters

This operation fetches a list of clusters. The result depends on the permissions of the user. A normal user will get all the clusters the user has access to, via organization membership.  A user with admin permission will receive a list of all existing clusters.  The result array items are sparse representations of the cluster objects. 


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **authorization** | **string**| Header to pass an authorization token. The value has to be in the form &#x60;giantswarm &lt;token&gt;&#x60;.  | 
 **xRequestID** | **string**| A randomly generated key that can be used to track a request throughout services of Giant Swarm  | [optional] 
 **xGiantSwarmActivity** | **string**| Name of an activity to track, like \&quot;list-clusters\&quot; | [optional] 
 **xGiantSwarmCmdLine** | **string**| If activity has been issued by a CLI, this header can contain the command line  | [optional] 

### Return type

[**[]V4ClusterListItem**](V4ClusterListItem.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetKeyPairs**
> []KeyPairModel GetKeyPairs($authorization, $clusterId, $xRequestID, $xGiantSwarmActivity, $xGiantSwarmCmdLine)

Get key-pairs for cluster


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **authorization** | **string**| Header to pass an authorization token. The value has to be in the form &#x60;giantswarm &lt;token&gt;&#x60;.  | 
 **clusterId** | **string**|  | 
 **xRequestID** | **string**| A randomly generated key that can be used to track a request throughout services of Giant Swarm  | [optional] 
 **xGiantSwarmActivity** | **string**| Name of an activity to track, like \&quot;list-clusters\&quot; | [optional] 
 **xGiantSwarmCmdLine** | **string**| If activity has been issued by a CLI, this header can contain the command line  | [optional] 

### Return type

[**[]KeyPairModel**](KeyPairModel.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetReleases**
> []V4ReleaseListItem GetReleases($authorization, $xRequestID, $xGiantSwarmActivity, $xGiantSwarmCmdLine)

Get releases


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **authorization** | **string**| Header to pass an authorization token. The value has to be in the form &#x60;giantswarm &lt;token&gt;&#x60;.  | 
 **xRequestID** | **string**| A randomly generated key that can be used to track a request throughout services of Giant Swarm  | [optional] 
 **xGiantSwarmActivity** | **string**| Name of an activity to track, like \&quot;list-clusters\&quot; | [optional] 
 **xGiantSwarmCmdLine** | **string**| If activity has been issued by a CLI, this header can contain the command line  | [optional] 

### Return type

[**[]V4ReleaseListItem**](V4ReleaseListItem.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetUserOrganizations**
> []V4OrganizationListItem GetUserOrganizations($authorization, $xRequestID, $xGiantSwarmActivity, $xGiantSwarmCmdLine)

Get organizations for user

This operation allows to fetch a list of organizations the user is a member of. In the case of an admin user, the result includes all existing organizations. 


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **authorization** | **string**| Header to pass an authorization token. The value has to be in the form &#x60;giantswarm &lt;token&gt;&#x60;.  | 
 **xRequestID** | **string**| A randomly generated key that can be used to track a request throughout services of Giant Swarm  | [optional] 
 **xGiantSwarmActivity** | **string**| Name of an activity to track, like \&quot;list-clusters\&quot; | [optional] 
 **xGiantSwarmCmdLine** | **string**| If activity has been issued by a CLI, this header can contain the command line  | [optional] 

### Return type

[**[]V4OrganizationListItem**](V4OrganizationListItem.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **ModifyCluster**
> V4ClusterDetailsModel ModifyCluster($authorization, $clusterId, $body, $xRequestID, $xGiantSwarmActivity, $xGiantSwarmCmdLine)

Modify cluster


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **authorization** | **string**| Header to pass an authorization token. The value has to be in the form &#x60;giantswarm &lt;token&gt;&#x60;.  | 
 **clusterId** | **string**| Cluster ID | 
 **body** | [**V4ModifyClusterRequest**](V4ModifyClusterRequest.md)| Modified cluster definition (JSON merge-patch) | 
 **xRequestID** | **string**| A randomly generated key that can be used to track a request throughout services of Giant Swarm  | [optional] 
 **xGiantSwarmActivity** | **string**| Name of an activity to track, like \&quot;list-clusters\&quot; | [optional] 
 **xGiantSwarmCmdLine** | **string**| If activity has been issued by a CLI, this header can contain the command line  | [optional] 

### Return type

[**V4ClusterDetailsModel**](V4ClusterDetailsModel.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UserLogin**
> LoginResponseModel UserLogin($email, $payload, $xRequestID, $xGiantSwarmActivity, $xGiantSwarmCmdLine)

Log in as a user

This method takes email and password of a user and returns a new session token. The token can be found in the `data.Id` field of the response object. 


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **email** | **string**| User email address | 
 **payload** | [**LoginBodyModel**](LoginBodyModel.md)| base64 encoded password | 
 **xRequestID** | **string**| A randomly generated key that can be used to track a request throughout services of Giant Swarm  | [optional] 
 **xGiantSwarmActivity** | **string**| Name of an activity to track, like \&quot;list-clusters\&quot; | [optional] 
 **xGiantSwarmCmdLine** | **string**| If activity has been issued by a CLI, this header can contain the command line  | [optional] 

### Return type

[**LoginResponseModel**](LoginResponseModel.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UserLogout**
> GenericResponseModel UserLogout($authorization, $xRequestID, $xGiantSwarmActivity, $xGiantSwarmCmdLine)

Expire the currently used auth token


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **authorization** | **string**| Header to pass an authorization token. The value has to be in the form &#x60;giantswarm &lt;token&gt;&#x60;.  | 
 **xRequestID** | **string**| A randomly generated key that can be used to track a request throughout services of Giant Swarm  | [optional] 
 **xGiantSwarmActivity** | **string**| Name of an activity to track, like \&quot;list-clusters\&quot; | [optional] 
 **xGiantSwarmCmdLine** | **string**| If activity has been issued by a CLI, this header can contain the command line  | [optional] 

### Return type

[**GenericResponseModel**](GenericResponseModel.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, text/plain

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

