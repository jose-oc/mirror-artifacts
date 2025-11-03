package types

// Image represents an entry in images.yaml
type Image struct {
	Name   string `yaml:"name"`
	Source string `yaml:"source"`
}

// ImagesList represents the structure of images.yaml
type ImagesList struct {
	Images []Image `yaml:"images"`
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
