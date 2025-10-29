package charts

import (
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/appcontext"
	"github.com/rs/zerolog/log"
)

// MirrorHelmCharts mirrors Helm charts to GAR
func MirrorHelmCharts(ctx *appcontext.AppContext, chartsFile string) error {
	chartsList, err := loadChartsList(chartsFile)
	if err != nil {
		return err
	}

	for _, ch := range chartsList.Charts {
		if err := mirrorHelmChart(ctx, ch); err != nil {
			log.Error().Err(err).Str("chart", ch.Name).Msg("Failed to mirror chart")
			continue
		}
	}

	return nil
}

func mirrorHelmChart(ctx *appcontext.AppContext, chart Chart) error {
	log.Debug().Str("chart", chart.Name).Str("version", chart.Version).Msg("Mirroring chart")

	tmpDir, err := createTempDir(ctx)
	if err != nil {
		return err
	}

	srcChartPath, err := pullHelmChart(chart, tmpDir)
	if err != nil {
		return err
	}

	dstChartPath, err := TransformHelmChart(ctx, chart, srcChartPath)
	if err != nil {
		return err
	}

	pkgChartPath, err := packageHelmChart(dstChartPath)
	if err != nil {
		return err
	}

	if err := pushHelmChart(ctx, pkgChartPath, chart.Name, chart.Version); err != nil {
		return err
	}

	log.Info().Str("chart", chart.Name).Str("version", chart.Version).Msg("Chart successfully mirrored")
	return nil
}
