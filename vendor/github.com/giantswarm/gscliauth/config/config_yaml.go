package config

import (
	"github.com/giantswarm/microerror"
)

func (c *configStruct) MarshalYAML() (interface{}, error) {
	type configStructClone configStruct

	v := struct {
		ConfigStruct configStructClone          `yaml:",inline"`
		Endpoints    map[string]*endpointConfig `yaml:"endpoints"`
	}{
		// Due to yaml library quirks it has to be a value:
		// https://github.com/go-yaml/yaml/issues/55. Otherwise we
		// have:
		//
		//	panic: Option ,inline needs a struct value field.
		//
		ConfigStruct: (configStructClone)(*c),
		Endpoints:    c.endpoints,
	}

	return &v, nil
}

func (c *configStruct) UnmarshalYAML(unmarshal func(interface{}) error) error {
	v := struct {
		Endpoints map[string]*endpointConfig `yaml:"endpoints"`
	}{}

	err := unmarshal(&v)
	if err != nil {
		return microerror.Mask(err)
	}

	type configStructClone configStruct

	err = unmarshal((*configStructClone)(c))
	if err != nil {
		return microerror.Mask(err)
	}

	c.endpoints = v.Endpoints

	return nil
}
