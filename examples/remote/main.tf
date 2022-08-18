terraform {
  required_providers {
    goplugin = {
      source = "slok/goplugin"
    }
  }
}

provider goplugin { 
  resource_plugins_v1 = {
    "os_file": {
      source_code = {
        git = {
          url = "https://github.com/slok/terraform-provider-goplugin"
          paths_regex = ["examples/os_file/plugins/.*\\.go"]
        } 
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
      configuration =  jsonencode({}) // TF_GITHUB_TOKEN from env var.
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

resource "goplugin_plugin_v1" "os_file_test" {
  for_each = local.files
  
  plugin_id = "os_file"
  resource_data = jsonencode({
    path = "/tmp/${each.key}"
    content = each.value
    mode = 644
  })
}

resource "goplugin_plugin_v1" "github_gist_test" {
  for_each = local.files
  
  plugin_id = "github_gist"
  resource_data = jsonencode({
    description = "Managed by terraform."
    public = true
    files = {"${each.key}": each.value}
  })
}
