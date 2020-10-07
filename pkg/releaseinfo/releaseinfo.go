package releaseinfo

import (
	"fmt"
	"strings"
	"time"

	"github.com/giantswarm/gsclientgen/v2/models"
	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/microerror"
	"github.com/go-openapi/strfmt"
)

const (
	listReleasesActivityName = "list-releases"
	infoActivityName         = "info"

	dateFormat = strfmt.RFC3339FullDate
)

type Config struct {
	ClientWrapper *client.Wrapper
}

// ReleaseInfo is an utility data structure for collecting
// common information about a GS release version.
type ReleaseInfo struct {
	clientWrapper *client.Wrapper

	releases     []*models.V4ReleaseListItem
	infoResponse *models.V4InfoResponse
}

type ReleaseData struct {
	Version           string
	K8sVersion        string
	K8sVersionEOLDate string
	IsK8sVersionEOL   bool
}

func New(c Config) (*ReleaseInfo, error) {
	if c.ClientWrapper == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClientWrapper must not be empty", c)
	}

	ri := &ReleaseInfo{
		clientWrapper: c.ClientWrapper,
	}

	var err error
	ri.releases, err = getReleases(ri.clientWrapper)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	ri.infoResponse, err = getInfo(ri.clientWrapper)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return ri, nil
}

func (ri *ReleaseInfo) GetReleaseData(version string) (ReleaseData, error) {
	release, err := ri.getReleaseForVersion(version)
	if err != nil {
		return ReleaseData{}, microerror.Mask(err)
	}

	k8sComponent, err := ri.getReleaseComponent("kubernetes", release.Components)
	if err != nil {
		return ReleaseData{}, microerror.Mask(err)
	}

	rd := ReleaseData{
		Version:    version,
		K8sVersion: *k8sComponent.Version,
	}

	k8sEolDate := ri.getKubernetesVersionEOLDate(*k8sComponent.Version)
	if k8sEolDate == nil {
		rd.IsK8sVersionEOL = false
		rd.K8sVersionEOLDate = ""
	} else {
		rd.IsK8sVersionEOL = time.Now().After(*k8sEolDate)
		rd.K8sVersionEOLDate = k8sEolDate.Format(dateFormat)
	}

	return rd, nil
}

func (ri *ReleaseInfo) getReleaseForVersion(version string) (*models.V4ReleaseListItem, error) {
	var currentRelease *models.V4ReleaseListItem
	{
		for _, release := range ri.releases {
			if release.Version != nil && *release.Version == version {
				currentRelease = release
				break
			}
		}
	}
	if currentRelease == nil {
		return nil, microerror.Mask(versionNotFoundError)
	}

	return currentRelease, nil
}

func (ri *ReleaseInfo) getReleaseComponent(componentName string, components []*models.V4ReleaseListItemComponentsItems) (*models.V4ReleaseListItemComponentsItems, error) {
	for _, component := range components {
		if component.Name != nil && *component.Name == componentName {
			return component, nil
		}
	}

	return nil, microerror.Mask(componentNotFoundError)
}

func (ri *ReleaseInfo) getKubernetesVersionEOLDate(version string) *time.Time {
	var currentKubeVersion *models.V4InfoResponseGeneralKubernetesVersionsItems
	{
		for _, kubeVersion := range ri.infoResponse.General.KubernetesVersions {
			versionParts := strings.SplitN(version, ".", 3)
			if len(versionParts) < 2 {
				continue
			}

			minor := fmt.Sprintf("%s.%s", versionParts[0], versionParts[1])
			if kubeVersion.MinorVersion != nil && *kubeVersion.MinorVersion == minor {
				currentKubeVersion = kubeVersion
				break
			}
		}
	}

	if currentKubeVersion != nil && currentKubeVersion.EolDate != nil {
		eolDate := time.Time(*currentKubeVersion.EolDate)
		return &eolDate
	}

	return nil
}

func getReleases(wrapper *client.Wrapper) ([]*models.V4ReleaseListItem, error) {
	auxParams := wrapper.DefaultAuxiliaryParams()
	auxParams.ActivityName = listReleasesActivityName
	releasesResponse, err := wrapper.GetReleases(auxParams)
	if err != nil {
		return nil, microerror.Mask(handleClientError(err))
	}

	return releasesResponse.Payload, nil
}

func getInfo(wrapper *client.Wrapper) (*models.V4InfoResponse, error) {
	auxParams := wrapper.DefaultAuxiliaryParams()
	auxParams.ActivityName = infoActivityName
	infoResponse, err := wrapper.GetInfo(auxParams)
	if err != nil {
		return nil, microerror.Mask(handleClientError(err))
	}

	return infoResponse.Payload, nil
}

func handleClientError(err error) error {
	if clienterror.IsInternalServerError(err) {
		return microerror.Maskf(internalServerError, err.Error())
	} else if clienterror.IsUnauthorizedError(err) {
		return microerror.Mask(notAuthorizedError)
	} else if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
