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
          role            = "editor"
          all_deployments = true
        },

        # Or just for specific deployments
        {
          role           = "editor"
          deployment_ids = ["ce03a623751b4fc98d48400fec58b9c0"]
        }
      ]

      # Define roles for elasticsearch projects (Docs: https://www.elastic.co/docs/current/serverless/general/assign-user-roles#es)
      project_elasticsearch_roles = [
        # A role can be given for all projects
        {
          role         = "admin"
          all_projects = true
        },

        # Or just for specific projects
        {
          role        = "admin"
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
