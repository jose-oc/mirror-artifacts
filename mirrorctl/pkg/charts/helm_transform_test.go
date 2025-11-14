package charts

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/appcontext"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/config"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/types"
	"github.com/rs/zerolog/log"
)

var ctx = appcontext.AppContext{
	Config: &config.Config{
		GCP: config.GCPConfig{
			ProjectID:         "poc-development-123456",
			Region:            "europe-southwest1",
			GARRepoCharts:     "europe-southwest1-docker.pkg.dev/poc-development-123456/test-helm-charts",
			GARRepoContainers: "europe-southwest1-docker.pkg.dev/poc-development-123456/test-container-images",
		},
		Options: config.OptionsConfig{
			Suffix:             "poc",
			NotifyTagMutations: false,
			KeepTempDir:        false,
		},
	},
	DryRun: false,
}

func runMirrorChartTest(t *testing.T, inputDir, expectedDir, outputDir string, chart types.Chart) {
	// Clean up output directory before test
	os.RemoveAll(outputDir)
	defer os.RemoveAll(outputDir) // Clean up after test

	// Execute the mirroring
	transformHelmChartDir, err := TransformHelmChart(&ctx, chart, inputDir, outputDir)
	if err != nil {
		t.Fatalf("TransformHelmChart failed: %v", err)
	}

	// Verify the output matches the expected output
	err = compareDirectories(expectedDir, transformHelmChartDir)
	if err != nil {
		t.Fatalf("Output does not match expected: %v", err)
	}
}

func TestMirrorCharts(t *testing.T) {
	tests := []struct {
		name        string
		inputDir    string
		expectedDir string
		outputDir   string
		chart       types.Chart
	}{
		{
			name:        "GrafanaAgentOperatorChart",
			inputDir:    "../../resources/data_test/input_charts/grafana-agent-operator",
			expectedDir: "../../resources/data_test/expected_charts/grafana-agent-operator",
			outputDir:   "../../resources/data_test/output_charts/grafana-agent-operator",
			chart: types.Chart{
				Name:    "grafana-agent-operator",
				Source:  "https://grafana.github.io/helm-charts",
				Version: "0.5.1",
			},
		},
		{
			name:        "LokiChart",
			inputDir:    "../../resources/data_test/input_charts/loki",
			expectedDir: "../../resources/data_test/expected_charts/loki",
			outputDir:   "../../resources/data_test/output_charts/loki",
			chart: types.Chart{
				Name:    "loki",
				Source:  "https://grafana.github.io/helm-charts",
				Version: "5.5.2",
			},
		},
		{
			name:        "MariadbChart",
			inputDir:    "../../resources/data_test/input_charts/mariadb",
			expectedDir: "../../resources/data_test/expected_charts/mariadb",
			outputDir:   "../../resources/data_test/output_charts/mariadb",
			chart: types.Chart{
				Name:    "mariadb",
				Source:  "https://charts.bitnami.com/bitnami",
				Version: "12.2.4",
			},
		},
		{
			name:        "RabbitChart",
			inputDir:    "../../resources/data_test/input_charts/rabbitmq",
			expectedDir: "../../resources/data_test/expected_charts/rabbitmq",
			outputDir:   "../../resources/data_test/output_charts/rabbitmq",
			chart: types.Chart{
				Name:    "rabbitmq",
				Source:  "https://charts.bitnami.com/bitnami",
				Version: "11.1.5",
			},
		},
		{
			name:        "RedisChart",
			inputDir:    "../../resources/data_test/input_charts/redis",
			expectedDir: "../../resources/data_test/expected_charts/redis",
			outputDir:   "../../resources/data_test/output_charts/redis",
			chart: types.Chart{
				Name:    "redis",
				Source:  "https://charts.bitnami.com/bitnami",
				Version: "17.3.11",
			},
		},
		{
			name:        "TelegrafChart",
			inputDir:    "../../resources/data_test/input_charts/telegraf",
			expectedDir: "../../resources/data_test/expected_charts/telegraf",
			outputDir:   "../../resources/data_test/output_charts/telegraf",
			chart: types.Chart{
				Name:    "telegraf",
				Source:  "https://helm.influxdata.com/",
				Version: "1.8.28",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runMirrorChartTest(t, tt.inputDir, tt.expectedDir, tt.outputDir, tt.chart)
		})
	}
}

func compareDirectories(expectedDir, actualDir string) error {
	// Walk through the expected directory
	return filepath.Walk(expectedDir, func(expectedPath string, expectedInfo fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(expectedDir, expectedPath)
		if err != nil {
			return err
		}

		// Calculate actual path
		actualPath := filepath.Join(actualDir, relPath)

		// Check if file/directory exists in actual output
		actualInfo, err := os.Stat(actualPath)
		if err != nil {
			return err
		}

		// Compare file types
		if expectedInfo.IsDir() != actualInfo.IsDir() {
			return os.ErrInvalid
		}

		// If it's a file, compare contents
		if !expectedInfo.IsDir() {
			// For Chart.yaml, we need special handling due to timestamp
			if filepath.Base(expectedPath) == "Chart.yaml" {
				return compareChartYaml(expectedPath, actualPath)
			}

			expectedContent, err := os.ReadFile(expectedPath)
			if err != nil {
				return err
			}

			actualContent, err := os.ReadFile(actualPath)
			if err != nil {
				return err
			}

			if string(expectedContent) != string(actualContent) {
				// log.Printf("Err - Contents differ\nExpectedContent: %s\nActualContent: %s\n", string(expectedContent), string(actualContent))
				log.Error().
					Str("actualPath", string(actualPath)).
					Str("actualContent", string(actualContent)).
					Str("expectedPath", string(expectedPath)).
					Str("expectedContent", string(expectedContent)).
					Msg("Contents differ")
				return os.ErrInvalid
			}
		}

		return nil
	})
}

func compareChartYaml(expectedPath, actualPath string) error {
	expectedContent, err := os.ReadFile(expectedPath)
	if err != nil {
		return err
	}

	actualContent, err := os.ReadFile(actualPath)
	if err != nil {
		return err
	}

	// Normalize timestamps before comparison (since they will be different on each run)
	timestampRegex := regexp.MustCompile(`repackage\.provenance/timestamp: "[^"]*"`)
	expectedNormalized := timestampRegex.ReplaceAllString(string(expectedContent), `repackage.provenance/timestamp: "TIMESTAMP"`)
	actualNormalized := timestampRegex.ReplaceAllString(string(actualContent), `repackage.provenance/timestamp: "TIMESTAMP"`)

	if strings.TrimSpace(expectedNormalized) != strings.TrimSpace(actualNormalized) {
		//log.Printf("Err - Chart.yaml contents differ\nExpectedContent: %s\nActualContent: %s\n", string(expectedContent), string(actualContent))
		log.Warn().Str("actualContent", string(actualContent)).Str("expectedContent", string(expectedContent)).Msg("Chart.yaml contents differ")
		return os.ErrInvalid
	}

	return nil
}

func TestProcessChartYaml(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		suffix           string
		originalChartURL string
		checkResult      func(t *testing.T, output string)
	}{
		{
			name: "chart with existing annotations",
			input: `annotations:
  category: Database
  licenses: Apache-2.0
apiVersion: v2
version: 1.2.3
name: test-chart`,
			suffix:           "build.1",
			originalChartURL: "https://charts.bitnami.com/bitnami",
			checkResult: func(t *testing.T, output string) {
				if !strings.Contains(output, "version: 1.2.3-build.1") {
					t.Errorf("Version was not updated correctly")
				}
				checkProvenanceAnnotations(t, output, "test-chart", "1.2.3")
			},
		},
		{
			name: "chart without annotations (Loki style)",
			input: `apiVersion: v2
name: loki
version: 5.0.0
description: Loki is a horizontally scalable log aggregation system`,
			suffix:           "poc",
			originalChartURL: "https://charts.bitnami.com/bitnami",
			checkResult: func(t *testing.T, output string) {
				if !strings.Contains(output, "version: 5.0.0-poc") {
					t.Errorf("Version was not updated correctly")
				}
				if !strings.Contains(output, "annotations:") {
					t.Errorf("Annotations section was not created")
				}
				checkProvenanceAnnotations(t, output, "loki", "5.0.0")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			srcPath := filepath.Join(tmpDir, "Chart.yaml")
			destPath := filepath.Join(tmpDir, "Chart_modified.yaml")

			err := os.WriteFile(srcPath, []byte(tt.input), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			err = processChartYAML(srcPath, destPath, tt.suffix, tt.originalChartURL)
			if err != nil {
				t.Fatalf("processChartYaml failed: %v", err)
			}

			modifiedContent, err := os.ReadFile(destPath)
			if err != nil {
				t.Fatalf("Failed to read modified file: %v", err)
			}

			tt.checkResult(t, string(modifiedContent))
		})
	}
}

func checkProvenanceAnnotations(t *testing.T, content, expectedName, expectedVersion string) {
	provenanceChecks := []string{
		"# --- Provenance Metadata ---",
		"repackage.provenance/repackaged-by:",
		"repackage.provenance/original-chart-url:",
		"repackage.provenance/original-chart-name:",
		"repackage.provenance/original-chart-version:",
		"repackage.provenance/timestamp:",
	}

	for _, check := range provenanceChecks {
		if !strings.Contains(content, check) {
			t.Errorf("Missing provenance annotation: %s", check)
		}
	}

	if !strings.Contains(content, fmt.Sprintf(`repackage.provenance/original-chart-name: "%s"`, expectedName)) {
		t.Errorf("Original chart name not preserved correctly, expected %s", expectedName)
	}
	if !strings.Contains(content, fmt.Sprintf(`repackage.provenance/original-chart-version: "%s"`, expectedVersion)) {
		t.Errorf("Original chart version not preserved correctly, expected %s", expectedVersion)
	}
}

func TestProcessValuesYaml(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		registryURL string
		expected    string
	}{
		{
			name: "chart with global.imageRegistry",
			input: `global:
  imageRegistry: "old-registry.com"
  imagePullSecrets: []`,
			registryURL: "my-new-registry.com",
			expected: `global:
  imageRegistry: "my-new-registry.com"
  imagePullSecrets: []`,
		},
		{
			name: "chart with global.image.registry",
			input: `global:
  image:
    registry: "old-loki-registry.com"
    repository: loki/loki
    tag: "2.9.0"
  imagePullSecrets: []`,
			registryURL: "my-new-loki-registry.com",
			expected: `global:
  image:
    registry: "my-new-loki-registry.com"
    repository: loki/loki
    tag: "2.9.0"
  imagePullSecrets: []`,
		},
		{
			name: "chart with top-level image.repo and no global registry",
			input: `image:
  repo: "telegraf"
  tag: "1.26-alpine"
  pullPolicy: IfNotPresent`,
			registryURL: "europe-southwest1-docker.pkg.dev/poc-development-123456/test-helm-charts",
			expected: `image:
  repo: "europe-southwest1-docker.pkg.dev/poc-development-123456/test-helm-charts/telegraf"
  tag: "1.26-alpine"
  pullPolicy: IfNotPresent`,
		},
		{
			name: "chart with top-level image.repository and no global registry",
			input: `image:
  repository: "influxdb"
  tag: "1.8.10-alpine"
  pullPolicy: IfNotPresent`,
			registryURL: "europe-southwest1-docker.pkg.dev/poc-development-123456/test-helm-charts",
			expected: `image:
  repository: "europe-southwest1-docker.pkg.dev/poc-development-123456/test-helm-charts/influxdb"
  tag: "1.8.10-alpine"
  pullPolicy: IfNotPresent`,
		},
		{
			name: "chart with top-level image.registry and image.repository (grafana-agent-operator style)",
			input: `image:
  # -- Image registry
  registry: docker.io
  # -- Image repo
  repository: grafana/agent-operator
  tag: v0.44.2`,
			registryURL: "europe-southwest1-docker.pkg.dev/poc-development-123456/test-helm-charts",
			expected: `image:
  # -- Image registry
  registry: "europe-southwest1-docker.pkg.dev/poc-development-123456/test-helm-charts"
  # -- Image repo
  repository: grafana/agent-operator
  tag: v0.44.2`,
		},
		{
			name: "chart with nested test.image.registry (grafana-agent-operator nested style)",
			input: `test:
  image:
    # -- Test image registry
    registry: docker.io
    # -- Test image repo
    repository: library/busybox
    tag: latest`,
			registryURL: "europe-southwest1-docker.pkg.dev/poc-development-123456/test-helm-charts",
			expected: `test:
  image:
    # -- Test image registry
    registry: "europe-southwest1-docker.pkg.dev/poc-development-123456/test-helm-charts"
    # -- Test image repo
    repository: library/busybox
    tag: latest`,
		},
		{
			name: "chart with top-level prefixed image (minio nested style)",
			input: `image:
  repository: quay.io/minio/minio
  tag: RELEASE.2022-08-13T21-54-44Z
  pullPolicy: IfNotPresent

imagePullSecrets: []
# - name: "image-pull-secret"

mcImage:
  repository: quay.io/minio/mc
  tag: RELEASE.2022-08-11T00-30-48Z
  pullPolicy: IfNotPresent`,
			registryURL: "europe-southwest1-docker.pkg.dev/poc-development-123456/test-helm-charts",
			expected: `image:
  repository: "europe-southwest1-docker.pkg.dev/poc-development-123456/test-helm-charts/quay.io/minio/minio"
  tag: RELEASE.2022-08-13T21-54-44Z
  pullPolicy: IfNotPresent

imagePullSecrets: []
# - name: "image-pull-secret"

mcImage:
  repository: "europe-southwest1-docker.pkg.dev/poc-development-123456/test-helm-charts/quay.io/minio/mc"
  tag: RELEASE.2022-08-11T00-30-48Z
  pullPolicy: IfNotPresent`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file for testing
			tmpDir := t.TempDir()
			srcPath := filepath.Join(tmpDir, "values.yaml")
			destPath := filepath.Join(tmpDir, "values_modified.yaml")

			// Write test content
			err := os.WriteFile(srcPath, []byte(tt.input), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Process the file
			err = processValuesYAML(srcPath, destPath, tt.registryURL)
			if err != nil {
				t.Fatalf("processValuesYaml failed: %v", err)
			}

			// Read the modified content
			modifiedContent, err := os.ReadFile(destPath)
			if err != nil {
				t.Fatalf("Failed to read modified file: %v", err)
			}

			if strings.TrimSpace(string(modifiedContent)) != tt.expected {
				t.Errorf("Expected:\n%s\n\nGot:\n%s", tt.expected, string(modifiedContent))
			}
		})
	}
}

func TestReplaceVersion(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		versionSuffix string
		expected      string
	}{
		{
			name:          "simple version",
			input:         "version: 1.2.3\nname: test",
			versionSuffix: "build.1",
			expected:      "version: 1.2.3-build.1\nname: test",
		},
		{
			name:          "version with spaces",
			input:         "version:    1.2.3\nname: test",
			versionSuffix: "poc",
			expected:      "version: 1.2.3-poc\nname: test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := replaceVersion(tt.input, tt.versionSuffix)
			if result != tt.expected {
				t.Errorf("Expected:\n%s\n\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestExtractVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple version",
			input:    "version: 1.2.3\nname: test",
			expected: "1.2.3",
		},
		{
			name:     "version with spaces",
			input:    "version:    1.2.3   \nname: test",
			expected: "1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractVersion(tt.input)
			if result != tt.expected {
				t.Errorf("Expected: %s, Got: %s", tt.expected, result)
			}
		})
	}
}

func TestExtractChartName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple name",
			input:    "name: mariadb\nversion: 1.2.3",
			expected: "mariadb",
		},
		{
			name:     "name with spaces",
			input:    "name:    test-chart   \nversion: 1.2.3",
			expected: "test-chart",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractChartName(tt.input)
			if result != tt.expected {
				t.Errorf("Expected: %s, Got: %s", tt.expected, result)
			}
		})
	}
}
