data "ec_deployments" "example" {
  name_prefix            = "test"
  deployment_template_id = "azure-compute-optimized"

  size = 200

  tags = {
    "foo" = "bar"
  }

  elasticsearch {
    healthy = "true"
  }

  kibana {
    status = "started"
  }

  integrations_server {
    version = "8.0.0"
  }

  enterprise_search {
    healthy = "true"
  }
}
