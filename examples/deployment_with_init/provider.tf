terraform {
  required_version = ">= 1.0"

  required_providers {
    ec = {
      source  = "elastic/ec"
      version = "0.12.4"
    }
  }
}

provider "ec" {
  apikey = "<api key>"
}

