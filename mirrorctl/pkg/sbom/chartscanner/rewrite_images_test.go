package chartscanner

import (
	"reflect"
	"testing"

	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/config"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/types"
)

func TestRewriteImages(t *testing.T) {
	imageRewriteRules := []config.ImageRewrite{
		{Old: "/bitnami/", New: "/bitnamilegacy/"},
		{Old: "/bitnami-shell:", New: "/bitnami-shell-archived:"},
	}

	tests := []struct {
		name              string
		imageRewriteRules []config.ImageRewrite
		images            []types.Image
		expected          []types.Image
	}{
		{
			name:              "No replacement rules",
			imageRewriteRules: nil,
			images: []types.Image{
				{Source: "docker.io/bitnami/redis:7.0.5-debian-11-r15"},
				{Source: "docker.io/bitnami/bitnami-shell:11-debian-11-r48"},
				{Source: "docker.io/grafana/agent-operator:v0.25.1"},
			},
			expected: []types.Image{
				{Source: "docker.io/bitnami/redis:7.0.5-debian-11-r15"},
				{Source: "docker.io/bitnami/bitnami-shell:11-debian-11-r48"},
				{Source: "docker.io/grafana/agent-operator:v0.25.1"},
			},
		},
		{
			name:              "Basic replacements",
			imageRewriteRules: imageRewriteRules,
			images: []types.Image{
				{Source: "docker.io/bitnami/redis:7.0.5-debian-11-r15"},
				{Source: "docker.io/bitnami/bitnami-shell:11-debian-11-r48"},
				{Source: "docker.io/grafana/agent-operator:v0.25.1"},
			},
			expected: []types.Image{
				{Source: "docker.io/bitnamilegacy/redis:7.0.5-debian-11-r15"},
				{Source: "docker.io/bitnamilegacy/bitnami-shell-archived:11-debian-11-r48"},
				{Source: "docker.io/grafana/agent-operator:v0.25.1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RewriteImages(tt.images, tt.imageRewriteRules)
			if err != nil {
				t.Fatalf("RewriteImages() error = %v", err)
			}

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("RewriteImages() = %v, want %v", got, tt.expected)
			}
		})
	}
}
