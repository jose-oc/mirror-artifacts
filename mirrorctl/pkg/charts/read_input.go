package charts

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Chart represents an entry in the charts.yaml file
type Chart struct {
	Name    string `yaml:"name"`
	Source  string `yaml:"source"`
	Version string `yaml:"version"`
}

// ChartsList represents the structure of the charts.yaml file
type ChartsList struct {
	Charts []Chart `yaml:"charts"`
}

func loadChartsList(filePath string) (*ChartsList, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read charts file: %w", err)
	}

	var chartsList ChartsList
	if err := yaml.Unmarshal(data, &chartsList); err != nil {
		return nil, fmt.Errorf("failed to unmarshal charts file: %w", err)
	}

	return &chartsList, nil
}
