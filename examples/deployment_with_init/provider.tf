terraform {
  required_version = ">= 1.0"

  required_providers {
    ec = {
      source  = "elastic/ec"
      version = "0.6.0"
    }
  }
}

provider "ec" {
  apikey = "<api key>"
}

