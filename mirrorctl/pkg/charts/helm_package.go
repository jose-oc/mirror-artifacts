package charts

import (
	"path/filepath"

	"github.com/rs/zerolog/log"
	"helm.sh/helm/v3/pkg/action"
)

// packageHelmChart packages a Helm chart into a .tgz file.
// It takes the path to the chart as input.
// It returns the path to the packaged chart and an error if the packaging fails.
func packageHelmChart(chartPath string) (string, error) {
	tmpDir := filepath.Dir(chartPath)
	p := action.NewPackage()
	p.Destination = tmpDir
	packagedChartPath, err := p.Run(chartPath, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to package chart")
		return "", err
	}

	return packagedChartPath, nil
}
