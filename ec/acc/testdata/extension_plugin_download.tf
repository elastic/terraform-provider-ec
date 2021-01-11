resource "ec_extension" "my_extension" {
  name           = "%s"
  version        = "7.10.1"
  extension_type = "plugin"
  download_url   = "%s"
}
