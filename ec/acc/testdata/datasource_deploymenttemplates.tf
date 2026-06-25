data "ec_deployment_templates" "test" {
  region        = "%s"
  stack_version = "7.17.0"
}

data "ec_deployment_templates" "by_id" {
  region = "%s"
  id     = data.ec_deployment_templates.test.templates.0.id
}