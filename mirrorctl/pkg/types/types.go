package types

// Image represents a container image with its name and source.
// The source is the full image reference, including the registry, repository, and tag.
// The name is the short name of the image.
type Image struct {
	Name   string `yaml:"name" json:"name"`
	Source string `yaml:"source" json:"source"`
}

// ImagesList represents a list of container images.
// It is used to unmarshal the images.yaml file.
type ImagesList struct {
	Images []Image `yaml:"images" json:"images"`
}

// Chart represents a Helm chart with its name, source, and version.
// The source is the URL of the Helm repository.
// The name is the name of the chart in the repository.
// The version is the version of the chart to be downloaded.
type Chart struct {
	Name    string `yaml:"name"`
	Source  string `yaml:"source"`
	Version string `yaml:"version"`
}

// ChartsList represents a list of Helm charts.
// It is used to unmarshal the charts.yaml file.
type ChartsList struct {
	Charts []Chart `yaml:"charts"`
}
