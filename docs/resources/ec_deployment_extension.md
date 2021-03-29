---
page_title: "Elastic Cloud: ec_deployment_extension"
description: |-
  Provides an Elastic Cloud extension resource, which allows extensions to be created, updated, and deleted.
---

# Resource: ec_deployment_extension
Provides an Elastic Cloud extension resource, which allows extensions to be created, updated, and deleted.

Extensions allow users of Elastic Cloud to use custom plugins, scripts, or dictionaries to enhance the core functionality of Elasticsearch. Before you install an extension, be sure to check out the supported and official [Elasticsearch plugins](https://www.elastic.co/guide/en/elasticsearch/plugins/current/index.html) already available.

## Example Usage
### With extension file

```hcl
locals {
  file_path = "/path/to/plugin.zip"
}

resource "ec_deployment_extension" "example_extension" {
  name           = "my_extension"
  description    = "my extension"
  version        = "*"
  extension_type = "bundle"

  file_path = local.file_path
  file_hash = filebase64sha256(local.file_path)
}
```

### With download URL
```hcl
resource "ec_deployment_extension" "example_extension" {
  name           = "my_extension"
  description    = "my extension"
  version        = "*"
  extension_type = "bundle"
  download_url   = "https://example.net"
}
```

### Using extension in ec_deployment
```hcl
resource "ec_deployment_extension" "example_extension" {
  name           = "my_extension"
  description    = "my extension"
  version        = "*"
  extension_type = "bundle"
  download_url   = "https://example.net"
}

data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "us-east-1"
}

resource "ec_deployment" "with_extension" {
  # Optional name.
  name = "my_example_deployment"

  # Mandatory fields
  region                 = "us-east-1"
  version                = data.ec_stack.latest.version
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch {
    extension {
      name    = ec_deployment_extension.example_extension.name
      type    = "bundle"
      version = data.ec_stack.latest.version
      url     = ec_deployment_extension.example_extension.url
    }
  }
}
```

## Argument Reference
The following arguments are supported:

* `name` - (Required) Name of the extension. 
* `description` - (Optional) Description of the extension.
* `extension_type` - (Required) `bundle` or `plugin` allowed. A `bundle` will usually contain a dictionary or script, where a `plugin` is compiled from source.
* `version` - (Required) Elastic stack version, a numeric version for plugins, e.g. 2.3.0 should be set. Major version e.g. 2.*, or wildcards e.g. * for bundles.
* `download_url` - (Optional) The URL to download the extension archive.
* `file_path` - (Optional) File path of the extension uploaded.
* `file_hash` - (Optional) Hash value of the file. If it is changed, the file is reuploaded. 


## Attributes Reference
In addition to all the arguments above, the following attributes are exported:

* `id` - Extension identifier.
* `url` - The extension URL to be used in the plan.
* `last_modified` - The datetime the extension was last modified.
* `size` - The extension file size in bytes.

## Import

You can import extensions using the `id`, for example:

```
$ terraform import ec_deployment_extension.name 320b7b540dfc967a7a649c18e2fce4ed
```
