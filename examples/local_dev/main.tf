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
  resource_data = jsonencode({
    path = each.key
    content = each.value
    mode = 644
  })
}
