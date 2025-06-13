# Provider configuration
provider "google" {
  project = var.project_id
  region  = var.region
  zone    = var.zone
}

# Data source for existing PostgreSQL database
# This references your existing database without attempting to manage it
data "google_sql_database_instance" "existing_db" {
  name = var.db_instance_name
}

# Cloud Storage bucket for storing user images
resource "google_storage_bucket" "app_images" {
  name          = "${var.project_id}-${var.app_name}-images"
  location      = var.region
  storage_class = "STANDARD"  # This is cost-effective and billed only for what you use

  # Optional settings for your bucket
  uniform_bucket_level_access = true
  public_access_prevention    = "inherited"  # Allows public access for images through object ACLs
  
  # CORS configuration
  cors {
    origin          = ["http://localhost:8080", "https://kanji-go-pdjzxrqjaq-uc.a.run.app", "https://kanji-go-111333019928.us-central1.run.app"]
    method          = ["GET", "POST", "PUT", "DELETE"]
    response_header = ["Content-Type"]
    max_age_seconds = 3600
  }
}

# Make bucket publicly accessible
resource "google_storage_bucket_iam_binding" "public_access" {
  bucket = google_storage_bucket.app_images.name
  role   = "roles/storage.objectViewer"
  members = [
    "allUsers",
  ]
}

# Service account for Cloud Run to access DB and Cloud Storage
resource "google_service_account" "app_service_account" {
  account_id   = "${var.app_name}-service-account"
  display_name = "${var.app_name} Service Account"
  description  = "Service account for ${var.app_name} to access Cloud SQL and Cloud Storage"
}

# Grant necessary permissions to service account

# DB permissions
resource "google_project_iam_member" "cloud_sql_client" {
  project = var.project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${google_service_account.app_service_account.email}"
}

# Storage permissions
resource "google_storage_bucket_iam_member" "storage_object_admin" {
  bucket = google_storage_bucket.app_images.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${google_service_account.app_service_account.email}"
}

# Secret Manager access
resource "google_project_iam_member" "secret_accessor" {
  project = var.project_id
  role    = "roles/secretmanager.secretAccessor"
  member  = "serviceAccount:${google_service_account.app_service_account.email}"
}

# Reference the secrets (but don't create or manage them)
data "google_secret_manager_secret" "db_password" {
  secret_id = "db-password"
}

data "google_secret_manager_secret" "csrf_key" {
  secret_id = "csrf-key"
}

# Cloud Run service
resource "google_cloud_run_service" "app" {
  name     = var.app_name
  location = var.region

  template {
    spec {
      containers {
        image = var.container_image

        # Environment variables
        env {
          name  = "BUCKET_NAME"
          value = google_storage_bucket.app_images.name
        }
        env {
          name  = "GCP_PROJECT_ID"
          value = var.project_id
        }
        env {
          name  = "GCP_REGION"
          value = var.region
        }
        env {
          name  = "DB_SCHEMA"
          value = "kanji_go"
        }
        env {
          name  = "APP_ENV"
          value = var.app_env
        }

        # Set up secret environment variables from Secret Manager
        env {
          name = "DB_HOST"
          value = data.google_sql_database_instance.existing_db.connection_name
        }
        env {
          name = "DB_PORT"
          value = "5432"
        }
        env {
          name = "DB_USER"
          value = var.db_user
        }
        env {
          name = "DB_NAME"
          value = var.db_name
        }
        env {
          name = "DB_PASSWORD"
          value_from {
            secret_key_ref {
              name = "db-password"
              key  = "latest"
            }
          }
        }
        env {
          name = "CSRF_KEY"
          value_from {
            secret_key_ref {
              name = "csrf-key"
              key  = "latest"
            }
          }
        }
      }

      # Use the service account
      service_account_name = google_service_account.app_service_account.email
    }

    # Configure the connection to Cloud SQL
    metadata {
      annotations = {
        "run.googleapis.com/cloudsql-instances" = data.google_sql_database_instance.existing_db.connection_name
      }
    }
  }

  # Traffic configuration
  traffic {
    percent         = 100
    latest_revision = true
  }
}

# Make the Cloud Run service publicly accessible
resource "google_cloud_run_service_iam_member" "public_access" {
  service  = google_cloud_run_service.app.name
  location = google_cloud_run_service.app.location
  role     = "roles/run.invoker"
  member   = "allUsers"  # Can be restricted to specific users/groups in production
}
