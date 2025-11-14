package chartscanner

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/types"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

var (
	templatedValueRegex = regexp.MustCompile(`{{.*}}`)
)

// extractGlobalRegistry looks for global.imageRegistry or global.image.registry in values.yaml/yml
func extractGlobalRegistry(chartPath string) string {
	// Try values.yaml first, then values.yml
	valuesFiles := []string{"values.yaml", "values.yml"}

	for _, fileName := range valuesFiles {
		valuesPath := filepath.Join(chartPath, fileName)
		if _, err := os.Stat(valuesPath); err == nil {
			log.Debug().Msgf("Checking for global registry in: %s", valuesPath)

			data, err := os.ReadFile(valuesPath)
			if err != nil {
				log.Warn().Err(err).Msgf("Failed to read values file: %s", valuesPath)
				continue
			}

			var node yaml.Node
			if err := yaml.Unmarshal(data, &node); err != nil {
				log.Warn().Err(err).Msgf("Failed to parse values file: %s", valuesPath)
				continue
			}

			registry := findGlobalRegistry(&node)
			if registry != "" {
				return registry
			}
		}
	}

	return ""
}

// findGlobalRegistry searches for global.imageRegistry or global.image.registry in YAML
func findGlobalRegistry(node *yaml.Node) string {
	if node == nil || node.Kind != yaml.DocumentNode {
		return ""
	}

	for _, content := range node.Content {
		if content.Kind == yaml.MappingNode {
			var imageRegistry, imageRegistryNested string

			for i := 0; i < len(content.Content); i += 2 {
				keyNode := content.Content[i]
				valueNode := content.Content[i+1]

				if strings.ToLower(keyNode.Value) == "global" && valueNode.Kind == yaml.MappingNode {
					// Look for both patterns
					for j := 0; j < len(valueNode.Content); j += 2 {
						globalKey := valueNode.Content[j]
						globalValue := valueNode.Content[j+1]

						if strings.ToLower(globalKey.Value) == "imageregistry" {
							if globalValue.Kind == yaml.ScalarNode && !isTemplatedValue(globalValue.Value) {
								imageRegistry = globalValue.Value
							}
						} else if strings.ToLower(globalKey.Value) == "image" && globalValue.Kind == yaml.MappingNode {
							// Look for nested registry under global.image
							for k := 0; k < len(globalValue.Content); k += 2 {
								imageKey := globalValue.Content[k]
								imageValue := globalValue.Content[k+1]

								if strings.ToLower(imageKey.Value) == "registry" {
									if imageValue.Kind == yaml.ScalarNode && !isTemplatedValue(imageValue.Value) {
										imageRegistryNested = imageValue.Value
									}
								}
							}
						}
					}

					// Determine which registry to use
					if imageRegistry != "" && imageRegistryNested != "" {
						log.Warn().Msg("Both global.imageRegistry and global.image.registry found. Using global.imageRegistry")
						return imageRegistry
					} else if imageRegistry != "" {
						return imageRegistry
					} else if imageRegistryNested != "" {
						return imageRegistryNested
					}
				}
			}
		}
	}

	return ""
}

// applyGlobalRegistry replaces the registry portion of an image with the global registry
func applyGlobalRegistry(img types.Image, globalRegistry string) types.Image {
	source := img.Source

	// Split source into registry, repository, and tag
	var repository, tag string

	// Extract tag/digest first
	if lastColon := strings.LastIndex(source, ":"); lastColon != -1 {
		// Check if it's a tag (not part of registry port)
		if !strings.Contains(source[lastColon:], "/") {
			tag = source[lastColon:]
			source = source[:lastColon]
		}
	}

	// Extract repository (everything after the first slash, or entire string if no slash)
	if firstSlash := strings.Index(source, "/"); firstSlash != -1 {
		repository = source[firstSlash+1:]
	} else {
		// No registry specified, entire string is repository
		repository = source
	}

	// Construct new source with global registry
	newSource := globalRegistry + "/" + repository + tag

	log.Debug().
		Str("original", img.Source).
		Str("new", newSource).
		Msg("Applied global registry override")

	return types.Image{
		Name:   img.Name,
		Source: newSource,
	}
}

// parseYAML parses a YAML file and returns a list of container images found in the file.
// It takes the path to the YAML file as input.
// It returns a slice of Image objects and an error if the file cannot be parsed.
func parseYAML(filePath string) ([]types.Image, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return nil, err
	}

	var images []types.Image
	extractImagesFromNode(&node, &images, "")

	return images, nil
}

// extractImagesFromNode recursively traverses a YAML node and extracts container image information.
// It takes a YAML node, a pointer to a slice of images, and the parent key as input.
func extractImagesFromNode(node *yaml.Node, images *[]types.Image, parentKey string) {
	if node == nil {
		return
	}

	switch node.Kind {
	case yaml.DocumentNode, yaml.SequenceNode:
		for _, content := range node.Content {
			extractImagesFromNode(content, images, parentKey)
		}
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]
			currentKey := keyNode.Value

			// Check for image patterns
			if strings.HasSuffix(strings.ToLower(currentKey), "image") && !isTemplatedValue(valueNode.Value) {
				if valueNode.Kind == yaml.ScalarNode {
					// Simple string image: image: docker.io/bitnami/minio-client:2024.1.31-debian-11-r2
					// Image without tag: image: amazon/aws-cli
					imageSource := valueNode.Value
					if imageSource == "" || imageSource == "null" { // Skip if image source is empty or "null"
						return
					}
					if !strings.Contains(imageSource, ":") {
						imageSource = imageSource + ":"
					}
					imageName := extractImageName(imageSource)
					if imageName == "" || imageName == "null" { // Skip if image name is invalid
						return
					}
					*images = append(*images, types.Image{Name: imageName, Source: imageSource})
				} else if valueNode.Kind == yaml.MappingNode {
					// Block with repository field:
					// image:
					//   repository: influxdb
					//   tag: 1.8.10-alpine
					// Block with registry + repository:
					// image:
					//   registry: docker.io
					//   repository: grafana/agent-operator
					//   tag: v0.25.1
					extractImageFromMapping(valueNode, images)
				}
			} else if strings.ToLower(currentKey) == "images" && valueNode.Kind == yaml.SequenceNode {
				// Image lists:
				// images:
				//   - repository: alpine
				//     tag: latest
				//   - repository: busybox
				for _, imgNode := range valueNode.Content {
					if imgNode.Kind == yaml.MappingNode {
						extractImageFromMapping(imgNode, images)
					}
				}
			}

			// Recursively call for nested nodes
			extractImagesFromNode(valueNode, images, currentKey)
		}
	}
}

// extractImageFromMapping extracts container image information from a YAML mapping node.
// It takes a YAML mapping node and a pointer to a slice of images as input.
func extractImageFromMapping(node *yaml.Node, images *[]types.Image) {
	var repository, registry, tag string
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]
		if valueNode.Kind == yaml.ScalarNode && !isTemplatedValue(valueNode.Value) {
			switch strings.ToLower(keyNode.Value) {
			case "repository", "repo":
				repository = valueNode.Value
			case "registry":
				registry = valueNode.Value
			case "tag":
				tag = valueNode.Value
			}
		}
	}

	// Only proceed if a valid repository is found
	if repository == "" || repository == "null" { // Also check for "null" string
		return // Skip if no valid repository is found
	}

	source := repository
	if registry != "" && registry != "null" { // Also check for "null" string
		source = registry + "/" + repository
	}

	source = source + ":" + tag

	imageName := extractImageName(source)
	// Ensure imageName is not empty or "null" before appending
	if imageName == "" || imageName == "null" { // Skip if image name is invalid
		return
	}

	*images = append(*images, types.Image{Name: imageName, Source: source})

	log.Debug().Str("name", imageName).Str("source", source).Msg("extracted image")
}

// isTemplatedValue checks if a string is a Helm template value.
// It takes a string as input and returns true if the string is a template value, false otherwise.
func isTemplatedValue(value string) bool {
	return templatedValueRegex.MatchString(value)
}

// extractImageName extracts the image name from a container image source string.
// It takes an image source string as input and returns the image name.
func extractImageName(source string) string {
	// Remove tag/digest
	name := source
	if lastIndex := strings.LastIndex(name, ":"); lastIndex != -1 {
		// Check if it is a tag or part of the port in a registry
		if !strings.Contains(name[lastIndex:], "/") {
			name = name[:lastIndex]
		}
	}

	// Get the last part of the path
	if lastIndex := strings.LastIndex(name, "/"); lastIndex != -1 {
		name = name[lastIndex+1:]
	}

	return name
}
