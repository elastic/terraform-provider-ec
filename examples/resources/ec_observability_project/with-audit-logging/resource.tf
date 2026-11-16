# The destination must be an existing observability or security project.
resource "ec_observability_project" "audit_logs" {
  name      = "my_audit_logs"
  region_id = "aws-us-east-1"
}

resource "ec_observability_project" "my_project" {
  name      = "my_project"
  region_id = "aws-us-east-1"

  monitoring = {
    logging = {
      audit = {
        # enabled defaults to true. Set it to false to pause delivery while
        # keeping the configuration.
        destination = {
          project_id   = ec_observability_project.audit_logs.id
          project_type = "observability"
        }
        # Optional. Ids of logging ignore filters that drop matching audit
        # events before delivery.
        ignore_filter_ids = ["audit-ignore-data-reads"]
      }
    }
  }
}
