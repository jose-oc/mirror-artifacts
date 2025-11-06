package cmdutils

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/appcontext"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/types"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func PrintConfigFileInfo(ctx *appcontext.AppContext) {
	if viper.GetBool("quiet") {
		return
	}
	// Handle color disabling if needed
	color.NoColor = viper.GetBool("no_color")

	_, err := color.New(color.Faint).Println("Using config file:", viper.ConfigFileUsed())
	if err != nil {
		log.Error().Err(err).Msg("Failed to print config file info message")
	}
}

func PrintDryRunMessage(ctx *appcontext.AppContext) {
	if viper.GetBool("quiet") {
		return
	}
	// Handle color disabling if needed
	color.NoColor = viper.GetBool("no_color")

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
	// Handle color disabling if needed
	color.NoColor = viper.GetBool("no_color")

	green := color.New(color.FgGreen).SprintFunc()
	greenBold := color.New(color.FgGreen, color.Bold).SprintFunc()
	redBold := color.New(color.FgHiRed).Add(color.Bold).SprintFunc()
	red := color.New(color.FgHiRed).SprintFunc()

	fmt.Printf("%s: \n %s\n", greenBold("Charts pushed"), green(strings.Join(successfulCharts, "\n ")))
	if len(failedCharts) > 0 {
		fmt.Printf("%s: \n %s\n", redBold("Charts failed to push"), red(strings.Join(failedCharts, "\n ")))
	}
}

func PrintImagesPushed(imagesPushed, imagesFailed []string) {
	if viper.GetBool("quiet") {
		return
	}
	// Handle color disabling if needed
	color.NoColor = viper.GetBool("no_color")

	green := color.New(color.FgGreen).SprintFunc()
	greenBold := color.New(color.FgGreen, color.Bold).SprintFunc()
	redBold := color.New(color.FgHiRed).Add(color.Bold).SprintFunc()
	red := color.New(color.FgHiRed).SprintFunc()

	fmt.Printf("%s: \n %s\n", greenBold("Images pushed"), green(strings.Join(imagesPushed, "\n ")))
	if len(imagesFailed) > 0 {
		fmt.Printf("%s: \n %s\n", redBold("Images failed"), red(strings.Join(imagesFailed, "\n ")))
	}
}

// PrintImageListByChart prints a map of images grouped by chart in a formatted, readable way.
func PrintImageListByChart(imagesByChart map[string][]types.Image) {
	if viper.GetBool("quiet") {
		return
	}
	// Handle color disabling if needed
	color.NoColor = viper.GetBool("no_color")

	// 1. Initialize color functions
	green := color.New(color.FgGreen).SprintFunc()
	greenBold := color.New(color.FgGreen, color.Bold).SprintFunc()
	greenBoldUnderlined := color.New(color.FgGreen, color.Bold, color.Underline).SprintFunc()

	fmt.Printf("%s:\n", greenBoldUnderlined("Images by chart"))

	// 2. Sort keys for deterministic output (Best Practice)
	chartNames := make([]string, 0, len(imagesByChart))
	for name := range imagesByChart {
		chartNames = append(chartNames, name)
	}
	sort.Strings(chartNames)

	// 3. Iterate over sorted chart names
	for _, chartName := range chartNames {
		images := imagesByChart[chartName]

		// Print the Chart Name (Key) in green bold, followed by a newline
		fmt.Printf("\n%s:\n", greenBold(chartName))

		// Iterate over the images for the current chart
		for _, img := range images {
			// Print the Image Source (Value) in green with 4 spaces of indentation
			// We use %s to ensure the output is a single argument, preventing the
			// extra space fmt.Println would add.
			fmt.Printf("    %s\n", green(img.Source))
		}
	}

	// Add a final newline for clean terminal output after the list finishes
	fmt.Println()
}
