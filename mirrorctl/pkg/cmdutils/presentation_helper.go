package cmdutils

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/appcontext"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func PrintConfigFileInfo(ctx *appcontext.AppContext) {
	if viper.GetBool("quiet") {
		return
	}
	_, err := color.New(color.Faint).Println("Using config file:", viper.ConfigFileUsed())
	if err != nil {
		log.Error().Err(err).Msg("Failed to print config file info message")
	}
}

func PrintDryRunMessage(ctx *appcontext.AppContext) {
	if viper.GetBool("quiet") {
		return
	}
	if ctx.DryRun {
		yellow := color.New(color.FgYellow).SprintFunc()
		red := color.New(color.FgHiMagenta).Add(color.Bold).Add(color.Underline).SprintFunc()
		fmt.Printf("%s: %s.\n", red("Dry-run mode"), yellow("nothing pushed"))
	}
}
