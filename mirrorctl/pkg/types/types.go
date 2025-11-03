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
