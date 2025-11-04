package chartscanner

import (
	"os"
	"regexp"
	"strings"

	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/types"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

var (
	templatedValueRegex = regexp.MustCompile(`{{.*}}`)
)

// parseYAMLFile parses a YAML file and extracts image references.
func parseYAMLFile(filePath string) ([]types.Image, error) {
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

// extractImagesFromNode recursively traverses a YAML node and extracts image references.
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
					if !strings.Contains(imageSource, ":") || strings.HasSuffix(imageSource, ":null") { // If no tag is present or tag is :null, append :latest
						imageSource = strings.TrimSuffix(imageSource, ":null") + ":latest"
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

// extractImageFromMapping extracts image details from a mapping node (e.g., repository, registry, tag).
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

	// If tag is empty or "null", append :latest
	if tag == "" || tag == "null" {
		source = source + ":latest"
	} else {
		source = source + ":" + tag
	}

	imageName := extractImageName(source)
	// Ensure imageName is not empty or "null" before appending
	if imageName == "" || imageName == "null" { // Skip if image name is invalid
		return
	}

	*images = append(*images, types.Image{Name: imageName, Source: source})

	log.Debug().Str("name", imageName).Str("source", source).Msg("extracted image")
}

// isTemplatedValue checks if a string contains a Helm templated value.
func isTemplatedValue(value string) bool {
	return templatedValueRegex.MatchString(value)
}

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
