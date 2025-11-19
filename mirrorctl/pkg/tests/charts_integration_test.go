package tests

import (
	"path/filepath"
	"sort"
	"testing"

	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/appcontext"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/config"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/images"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/sbom/chartscanner"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/types"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// TestTransformedChartImagesMatchMirroredImages tests that the transformed chart images match the mirrored images.
func TestTransformedChartImagesMatchMirroredImages(t *testing.T) {
	// 1. Set up application context
	// Setup viper for the test
	viper.Set("skip_image_mirroring", false)
	viper.Set("dry_run", true)

	// Add other configuration settings from sample.mirrorctl.yaml
	viper.Set("gcp.project_id", "poc-development-123456")
	viper.Set("gcp.region", "europe-southwest1")
	viper.Set("gcp.gar_repo_charts", "europe-southwest1-docker.pkg.dev/poc-development-123456/test-helm-charts")
	viper.Set("gcp.gar_repo_containers", "europe-southwest1-docker.pkg.dev/poc-development-123456/test-container-images")
	viper.Set("options.suffix", "devopstest")
	viper.Set("options.keep_temp_dir", false)
	viper.Set("options.notify_tag_mutations", true)
	viper.Set("prod-mode", false)
	viper.Set("log-color", true)
	viper.Set("log-level", "debug")
	viper.Set("log-file", "")

	appCtx := &appcontext.AppContext{
		DryRun: viper.GetBool("dry_run"),
		Config: &config.Config{
			GCP: config.GCPConfig{
				ProjectID:         viper.GetString("gcp.project_id"),
				Region:            viper.GetString("gcp.region"),
				GARRepoCharts:     viper.GetString("gcp.gar_repo_charts"),
				GARRepoContainers: viper.GetString("gcp.gar_repo_containers"),
			},
			Options: config.OptionsConfig{
				Suffix:             viper.GetString("options.suffix"),
				KeepTempDir:        viper.GetBool("options.keep_temp_dir"),
				NotifyTagMutations: viper.GetBool("options.notify_tag_mutations"),
			},
		},
	}

	// Define chart paths
	// TODO this test is checking only telegraf chart, but it should be extended to check all charts
	inputChartPath := filepath.Join("..", "..", "resources", "data_test", "input_charts", "telegraf")
	transformedChartPath := filepath.Join("..", "..", "resources", "data_test", "expected_charts", "telegraf")

	// Ensure paths are absolute for chartscanner
	absInputChartPath, err := filepath.Abs(inputChartPath)
	assert.NoError(t, err)
	absTransformedChartPath, err := filepath.Abs(transformedChartPath)
	assert.NoError(t, err)

	// 2. Get Source Images from the input chart
	sourceImages, err := chartscanner.ScanChart(absInputChartPath)
	assert.NoError(t, err)
	assert.NotEmpty(t, sourceImages, "No source images found in the input chart")

	// Convert chartscanner.Image to types.Image for images.MirrorImages
	var sourceImagesForMirroring types.ImagesList
	sourceImagesForMirroring.Images = sourceImages

	// 3. Get Mirrored (Target) Images
	imagesPushed, _, err := images.MirrorImages(appCtx, sourceImagesForMirroring)
	assert.NoError(t, err)
	assert.NotEmpty(t, imagesPushed, "No images to be mirrored")

	var mirroredImagePaths []string
	for _, img := range imagesPushed {
		mirroredImagePaths = append(mirroredImagePaths, img)
	}
	sort.Strings(mirroredImagePaths)

	// 4. Get Transformed (Actual) Images from the expected transformed chart
	transformedImages, err := chartscanner.ScanChart(absTransformedChartPath)
	assert.NoError(t, err)
	assert.NotEmpty(t, transformedImages, "No images found in the transformed chart")

	var imagePathsInTransformedChart []string
	for _, img := range transformedImages {
		imagePathsInTransformedChart = append(imagePathsInTransformedChart, img.Source)
	}
	sort.Strings(imagePathsInTransformedChart)

	// 5. Compare and Assert
	if !assert.Equal(t, mirroredImagePaths, imagePathsInTransformedChart, "Mirrored image paths should match transformed chart image paths") {
		// Print diff in a clear format for long lists. Show one element per line
		// Call the helper function to print the difference
		printSliceDiff(t, mirroredImagePaths, imagePathsInTransformedChart)
	}
}

// Helper function to print the difference between two string slices
func printSliceDiff(t *testing.T, expected, actual []string) {
	// Use a map to track elements in the actual slice for quick lookup
	actualSet := make(map[string]struct{})
	for _, path := range actual {
		actualSet[path] = struct{}{}
	}

	// 1. Find elements that are MISSING in the actual slice (Expected but not found)
	t.Logf("--- Missing from actualTransformedImagePaths (Expected, but NOT Found) ---")
	missingCount := 0
	for _, path := range expected {
		if _, exists := actualSet[path]; !exists {
			t.Logf("- %s", path)
			missingCount++
		}
	}
	if missingCount == 0 {
		t.Logf("(None)")
	}
	t.Logf("-------------------------------------------------------------------------")

	// Use a map to track elements in the expected slice for quick lookup
	expectedSet := make(map[string]struct{})
	for _, path := range expected {
		expectedSet[path] = struct{}{}
	}

	// 2. Find elements that are EXTRA in the actual slice (Found, but NOT Expected)
	t.Logf("--- Extra in actualTransformedImagePaths (Found, but NOT Expected) ---")
	extraCount := 0
	for _, path := range actual {
		if _, exists := expectedSet[path]; !exists {
			t.Logf("+ %s", path)
			extraCount++
		}
	}
	if extraCount == 0 {
		t.Logf("(None)")
	}
	t.Logf("----------------------------------------------------------------------")
}
