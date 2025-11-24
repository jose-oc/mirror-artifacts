package chartscanner

import (
	"os"
	"reflect"
	"testing"
)

func TestRewriteImages(t *testing.T) {
	// Create temporary config file
	configContent := `
rewrites:
  - old: "/bitnami/"
    new: "/bitnamilegacy/"
  - old: "/bitnami-shell:"
    new: "/bitnami-shell-archived:"
`
	tmpfile, err := os.CreateTemp("", "rewrite_config_*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	if _, err := tmpfile.Write([]byte(configContent)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		images   []string
		expected []string
	}{
		{
			name: "Basic replacements",
			images: []string{
				"docker.io/bitnami/redis:7.0.5-debian-11-r15",
				"docker.io/bitnami/bitnami-shell:11-debian-11-r48",
				"docker.io/grafana/agent-operator:v0.25.1",
			},
			expected: []string{
				"docker.io/bitnamilegacy/redis:7.0.5-debian-11-r15",
				"docker.io/bitnamilegacy/bitnami-shell-archived:11-debian-11-r48",
				"docker.io/grafana/agent-operator:v0.25.1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RewriteImages(tt.images, tmpfile.Name())
			if err != nil {
				t.Fatalf("RewriteImages() error = %v", err)
			}

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("RewriteImages() = %v, want %v", got, tt.expected)
			}
		})
	}
}
