# Create an Elastic Cloud Extension
resource "ec_extension" "my_extension" {
  name           = "%s"
  description    = "%s"
  version        = "*"
  extension_type = "bundle"
}
