package cluster

import (
	"bufio"
	"os"

	"github.com/giantswarm/gsctl/commands/errors"
	"github.com/giantswarm/gsctl/commands/types"
	"github.com/giantswarm/microerror"
	"github.com/spf13/afero"
	yaml "gopkg.in/yaml.v2"
)

// readDefinitionFromYAML reads a cluster definition from YAML data.
func readDefinitionFromYAML(yamlBytes []byte) (interface{}, error) {
	// First unmarshal into a map so we can detect v4 or v5 schema.
	rawMap := map[string]interface{}{}

	err := yaml.Unmarshal(yamlBytes, rawMap)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Detecting v5 purely based on the existence of the 'api_version' key.
	if _, apiVersionOK := rawMap["api_version"]; apiVersionOK {
		// v5
		def := &types.ClusterDefinitionV5{}
		err := yaml.Unmarshal(yamlBytes, def)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		return def, nil
	}

	// v4 (default)
	def := &types.ClusterDefinitionV4{}
	err = yaml.Unmarshal(yamlBytes, def)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return def, nil

}

// readDefinitionFromFile reads a cluster definition from a YAML file.
func readDefinitionFromFile(fs afero.Fs, path string) (interface{}, error) {
	data, err := afero.ReadFile(fs, path)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return readDefinitionFromYAML(data)
}

// readDefinitionFromSTDIN reads a YAML definition coming via standard input.
// TODO: provide unit test
func readDefinitionFromSTDIN() (interface{}, error) {
	yamlString := ""
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		yamlString += scanner.Text() + "\n"
	}

	if err := scanner.Err(); err != nil {
		return nil, microerror.Mask(err)
	}

	def, err := readDefinitionFromYAML([]byte(yamlString))
	if err != nil {
		return nil, microerror.Maskf(errors.YAMLFileNotReadableError, err.Error())
	}

	return def, nil
}
