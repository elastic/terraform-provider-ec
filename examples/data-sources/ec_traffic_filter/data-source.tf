data "ec_traffic_filter" "name" {
  name = "example-filter"
}

data "ec_traffic_filter" "id" {
  id = "41d275439f884ce89359039e53eac516"
}

data "ec_traffic_filter" "region" {
  region = "us-east-1"
}
