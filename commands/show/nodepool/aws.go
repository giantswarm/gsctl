package nodepool

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/giantswarm/columnize"
	"github.com/giantswarm/gsclientgen/v2/models"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/gsctl/formatting"
	"github.com/giantswarm/gsctl/nodespec"
)

func getOutputAWS(nodePool *models.V5GetNodePoolResponse) (string, error) {
	awsInfo, err := nodespec.NewAWS()
	if err != nil {
		return "", microerror.Mask(err)
	}

	instanceTypeDetails, err := awsInfo.GetInstanceTypeDetails(nodePool.NodeSpec.Aws.InstanceType)
	if nodespec.IsInstanceTypeNotFoundErr(err) {
		// We deliberately ignore "instance type not found", but respect all other errors.
	} else if err != nil {
		return "", microerror.Mask(err)
	}

	var instanceTypes string
	{
		if len(nodePool.Status.InstanceTypes) > 0 {
			instanceTypes = strings.Join(nodePool.Status.InstanceTypes, ",")
		} else {
			instanceTypes = nodePool.NodeSpec.Aws.InstanceType
		}
	}

	var table []string
	{
		table = append(table, color.YellowString("ID:")+"|"+nodePool.ID)
		table = append(table, color.YellowString("Name:")+"|"+nodePool.Name)
		table = append(table, color.YellowString("Node instance types:")+"|"+formatInstanceTypeAWS(instanceTypes, instanceTypeDetails))
		table = append(table, color.YellowString("Alike instances types:")+fmt.Sprintf("|%t", nodePool.NodeSpec.Aws.UseAlikeInstanceTypes))
		table = append(table, color.YellowString("Availability zones:")+"|"+formatting.AvailabilityZonesList(nodePool.AvailabilityZones))
		table = append(table, color.YellowString("On-demand base capacity:")+fmt.Sprintf("|%d", nodePool.NodeSpec.Aws.InstanceDistribution.OnDemandBaseCapacity))
		table = append(table, color.YellowString("Spot percentage above base capacity:")+fmt.Sprintf("|%d", 100-nodePool.NodeSpec.Aws.InstanceDistribution.OnDemandPercentageAboveBaseCapacity))
		table = append(table, color.YellowString("Node scaling:")+"|"+formatNodeScalingAWS(nodePool.Scaling))
		table = append(table, color.YellowString("Nodes desired:")+fmt.Sprintf("|%d", nodePool.Status.Nodes))
		table = append(table, color.YellowString("Nodes in state Ready:")+fmt.Sprintf("|%d", nodePool.Status.NodesReady))
		table = append(table, color.YellowString("Spot instances:")+fmt.Sprintf("|%d", nodePool.Status.SpotInstances))
		table = append(table, color.YellowString("CPUs:")+"|"+formatCPUsAWS(nodePool.Status.NodesReady, instanceTypeDetails))
		table = append(table, color.YellowString("RAM:")+"|"+formatRAMAWS(nodePool.Status.NodesReady, instanceTypeDetails))
	}

	return columnize.SimpleFormat(table), nil
}

func formatInstanceTypeAWS(instanceTypeName string, details *nodespec.InstanceType) string {
	if details != nil {
		return fmt.Sprintf("%s - %d GB RAM, %d CPUs each",
			instanceTypeName,
			details.MemorySizeGB,
			details.CPUCores)
	}

	return fmt.Sprintf("%s %s", instanceTypeName, color.RedString("(no information available on this instance type)"))
}

func formatCPUsAWS(numNodes int64, details *nodespec.InstanceType) string {
	if details != nil {
		return fmt.Sprintf("%d", numNodes*int64(details.CPUCores))
	}

	return "n/a"
}

func formatRAMAWS(numNodes int64, details *nodespec.InstanceType) string {
	if details != nil {
		return fmt.Sprintf("%d GB", numNodes*int64(details.MemorySizeGB))
	}

	return "n/a"
}

func formatNodeScalingAWS(scaling *models.V5GetNodePoolResponseScaling) string {

	minScale := int64(0)
	if scaling.Min != nil {
		minScale = *scaling.Min
	}

	if minScale == scaling.Max {
		return fmt.Sprintf("Pinned to %d", minScale)
	}

	return fmt.Sprintf("Autoscaling between %d and %d", minScale, scaling.Max)
}
