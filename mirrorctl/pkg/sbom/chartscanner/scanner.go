package chartscanner

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/types"
	"github.com/rs/zerolog/log"
)

// ScanChart scans a Helm chart directory for container images.
// It walks through the directory, parses YAML files, and extracts image references.
// It returns a slice of unique images found in the chart.
func ScanChart(chartPath string) ([]types.Image, error) {
	uniqueImages := make(map[string]types.Image)

	log.Debug().Msgf("Scanning chart directory: %s", chartPath)

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
				uniqueImages[img.Source] = img
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
