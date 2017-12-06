# V4ReleaseListItem

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Version** | **string** | The semantic version number | [default to null]
**Timestamp** | **string** | Date and time of the release creation | [default to null]
**Active** | **bool** | If true, the version is available for new clusters and cluster upgrades. Older versions become unavailable and thus have the value &#x60;false&#x60; here.  | [optional] [default to null]
**Changelog** | [**[]V4ReleaseChangelogItem**](V4ReleaseChangelogItem.md) | Structured list of changes in this release, in comparison to the previous version, with respect to the contained components.  | [default to null]
**Components** | [**[]V4ReleaseComponent**](V4ReleaseComponent.md) | List of components and their version contained in the release  | [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


