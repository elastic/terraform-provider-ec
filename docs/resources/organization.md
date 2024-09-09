---
page_title: "Elastic Cloud: ec_organization Resource"
description: |-
  Manages an Elastic Cloud organization membership.

  ~> **This resource can only be used with Elastic Cloud SaaS**
---

# Resource: ec_organization

Manages an Elastic Cloud organization membership.

  ~> **This resource can only be used with Elastic Cloud SaaS**

## Example Usage

### Import

To import an organization into terraform, first define your organization configuration in your terraform file. For example:
```terraform
resource "ec_organization" "myorg" {
}
```

Then import the organization using your organization-id (The organization id can be found on [the organization page](https://cloud.elastic.co/account/members))
```bash
terraform import ec_organization.myorg <organization-id>
```

Now you can run `terraform plan` to see if there are any diffs between your config and how your organization is currently configured.

### Basic

```terraform
resource "ec_organization" "my_org" {
  members = {
    "a.member@example.com" = {
      # All role definitions are optional

      # Define roles for the whole organization
      # Available roles are documented here: https://www.elastic.co/guide/en/cloud/current/ec-user-privileges.html#ec_organization_level_roles
      organization_role = "billing-admin"

      # Define deployment-specific roles
      # Available roles are documented here: https://www.elastic.co/guide/en/cloud/current/ec-user-privileges.html#ec_instance_access_roles
      deployment_roles = [
        # A role can be given for all deployments
        {
          role = "editor"
          for_all_deployments = true
        },

        # Or just for specific deployments
        {
          role = "editor"
          deployment_ids = ["ce03a623751b4fc98d48400fec58b9c0"]
        }
      ]

      # Define roles for elasticsearch projects (Docs: https://www.elastic.co/docs/current/serverless/general/assign-user-roles#es)
      project_elasticsearch_roles = [
        # A role can be given for all projects
        {
          role = "admin"
          for_all_projects = true
        },

        # Or just for specific projects
        {
          role = "admin"
          project_ids = ["c866244b611442d585e23a0cc8c9434c"]
        }
      ]

      project_observability_roles = [
        # Same as for an elasticsearch project
        # Available roles are documented here: https://www.elastic.co/docs/current/serverless/general/assign-user-roles#observability
      ]

      project_security_roles = [
        # Same as for an elasticsearch project
        # Available roles are documented here: https://www.elastic.co/docs/current/serverless/general/assign-user-roles#security
      ]
    }
  }
}
```

### Use variables to give the same roles to multiple users

```terraform
# To simplify managing multiple members with the same roles, the roles can be assigned to local variables
locals {
  deployment_admin = {
    deployment_roles = [
      {
        role = "admin"
        for_all_deployments = true
      }
    ]
  }

  deployment_viewer = {
    deployment_roles = [
      {
        role = "viewer"
        for_all_deployments = true
      }
    ]
  }
}

resource "ec_organization" "my_org" {
  members = {
    "admin@example.com" = local.deployment_admin
    "viewer@example.com" = local.deployment_viewer
    "another.viewer@example.com" = local.deployment_viewer
  }
}
```