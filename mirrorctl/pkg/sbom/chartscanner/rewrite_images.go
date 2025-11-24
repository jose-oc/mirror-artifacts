package chartscanner

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// RewriteConfig defines the structure of the rewrite configuration file
type RewriteConfig struct {
	Rewrites []RewriteRule `yaml:"rewrites"`
}

// RewriteRule defines a single replacement rule
type RewriteRule struct {
	Old string `yaml:"old"`
	New string `yaml:"new"`
}

// RewriteImages takes a list of images and a path to a configuration file,
// and returns a new list of images with replacements applied.
func RewriteImages(images []string, configPath string) ([]string, error) {
	// Read configuration
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var config RewriteConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	var rewrittenImages []string

	for _, image := range images {
		currentImage := image
		for _, rule := range config.Rewrites {
			if strings.Contains(currentImage, rule.Old) {
				newImage := strings.ReplaceAll(currentImage, rule.Old, rule.New)
				log.Printf("Rewriting image: %s -> %s (rule: %s -> %s)", currentImage, newImage, rule.Old, rule.New)
				currentImage = newImage
			}
		}
		rewrittenImages = append(rewrittenImages, currentImage)
	}

	return rewrittenImages, nil
}
