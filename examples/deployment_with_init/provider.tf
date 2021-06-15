terraform {
  required_version = ">= 0.12.29"

  required_providers {
    ec = {
      source  = "elastic/ec"
      version = "0.2.1"
    }
  }
}

provider "ec" {
  apikey = "<api key>"
}

