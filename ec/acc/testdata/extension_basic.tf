resource "ec_deployment_extension" "my_extension" {
  name           = "%s"
  description    = "%s"
  version        = "*"
  extension_type = "bundle"
}
