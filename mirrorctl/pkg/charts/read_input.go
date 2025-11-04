package charts

import (
	"fmt"
	"os"

	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/types"
	"gopkg.in/yaml.v3"
)

// LoadChartsList reads a YAML file containing a list of charts and returns a ChartsList object.
// It takes the path to the YAML file as input.
// It returns a pointer to a ChartsList object and an error if the file cannot be read or unmarshalled.
func LoadChartsList(filePath string) (*types.ChartsList, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read charts file: %w", err)
	}

	var chartsList types.ChartsList
	if err := yaml.Unmarshal(data, &chartsList); err != nil {
		return nil, fmt.Errorf("failed to unmarshal charts file: %w", err)
	}

	return &chartsList, nil
}
