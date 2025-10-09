package cmdutils

import (
	"github.com/jose-oc/poc-mirror-artifacts/jocmirrorctl/pkg/appcontext"
	"github.com/jose-oc/poc-mirror-artifacts/jocmirrorctl/pkg/images"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// MirrorImages handles the `mirror images` subcommand
func MirrorImages(ctx *appcontext.AppContext, cmd *cobra.Command) {
	imagesFile := viper.GetString("images")
	if imagesFile == "" {
		ctx.Logger.Fatal().Msg("Images file path is required")
	}
	if err := images.MirrorImages(ctx, imagesFile); err != nil {
		ctx.Logger.Fatal().Err(err).Msg("Failed to mirror images")
	}
}

// MirrorCharts handles the `mirror charts` subcommand
func MirrorCharts(ctx *appcontext.AppContext, cmd *cobra.Command) {
	chartsFile := viper.GetString("charts")
	ctx.Logger.Info().Str("charts_file", chartsFile).Msg("Mirroring charts (stub)")
	if ctx.DryRun {
		ctx.Logger.Info().Msg("Dry-run: Would mirror charts to GAR")
	}
}

// MirrorAll handles the `mirror all` subcommand
func MirrorAll(ctx *appcontext.AppContext, cmd *cobra.Command) {
	imagesFile := viper.GetString("images")
	chartsFile := viper.GetString("charts")
	ctx.Logger.Info().
		Str("images_file", imagesFile).
		Str("charts_file", chartsFile).
		Msg("Mirroring all artifacts")
	if imagesFile != "" {
		if err := images.MirrorImages(ctx, imagesFile); err != nil {
			ctx.Logger.Error().Err(err).Msg("Failed to mirror images, continuing with charts")
		}
	}
	ctx.Logger.Info().Str("charts_file", chartsFile).Msg("Mirroring charts (stub)")
	if ctx.DryRun {
		ctx.Logger.Info().Msg("Dry-run: Would mirror charts to GAR")
	}
}

// Verify handles the `verify` subcommand
func Verify(ctx *appcontext.AppContext, cmd *cobra.Command) {
	ctx.Logger.Info().Msg("Verifying artifacts (stub)")
	if ctx.DryRun {
		ctx.Logger.Info().Msg("Dry-run: Would verify artifacts in GAR")
	}
}
