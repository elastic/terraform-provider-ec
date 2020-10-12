terraform {
  required_version = ">= 0.12"

  required_providers {
    ec = {
      source = "elastic/ec"
    }
  }
}

provider "ec" {
  apikey = "azFXa0pIUUJaeWRjMmUzVXpCOG06VUJ1NEtFREtSNE9BMThxeG9mM1A3dw=="
}

