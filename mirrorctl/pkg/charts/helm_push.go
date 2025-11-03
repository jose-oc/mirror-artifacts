package charts

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/appcontext"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/version"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/rs/zerolog/log"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

// pushHelmChart pushes a packaged Helm chart (.tgz or .tar.gz) to Google Artifact Registry (GAR) using ORAS.
// Authentication is done with the `gcloud access token` command.
// It pushes the chart as an OCI artifact with manifest annotations.
// It can also run in a dry-run mode, in which no data is pushed.
func pushHelmChart(ctx *appcontext.AppContext, packagedChartPath string, chartName string, chartVersion string) error {
	log.Debug().Str("chart_path", packagedChartPath).Msg("Pushing chart to GAR")

	chartFilename := filepath.Base(packagedChartPath)
	imageName := stripArchiveExtension(chartFilename)
	if imageName == "" {
		return fmt.Errorf("unable to derive image name from packaged chart filename %q", chartFilename)
	}

	repoRef := buildRepositoryReference(ctx.Config.GCP.GARRepoCharts, chartName)
	tag := fmt.Sprintf("%s-%s", chartVersion, ctx.Config.Options.Suffix)

	if ctx.DryRun {
		log.Info().
			Str("chart_path", packagedChartPath).
			Str("repo", repoRef).
			Str("tag", tag).
			Msg("Running in dry-run mode: chart push to GAR omitted.")
		log.Info().
			Msgf("To push manually, run:\noras push %s %s:application/vnd.cncf.helm.chart.content.v1.tar+gzip --annotation mirrorctl/repackaged-by=%s/%s",
				repoRef,
				filepath.Base(packagedChartPath),
				version.AppName,
				version.Version)
		return nil
	}

	log.Debug().Str("repo_ref", repoRef).Msg("Normalized repository reference for ORAS")

	// Create the ORAS remote repository client with the normalized reference.
	repo, err := remote.NewRepository(repoRef)
	if err != nil {
		return fmt.Errorf("failed to create remote repository for %q: %w", repoRef, err)
	}

	// Authenticate with Google Artifact Registry using gcloud access token
	cmd := exec.Command("gcloud", "auth", "print-access-token")
	tokenBytes, err := cmd.Output()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get gcloud access token")
		return fmt.Errorf("failed to get gcloud access token: %w", err)
	}
	token := strings.TrimSpace(string(tokenBytes))

	// Configure the ORAS repository client auth using the gcloud access token.
	repo.Client = &auth.Client{
		Client: http.DefaultClient,
		Cache:  auth.NewCache(),
		Credential: func(ctx context.Context, s string) (auth.Credential, error) {
			return auth.Credential{AccessToken: token}, nil
		},
	}

	// TODO add more annotations? change the key to be aligned with the ones in the Chart.yaml?
	// Annotations attached to the manifest so we can trace origin / repackager
	annotations := map[string]string{
		"mirrorctl/repackaged-by": fmt.Sprintf("%s/%s", version.AppName, version.Version),
	}
	log.Debug().Interface("annotations", annotations).Msg("Setting chart annotations")

	fs, err := file.New(filepath.Dir(packagedChartPath))
	if err != nil {
		return fmt.Errorf("failed to create file store: %w", err)
	}
	defer fs.Close()

	// Add chart as a blob to the local file store
	fileDesc, err := fs.Add(context.Background(),
		filepath.Base(packagedChartPath),
		"application/vnd.cncf.helm.chart.content.v1.tar+gzip",
		packagedChartPath)
	if err != nil {
		return fmt.Errorf("failed to add chart file to store: %w", err)
	}

	// Push the chart blob itself to GAR
	log.Debug().Str("digest", fileDesc.Digest.String()).Msg("Pushing chart blob to GAR")
	chartData, err := os.Open(packagedChartPath)
	if err != nil {
		return fmt.Errorf("failed to open chart file for upload: %w", err)
	}
	defer chartData.Close()

	if err := repo.Push(context.Background(), fileDesc, chartData); err != nil {
		return fmt.Errorf("failed to push chart blob: %w", err)
	}

	// Create a minimal Helm config blob (Helm requires this)
	configJSON := []byte(`{"mediaType":"application/vnd.cncf.helm.config.v1+json","annotations":{}}`)
	hash := sha256.Sum256(configJSON)
	configDesc := v1.Descriptor{
		MediaType: "application/vnd.cncf.helm.config.v1+json",
		Digest:    digest.Digest("sha256:" + hex.EncodeToString(hash[:])),
		Size:      int64(len(configJSON)),
	}

	// Push config blob
	if err := repo.Push(context.Background(), configDesc, bytes.NewReader(configJSON)); err != nil {
		return fmt.Errorf("failed to push Helm config blob: %w", err)
	}

	// Pack manifest referencing config + layer
	packOpts := oras.PackManifestOptions{
		ConfigDescriptor:    &configDesc,
		Layers:              []v1.Descriptor{fileDesc},
		ManifestAnnotations: annotations,
	}
	manifestDesc, err := oras.PackManifest(
		context.Background(),
		fs,
		oras.PackManifestVersion1_1,
		"application/vnd.oci.image.manifest.v1+json",
		packOpts,
	)
	if err != nil {
		return fmt.Errorf("failed to pack manifest: %w", err)
	}

	// Push manifest itself
	manifestBytes, err := fs.Fetch(context.Background(), manifestDesc)
	if err != nil {
		return fmt.Errorf("failed to fetch manifest content from store: %w", err)
	}
	if err := repo.Push(context.Background(), manifestDesc, io.Reader(manifestBytes)); err != nil {
		return fmt.Errorf("failed to push manifest to GAR: %w", err)
	}

	if err := repo.Tag(context.Background(), manifestDesc, tag); err != nil {
		return fmt.Errorf("failed to tag manifest %q: %w", tag, err)
	}

	log.Info().
		Str("repo", repoRef).
		Str("tag", tag).
		Msg("Successfully pushed chart to GAR")
	return nil
}

// stripArchiveExtension removes common archive extensions from a filename.
// Example: "mychart-1.2.3.tgz" -> "mychart-1.2.3"
func stripArchiveExtension(name string) string {
	name = strings.TrimSpace(name)
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, ".tar.gz"):
		return name[:len(name)-len(".tar.gz")]
	case strings.HasSuffix(lower, ".tgz"):
		return name[:len(name)-len(".tgz")]
	case strings.HasSuffix(lower, ".zip"):
		return name[:len(name)-len(".zip")]
	case strings.HasSuffix(lower, ".tar"):
		return name[:len(name)-len(".tar")]
	default:
		// If there's no known archive suffix, remove a trailing dot-extension if present
		if idx := strings.LastIndex(name, "."); idx > 0 {
			return name[:idx]
		}
		return name
	}
}

// buildRepositoryReference normalizes a base repository string (possibly containing scheme
// or a "/v2" prefix) and guarantees a reference that includes an image name at the end.
// baseRepo may be any of:
//   - "europe-southwest1-docker.pkg.dev/poc-dev-123/my-repo"
//   - "https://europe-southwest1-docker.pkg.dev/v2/poc-dev-123/my-repo"
//   - "europe-southwest1-docker.pkg.dev/poc-dev-123/my-repo/my-image" (already contains image)
//
// If baseRepo already has an "image" segment (i.e., at least 4 path segments after host),
// it will be returned (after stripping scheme and /v2). Otherwise, imageName will be appended.
//
// This ensures ORAS will address: HOST/PROJECT/REPOSITORY/IMAGE
func buildRepositoryReference(baseRepo string, imageName string) string {
	ref := strings.TrimSpace(baseRepo)
	ref = strings.TrimPrefix(ref, "https://")
	ref = strings.TrimPrefix(ref, "http://")
	ref = strings.TrimPrefix(ref, "/v2/")
	ref = strings.Replace(ref, "/v2/", "/", 1)
	ref = strings.ReplaceAll(ref, "/v2/", "/")

	// Split into host + path segments
	parts := strings.Split(ref, "/")
	// Typical: parts[0] = host, parts[1] = project, parts[2] = repo, optional parts[3:] = image or deeper
	if len(parts) >= 4 {
		// Already has an image segment, return as-is
		return ref
	}
	// Append the imageName to ensure we have HOST/PROJECT/REPO/IMAGE
	if strings.HasSuffix(ref, "/") {
		return ref + imageName
	}
	return ref + "/" + imageName
}
