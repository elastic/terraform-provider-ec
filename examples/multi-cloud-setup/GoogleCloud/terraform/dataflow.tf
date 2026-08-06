# -------------------------------------------------------------
# Create a Dataflow job to read from BigQuery and write to Elastic
# -------------------------------------------------------------
# resource "google_dataflow_flex_template_job" "read_from_bigquery_to_elasticserach" {
#   project                 = var.google_cloud_project
#   provider                = google-beta
#   name                    = var.google_cloud_dataflow_job_name
#   region                  = var.google_cloud_region
#   container_spec_gcs_path = var.google_cloud_container_spec_gcs_path
#   parameters = {
#     connectionUrl         = ec_deployment.elastic_gc_deployment.elasticsearch[0].cloud_id
#     apiKey                = data.external.elastic_generate_api_key.result.encoded
#     index                 = var.elastic_index_name
#     inputTableSpec        = var.google_cloud_inputTableSpec
#     maxNumWorkers         = var.google_cloud_maxNumWorkers
#     maxRetryAttempts      = var.google_cloud_maxRetryAttempts
#     maxRetryDuration      = var.google_cloud_maxRetryDuration
#   }
#   depends_on = [data.external.elastic_generate_api_key]
# }