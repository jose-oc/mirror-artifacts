package charts

/*
// GEMINI:
// Modify this code to cover a new helm chart structure we are not covering yet.
// I've written a test to cover it: TestMirrorGrafanaAgentOperatorChart, you can see the values.yaml file I'm trying to cover there
// (the file is pkg/charts/tests/data_test/input_charts/grafana-agent-operator/values.yaml)
// The current behaviour should work as it is: this is, priority is the global values if they are present.
// if not, look for the image sections. In these `image` sections you can find `registry`, then change its value.
// If not, look for `repo` or `repository` in each `image` sections and replace its value.
//
// Also, please add doc comments.

*/

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/appcontext"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/version"
	"github.com/rs/zerolog/log"
)

var (
	// annotationsRegex matches the 'annotations:' line in a Chart.yaml.
	annotationsRegex = regexp.MustCompile(`(?m)^annotations:\s*`)
	// versionRegex extracts the version from a Chart.yaml.
	versionRegex = regexp.MustCompile(`(?m)^version:\s*(.+)`)
	// globalSectionRegex matches the 'global:' section in a values.yaml.
	globalSectionRegex = regexp.MustCompile(`^global:\s*`)
	// imageRegistryRegex matches 'imageRegistry:' under the 'global:' section.
	imageRegistryRegex = regexp.MustCompile(`^\s+imageRegistry:\s*`) // global.imageRegistry
	// globalImageSectionRegex matches 'image:' under the 'global:' section.
	globalImageSectionRegex = regexp.MustCompile(`^\s+image:\s*`) // global.image
	// registryRegex matches a 'registry:' line within an image section.
	registryRegex = regexp.MustCompile(`^\s+registry:\s*`) // any *.registry
	// imageSectionRegex matches a top-level 'image:' section.
	imageSectionRegex = regexp.MustCompile(`^image:\s*`) // top-level image:
	// anyImageSectionRegex matches any 'image:' section, capturing indentation.
	anyImageSectionRegex = regexp.MustCompile(`^(\s*)image:\s*$`)
	// repoRegex matches 'repo:' or 'repository:' lines.
	repoRegex = regexp.MustCompile(`^\s+(repo|repository):\s*(.*)`)
)

// ProvenanceMetadata holds information about the original chart before it was repackaged.
type ProvenanceMetadata struct {
	RepackagedBy         string
	OriginalChartURL     string
	OriginalChartName    string
	OriginalChartVersion string
	Timestamp            string
}

// TransformHelmChart processes a Helm chart from a source path, modifies it, and saves it to an output path.
// Modifications include updating the Chart.yaml with a version suffix and provenance annotations,
// and updating the values.yaml to point to a new container registry.
func TransformHelmChart(ctx *appcontext.AppContext, chart Chart, srcChartPath string, outputPath ...string) (string, error) {
	var transformedChartPath string
	if len(outputPath) == 0 {
		transformedChartPath = fmt.Sprintf("%s-%s", srcChartPath, time.Now().Format("20060102150405.1234"))
	} else {
		transformedChartPath = path.Join(outputPath[0], fmt.Sprintf("%s-%s", chart.Name, time.Now().Format("20060102150405.1234")))
	}

	// Create output directory
	if err := os.MkdirAll(transformedChartPath, 0755); err != nil {
		log.Error().Err(err).Str("path", transformedChartPath).Msg("Failed to create output directory")
		return "", err
	}

	err := filepath.Walk(srcChartPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(srcChartPath, path)
		if err != nil {
			return err
		}

		// Calculate destination path
		destPath := filepath.Join(transformedChartPath, relPath)

		// Handle directories
		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		// Handle files
		baseName := filepath.Base(path)

		switch baseName {
		case "Chart.yaml":
			// Only process the root Chart.yaml
			if filepath.Dir(relPath) == "." {
				return processChartYaml(path, destPath, ctx.Config.Options.Suffix, chart.Source)
			}
			return copyFile(path, destPath)
		case "values.yaml":
			// Only process the root values.yaml, skip sub-chart values.yaml files
			if filepath.Dir(relPath) == "." {
				return processValuesYaml(path, destPath, ctx.Config.GCP.GARRepoCharts)
			}
			return copyFile(path, destPath)
		default:
			return copyFile(path, destPath)
		}
	})
	if err != nil {
		log.Error().Err(err).Str("path", transformedChartPath).Msg("Failed to process chart")
		return "", err
	}

	log.Debug().
		Str("original chart path", srcChartPath).
		Str("transformed chart path", transformedChartPath).
		Msg("Helm chart transformed")

	return transformedChartPath, nil
}

// processChartYaml reads a Chart.yaml file, updates its version, adds provenance annotations,
// and writes the modified content to a destination path.
func processChartYaml(srcPath, destPath, versionSuffix, originalChartURL string) error {
	content, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read Chart.yaml: %w", err)
	}

	// Extract original chart metadata before modification
	originalVersion := extractVersion(string(content))
	originalName := extractChartName(string(content))

	// Replace version with suffix
	modified := replaceVersion(string(content), versionSuffix)

	// Add provenance metadata
	// TODO Think about getting the original helm chart digest and store it here too
	provenance := ProvenanceMetadata{
		RepackagedBy:         fmt.Sprintf("%s %s", version.AppName, version.Version),
		OriginalChartURL:     originalChartURL,
		OriginalChartName:    originalName,
		OriginalChartVersion: originalVersion,
		Timestamp:            time.Now().UTC().Format(time.RFC3339),
	}
	modified = addProvenanceAnnotations(modified, provenance)

	return os.WriteFile(destPath, []byte(modified), 0644)
}

// extractVersion extracts the chart version from the content of a Chart.yaml file.
func extractVersion(content string) string {
	matches := versionRegex.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// extractChartName extracts the chart name from the content of a Chart.yaml file.
func extractChartName(content string) string {
	nameRegex := regexp.MustCompile(`(?m)^name:\s*(.+)$`)
	matches := nameRegex.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// replaceVersion appends a suffix to the chart version in the content of a Chart.yaml file.
func replaceVersion(content string, versionSuffix string) string {
	return versionRegex.ReplaceAllStringFunc(content, func(match string) string {
		parts := strings.SplitN(match, ":", 2)
		if len(parts) != 2 {
			return match
		}
		currentVersion := strings.TrimSpace(parts[1])
		newVersion := currentVersion + "-" + versionSuffix
		return fmt.Sprintf("version: %s", newVersion)
	})
}

// addProvenanceAnnotations adds a set of provenance annotations to the Chart.yaml content.
// It either adds them to an existing 'annotations' section or creates one if it doesn't exist.
func addProvenanceAnnotations(content string, provenance ProvenanceMetadata) string {
	lines := strings.Split(content, "\n")
	var result []string
	annotationsFound := false
	annotationsProcessed := false
	apiVersionIndex := -1

	for i, line := range lines {
		// Track where apiVersion is for potential insertion point
		if strings.HasPrefix(line, "apiVersion:") {
			apiVersionIndex = i
		}

		result = append(result, line)

		// Check if this is the annotations line
		if annotationsRegex.MatchString(line) && !annotationsProcessed {
			annotationsFound = true
			annotationsProcessed = true

			// Look ahead to find the indentation level of existing annotations
			indent := "  "
			if i+1 < len(lines) {
				nextLine := lines[i+1]
				if len(nextLine) > 0 && nextLine[0] == ' ' {
					// Extract indentation from the next line
					for j, ch := range nextLine {
						if ch != ' ' {
							indent = nextLine[:j]
							break
						}
					}
				}
			}

			// Add provenance metadata annotations
			result = append(result, fmt.Sprintf("%s# --- Provenance Metadata ---", indent))
			result = append(result, fmt.Sprintf("%srepackage.provenance/repackaged-by: \"%s\"", indent, provenance.RepackagedBy))
			result = append(result, fmt.Sprintf("%srepackage.provenance/original-chart-url: \"%s\"", indent, provenance.OriginalChartURL))
			result = append(result, fmt.Sprintf("%srepackage.provenance/original-chart-name: \"%s\"", indent, provenance.OriginalChartName))
			result = append(result, fmt.Sprintf("%srepackage.provenance/original-chart-version: \"%s\"", indent, provenance.OriginalChartVersion))
			result = append(result, fmt.Sprintf("%srepackage.provenance/timestamp: \"%s\"", indent, provenance.Timestamp))
		}
	}

	// If annotations section wasn't found, create it after apiVersion
	if !annotationsFound {
		if apiVersionIndex >= 0 {
			// Insert annotations section after apiVersion line
			newResult := make([]string, 0, len(result)+8)
			newResult = append(newResult, result[:apiVersionIndex+1]...)
			newResult = append(newResult, "annotations:")
			newResult = append(newResult, "  # --- Provenance Metadata ---")
			newResult = append(newResult, fmt.Sprintf("  repackage.provenance/repackaged-by: \"%s\"", provenance.RepackagedBy))
			newResult = append(newResult, fmt.Sprintf("  repackage.provenance/original-chart-url: \"%s\"", provenance.OriginalChartURL))
			newResult = append(newResult, fmt.Sprintf("  repackage.provenance/original-chart-name: \"%s\"", provenance.OriginalChartName))
			newResult = append(newResult, fmt.Sprintf("  repackage.provenance/original-chart-version: \"%s\"", provenance.OriginalChartVersion))
			newResult = append(newResult, fmt.Sprintf("  repackage.provenance/timestamp: \"%s\"", provenance.Timestamp))
			newResult = append(newResult, result[apiVersionIndex+1:]...)
			result = newResult
		} else {
			log.Printf("Warning: Could not find suitable location to add annotations in Chart.yaml")
		}
	}

	return strings.Join(result, "\n")
}

// handleGlobalImageDotRegistry handles the case of 'global.image.registry'.
func handleGlobalImageDotRegistry(line string, registryURL string, inImageSection *bool) (string, bool) {
	if globalImageSectionRegex.MatchString(line) {
		*inImageSection = true
	}

	if *inImageSection && registryRegex.MatchString(line) {
		indent := line[:strings.Index(line, "registry")]
		return fmt.Sprintf("%sregistry: \"%s\"", indent, registryURL), true
	}

	// If we are in image section and encounter a line that is not indented as part of image section,
	// then we are out of image section.
	trimmed := strings.TrimSpace(line)
	if *inImageSection && len(line) > 0 && line[0] != ' ' && line[0] != '#' && trimmed != "" {
		*inImageSection = false
	}

	return line, false
}

// handleGlobalImageRegistry handles the case of 'global.imageRegistry'.
func handleGlobalImageRegistry(line string, registryURL string) (string, bool) {
	if imageRegistryRegex.MatchString(line) {
		indent := line[:strings.Index(line, "imageRegistry")]
		return fmt.Sprintf("%simageRegistry: \"%s\"", indent, registryURL), true
	}
	return line, false
}

// handleImageRepo handles 'repo' or 'repository' keys within an image section,
// prepending the new registry URL to their values.
func handleImageRepo(line string, registryURL string) (string, bool) {
	matches := repoRegex.FindStringSubmatch(line)
	if len(matches) == 3 {
		key := matches[1]
		valueStr := strings.TrimSpace(matches[2])
		currentValue := strings.Trim(valueStr, "\"'")
		newValue := registryURL + "/" + currentValue
		indent := line[:strings.Index(line, key)]
		return fmt.Sprintf("%s%s: \"%s\"", indent, key, newValue), true
	}
	return line, false
}

// hasGlobalRegistry checks if a values.yaml file contains a global image registry definition
// (either 'global.imageRegistry' or 'global.image.registry').
func hasGlobalRegistry(srcPath string) (bool, error) {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return false, fmt.Errorf("failed to open values.yaml: %w", err)
	}
	defer srcFile.Close()

	scanner := bufio.NewScanner(srcFile)
	inGlobalSection := false
	inImageSection := false

	for scanner.Scan() {
		line := scanner.Text()

		// Check if we're entering the global section
		if globalSectionRegex.MatchString(line) {
			inGlobalSection = true
			continue
		}

		// Check if we're in the global section
		if inGlobalSection {
			// Check for global.imageRegistry
			if imageRegistryRegex.MatchString(line) {
				return true, nil
			}

			// Check for global.image.registry
			if globalImageSectionRegex.MatchString(line) {
				inImageSection = true
				continue
			}
			if inImageSection && registryRegex.MatchString(line) {
				return true, nil
			}

			// Detect if we've left the global section
			trimmed := strings.TrimSpace(line)
			if len(line) > 0 && line[0] != ' ' && line[0] != '#' && trimmed != "" {
				inGlobalSection = false
				inImageSection = false
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("error reading values.yaml: %w", err)
	}

	return false, nil
}

// processValuesYaml reads a values.yaml file and updates container image references.
// It prioritizes global registry settings. If not present, it processes all 'image:' sections.
func processValuesYaml(srcPath, destPath, registryURL string) error {
	// Check if values.yaml contains global.imageRegistry or global.image.registry
	hasGlobal, err := hasGlobalRegistry(srcPath)
	if err != nil {
		return err
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open values.yaml: %w", err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create output values.yaml: %w", err)
	}
	defer destFile.Close()

	scanner := bufio.NewScanner(srcFile)
	writer := bufio.NewWriter(destFile)
	defer writer.Flush()

	// Track if we found the imageRegistry field
	foundRegistry := false
	inGlobalSection := false
	inImageSection := false
	inTopImageSection := false

	for scanner.Scan() {
		line := scanner.Text()
		originalLine := line

		// Check if we're entering the global section
		if globalSectionRegex.MatchString(line) {
			inGlobalSection = true
			writer.WriteString(line + "\n")
			continue
		}

		// Check if we're entering the top-level image section
		if !hasGlobal && imageSectionRegex.MatchString(line) {
			inTopImageSection = true
			writer.WriteString(line + "\n")
			continue
		}

		// Check if we're in the global section
		if inGlobalSection {
			// Try to handle global.imageRegistry
			if modifiedLine, replaced := handleGlobalImageRegistry(line, registryURL); replaced {
				foundRegistry = true
				writer.WriteString(modifiedLine + "\n")
				continue
			}

			// Try to handle global.image.registry
			if modifiedLine, replaced := handleGlobalImageDotRegistry(line, registryURL, &inImageSection); replaced {
				foundRegistry = true
				writer.WriteString(modifiedLine + "\n")
				continue
			}

			// Detect if we've left the global section
			trimmed := strings.TrimSpace(line)
			if len(line) > 0 && line[0] != ' ' && line[0] != '#' && trimmed != "" {
				inGlobalSection = false
				inImageSection = false // Reset inImageSection when leaving global section
			}
		}

		// Check if we're in the top-level image section and no global registry was found
		if inTopImageSection && !hasGlobal {
			if modifiedLine, replaced := handleImageRepo(line, registryURL); replaced {
				foundRegistry = true
				writer.WriteString(modifiedLine + "\n")
				continue
			}

			// Detect if we've left the top-level image section
			trimmed := strings.TrimSpace(line)
			if len(line) > 0 && line[0] != ' ' && line[0] != '#' && trimmed != "" {
				inTopImageSection = false
			}
		}

		writer.WriteString(originalLine + "\n")
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading values.yaml: %w", err)
	}

	if !foundRegistry {
		log.Printf("Warning: No global image registry field (global.imageRegistry or global.image.registry) or image repo found in values.yaml")
	}

	return nil
}

// copyFile copies a single file from a source to a destination.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}
