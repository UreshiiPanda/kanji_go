variable "project_id" {
  description = "Google Cloud Project ID"
  type        = string
}

variable "region" {
  description = "GCP region to deploy resources"
  type        = string
  default     = "us-central1"
}

variable "zone" {
  description = "GCP zone for resources that require it"
  type        = string
  default     = "us-central1-a"
}

variable "app_name" {
  description = "Name of the application"
  type        = string
  default     = "kanji-go"
}

variable "db_instance_name" {
  description = "Name of the existing Cloud SQL instance"
  type        = string
}

variable "db_name" {
  description = "Database name in the existing instance"
  type        = string
}

variable "db_user" {
  description = "Database username"
  type        = string
}

variable "container_image" {
  description = "Container image URL (e.g., gcr.io/project-id/image:tag)"
  type        = string
}

variable "app_env" {
  description = "Application environment (LOCAL or PROD)"
  type        = string
  default     = "PROD"
}
