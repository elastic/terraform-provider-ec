---
page_title: "Elastic Cloud: ec_extension"
description: |-
  Provides an Elastic Cloud extension resource, which allows extension to be created, updated, and deleted.
---

# Resource: ec_extension
Provides an Elastic Cloud extension resource, which allows extension to be created, updated, and deleted.

## Example Usage
### with extension file

```hcl
locals {
  file_path = "/path/to/plugin.zip"
}

resource "ec_extension" "example_extension" {
  name           = "my_extension"
  description    = "my extension"
  version        = "*"
  extension_type = "bundle"

  file_path = local.file_path
  file_hash = filebase64sha256(local.file_path)
}
```

### with download URL
```hcl
resource "ec_extension" "example_extension" {
  name           = "my_extension"
  description    = "my extension"
  version        = "*"
  extension_type = "bundle"
  download_url   = "https://example.net"
}
```

## Argument Reference
The following arguments are supported:

* `name` - (Required) Name of the extension. 
* `description` - (Optional) Description of the extension.
* `extension_type` - (Required) `bundle` or `plugin` allowed.
* `version` - (Required) Elastic version.
* `download_url` - (Optional) The URL to download the extension archive.
* `file_path` - (Optional) File path of the extension uploaded.
* `file_hash` - (Optional) Hash value of the file. If it is changed, the file is reuploaded. 


## Attributes Reference
In addition to all the arguments above, the following attributes are exported:

* `id` - Extension identifier.
* `url` - The extension URL to be used in the plan.
* `last_modified` - The datetime the extension was last modified.
* `size` - The extension file size in bytes.
