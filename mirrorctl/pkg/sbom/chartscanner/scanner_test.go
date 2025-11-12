package chartscanner

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestScanChart(t *testing.T) {
	testCases := []struct {
		name           string
		chartPath      string
		expectedImages []types.Image
	}{
		{
			name:      "redis chart",
			chartPath: "../../../resources/data_test/input_charts/redis",
			expectedImages: []types.Image{
				{Name: "redis", Source: "docker.io/bitnami/redis:7.0.5-debian-11-r15"},
				{Name: "redis-sentinel", Source: "docker.io/bitnami/redis-sentinel:7.0.5-debian-11-r14"},
				{Name: "redis-exporter", Source: "docker.io/bitnami/redis-exporter:1.45.0-debian-11-r1"},
				{Name: "bitnami-shell", Source: "docker.io/bitnami/bitnami-shell:11-debian-11-r48"},
			},
		},
		{
			name:      "loki chart",
			chartPath: "../../../resources/data_test/input_charts/loki",
			expectedImages: []types.Image{
				{Name: "agent-operator", Source: "docker.io/grafana/agent-operator:v0.25.1"},
				{Name: "kubectl", Source: "docker.io/bitnami/kubectl:latest"},
				{Name: "enterprise-logs-provisioner", Source: "docker.io/grafana/enterprise-logs-provisioner:latest"},
				{Name: "enterprise-logs", Source: "docker.io/grafana/enterprise-logs:latest"},
				{Name: "loki-canary", Source: "docker.io/grafana/loki-canary:latest"},
				{Name: "loki-helm-test", Source: "docker.io/grafana/loki-helm-test:latest"},
				{Name: "loki", Source: "docker.io/grafana/loki:latest"},
				{Name: "nginx-unprivileged", Source: "docker.io/nginxinc/nginx-unprivileged:1.19-alpine"},
				{Name: "mc", Source: "quay.io/minio/mc:RELEASE.2022-08-11T00-30-48Z"},
				{Name: "minio", Source: "quay.io/minio/minio:RELEASE.2022-08-13T21-54-44Z"},
			},
		},
		{
			name: "loki chart (expected)",
			// TODO fix mc - the root problem is that transform isn't changing it
			chartPath: "../../../resources/data_test/expected_charts/loki",
			expectedImages: []types.Image{
				{Name: "agent-operator", Source: "europe-southwest1-docker.pkg.dev/poc-development-123456/test-container-images/grafana/agent-operator:v0.25.1"},
				{Name: "kubectl", Source: "europe-southwest1-docker.pkg.dev/poc-development-123456/test-container-images/bitnami/kubectl:latest"},
				{Name: "enterprise-logs-provisioner", Source: "europe-southwest1-docker.pkg.dev/poc-development-123456/test-container-images/grafana/enterprise-logs-provisioner:latest"},
				{Name: "enterprise-logs", Source: "europe-southwest1-docker.pkg.dev/poc-development-123456/test-container-images/grafana/enterprise-logs:latest"},
				{Name: "loki-canary", Source: "europe-southwest1-docker.pkg.dev/poc-development-123456/test-container-images/grafana/loki-canary:latest"},
				{Name: "loki-helm-test", Source: "europe-southwest1-docker.pkg.dev/poc-development-123456/test-container-images/grafana/loki-helm-test:latest"},
				{Name: "loki", Source: "europe-southwest1-docker.pkg.dev/poc-development-123456/test-container-images/grafana/loki:latest"},
				{Name: "nginx-unprivileged", Source: "europe-southwest1-docker.pkg.dev/poc-development-123456/test-container-images/nginxinc/nginx-unprivileged:1.19-alpine"},
				{Name: "mc", Source: "quay.io/minio/mc:RELEASE.2022-08-11T00-30-48Z"},
				{Name: "minio", Source: "europe-southwest1-docker.pkg.dev/poc-development-123456/test-container-images/quay.io/minio/minio:RELEASE.2022-08-13T21-54-44Z"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Construct the absolute path to the test chart
			absChartPath, err := filepath.Abs(tc.chartPath)
			assert.NoError(t, err)

			images, err := ScanChart(absChartPath)
			assert.NoError(t, err)

			// Sort both actual and expected slices for consistent comparison
			sort.Slice(images, func(i, j int) bool {
				return strings.Compare(images[i].Source, images[j].Source) < 0
			})
			sort.Slice(tc.expectedImages, func(i, j int) bool {
				return strings.Compare(tc.expectedImages[i].Source, tc.expectedImages[j].Source) < 0
			})

			assert.Equal(t, tc.expectedImages, images)

			// TODO remove or put this as optional, it is to read it easier
			imgJson, err := json.MarshalIndent(images, "", "  ")
			fmt.Println(string(imgJson))
		})
	}
}
