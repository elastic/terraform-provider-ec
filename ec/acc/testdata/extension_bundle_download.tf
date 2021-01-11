resource "ec_extension" "my_extension" {
  name           = "%s"
  version        = "*"
  extension_type = "bundle"
  download_url   = "%s"
}
