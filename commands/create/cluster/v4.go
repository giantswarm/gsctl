package cluster

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/fatih/color"
	"github.com/giantswarm/gsclientgen/v2/models"
	"github.com/giantswarm/microerror"
	yaml "gopkg.in/yaml.v2"

	"github.com/giantswarm/gsctl/client"
	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/commands/types"
	"github.com/giantswarm/gsctl/formatting"
)

// updateDefinitionFromFlagsV4 extend/overwrites a clusterDefinition based on the
// flags/arguments the user has given.
func updateDefinitionFromFlagsV4(def *types.ClusterDefinitionV4, clusterName, releaseVersion, owner string) {
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

// createAddClusterBodyV4 creates a models.V4AddClusterRequest from cluster definition.
func createAddClusterBodyV4(d *types.ClusterDefinitionV4) *models.V4AddClusterRequest {
	a := &models.V4AddClusterRequest{}

	if d != nil {
		a.AvailabilityZones = int64(d.AvailabilityZones)
		a.Name = d.Name
		a.Owner = &d.Owner
		a.ReleaseVersion = d.ReleaseVersion

		if d.Scaling.Min > 0 {
			a.Scaling = &models.V4AddClusterRequestScaling{
				Min: &d.Scaling.Min,
				Max: d.Scaling.Max,
			}
		}

		// We accept only exactly one worker item, as the number of worker nodes is
		// determined via the scaling key.
		if len(d.Workers) == 1 {
			worker := &models.V4AddClusterRequestWorkersItems{
				Memory: &models.V4AddClusterRequestWorkersItemsMemory{
					SizeGb: float64(d.Workers[0].Memory.SizeGB),
				},
				CPU: &models.V4AddClusterRequestWorkersItemsCPU{
					Cores: int64(d.Workers[0].CPU.Cores),
				},
				Storage: &models.V4AddClusterRequestWorkersItemsStorage{
					SizeGb: float64(d.Workers[0].Storage.SizeGB),
				},
				Aws: &models.V4AddClusterRequestWorkersItemsAws{
					InstanceType: d.Workers[0].AWS.InstanceType,
				},
				Azure: &models.V4AddClusterRequestWorkersItemsAzure{
					VMSize: d.Workers[0].Azure.VMSize,
				},
			}

			a.Workers = append(a.Workers, worker)
		}
	}

	return a
}

func addClusterV4(def *types.ClusterDefinitionV4, args Arguments, clientWrapper *client.Wrapper, auxParams *client.AuxiliaryParams) (id, location string, err error) {
	// Let user-provided arguments (flags) overwrite/extend definition from YAML.

	// Validate definition
	if def.Owner == "" {
		return "", "", microerror.Mask(errors.ClusterOwnerMissingError)
	}

	// create JSON API call payload to catch and handle errors early
	addClusterBody := createAddClusterBodyV4(def)
	_, marshalErr := json.Marshal(addClusterBody)
	if marshalErr != nil {
		return "", "", microerror.Maskf(errors.CouldNotCreateJSONRequestBodyError, marshalErr.Error())
	}

	// Preview in YAML format
	if args.Verbose {
		fmt.Println("\nDefinition for the requested cluster:")
		d, marshalErr := yaml.Marshal(addClusterBody)
		if marshalErr != nil {
			log.Fatalf("error: %v", marshalErr)
		}
		fmt.Printf(color.CyanString(string(d)))
		fmt.Println()
	}

	if args.OutputFormat != formatting.OutputFormatJSON {
		fmt.Printf("Requesting new cluster for organization '%s'\n", color.CyanString(def.Owner))
	}

	// perform API call
	response, err := clientWrapper.CreateClusterV4(addClusterBody, auxParams)
	if err != nil {
		return "", "", microerror.Mask(err)
	}

	// success
	location = response.Location
	id = strings.Split(location, "/")[3]

	return id, location, nil
}
