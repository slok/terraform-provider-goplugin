terraform {
  required_providers {
    goplugin = {
      source = "slok.dev/tf/goplugin"
    }
  }
}

provider goplugin { 
  resource_plugins_v1 = {
    "os_file": {
      source_code = {
        data = [for f in fileset("./", "../os_file/plugins/resource_os_file/*"): file(f)]
        # git = {
        #   url = "https://github.com/slok/terraform-provider-goplugin"
        #   paths_regex = ["examples/os_file/plugins/.*\\.go"]
        # } 
      }
      configuration =  jsonencode({})
    }
  }
    data_source_plugins_v1 = {
    "os_file": {
      source_code = {
        data = [for f in fileset("./", "../os_file/plugins/resource_os_file/*"): file(f)]
        # git = {
        #   url = "https://github.com/slok/terraform-provider-goplugin"
        #   paths_regex = ["examples/os_file/plugins/.*\\.go"]
        # } 
      }
      configuration =  jsonencode({})
    }
  }
}

locals {
    files = {
        "/tmp/tf-local-dev1.txt": "test-1"
        "/tmp/tf-local-dev2.txt": "test-2"
        "/tmp/tf-local-dev3.txt": "test-3"
    }
}

resource "goplugin_plugin_v1" "os_file_test" {
  for_each = local.files
  
  plugin_id = "os_file"
  attributes = jsonencode({
    path = each.key
    content = each.value
    mode = 644
  })
}
