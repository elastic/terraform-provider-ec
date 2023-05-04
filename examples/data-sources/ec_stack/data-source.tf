data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "us-east-1"
  lock          = true
}

data "ec_stack" "latest_patch" {
  version_regex = "7.9.?"
  region        = "us-east-1"
}
