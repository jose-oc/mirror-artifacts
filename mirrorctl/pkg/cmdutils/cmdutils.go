package cmdutils

import (
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/appcontext"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/charts"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/images"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// MirrorImages handles the `mirror images` subcommand
func MirrorImages(ctx *appcontext.AppContext, _ *cobra.Command) {
	imagesFile := viper.GetString("images")
	if imagesFile == "" {
		log.Fatal().Msg("Images file path is required, please provide via --images flag")
	}
	if ctx.DryRun {
		log.Info().Msg("Dry-run: Would mirror images to GAR")
	}
	if _, _, err := images.MirrorImages(ctx, imagesFile); err != nil {
		log.Fatal().Err(err).Msg("Failed to mirror images")
	}
}

// MirrorCharts handles the `mirror charts` subcommand
func MirrorCharts(ctx *appcontext.AppContext, cmd *cobra.Command) {
	chartsFile := viper.GetString("charts")
	if chartsFile == "" {
		log.Fatal().Msg("Charts file path is required")
	}
	if ctx.DryRun {
		log.Info().Msg("Dry-run: Would mirror charts to GAR")
	}
	if err := charts.MirrorHelmCharts(ctx, chartsFile); err != nil {
		log.Fatal().Err(err).Msg("Failed to mirror charts")
	}
}
