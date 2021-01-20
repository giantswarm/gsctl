package nodepool

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/giantswarm/columnize"
	"github.com/giantswarm/gsclientgen/v2/models"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/gsctl/nodespec"
)

func getOutputAzure(nodePool *models.V5GetNodePoolResponse) (string, error) {
	azureInfo, err := nodespec.NewAzureProvider()
	if err != nil {
		return "", microerror.Mask(err)
	}

	vmSizeDetails, err := azureInfo.GetVMSizeDetails(nodePool.NodeSpec.Azure.VMSize)
	if nodespec.IsVMSizeNotFoundErr(err) {
		// We deliberately ignore "vm size not found", but respect all other errors.
	} else if err != nil {
		return "", microerror.Mask(err)
	}

	var vmSizes string
	{
		if len(nodePool.Status.InstanceTypes) > 0 {
			vmSizes = strings.Join(nodePool.Status.InstanceTypes, ",")
		} else {
			vmSizes = nodePool.NodeSpec.Azure.VMSize
		}
	}

	var table []string
	{
		table = append(table, color.YellowString("ID:")+"|"+nodePool.ID)
		table = append(table, color.YellowString("Name:")+"|"+nodePool.Name)
		table = append(table, color.YellowString("Node VM sizes:")+"|"+formatVMSizeAzure(vmSizes, vmSizeDetails))
		table = append(table, color.YellowString("Availability zones:")+"|"+formatAZsAzure(nodePool.AvailabilityZones))
		table = append(table, color.YellowString("Node scaling:")+"|"+formatNodeScalingAzure(nodePool.Scaling))
		table = append(table, color.YellowString("Nodes desired:")+fmt.Sprintf("|%d", nodePool.Status.Nodes))
		table = append(table, color.YellowString("Nodes in state Ready:")+fmt.Sprintf("|%d", nodePool.Status.NodesReady))

		if nodePool.NodeSpec.Azure.SpotInstances != nil && nodePool.NodeSpec.Azure.SpotInstances.Enabled {
			table = append(table, color.YellowString("Spot instances:")+"|Enabled")
			table = append(table, color.YellowString("Spot instances max price:")+fmt.Sprintf("|$%.5f", nodePool.NodeSpec.Azure.SpotInstances.MaxPrice))
		} else {
			table = append(table, color.YellowString("Spot instances:")+"|Disabled")
		}

		table = append(table, color.YellowString("CPUs:")+"|"+formatCPUsAzure(nodePool.Status.NodesReady, vmSizeDetails))
		table = append(table, color.YellowString("RAM:")+"|"+formatRAMAzure(nodePool.Status.NodesReady, vmSizeDetails))
	}

	return columnize.SimpleFormat(table), nil
}

func formatVMSizeAzure(vmSize string, details *nodespec.VMSize) string {
	if details != nil {
		return fmt.Sprintf("%s - %.1f GB RAM, %d CPUs each",
			vmSize,
			details.MemoryInMB/1000,
			details.NumberOfCores)
	}

	return fmt.Sprintf("%s %s", vmSize, color.RedString("(no information available on this vm size)"))
}

func formatNodeScalingAzure(scaling *models.V5GetNodePoolResponseScaling) string {
	minScale := int64(0)
	if scaling.Min != nil {
		minScale = *scaling.Min
	}

	if minScale == scaling.Max {
		return fmt.Sprintf("Pinned to %d", minScale)
	}

	return fmt.Sprintf("Autoscaling between %d and %d", minScale, scaling.Max)
}

func formatCPUsAzure(nodesReady int64, details *nodespec.VMSize) string {
	if details != nil {
		return strconv.FormatInt(nodesReady*details.NumberOfCores, 10)
	}

	return "n/a"
}

func formatRAMAzure(nodesReady int64, details *nodespec.VMSize) string {
	if details != nil {
		return strconv.FormatFloat(float64(nodesReady)*details.MemoryInMB/1000, 'f', 1, 64)
	}

	return "n/a"
}

func formatAZsAzure(azs []string) string {
	if len(azs) > 0 {
		return strings.Join(azs, ",")
	}

	return "n/a"
}
