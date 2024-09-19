# To simplify managing multiple members with the same roles, the roles can be assigned to local variables
locals {
  deployment_admin = {
    deployment_roles = [
      {
        role            = "admin"
        all_deployments = true
      }
    ]
  }

  deployment_viewer = {
    deployment_roles = [
      {
        role            = "viewer"
        all_deployments = true
      }
    ]
  }
}

resource "ec_organization" "my_org" {
  members = {
    "admin@example.com"          = local.deployment_admin
    "viewer@example.com"         = local.deployment_viewer
    "another.viewer@example.com" = local.deployment_viewer
  }
}