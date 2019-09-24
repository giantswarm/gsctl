package cluster

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/giantswarm/gsclientgen/models"
	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/commands/types"
	"github.com/giantswarm/microerror"
)

// updateDefinitionFromFlagsV5 extend/overwrites a clusterDefinition based on the
// flags/arguments the user has given.
func updateDefinitionFromFlagsV5(def *types.ClusterDefinitionV5, clusterName, releaseVersion, owner string) {
	if def == nil {
		return
	}

	if clusterName != "" {
		def.Name = clusterName
	}

	if releaseVersion != "" {
		def.ReleaseVersion = releaseVersion
	}

	if owner != "" {
		def.Owner = owner
	}
}

func createAddClusterBodyV5(def *types.ClusterDefinitionV5) *models.V5AddClusterRequest {
	b := &models.V5AddClusterRequest{
		Owner:          &def.Owner,
		Name:           def.Name,
		ReleaseVersion: def.ReleaseVersion,
	}

	if def.Master != nil {
		b.Master = &models.V5AddClusterRequestMaster{
			AvailabilityZone: def.Master.AvailabilityZone,
		}
	}

	return b
}

func createAddNodePoolBody(def *types.NodePoolDefinition) *models.V5AddNodePoolRequest {
	b := &models.V5AddNodePoolRequest{
		Name:              def.Name,
		AvailabilityZones: &models.V5AddNodePoolRequestAvailabilityZones{},
		Scaling:           &models.V5AddNodePoolRequestScaling{},
		NodeSpec:          &models.V5AddNodePoolRequestNodeSpec{},
	}

	if def.AvailabilityZones != nil {
		if def.AvailabilityZones.Number != 0 {
			b.AvailabilityZones.Number = def.AvailabilityZones.Number
		}
		if len(def.AvailabilityZones.Zones) != 0 {
			b.AvailabilityZones.Zones = def.AvailabilityZones.Zones
		}
	}

	if def.Scaling != nil {
		if def.Scaling.Min != 0 {
			b.Scaling.Min = def.Scaling.Min
		}
		if def.Scaling.Max != 0 {
			b.Scaling.Max = def.Scaling.Max
		}
	}

	if def.NodeSpec != nil && def.NodeSpec.AWS != nil {
		b.NodeSpec.Aws = &models.V5AddNodePoolRequestNodeSpecAws{
			InstanceType: def.NodeSpec.AWS.InstanceType,
		}
	}

	return b
}

func addClusterV5(def *types.ClusterDefinitionV5, args Arguments, clientWrapper *client.Wrapper, auxParams *client.AuxiliaryParams) (string, bool, error) {
	// Validate definition
	if def.Owner == "" {
		return "", true, microerror.Mask(errors.ClusterOwnerMissingError)
	}

	clusterRequestBody := createAddClusterBodyV5(def)

	fmt.Printf("Requesting new cluster for organization '%s'\n", color.CyanString(def.Owner))

	response, err := clientWrapper.CreateClusterV5(clusterRequestBody, auxParams)
	if err != nil {
		return "", true, microerror.Mask(err)
	}

	hasErrors := false

	// Create node pools.
	if def.NodePools != nil && len(def.NodePools) > 0 {
		for i, np := range def.NodePools {
			nodePoolRequestBody := createAddNodePoolBody(np)

			if args.Verbose {
				fmt.Println(color.WhiteString("Adding node pool %d", i+1))
			}

			// TODO: fire creation request, store result
			npResponse, err := clientWrapper.CreateNodePool(response.Payload.ID, nodePoolRequestBody, auxParams)
			if err != nil {
				fmt.Println(color.RedString("Error creating node pool %d: %s", i+1, err.Error()))
				hasErrors = true
			} else if args.Verbose {
				fmt.Println(color.WhiteString("Added node pool with ID %s named '%s'", i+1, npResponse.Payload.ID, npResponse.Payload.Name))
			}
		}
	}

	return response.Payload.ID, hasErrors, nil

}
