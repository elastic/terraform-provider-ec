provider "google" {
  project     = var.google_cloud_project
  region      = var.google_cloud_region
  zone    	  = "${var.google_cloud_region}-a"
  credentials = var.google_cloud_service_account_path
}
