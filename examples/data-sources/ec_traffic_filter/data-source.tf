data "ec_trafficfilter" "name" {
  name = "example-filter"
}

data "ec_trafficfilter" "id" {
  id = "41d275439f884ce89359039e53eac516"
}

data "ec_trafficfilter" "region" {
  region = "us-east-1"
}
