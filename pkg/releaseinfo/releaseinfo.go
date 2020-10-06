package releaseinfo

import (
	"fmt"
	"github.com/giantswarm/gsclientgen/v2/models"
	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/microerror"
	"strings"
	"time"
)

const (
	listReleasesActivityName = "list-releases"
	infoActivityName         = "info"
)

type Config struct {
	ClientWrapper  *client.Wrapper
	ReleaseVersion string
}

type ReleaseInfo struct {
	clientWrapper  *client.Wrapper
	releaseVersion string

	k8sVersion        string
	k8sVersionEOLDate *time.Time
}

func New(c Config) (*ReleaseInfo, error) {
	if c.ClientWrapper == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClientWrapper must not be empty", c)
	}
	if len(c.ReleaseVersion) < 1 {
		return nil, microerror.Maskf(invalidConfigError, "%T.ReleaseVersion must not be empty", c)
	}

	releases, err := getReleases(c.ClientWrapper)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	infoResponse, err := getInfo(c.ClientWrapper)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	ri := &ReleaseInfo{
		clientWrapper:  c.ClientWrapper,
		releaseVersion: c.ReleaseVersion,
	}
	err = ri.parseCapabilities(releases, infoResponse)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return ri, nil
}

func (ri *ReleaseInfo) IsK8sVersionEOL() bool {
	if ri.k8sVersionEOLDate == nil {
		return false
	}

	return time.Now().After(*ri.k8sVersionEOLDate)
}

func (ri *ReleaseInfo) parseCapabilities(releases []*models.V4ReleaseListItem, infoResponse *models.V4InfoResponse) error {
	err := ri.parseReleaseCapabilities(releases)
	if err != nil {
		return microerror.Mask(err)
	}

	err = ri.parseInfoCapabilities(infoResponse)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (ri *ReleaseInfo) parseReleaseCapabilities(releases []*models.V4ReleaseListItem) error {
	var currentRelease *models.V4ReleaseListItem
	{
		for _, release := range releases {
			if release.Version != nil && *release.Version == ri.releaseVersion {
				currentRelease = release
				break
			}
		}
	}
	if currentRelease == nil {
		return microerror.Mask(versionNotFoundError)
	}

	for _, component := range currentRelease.Components {
		if component.Name != nil && *component.Name == "kubernetes" && component.Version != nil {
			ri.k8sVersion = *component.Version
			break
		}
	}

	return nil
}

func (ri *ReleaseInfo) parseInfoCapabilities(infoResponse *models.V4InfoResponse) error {
	var currentKubeVersion *models.V4InfoResponseGeneralKubernetesVersionsItems
	{
		for _, kubeVersion := range infoResponse.General.KubernetesVersions {
			versionParts := strings.SplitN(ri.k8sVersion, ".", 2)
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
		ri.k8sVersionEOLDate = &eolDate
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
