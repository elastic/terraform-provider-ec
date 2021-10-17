terraform {
  required_version = ">= 0.12.29"

  required_providers {
    ec = {
      source  = "elastic/ec"
      version = "v0.4.0"
    }
  }
}

provider "ec" {
  apikey = "<api key>"
}

