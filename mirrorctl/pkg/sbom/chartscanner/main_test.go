package chartscanner

import (
	"os"
	"sort"
	"testing"

	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/appcontext"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/config"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestExtractImagesFromCharts(t *testing.T) {
	// Create a temporary charts.yaml file
	chartsYAML := `
charts:
  - name: grafana-agent-operator
    source: https://grafana.github.io/helm-charts
    version: 0.5.1
  - name: loki
    source: https://grafana.github.io/helm-charts
    version: 5.5.2
  - name: mariadb
    source: oci://registry-1.docker.io/bitnamicharts
    version: 12.2.4
`
	tmpfile, err := os.CreateTemp("", "charts-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name()) // clean up

	_, err = tmpfile.WriteString(chartsYAML)
	assert.NoError(t, err)
	err = tmpfile.Close()
	assert.NoError(t, err)

	// Create a dummy AppContext
	ctx := &appcontext.AppContext{
		Config: &config.Config{
			Options: config.OptionsConfig{
				KeepTempDir: false,
			},
		},
	}

	expectedImagesByChart := map[string][]types.Image{
		"grafana-agent-operator": {
			{Name: "agent-operator", Source: "docker.io/grafana/agent-operator:v0.44.2"},
			{Name: "busybox", Source: "docker.io/library/busybox:latest"},
		},
		"loki": {
			{Name: "kubectl", Source: "docker.io/bitnami/kubectl:latest"},
			{Name: "agent-operator", Source: "docker.io/grafana/agent-operator:v0.25.1"},
			{Name: "enterprise-logs-provisioner", Source: "docker.io/grafana/enterprise-logs-provisioner:latest"},
			{Name: "enterprise-logs", Source: "docker.io/grafana/enterprise-logs:latest"},
			{Name: "loki-canary", Source: "docker.io/grafana/loki-canary:latest"},
			{Name: "loki-helm-test", Source: "docker.io/grafana/loki-helm-test:latest"},
			{Name: "loki", Source: "docker.io/grafana/loki:latest"},
			{Name: "nginx-unprivileged", Source: "docker.io/nginxinc/nginx-unprivileged:1.19-alpine"},
			{Name: "mc", Source: "quay.io/minio/mc:RELEASE.2022-08-11T00-30-48Z"},
			{Name: "minio", Source: "quay.io/minio/minio:RELEASE.2022-08-13T21-54-44Z"},
		},
		"mariadb": {
			{Name: "bitnami-shell", Source: "docker.io/bitnami/bitnami-shell:11-debian-11-r118"},
			{Name: "mariadb", Source: "docker.io/bitnami/mariadb:10.11.3-debian-11-r5"},
			{Name: "mysqld-exporter", Source: "docker.io/bitnami/mysqld-exporter:0.14.0-debian-11-r119"},
		},
	}

	imagesByChart, err := ExtractImagesFromCharts(ctx, tmpfile.Name())
	assert.NoError(t, err)

	// Sort images within each chart for consistent comparison
	for _, images := range expectedImagesByChart {
		sort.Slice(images, func(i, j int) bool {
			return images[i].Source < images[j].Source
		})
	}

	assert.Equal(t, expectedImagesByChart, imagesByChart)
}
