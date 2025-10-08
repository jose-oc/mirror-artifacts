# Google Artifact Registry (GAR) Test Repositories

This Terraform configuration provisions two new repositories in Google Artifact Registry (GAR) for testing purposes, configured to align with best practices for OCI artifacts and container images.

## What It Does

This configuration provisions the following resources in your specified GCP project:

1.  **GCS Backend:** Initializes a Terraform state backend in the bucket `gs://provisioning-tfstate-tst/gar/test-repositories/`.
2.  **`test-helm-charts` Repository:** A `DOCKER` format repository intended for storing Helm charts as OCI artifacts.
3.  **`test-container-images` Repository:** A `DOCKER` format repository for standard container images.

Both repositories are created with the following metadata (labels) and settings:

  * **Description:** "Testing a repository for mirroring Helm charts/docker images"
  * **Immutable Tags:** Disabled (`false`)
  * **Labels:** `team: poc-platform`, `purpose: test`, `created: terraform`

## Prerequisites

1.  **Google Cloud CLI:** Installed and authenticated.
2.  **Terraform:** Installed (version 0.13+ recommended).
3.  **Permissions:** Your GCP credentials must have permissions to create Artifact Registry repositories, and read/write to the GCS bucket for state management (`provisioning-tfstate-tst`).

## How to Run

### Step 1: Initialize Terraform

The initialization step prepares the local directory and sets up the remote GCS backend where the state file will be stored.

```bash
terraform init
```

### Step 2: Plan and Apply

The Terraform configuration requires one essential variable: `gcp_project_id`. **You must provide this value at runtime to avoid committing your Project ID to your repository.**

Choose one of the methods below to supply your Project ID:

#### Option 1: Using a Command Line Flag (Recommended)

Pass the value directly via the command line. This keeps the secret out of all local files.

```bash
# Run Plan
terraform plan -var="gcp_project_id=YOUR_PROJECT_ID_HERE"

# Run Apply
terraform apply -var="gcp_project_id=YOUR_PROJECT_ID_HERE"
```

#### Option 2: Using an Environment Variable (Ideal for CI/CD)

Set the variable as a `TF_VAR_` environment variable before running the command.

```bash
# 1. Set the variable (e.g., in your shell or CI pipeline)
export TF_VAR_gcp_project_id="YOUR_PROJECT_ID_HERE"

# 2. Run Terraform (it will automatically read the environment variable)
terraform apply
```

#### Option 3: Interactive Prompt

If you run `terraform plan` or `terraform apply` without providing the variable, Terraform will prompt you for the value.

```bash
terraform apply

# Console Prompt:
# var.gcp_project_id
#   The ID of the GCP project where resources will be created.
#
#   Enter a value: [Type your project ID here and press Enter]
```
