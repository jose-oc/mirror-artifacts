package images

import (
	"os"
	"testing"

	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/appcontext"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestMirrorImages_NoImagesFile(t *testing.T) {
	appCtx := &appcontext.AppContext{}
	_, _, err := MirrorImagesFromFile(appCtx, "")
	assert.Error(t, err)
}

func TestMirrorImages_ImagesFileNotFound(t *testing.T) {
	appCtx := &appcontext.AppContext{}
	_, _, err := MirrorImagesFromFile(appCtx, "non-existent-file.yaml")
	assert.Error(t, err)
}

func TestMirrorImages_InvalidYAML(t *testing.T) {
	appCtx := &appcontext.AppContext{}
	file, err := os.CreateTemp(t.TempDir(), "images.yaml")
	assert.NoError(t, err)
	_, err = file.WriteString("invalid yaml")
	assert.NoError(t, err)
	file.Close()

	_, _, err = MirrorImagesFromFile(appCtx, file.Name())
	assert.Error(t, err)
}

func TestMirrorImages_DryRun(t *testing.T) {
	appCtx := &appcontext.AppContext{
		DryRun: true,
		Config: &config.Config{
			GCP: config.GCPConfig{
				GARRepoContainers: "us-central1-docker.pkg.dev/my-project/my-repo",
			},
		},
	}

	file, err := os.CreateTemp(t.TempDir(), "images.yaml")
	assert.NoError(t, err)
	_, err = file.WriteString(`
images:
  - name: ubuntu
    source: ubuntu:22.04
`)
	assert.NoError(t, err)
	file.Close()

	mirrored, failed, err := MirrorImagesFromFile(appCtx, file.Name())
	assert.NoError(t, err)
	assert.Equal(t, "us-central1-docker.pkg.dev/my-project/my-repo/ubuntu:22.04", mirrored["ubuntu:22.04"])
	assert.Equal(t, 0, len(failed))
}

func TestMirrorImages_GetTagError(t *testing.T) {
	appCtx := &appcontext.AppContext{
		DryRun: true,
		Config: &config.Config{
			GCP: config.GCPConfig{
				GARRepoContainers: "us-central1-docker.pkg.dev/my-project/my-repo",
			},
		},
	}

	file, err := os.CreateTemp(t.TempDir(), "images.yaml")
	assert.NoError(t, err)
	_, err = file.WriteString(`
images:
  - name: ubuntu
    source: ubuntu
`)
	assert.NoError(t, err)
	file.Close()

	_, _, err = MirrorImagesFromFile(appCtx, file.Name())
	assert.NoError(t, err) // The function itself doesn't return an error, it logs it
}
