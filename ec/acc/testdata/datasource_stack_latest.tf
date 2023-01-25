data "ec_stack" "latest" {
  version_regex = "latest"
  lock          = true
  region        = "%s"
}
