# Terraform Provider for Elastic Cloud

![Go](https://github.com/elastic/terraform-provider-ec/workflows/Go/badge.svg?branch=master)
[![Acceptance Status](https://devops-ci.elastic.co/job/elastic+terraform-provider-ec+master/badge/icon?subject=acceptance&style=plastic)](https://devops-ci.elastic.co/job/elastic+terraform-provider-ec+master/)

Terraform provider for the Elastic Cloud API, including:

* Elasticsearch Service (ESS).
* Elastic Cloud Enterprise (ECE).
* Elasticsearch Service Private (ESSP).

_Model changes might be introduced between minors until version 1.0.0 is released. Such changes and the expected impact will be detailed in the change log and the individual release notes._

## Terraform provider scope

The goal for a Terraform provider is to orchestrate lifecycle for deployments via common set of APIs across ESS, ESSP and ECE (see https://www.elastic.co/guide/en/cloud/current/ec-restful-api.html for API examples)

Things which are out of scope for provider:
- Configuring individual Elastic Stack components (Elasticsearch, Kibana, etc)
- Configuring snapshots settings for deployment (since they are using Elasticsearch SLM for this now, see https://www.elastic.co/guide/en/elasticsearch/reference/current/snapshot-lifecycle-management.html)

We now have Terraform provider for Elastic Stack https://github.com/elastic/terraform-provider-elasticstack which should be used for any operations on Elastic Stack products.

## Example usage

_These examples are forward looking and might use an unreleased version, for a current view of working examples, please refer to the [Terraform registry documentation](https://registry.terraform.io/providers/elastic/ec/latest/docs)._

```hcl
terraform {
  required_version = ">= 0.12.29"

  required_providers {
    ec = {
      source  = "elastic/ec"
      version = "0.4.0"
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

data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "us-east-1"
}

# Create an Elastic Cloud deployment
resource "ec_deployment" "example_minimal" {
  # Optional name.
  name = "my_example_deployment"

  # Mandatory fields
  region                 = "us-east-1"
  version                = data.ec_stack.latest.version
  deployment_template_id = "aws-io-optimized-v2"

  # Use the deployment template defaults
  elasticsearch {}

  kibana {}
}
```

## Developer Requirements

- [Terraform](https://www.terraform.io/downloads.html) 0.13+
- [Go](https://golang.org/doc/install) 1.16+ (to build the provider plugin)

### Installing the provider via the source code

Clone the repository to a folder on your machine and run `make install`:

```sh
$ mkdir -p ~/development; cd ~/development
$ git clone https://github.com/elastic/terraform-provider-ec
$ cd terraform-provider-ec
$ make install
```

### Generating an Elasticsearch Service (ESS) API Key

To generate an API key, follow these steps:

  1. Open your browser and navigate to <https://cloud.elastic.co/login>.
  2. Log in with your email and password.
  3. Click on [Elasticsearch Service](https://cloud.elastic.co/deployments).
  4. Navigate to [Features > API Keys](https://cloud.elastic.co/deployment-features/keys) and click on **Generate API Key**.
  5. Choose a name for your API key.
  6. Save your API key somewhere safe.

### Using your API Key on the Elastic Cloud terraform provider

After you've generated your API Key, you can make it available to the Terraform provider by exporting it as an environment variable:

```sh
$ export EC_API_KEY="<apikey value>"
```

After doing so, you can navigate to any of our examples in `./examples` and try one.
