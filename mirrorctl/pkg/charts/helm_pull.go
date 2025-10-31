package charts

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

func pullHelmChart(ch Chart, tmpDir string) (string, error) {
	log.Debug().Str("chart", ch.Name).Str("version", ch.Version).Msg("Pulling chart")

	chartPath, err := downloadChart(ch, tmpDir)
	if err != nil {
		return "", err
	}

	log.Info().Str("chart", ch.Name).Str("version", ch.Version).Str("temporary path", tmpDir).
		Msg("Helm chart pulled")
	return chartPath, nil
}

func downloadChart(chart Chart, destDir string) (string, error) {
	log.Debug().Str("chart", chart.Name).Str("source", chart.Source).Msg("Downloading chart")

	settings := cli.New()
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), func(format string, v ...interface{}) {
		log.Debug().Msgf(format, v...)
	}); err != nil {
		return "", fmt.Errorf("failed to init action config: %w", err)
	}

	client := action.NewPullWithOpts(action.WithConfig(actionConfig))
	client.Settings = settings
	client.RepoURL = chart.Source
	client.Version = chart.Version
	client.DestDir = destDir
	client.Untar = true
	client.UntarDir = destDir

	_, err := client.Run(chart.Name)
	if err != nil {
		return "", fmt.Errorf("failed to download chart: %w", err)
	}

	return filepath.Join(destDir, chart.Name), nil
}
