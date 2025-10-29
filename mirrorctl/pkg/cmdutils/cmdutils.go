package cmdutils

import (
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/appcontext"
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
	if _, _, err := images.MirrorImages(ctx, imagesFile); err != nil {
		log.Fatal().Err(err).Msg("Failed to mirror images")
	}
}
