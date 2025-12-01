package chartscanner

import (
	"strings"

	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/config"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/types"
	"github.com/rs/zerolog/log"
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
func RewriteImages(images []types.Image, imageRewriteRules []config.ImageRewrite) ([]types.Image, error) {
	var rewrittenImages []types.Image

	for _, image := range images {
		currentImage := image
		for _, rule := range imageRewriteRules {
			if strings.Contains(currentImage.Source, rule.Old) {
				newImage := strings.ReplaceAll(currentImage.Source, rule.Old, rule.New)
				log.Debug().Str("image", currentImage.Source).Str("newImage", newImage).Msg("Rewriting image")
				currentImage.Source = newImage
			}
		}
		rewrittenImages = append(rewrittenImages, currentImage)
	}

	return rewrittenImages, nil
}
