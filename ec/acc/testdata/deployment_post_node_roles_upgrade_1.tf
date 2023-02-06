data "ec_stack" "post_node_roles_upgrade" {
  version_regex = "7.12.?"
  region        = "%s"
}

resource "ec_deployment" "post_nr_upgrade" {
  name                   = "%s"
  region                 = "%s"
  version                = data.ec_stack.post_node_roles_upgrade.version
  deployment_template_id = "%s"

  elasticsearch = {
    topology = {
      "hot_content" = {
        size        = "1g"
        zone_count  = 1
        autoscaling = {}
      }
    }
  }
}
