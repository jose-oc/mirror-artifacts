package chartscanner

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/types"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3" // Import yaml.v3
)

// chartYaml represents a simplified structure of Chart.yaml to extract appVersion.
type chartYaml struct {
	AppVersion string `yaml:"appVersion"`
}

// getAppVersionFromChartYaml reads a Chart.yaml file and extracts the appVersion.
// It returns the appVersion as a string, or an empty string if not found or an error occurs.
func getAppVersionFromChartYaml(chartYamlPath string) string {
	log.Debug().Str("chartYamlPath", chartYamlPath).Msg("Attempting to read Chart.yaml for appVersion")
	content, err := os.ReadFile(chartYamlPath)
	if err != nil {
		log.Debug().Err(err).Str("chartYamlPath", chartYamlPath).Msg("Failed to read Chart.yaml")
		return ""
	}

	var chart chartYaml
	if err := yaml.Unmarshal(content, &chart); err != nil {
		log.Debug().Err(err).Str("chartYamlPath", chartYamlPath).Msg("Failed to unmarshal Chart.yaml")
		return ""
	}
	log.Debug().Str("chartYamlPath", chartYamlPath).Str("appVersion", chart.AppVersion).Msg("Extracted appVersion from Chart.yaml")
	return chart.AppVersion
}

// parseImageSource extracts the repository and tag from a full image source string.
// It assumes the format "repository:tag". If no tag is present, "latest" is assumed.
func parseImageSource(source string) (repository, tag string) {
	lastColon := strings.LastIndex(source, ":")
	if lastColon == -1 {
		return source, "latest" // No tag found, assume latest
	}
	return source[:lastColon], source[lastColon+1:]
}

// ScanChart scans a Helm chart directory for container images.
// It walks through the directory, parses YAML files, and extracts image references.
// For images with the "latest" tag, it also returns an image tagged with the chart's appVersion.
// It returns a slice of unique images found in the chart.
func ScanChart(chartPath string) ([]types.Image, error) {
	uniqueImages := make(map[string]types.Image)

	log.Debug().Msgf("Scanning chart directory: %s", chartPath)

	// First, extract global registry from values.yaml or values.yml
	globalRegistry := extractGlobalRegistry(chartPath)
	if globalRegistry != "" {
		log.Debug().Msgf("Found global registry: %s", globalRegistry)
	}

	err := filepath.Walk(chartPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml" {
			log.Trace().Msgf("Parsing YAML file: %s", path)
			images, err := parseYAML(path)
			if err != nil {
				log.Warn().Err(err).Msgf("Failed to parse YAML file: %s", path)
				return nil // Don't stop the walk, just skip this file
			}
			for _, img := range images {
				// Apply global registry override if present
				isSubChart, err := isSubChart(path, chartPath)
				if err != nil {
					continue
				}
				if globalRegistry != "" && globalRegistry != "null" && !isSubChart {
					img = applyGlobalRegistry(img, globalRegistry)
				}
				uniqueImages[img.Source] = img

				// Handle "latest" tag with appVersion
				repo, tag := parseImageSource(img.Source)
				if tag == "latest" {
					chartYamlFile := findChartYamlForPath(path, chartPath)
					if chartYamlFile != "" {
						appVersion := getAppVersionFromChartYaml(chartYamlFile)
						if appVersion != "" {
							appVersionImage := types.Image{
								Name:   img.Name, // Keep the original name
								Source: repo + ":" + appVersion,
							}
							uniqueImages[appVersionImage.Source] = appVersionImage
							log.Debug().Msgf("Added appVersion image: %s for latest image: %s", appVersionImage.Source, img.Source)
						}
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Convert map to slice and sort for consistent output
	var result []types.Image
	for _, img := range uniqueImages {
		result = append(result, img)
	}
	sort.Slice(result, func(i, j int) bool {
		return strings.Compare(result[i].Source, result[j].Source) < 0
	})

	return result, nil
}

// isSubChart returns true if the path belongs to a chart dependency or subchart
func isSubChart(path string, chartPath string) (bool, error) {
	relPath, err := filepath.Rel(chartPath, path)
	if err != nil {
		log.Warn().Err(err).Msgf("Failed to resolve relative path: %s", path)
		return false, err
	}

	return strings.HasPrefix(relPath, "charts"+string(os.PathSeparator)), nil
}

// findChartYamlForPath determines the correct Chart.yaml file for a given file path within a chart.
func findChartYamlForPath(filePath, chartRootPath string) string {
	log.Debug().Str("filePath", filePath).Str("chartRootPath", chartRootPath).Msg("Finding Chart.yaml for path")
	relPath, err := filepath.Rel(chartRootPath, filePath)
	if err != nil {
		log.Debug().Err(err).Str("filePath", filePath).Str("chartRootPath", chartRootPath).Msg("Failed to get relative path")
		return ""
	}

	// If the file is in the root of the chart, use the main Chart.yaml
	if filepath.Dir(relPath) == "." {
		chartYamlPath := filepath.Join(chartRootPath, "Chart.yaml")
		log.Debug().Str("resolvedChartYamlPath", chartYamlPath).Msg("Resolved Chart.yaml to main chart")
		return chartYamlPath
	}

	// If the file is in a subchart, find the Chart.yaml for that subchart
	parts := strings.Split(relPath, string(os.PathSeparator))
	for i := 0; i < len(parts); i++ {
		if parts[i] == "charts" && i+1 < len(parts) {
			subchartName := parts[i+1]
			subchartPath := filepath.Join(chartRootPath, "charts", subchartName)
			chartYamlPath := filepath.Join(subchartPath, "Chart.yaml")
			log.Debug().Str("resolvedChartYamlPath", chartYamlPath).Msg("Resolved Chart.yaml to subchart")
			return chartYamlPath
		}
	}

	// Fallback to main Chart.yaml if no specific subchart Chart.yaml is found
	chartYamlPath := filepath.Join(chartRootPath, "Chart.yaml")
	log.Debug().Str("resolvedChartYamlPath", chartYamlPath).Msg("Resolved Chart.yaml to main chart (fallback)")
	return chartYamlPath
}
