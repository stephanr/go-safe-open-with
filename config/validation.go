package config

import (
	"log"
	"os"

	"github.com/go-faster/yaml"
)

type Configuration struct {
	Allowed []Validation `yaml:"allowed"`
}

type Validation struct {
	Command   string          `yaml:"cmd"`
	Arguments []ValidationArg `yaml:"arguments"`
}

type ValidationArg struct {
	Type         string   `yaml:"type"`
	Value        string   `yaml:"value"`
	Values       []string `yaml:"values"`
	InsertBefore []string `yaml:"insertBefore"`
	InsertAfter  []string `yaml:"insertAfter"`
	TrimLeft     []string `yaml:"trimLeft"`
	TrimRight    []string `yaml:"trimRight"`
	SplitSpace   bool     `yaml:"splitSpace"`
	//RegexString string   `ymal:"regex"`
	//Regex       *regexp.Regexp
}

//func configurationName(fileName string) string {
//	return strings.TrimSuffix(fileName, filepath.Ext(fileName)) + ".yaml"
//}

func ReadConfiguration() Configuration {
	//executable, err := os.Executable()
	//if err != nil {
	//	log.Fatalf("Could not find executable location: %v", err)
	//}
	//configName := configurationName(executable)
	configName := "go-safe-open-with.yaml"

	data, err := os.ReadFile(configName)
	if err != nil {
		log.Fatalf("Could not read configuration file %s: %v", configName, err)
	}

	cfg := Configuration{}

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatalf("Could not load configuration from file %s: %v", configName, err)
	}

	/*
		for idxAllowed, allowed := range cfg.Allowed {
			for idxArg, arg := range allowed.Arguments {
				if arg.Type == "regex" {
					arg.Regex, err = regexp.Compile(arg.RegexString)
					if err != nil {
						log.Fatalf("Failed to load regex in allowed element number %v argument %v: %v", idxAllowed, idxArg, err)
					}
				}
			}
		}
			//*/

	return cfg
}
