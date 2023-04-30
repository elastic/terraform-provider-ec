data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "us-east-1"
}

resource "ec_deployment" "example_minimal" {
  region                 = "us-east-1"
  version                = data.ec_stack.latest.version
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch = {

    autoscale = "true"

    # If `autoscale` is set, all topology elements that
    # - either set `size` in the plan or
    # - have non-zero default `max_size` (that is read from the deployment templates's `autoscaling_max` value)
    # have to be listed even if their blocks don't specify other fields beside `id`

    cold = {
      autoscaling = {}
    }

    frozen = {
      autoscaling = {}
    }

    hot = {
      size = "8g"

      autoscaling = {
        max_size          = "128g"
        max_size_resource = "memory"
      }
    }

    ml = {
      autoscaling = {}
    }

    warm = {
      autoscaling = {}
    }
  }

  # Initial size for `hot_content` tier is set to 8g
  # so `hot_content`'s size has to be added to the `ignore_changes` meta-argument to ignore future modifications that can be made by the autoscaler
  lifecycle {
    ignore_changes = [
      elasticsearch.hot.size
    ]
  }

  kibana = {}

  integrations_server = {}

  enterprise_search = {}
}
