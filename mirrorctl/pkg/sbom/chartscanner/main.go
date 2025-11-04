package chartscanner

import (
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/appcontext"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/charts"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/helm"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/types"
	"github.com/rs/zerolog/log"
)

// ExtractImagesFromCharts extracts the container images from a list of Helm charts.
// It takes an application context and the path to the file containing the list of charts as input.
// It returns a map of strings to slices of images, where the keys are the chart names and the values are the slices of images.
// It also returns an error if the extraction fails.
func ExtractImagesFromCharts(ctx *appcontext.AppContext, chartsFile string) (map[string][]types.Image, error) {
	chartsList, err := charts.LoadChartsList(chartsFile)
	if err != nil {
		return nil, err
	}

	imagesByChart := make(map[string][]types.Image)

	tmpDir, err := helm.CreateTempDir(ctx)
	if err != nil {
		return nil, err
	}

	for _, ch := range chartsList.Charts {
		srcChartPath, err := helm.PullChart(ch, tmpDir)
		if err != nil {
			log.Error().Err(err).Str("chart", ch.Name).Msg("Failed to pull chart")
			continue
		}

		images, err := ScanChart(srcChartPath)
		if err != nil {
			log.Error().Err(err).Str("chart", ch.Name).Msg("Failed to extract images from chart")
			continue
		}

		imagesByChart[ch.Name] = images
	}

	return imagesByChart, nil
}
