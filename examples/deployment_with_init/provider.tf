terraform {
  required_version = ">= 0.12"

  required_providers {
    ec = {
      source = "elastic/ec"
    }
  }
}

provider "ec" {
  apikey = "<api key>"
}

