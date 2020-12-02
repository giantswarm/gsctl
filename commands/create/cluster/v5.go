package cluster

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/giantswarm/gsclientgen/v2/models"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/commands/types"
	"github.com/giantswarm/gsctl/formatting"
)

type definitionFromFlagsV5 struct {
	clusterName    string
	releaseVersion string
	owner          string
	isHAMaster     *bool
}

// updateDefinitionFromFlagsV5 extend/overwrites a clusterDefinition based on the
// flags/arguments the user has given.
func updateDefinitionFromFlagsV5(def *types.ClusterDefinitionV5, flags definitionFromFlagsV5) {
	if def == nil {
		return
	}

	if flags.clusterName != "" {
		def.Name = flags.clusterName
	}

	if flags.releaseVersion != "" {
		def.ReleaseVersion = flags.releaseVersion
	}

	if flags.owner != "" {
		def.Owner = flags.owner
	}

	if flags.isHAMaster != nil {
		def.MasterNodes = &types.MasterNodes{
			HighAvailability: *flags.isHAMaster,
		}
	}
}

func createAddClusterBodyV5(def *types.ClusterDefinitionV5) *models.V5AddClusterRequest {
	b := &models.V5AddClusterRequest{
		Owner:          &def.Owner,
		Name:           def.Name,
		ReleaseVersion: def.ReleaseVersion,
	}

	if def.MasterNodes != nil {
		b.MasterNodes = &models.V5AddClusterRequestMasterNodes{
			HighAvailability:  &def.MasterNodes.HighAvailability,
			AvailabilityZones: def.MasterNodes.AvailabilityZones,
			Azure: &models.V5AddClusterRequestMasterNodesAzure{
				AvailabilityZonesUnspecified: def.MasterNodes.Azure.AvailabilityZonesUnspecified,
			},
		}
	} else {
		b.MasterNodes = &models.V5AddClusterRequestMasterNodes{
			// We default to unspecified AZs on azure to match happa's behaviour.
			Azure: &models.V5AddClusterRequestMasterNodesAzure{
				AvailabilityZonesUnspecified: true,
			},
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
			b.Scaling.Min = &def.Scaling.Min
		}
		if def.Scaling.Max != 0 {
			b.Scaling.Max = def.Scaling.Max
		}
	}

	if def.NodeSpec != nil {
		if def.NodeSpec.AWS != nil {
			b.NodeSpec.Aws = &models.V5AddNodePoolRequestNodeSpecAws{}

			if def.NodeSpec.AWS.InstanceDistribution != nil {
				b.NodeSpec.Aws.InstanceDistribution = &models.V5AddNodePoolRequestNodeSpecAwsInstanceDistribution{
					OnDemandBaseCapacity:                &def.NodeSpec.AWS.InstanceDistribution.OnDemandBaseCapacity,
					OnDemandPercentageAboveBaseCapacity: &def.NodeSpec.AWS.InstanceDistribution.OnDemandPercentageAboveBaseCapacity,
				}
			}

			if def.NodeSpec.AWS.InstanceType != "" {
				b.NodeSpec.Aws.InstanceType = def.NodeSpec.AWS.InstanceType
			}

			b.NodeSpec.Aws.UseAlikeInstanceTypes = &def.NodeSpec.AWS.UseAlikeInstanceTypes
		}

		if def.NodeSpec.Azure != nil {
			b.NodeSpec.Azure = &models.V5AddNodePoolRequestNodeSpecAzure{}
			if def.NodeSpec.Azure.VMSize != "" {
				b.NodeSpec.Azure.VMSize = def.NodeSpec.Azure.VMSize
			}
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

	if args.OutputFormat != formatting.OutputFormatJSON {
		fmt.Printf("Requesting new cluster for organization '%s'\n", color.CyanString(def.Owner))
	}

	response, err := clientWrapper.CreateClusterV5(clusterRequestBody, auxParams)
	if err != nil {
		return "", true, microerror.Mask(err)
	}

	hasErrors := false

	// Create node pools.
	if def.NodePools != nil && len(def.NodePools) > 0 {
		for i, np := range def.NodePools {
			nodePoolRequestBody := createAddNodePoolBody(np)

			if args.OutputFormat != formatting.OutputFormatJSON {
				fmt.Printf("Adding node pool %d\n", i+1)
			}

			npResponse, err := clientWrapper.CreateNodePool(response.Payload.ID, nodePoolRequestBody, auxParams)
			if err != nil {
				fmt.Println(color.RedString("Error creating node pool %d: %s", i+1, err.Error()))
				hasErrors = true
			} else if args.Verbose {
				fmt.Println(color.WhiteString("Added node pool %d with ID %s named '%s'", i+1, npResponse.Payload.ID, npResponse.Payload.Name))
			}
		}
	} else if args.CreateDefaultNodePool {
		if args.OutputFormat != formatting.OutputFormatJSON {
			fmt.Println("Adding a default node pool")
		}

		nodePoolRequestBody := &models.V5AddNodePoolRequest{}
		npResponse, err := clientWrapper.CreateNodePool(response.Payload.ID, nodePoolRequestBody, auxParams)
		if err != nil {
			fmt.Println(color.RedString("Error creating default node pool: %s", err.Error()))
			hasErrors = true
		} else if args.Verbose {
			fmt.Println(color.WhiteString("Added default node pool with ID %s", npResponse.Payload.ID))
		}
	}

	// Create labels
	if def.Labels != nil && len(def.Labels) > 0 {
		labelsRequest := models.V5SetClusterLabelsRequest{Labels: def.Labels}
		_, err := clientWrapper.UpdateClusterLabels(response.Payload.ID, &labelsRequest, auxParams)
		if err != nil {
			fmt.Println(color.RedString("Error attaching labels %s", err.Error()))
			hasErrors = true
		} else if args.Verbose {
			fmt.Println(color.WhiteString("Attached labels to cluster with ID %s named '%s'", response.Payload.ID, response.Payload.Name))
		}
	}

	return response.Payload.ID, hasErrors, nil

}
