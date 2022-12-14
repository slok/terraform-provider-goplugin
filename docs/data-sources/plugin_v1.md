---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "goplugin_plugin_v1 Data Source - terraform-provider-goplugin"
subcategory: ""
description: |-
  Executes a Data source Go plugin v1.
  The requirements for a plugin are:
  Written in Go.No external dependencies, only Go standard library.Implemented in a single package.Implement the plugin v1 API https://github.com/slok/terraform-provider-goplugin/tree/main/pkg/api/v1.
  Check examples https://github.com/slok/terraform-provider-goplugin/tree/main/examples
---

# goplugin_plugin_v1 (Data Source)

Executes a Data source Go plugin v1.

The requirements for a plugin are:

- Written in Go.
- No external dependencies, only Go standard library.
- Implemented in a single package.
- Implement the [plugin v1 API](https://github.com/slok/terraform-provider-goplugin/tree/main/pkg/api/v1).

Check [examples](https://github.com/slok/terraform-provider-goplugin/tree/main/examples)

## Example Usage

```terraform
# We have a datasource plugin loaded that has
# been registered as `os_file`.
# This plugin gets file information from the sytesm.
data "goplugin_plugin_v1" "os_file_test" {
  plugin_id = "os_file"
  attributes = jsonencode({
    path = "/tmp/hello-world.txt"
  })
}

output "test" {
   value = data.goplugin_plugin_v1.os_file_test.result
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `attributes` (String) A JSON string object with the properties that will be passed to the data source
							  plugin, the plugin is responsible of knowing how to load and use these properties.
- `plugin_id` (String) The ID of the data source plugin to use, must be loaded and registered by the provider.

### Read-Only

- `id` (String) Not used (Used internally by the provider and Terraform), can be ignored.
- `result` (String) A JSON string object with the plugin result.


