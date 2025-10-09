package images

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jose-oc/poc-mirror-artifacts/jocmirrorctl/pkg/appcontext"
	"github.com/jose-oc/poc-mirror-artifacts/jocmirrorctl/pkg/config"
	"github.com/jose-oc/poc-mirror-artifacts/jocmirrorctl/pkg/logging"
	"github.com/rs/zerolog"
)

func TestParseImagesYAML(t *testing.T) {
	// Create a temporary images.yaml
	imagesContent := `
images:
  - name: nginx
    source: docker.io/bitnami/nginx:1.25.3
  - name: redis
    source: docker.io/bitnami/redis:7.2.0
`
	tmpDir := t.TempDir()
	imagesFile := filepath.Join(tmpDir, "images.yaml")
	err := os.WriteFile(imagesFile, []byte(imagesContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp images file: %v", err)
	}

	logger := logging.NewLogger()
	cfg := &config.Config{
		GCP: config.GCPConfig{
			GARRepoContainers: "europe-southwest1-docker.pkg.dev/my-gcp-project/test-container-images",
		},
	}
	ctx := appcontext.NewAppContext(logger, cfg, false)

	err = MirrorImages(ctx, imagesFile)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestDryRunMode(t *testing.T) {
	imagesContent := `
images:
  - name: nginx
    source: docker.io/bitnami/nginx:1.25.3
`
	tmpDir := t.TempDir()
	imagesFile := filepath.Join(tmpDir, "images.yaml")
	err := os.WriteFile(imagesFile, []byte(imagesContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp images file: %v", err)
	}

	var buf bytes.Buffer
	logger := zerolog.New(&buf).With().Timestamp().Logger()
	cfg := &config.Config{
		GCP: config.GCPConfig{
			GARRepoContainers: "europe-southwest1-docker.pkg.dev/my-gcp-project/test-container-images",
		},
	}
	ctx := appcontext.NewAppContext(&logger, cfg, true)

	err = MirrorImages(ctx, imagesFile)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !strings.Contains(buf.String(), "Dry-run: Would mirror image to GAR") {
		t.Errorf("Expected dry-run log, got '%s'", buf.String())
	}
	if !strings.Contains(buf.String(), "oras cp docker.io/bitnami/nginx:1.25.3 europe-southwest1-docker.pkg.dev/my-gcp-project/test-container-images/nginx") {
		t.Errorf("Expected oras cp command in log, got '%s'", buf.String())
	}
}
