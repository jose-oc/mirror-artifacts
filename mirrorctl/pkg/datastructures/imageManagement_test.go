package datastructures

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/types"
)

func TestDeduplicateAndSortImages(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string][]types.Image
		expected []types.Image
	}{
		{
			name:     "empty input",
			input:    map[string][]types.Image{},
			expected: []types.Image{},
		},
		{
			name: "single chart with duplicates",
			input: map[string][]types.Image{
				"chart1": {
					{Source: "nginx:latest"},
					{Source: "redis:alpine"},
					{Source: "nginx:latest"}, // duplicate
				},
			},
			expected: []types.Image{
				{Source: "nginx:latest"},
				{Source: "redis:alpine"},
			},
		},
		{
			name: "multiple charts with duplicates across charts",
			input: map[string][]types.Image{
				"chart1": {
					{Source: "nginx:latest"},
					{Source: "redis:alpine"},
				},
				"chart2": {
					{Source: "postgres:13"},
					{Source: "nginx:latest"}, // duplicate across charts
				},
			},
			expected: []types.Image{
				{Source: "nginx:latest"},
				{Source: "postgres:13"},
				{Source: "redis:alpine"},
			},
		},
		{
			name: "already sorted images",
			input: map[string][]types.Image{
				"chart1": {
					{Source: "alpine:latest"},
					{Source: "busybox:latest"},
					{Source: "centos:latest"},
				},
			},
			expected: []types.Image{
				{Source: "alpine:latest"},
				{Source: "busybox:latest"},
				{Source: "centos:latest"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DeduplicateAndSortImages(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d images, got %d", len(tt.expected), len(result))
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}

			// Verify the result is sorted
			for i := 1; i < len(result); i++ {
				if result[i].Source < result[i-1].Source {
					t.Errorf("result is not sorted: %s should come after %s", result[i].Source, result[i-1].Source)
				}
			}
		})
	}
}

func TestWriteImagesToFile(t *testing.T) {
	tests := []struct {
		name        string
		images      []types.Image
		filename    string
		expectError bool
	}{
		{
			name: "write JSON file",
			images: []types.Image{
				{Source: "nginx:latest"},
				{Source: "redis:alpine"},
			},
			filename:    "test_output.json",
			expectError: false,
		},
		{
			name: "write YAML file",
			images: []types.Image{
				{Source: "nginx:latest"},
			},
			filename:    "test_output.yaml",
			expectError: false,
		},
		{
			name: "unsupported format",
			images: []types.Image{
				{Source: "nginx:latest"},
			},
			filename:    "test_output.txt",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			outputPath := filepath.Join(tempDir, tt.filename)

			err := WriteImagesToFile(tt.images, outputPath)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify file exists
			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				t.Errorf("output file was not created: %s", outputPath)
			}

			// Verify file content based on format
			content, err := os.ReadFile(outputPath)
			if err != nil {
				t.Errorf("failed to read output file: %v", err)
			}

			if strings.HasSuffix(tt.filename, ".json") {
				if !strings.Contains(string(content), `"images"`) {
					t.Error("JSON output doesn't contain expected structure")
				}
			} else if strings.HasSuffix(tt.filename, ".yaml") {
				if !strings.Contains(string(content), "images:") {
					t.Error("YAML output doesn't contain expected structure")
				}
			}

			// Cleanup
			os.Remove(outputPath)
		})
	}
}

func TestGetExportFormat(t *testing.T) {
	tests := []struct {
		path     string
		expected ExportFormat
		hasError bool
	}{
		{"/path/to/file.json", FormatJSON, false},
		{"/path/to/file.yaml", FormatYAML, false},
		{"/path/to/file.yml", FormatYAML, false},
		{"/path/to/file.JSON", FormatJSON, false},
		{"/path/to/file.YAML", FormatYAML, false},
		{"/path/to/file.txt", "", true},
		{"/path/to/file", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result, err := getExportFormat(tt.path)

			if tt.hasError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func BenchmarkDeduplicateAndSortImages(b *testing.B) {
	// Create test data with some duplicates
	imagesByChart := map[string][]types.Image{
		"chart1": makeTestImages(1000),
		"chart2": makeTestImages(1000), // Will have duplicates with chart1
		"chart3": makeTestImages(500),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DeduplicateAndSortImages(imagesByChart)
	}
}

// makeTestImages generates a slice of test images
func makeTestImages(count int) []types.Image {
	images := make([]types.Image, count)
	for i := 0; i < count; i++ {
		images[i] = types.Image{Source: fmt.Sprintf("image%d:tag%d", i%100, i)} // Some duplicates
	}
	return images
}
