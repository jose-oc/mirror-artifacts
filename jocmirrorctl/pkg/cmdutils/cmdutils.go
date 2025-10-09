package cmdutils

import (
	"github.com/jose-oc/poc-mirror-artifacts/jocmirrorctl/pkg/appcontext"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// MirrorImages handles the `mirror images` subcommand
func MirrorImages(ctx *appcontext.AppContext, cmd *cobra.Command) {
	imagesFile := viper.GetString("images")
	ctx.Logger.Info().Str("images_file", imagesFile).Msg("Mirroring images (stub)")
	if ctx.DryRun {
		ctx.Logger.Info().Msg("Dry-run: Would mirror images to GAR")
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
		Msg("Mirroring all artifacts (stub)")
	if ctx.DryRun {
		ctx.Logger.Info().Msg("Dry-run: Would mirror images and charts to GAR")
	}
}

// Verify handles the `verify` subcommand
func Verify(ctx *appcontext.AppContext, cmd *cobra.Command) {
	ctx.Logger.Info().Msg("Verifying artifacts (stub)")
	if ctx.DryRun {
		ctx.Logger.Info().Msg("Dry-run: Would verify artifacts in GAR")
	}
}
