package horenso

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type handlers []string

type config struct {
	Reporter       handlers `yaml:"reporter"`
	Noticer        handlers `yaml:"noticer"`
	Timestamp      bool     `yaml:"timestamp"`
	Tag            string   `yaml:"tag"`
	OverrideStatus bool     `yaml:"overrideStatus"`
	Logfile        string   `yaml:"log"`
}

func (ha *handlers) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var aux interface{}
	if err := unmarshal(&aux); err != nil {
		return err
	}
	switch raw := aux.(type) {
	case string:
		*ha = []string{raw}
	case []interface{}:
		list := make([]string, len(raw))
		for i, r := range raw {
			v, ok := r.(string)
			if !ok {
				return fmt.Errorf("handlers should be a string or an array of string: %v", aux)
			}
			list[i] = v
		}
		*ha = list
	default:
		return fmt.Errorf("handlers should be a string or an array of string: %v", aux)
	}
	return nil
}

func loadConfig(file string) (*config, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	c := &config{}
	err = yaml.NewDecoder(f).Decode(c)
	return c, err
}
