terraform {
  required_providers {
    goplugin = {
      source = "slok/goplugin"
    }
  }
}

// Load all required plugins on the provider.
provider goplugin { 
  data_source_plugins_v1 = {
    "os_file" : {
      source_code = {
        data = [for f in fileset("./", "../os_file/plugins/resource_os_file/*"): file(f)]
      }
      configuration = jsonencode({})
    }
  }

  resource_plugins_v1 = {
    "os_file": {
      source_code = {
        data = [for f in fileset("./", "../os_file/plugins/resource_os_file/*"): file(f)]
      }
      configuration =  jsonencode({})
    }

    "github_gist": {
      source_code = {
        git = {
          url = "https://github.com/slok/terraform-provider-goplugin"
          paths_regex = ["examples/github_gist/plugins/.*\\.go"]
        } 
      }
      configuration =  jsonencode({
         // TF_GITHUB_TOKEN loaded from env var.
         api_url = "https://api.github.com"
      })
    }
  }
}

locals {
    files = {
        "tf-test1.txt": "test-1"
        "tf-test2.txt": "test-2"
        "tf-test3.txt": "test-3"
    }
}

// Use the plugins.
resource "goplugin_plugin_v1" "os_file_test" {
  for_each = local.files
  
  plugin_id = "os_file"
  attributes = jsonencode({
    path = "/tmp/${each.key}"
    content = each.value
    mode = 644
  })
}

resource "goplugin_plugin_v1" "github_gist_test" {
  for_each = local.files
  
  plugin_id = "github_gist"
  attributes = jsonencode({
    description = "Managed by terraform."
    public = true
    files = {"${each.key}": each.value}
  })
}

data "goplugin_plugin_v1" "os_file_test" {
  for_each = goplugin_plugin_v1.os_file_test

  plugin_id = "os_file"
  attributes = jsonencode({
    path = jsondecode(each.value.attributes).path
  })
}

output "test" {
  value = { for k, v in data.goplugin_plugin_v1.os_file_test : k => jsondecode(v.result) }
}

