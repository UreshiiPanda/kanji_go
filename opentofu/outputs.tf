output "cloud_run_url" {
  description = "URL of the deployed Cloud Run service"
  value       = google_cloud_run_service.app.status[0].url
}

output "storage_bucket_name" {
  description = "Name of the Cloud Storage bucket for images"
  value       = google_storage_bucket.app_images.name
}

output "service_account_email" {
  description = "Email of the service account used by Cloud Run"
  value       = google_service_account.app_service_account.email
}

output "db_connection_name" {
  description = "Cloud SQL connection name"
  value       = data.google_sql_database_instance.existing_db.connection_name
}
