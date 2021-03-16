# Terraform Provider for Elastic Cloud

![Go](https://github.com/elastic/terraform-provider-ec/workflows/Go/badge.svg?branch=master)
[![Acceptance Status](https://devops-ci.elastic.co/job/elastic+terraform-provider-ec+master/badge/icon?subject=acceptance&style=plastic)](https://devops-ci.elastic.co/job/elastic+terraform-provider-ec+master/)

#### This project is currently under active development. Source code is provided with no assurances, use at your own risk.

Terraform provider for the Elastic Cloud API, including:

* Elasticsearch Service (ESS).
* Elastic Cloud Enterprise (ECE).
* Elasticsearch Service Private (ESSP).

## Example usage

```hcl
terraform {
  required_version = ">= 0.12.29"

  required_providers {
    ec = {
      source  = "elastic/ec"
      version = "0.1.0"
    }
  }
}

provider "ec" {
  # ECE installation endpoint
  endpoint = "https://my.ece-environment.corp"

  # If the ECE installation has a self-signed certificate
  # setting "insecure" to true is required.
  insecure = true

  # APIKey is the recommended authentication mechanism. When
  # Targeting the Elasticsearch Service, APIKeys are the only
  # valid authentication mechanism.
  apikey = "my-apikey"

  # When targeting ECE installations, username and password
  # authentication is allowed.
  username = "my-username"
  password = "my-password"
}


# Create an Elastic Cloud deployment
resource "ec_deployment" "example_minimal" {
  # Optional name.
  name = "my_example_deployment"

  # Mandatory fields
  region                 = "us-east-1"
  version                = "7.9.2"
  deployment_template_id = "aws-io-optimized-v2"

  # Use the deployment template defaults
  elasticsearch {}

  kibana {}
}
```

You can find the full resource documentation in the [terraform-provider-ec/docs/](https://github.com/elastic/terraform-provider-ec/tree/master/docs) repo.

## Developer Requirements

- [Terraform](https://www.terraform.io/downloads.html) 0.13+
- [Go](https://golang.org/doc/install) 1.13+ (to build the provider plugin)

### Installing the provider via the source code

Clone the repository to a folder on your machine and run `make install`:

```sh
$ mkdir -p ~/development; cd ~/development
$ git clone https://github.com/elastic/terraform-provider-ec
$ cd terraform-provider-ec
$ make install
```

### Generating an Elasticsearch Service API Key

To generate an API key, follow these steps:

  1. Navigate to <https://cloud.elastic.co/login> with your browser
  2. Log in with your Email and Password.
  3. Click on [Elasticsearch Service](https://cloud.elastic.co/deployments).
  4. Navigate to [Account > API Keys](https://cloud.elastic.co/account/keys) and click on **Generate API Key**.
  5. Once you Re-Authenticate, you'll have to chose a name for your API key.
  6. Copy your API key somewhere safe.

### Using your API Key on the Elastic Cloud terraform provider

After you've generated your API Key, you can make it available to the Terraform provider by exporting it as an environment variable:

```sh
$ export EC_API_KEY="<apikey value>"
```

After doing so, you can navigate to any of our examples in `./examples` and try one.
