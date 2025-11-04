package charts

import (
	"fmt"

	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/appcontext"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/helm"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/types"
	"github.com/rs/zerolog/log"
)

// MirrorHelmCharts mirrors a list of Helm charts to a Google Artifact Registry.
// It takes an application context and the path to a file containing the list of charts to mirror.
// It returns three values:
// - A slice of strings containing the names of the charts that were successfully mirrored.
// - A slice of strings containing the names of the charts that failed to be mirrored.
// - An error if the file containing the list of charts could not be read.
func MirrorHelmCharts(ctx *appcontext.AppContext, chartsFile string) ([]string, []string, error) {
	chartsList, err := LoadChartsList(chartsFile)
	if err != nil {
		// Only return an error here if the failure prevents processing any chart
		return nil, nil, err
	}

	// Initialize the lists to be returned
	var successfulCharts []string
	var failedCharts []string

	for _, ch := range chartsList.Charts {
		// Format the chart identifier as "name:version" for the lists
		chartDetail := fmt.Sprintf("%s:%s", ch.Name, ch.Version)

		if err := mirrorChart(ctx, ch); err != nil {
			log.Error().Err(err).Str("chart", ch.Name).Msg("Failed to mirror chart")
			failedCharts = append(failedCharts, chartDetail) // Add to failed list
			continue
		}

		successfulCharts = append(successfulCharts, chartDetail) // Add to successful list
	}

	// Return the two lists and a nil error (since processing the loop was successful)
	return successfulCharts, failedCharts, nil
}

// mirrorChart mirrors a single Helm chart to a Google Artifact Registry.
// It takes an application context and a Chart object as input.
// It returns an error if the chart could not be mirrored.
func mirrorChart(ctx *appcontext.AppContext, chart types.Chart) error {
	log.Debug().Str("chart", chart.Name).Str("version", chart.Version).Msg("Mirroring chart")

	tmpDir, err := helm.CreateTempDir(ctx)
	if err != nil {
		return err
	}

	srcChartPath, err := helm.PullChart(chart, tmpDir)
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

	if err := pushChart(ctx, pkgChartPath, chart.Name, chart.Version); err != nil {
		return err
	}

	if ctx.DryRun {
		log.Info().Str("chart", chart.Name).Str("version", chart.Version).Msg("Running in dry-run, chart would have been mirrored")
	} else {
		log.Info().Str("chart", chart.Name).Str("version", chart.Version).Msg("Chart successfully mirrored")
	}
	return nil
}
