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
      version = "0.6.0"
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
  elasticsearch = {
    hot = {
      autoscaling = {}
    }
  }

  kibana = {
    topology = {}
  }
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

### Moving to TF Framework and schema change for `ec_deployment` resource.

v6.0.0 contains migration to [the TF Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework) and intoducing version 2 for `ec_deployment` resource:

- switching to attributes syntax instead of blocks for almost all definitions that used to be blocks. It means that, for example, a definition like `config {}` has to be changed to `config = {}`, e.g.

```hcl
resource "ec_deployment" "defaults" {
  name                   = "example"
  region                 = "us-east-1"
  version                = data.ec_stack.latest.version
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch = {
    hot = {
      autoscaling = {}
    }
  }

  kibana = {
    topology = {}
  }

  enterprise_search = {
    zone_count = 1
  }
}
```

- `topology` attribute of `elasticsearch` is replaced with a number of dedicated attributes, one per tier, e.g. 

```
  elasticsearch {
    topology {
      id         = "hot_content"
      size       = "1g"
      autoscaling {
        max_size = "8g"
      }
    }
    topology {
      id         = "warm"
      size       = "2g"
      autoscaling {
        max_size = "15g"
      }
    }
  }
```

has to be converted to 

```
  elasticsearch = {
    hot = {
      size = "1g"
      autoscaling = {
        max_size = "8g"
      }
    }

    warm = {
      size = "2g"
      autoscaling = {
        max_size = "15g"
      }
    }
  }

```

- due to some existing limitations of TF, nested attributes that are nested inside other nested attributes cannot be `Computed`. It means that all such attributes have to be mentioned in configurations even if they are empty. E.g., a definition of `elasticsearch` has to include all topology elements (tiers) that have non-zero size or can be scaled up (if autoscaling is enabled) in the corresponding template. For example, the simplest definition of `elasticsearch` for `aws-io-optimized-v2` template is 

```hcl
resource "ec_deployment" "defaults" {
  name                   = "example"
  region                 = "us-east-1"
  version                = data.ec_stack.latest.version
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch = {
    hot = {
      autoscaling = {}
    }
  }
}
```

Please note that the configuration explicitly mentions `hot` tier and the tier has `autoscaling` and `config` attributes even despite the fact that they are empty. If they were omitted, TF (at least up to version 1.3.3) could complain `Error: Provider produced inconsistent result after apply`.

- a lot of attributes that used to be collections (e.g. lists and sets) are converted to sigletons, e.g. `elasticsearch`, `apm`, `kibana`, `enterprise_search`, `observability`, `topology`, `autoscaling`, etc. Please note that, generally, users are not expected to make any change to their existing configuration to address this particular change (besides moving from block to attribute syntax). All these components used to exist in single instances, so the change is mostly syntactical, taking into account the switch to attributes instead of blocks (otherwise if we kept list for configs,  `config {}` had to be rewritten in `config = [{}]` with the move to the attribute syntax). However this change is a breaking one from the schema perspective and requires state upgrade for existing resources that is performed by TF (by calling the provider's API).

- [`strategy` attribute](https://registry.terraform.io/providers/elastic/ec/latest/docs/resources/ec_deployment#strategy) is converted to string with the same set of values that was used for its `type` attribute previously;

- switching to TF protocol 6. From user perspective it should not require any change in their existing configurations.

#### Migration guide.

The schema modifications means that a current TF state cannot work as is with the provider version 0.6.0 and higher.

There are 2 ways to tackle this

- import existing resource using deployment ID, e.g `terraform import 'ec_deployment.test' <deployment_id>`
- state upgrade that is performed by TF by calling the provider's API so no action is required from user perspective

Currently the state upgrade functionality is still in development so importing existing resources is the recommended way to deal with existing TF states.

#### Known issues.

For the migrated version (0.6.0 or higher), `terraform plan` output can contain more changes comparing to the older versions of the provider (that use TF SDK). 
This happens because TF Framework treats all `computed` attributes as `unknown` (known after apply) once configuration changes.
`ec_deployment` schema contains quite a few of such attributes, so `terraform plan`'s output can be quite big for the resource due to the mentioned reason.
However, it doesn't mean that all attributes that marked as `unknown` in the plan will get new values after apply.
To mitigitate the problem, the provider uses plan modifiers that is a recommended way to reduce plan output. 
However, currently plan modifiers don't cover the all `computed` attributes.
