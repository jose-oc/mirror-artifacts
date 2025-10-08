# --- 1. Terraform Backend Configuration (State Management) ---
terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }

  # Configure the GCS backend for state management
  backend "gcs" {
    bucket = "provisioning-tfstate-tst"
    prefix = "gar/test-repositories"
  }
}

# --- 2. Google Provider Configuration (Replace with your actual project/region) ---
provider "google" {
  project = var.gcp_project_id
  region  = "europe-southwest1"
}

# --- 3. Artifact Registry: Helm Charts Repository ---
resource "google_artifact_registry_repository" "helm_charts_repo" {
  repository_id = "test-helm-charts"
  location      = "europe-southwest1"
  format        = "DOCKER"        # Recommended format for OCI Helm charts

  # Custom configuration
  description           = "Testing a repository for mirroring Helm charts in OCI format"
  # Immutable image tags are disabled by default, but we explicitly set it to false
  docker_config {
    immutable_tags = false
  }

  # Add specified labels (metadata)
  labels = {
    team    = "joc-platform"
    purpose = "test"
    created = "terraform"
  }
}

# --- 4. Artifact Registry: Container Images Repository ---
resource "google_artifact_registry_repository" "container_images_repo" {
  repository_id = "test-container-images"
  location      = "europe-southwest1"
  format        = "DOCKER"

  # Custom configuration
  description           = "Testing a repository for mirroring docker images"
  # Immutable image tags are disabled by default, but we explicitly set it to false
  docker_config {
    immutable_tags = false
  }


  # Add specified labels (metadata)
  labels = {
    team    = "joc-platform"
    purpose = "test"
    created = "terraform"
  }
}
