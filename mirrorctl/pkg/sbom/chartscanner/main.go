package chartscanner

import (
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/appcontext"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/charts"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/helm"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/types"
	"github.com/rs/zerolog/log"
)

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
		srcChartPath, err := helm.PullHelmChart(ch, tmpDir)
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
