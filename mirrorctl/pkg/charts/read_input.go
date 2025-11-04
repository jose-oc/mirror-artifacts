package charts

import (
	"fmt"
	"os"

	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/types"
	"gopkg.in/yaml.v3"
)

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
