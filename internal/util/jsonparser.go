package util

import (
	"encoding/json"
	"os"
)

type TemplateReposConfig struct {
	LabEnvSetup struct {
		Repos []string `json:"repos"`
	} `json:"lab-env-setup"`
}

func LoadFromJsonFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config TemplateReposConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return config.LabEnvSetup.Repos, nil
}
