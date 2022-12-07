---
page_title: "Elastic Cloud: ec_deployment_elasticsearch_keystore"
description: |-
  Provides an Elastic Cloud Deployment Elasticsearch keystore resource, which allows creating and updating Elasticsearch Keystore  settings.
---

# Resource: ec_deployment_elasticsearch_keystore
Provides an Elastic Cloud Deployment Elasticsearch keystore resource, which allows you to create and update Elasticsearch keystore settings.

Elasticsearch keystore settings can be created and updated through this resource, **each resource represents a single Elasticsearch Keystore setting**. After adding a key and its secret value to the keystore, you can use the key in place of the secret value when you configure sensitive settings.

~> **Note on Elastic keystore settings** This resource offers weaker consistency guarantees and will not detect and update keystore setting values that have been modified outside of the scope of Terraform, usually referred to as _drift_. For example, consider the following scenario:
    1. A keystore setting is created using this resource.
    2. The keystore setting's value is modified to a different value using the Elasticsearch Service API.
    3. Running `terraform apply` fails to detect the changes and does not update the keystore setting to the value defined in the terraform configuration.
  To force the keystore setting to the value it is configured to hold, you may want to taint the resource and force its recreation.

Before you create Elasticsearch keystore settings, check the [official Elasticsearch keystore documentation](https://www.elastic.co/guide/en/elasticsearch/reference/master/elasticsearch-keystore.html) and the [Elastic Cloud specific documentation](https://www.elastic.co/guide/en/cloud/current/ec-configuring-keystore.html).

## Example Usage

These examples show how to use the resource at a basic level, and can be copied. This resource becomes really useful when combined with other data providers, like vault or similar.

### Adding a new keystore setting to your deployment

```hcl
data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "us-east-1"
}

# Create an Elastic Cloud deployment
resource "ec_deployment" "example_keystore" {
  region                 = "us-east-1"
  version                = data.ec_stack.latest.version
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch {}
}

# Create the keystore secret entry
resource "ec_deployment_elasticsearch_keystore" "secure_url" {
  deployment_id = ec_deployment.example_keystore.id
  setting_name  = "xpack.notification.slack.account.hello.secure_url"
  value         = "http://my-secure-url.com"
}

```

### Adding credentials to use GCS as a snapshot repository

For up-to-date documentation on the `setting_name`, refer to the [ESS documentation](https://www.elastic.co/guide/en/cloud/current/ec-gcs-snapshotting.html#ec-gcs-service-account-key).

```hcl
data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "us-east-1"
}

# Create an Elastic Cloud deployment
resource "ec_deployment" "example_keystore" {
  region                 = "us-east-1"
  version                = data.ec_stack.latest.version
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch = {
    hot = {
      autoscaling = {}
    }
  }
}

# Create the keystore secret entry
resource "ec_deployment_elasticsearch_keystore" "gcs_credential" {
  deployment_id = ec_deployment.example_keystore.id
  setting_name  = "gcs.client.default.credentials_file"
  value         = file("service-account-key.json")
  as_file       = true
}
```

## Argument reference
The following arguments are supported:

* `deployment_id` - (Required) Deployment ID of the deployment that holds the Elasticsearch cluster where the keystore setting is written to. 
* `setting_name` - (Required) Required name for the keystore setting, if the setting already exists in the Elasticsearch cluster, it will be overridden.
* `value` - (Required) Value of this setting. This can either be a string or a JSON object that is stored as a JSON string in the keystore.
* `as_file` - (Optional) if set to `true`, it stores the remote keystore setting as a file. The default value is `false`, which stores the keystore setting as string when value is a plain string.


## Attributes reference

There are no additional attributes exported by this resource other than the referenced arguments.

## Import

This resource cannot be imported.
