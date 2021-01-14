locals {
  file_path = "%s"
}

resource "ec_deployment_extension" "my_extension" {
  name           = "%s"
  description    = "%s"
  version        = "*"
  extension_type = "bundle"

  file_path = local.file_path
  file_hash = filebase64sha256(local.file_path)
}
