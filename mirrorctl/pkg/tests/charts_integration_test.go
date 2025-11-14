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
	"github.com/stretchr/testify/assert"
)

// TestTransformedChartImagesMatchMirroredImages tests that the transformed chart images match the mirrored images.
func TestTransformedChartImagesMatchMirroredImages(t *testing.T) {
	// 1. Set up application context
	appCtx := appcontext.AppContext{
		DryRun: true,
		Config: &config.Config{},
	}

	// Hardcode configuration values for the test
	appCtx.Config.GCP.GARRepoContainers = "europe-southwest1-docker.pkg.dev/poc-development-123456/test-container-images"
	appCtx.Config.GCP.ProjectID = "poc-development-123456"

	// Define chart paths
	inputChartPath := filepath.Join("..", "..", "resources", "data_test", "input_charts", "loki")
	transformedChartPath := filepath.Join("..", "..", "resources", "data_test", "expected_charts", "loki")

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
	imagesPushed, _, err := images.MirrorImages(&appCtx, sourceImagesForMirroring)
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

	var actualTransformedImagePaths []string
	for _, img := range transformedImages {
		actualTransformedImagePaths = append(actualTransformedImagePaths, img.Source)
	}
	sort.Strings(actualTransformedImagePaths)

	// 5. Compare and Assert
	assert.Equal(t, mirroredImagePaths, actualTransformedImagePaths, "Mirrored image paths should match transformed chart image paths")
}
