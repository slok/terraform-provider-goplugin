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
      source_code = [for f in fileset("./", "plugins/resource_os_file/*"): file(f)]
      configuration =  jsonencode({})
    }
  }
}

locals {
    files = {
        "/tmp/tf-test1.txt": "test-1"
        "/tmp/tf-test2.txt": "test-2"
        "/tmp/tf-test3.txt": "test-3"
        "/tmp/tf-test4.txt": "test-4"
        "/tmp/tf-test5.txt": "test-5"
        "/tmp/tf-test6.txt": "test-6"
        "/tmp/tf-test7.txt": "test-7"
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
