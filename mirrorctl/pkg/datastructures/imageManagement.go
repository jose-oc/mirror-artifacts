package datastructures

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/types"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// ExportFormat represents supported export formats
type ExportFormat string

const (
	FormatJSON ExportFormat = "json"
	FormatYAML ExportFormat = "yaml"
)

// WriteImagesToFile serializes and writes a list of images to a file in either JSON or YAML format.
// The format is determined by the file extension (.json or .yaml/.yml).
// It takes a slice of images and the output path as input.
// It returns an error if marshaling or file writing fails.
func WriteImagesToFile(images []types.Image, outputPath string) error {
	imageList := types.ImagesList{Images: images}

	format, err := getExportFormat(outputPath)
	if err != nil {
		return fmt.Errorf("invalid output format: %w", err)
	}

	var data []byte
	switch format {
	case FormatJSON:
		data, err = json.MarshalIndent(imageList, "", "  ")
	case FormatYAML:
		data, err = yaml.Marshal(imageList)
	}
	if err != nil {
		return fmt.Errorf("failed to marshal images: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		log.Error().Err(err).Str("file", outputPath).Msg("Failed to write images to file")
		return fmt.Errorf("failed to write images to file %s: %w", outputPath, err)
	}

	log.Info().Str("file", outputPath).Msg("Successfully wrote images to file")
	return nil
}

// getExportFormat determines the export format from a file path extension.
// It takes a file path as input.
// It returns the export format and an error if the extension is not supported.
func getExportFormat(filePath string) (ExportFormat, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".json":
		return FormatJSON, nil
	case ".yaml", ".yml":
		return FormatYAML, nil
	default:
		return "", fmt.Errorf("unsupported file extension: %s, must be .json, .yaml or .yml", ext)
	}
}

// DeduplicateAndSortImages deduplicates and sorts a map of images.
// It takes a map of strings to slices of images as input, where the keys are chart names and the values are slices of images.
// It returns a slice of images containing the deduplicated and sorted images.
func DeduplicateAndSortImages(imagesByChart map[string][]types.Image) []types.Image {
	uniqueImages := make(map[string]types.Image)

	// Flatten and deduplicate images
	for _, images := range imagesByChart {
		for _, image := range images {
			uniqueImages[image.Source] = image
		}
	}

	// Convert map to slice
	result := make([]types.Image, 0, len(uniqueImages))
	for _, image := range uniqueImages {
		result = append(result, image)
	}

	// Sort by source name
	sort.Slice(result, func(i, j int) bool {
		return result[i].Source < result[j].Source
	})

	return result
}

// WriteImagesToFilePerChart writes a list of images to a file for each chart.
// It takes a map of strings to slices of images as input, where the keys are chart names and the values are slices of images.
// It returns an error if writing the file fails.
func WriteImagesToFilePerChart(imageListByChart map[string][]types.Image, outputDir string) error {
	// 1. Ensure the output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
	}

	// Get the current time for the filename timestamp
	timestamp := time.Now().Format("20060102-150405")

	// 2. Iterate over the map, where key is chartName and value is the list of images
	for chartName, images := range imageListByChart {
		// 3. Marshal the list of images (the value) into JSON format
		// The output file should contain a JSON array of image objects.
		jsonData, err := json.MarshalIndent(images, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal images for chart %s: %w", chartName, err)
		}

		// 4. Construct the output filename
		// Format: sbom-images-for-chart-CHARTNAME-YYYYMMDD-HHMMSS.json
		filename := fmt.Sprintf("sbom-images-for-chart-%s-%s.json", chartName, timestamp)

		// 5. Construct the full file path
		filePath := filepath.Join(outputDir, filename)

		// 6. Write the JSON data to the file
		if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", filePath, err)
		}
	}

	return nil
}
