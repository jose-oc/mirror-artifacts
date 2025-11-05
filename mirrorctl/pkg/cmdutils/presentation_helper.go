package cmdutils

import (
	"fmt"
	"strings"

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
		fmt.Printf("\n%s: %s.\n", red("Dry-run mode"), yellow("nothing pushed"))
	}
}

func PrintChartsPushed(successfulCharts []string, failedCharts []string) {
	if viper.GetBool("quiet") {
		return
	}

	green := color.New(color.FgGreen).SprintFunc()
	greenBold := color.New(color.FgGreen, color.Bold).SprintFunc()
	redBold := color.New(color.FgHiRed).Add(color.Bold).SprintFunc()
	red := color.New(color.FgHiRed).SprintFunc()

	fmt.Printf("%s: \n %s\n", greenBold("Charts pushed"), green(strings.Join(successfulCharts, "\n ")))
	if len(failedCharts) > 0 {
		fmt.Printf("%s: \n %s\n", redBold("Charts failed to push"), red(strings.Join(failedCharts, "\n ")))
	}
}

func PrintImagesPushed(imagesPushed []string) {
	if viper.GetBool("quiet") {
		return
	}

	green := color.New(color.FgGreen).SprintFunc()
	greenBold := color.New(color.FgGreen, color.Bold).SprintFunc()
	fmt.Printf("%s: \n %s\n", greenBold("Images pushed"), green(strings.Join(imagesPushed, "\n ")))
}
