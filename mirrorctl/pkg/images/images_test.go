package images

import (
	"testing"

	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/appcontext"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/config"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/types"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestMirrorImages_DryRun(t *testing.T) {
	appCtx := &appcontext.AppContext{
		DryRun: true,
		Config: &config.Config{
			GCP: config.GCPConfig{
				GARRepoContainers: "us-central1-docker.pkg.dev/my-project/my-repo",
			},
		},
	}

	yamlContent := `
images:
  - name: myfolder/myubuntuimage
    source: sourcefolder/ubuntu:22.04
`
	var imagesList types.ImagesList
	err := yaml.Unmarshal([]byte(yamlContent), &imagesList)
	assert.NoError(t, err)

	mirrored, failed, err := MirrorImages(appCtx, imagesList)
	assert.NoError(t, err)
	assert.Equal(t, "us-central1-docker.pkg.dev/my-project/my-repo/myfolder/myubuntuimage:22.04", mirrored["sourcefolder/ubuntu:22.04"])
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

	yamlContent := `
images:
  - name: ubuntu
    source: ubuntu
`
	var imagesList types.ImagesList
	err := yaml.Unmarshal([]byte(yamlContent), &imagesList)
	assert.NoError(t, err)

	_, _, err = MirrorImages(appCtx, imagesList)
	assert.NoError(t, err) // The function itself doesn't return an error, it logs it
}
