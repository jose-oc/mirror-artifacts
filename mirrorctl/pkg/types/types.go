package types

// Image represents a container image
type Image struct {
	Name   string `yaml:"name" json:"name"`
	Source string `yaml:"source" json:"source"`
}

// ImagesList represents a collection of container images
type ImagesList struct {
	Images []Image `yaml:"images" json:"images"`
}

// Chart represents an entry in the charts.yaml file
type Chart struct {
	Name    string `yaml:"name"`
	Source  string `yaml:"source"`
	Version string `yaml:"version"`
}

// ChartsList represents the structure of the charts.yaml file
type ChartsList struct {
	Charts []Chart `yaml:"charts"`
}
