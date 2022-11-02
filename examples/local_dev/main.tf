terraform {
  required_providers {
    goplugin = {
      source = "slok.dev/tf/goplugin"
    }
  }
}

provider "goplugin" {
  resource_plugins_v1 = {
    "os_file" : {
      factory_name = "NewFileResourcePlugin"
      source_code = {
        git = {
          url  = "https://github.com/slok/terraform-go-plugins"
          auth = {} // Load `GOPLUGIN_GIT_PASSWORD`.
        }
      }
      configuration = jsonencode({})
    }
  }
  data_source_plugins_v1 = {
    "murmur3" : {
      source_code = {
        dir = "../with_dependencies/plugins/murmur3"
      }
      configuration = jsonencode({})
    }
  }
}


locals {
  files = {
    "/tmp/file1.txt" : "This is the file content for file1"
    "/tmp/file2.txt" : "This is the file content for file2"
    "/tmp/file3.txt" : "This is the file content for file3"
  }
}

resource "goplugin_plugin_v1" "os_file_test_files" {
  for_each = local.files

  plugin_id = "os_file"
  attributes = jsonencode({
    path    = each.key
    content = each.value
    mode    = 644
  })
}

data "goplugin_plugin_v1" "murmur3_hash_test_files" {
  for_each = goplugin_plugin_v1.os_file_test_files

  plugin_id = "murmur3"
  attributes = jsonencode({
    value = jsondecode(goplugin_plugin_v1.os_file_test_files[each.key].attributes).content
  })
}

output "test_files_hash" {
  value = {
    for k, hash in data.goplugin_plugin_v1.murmur3_hash_test_files : k => jsondecode(hash.result)
  }
}
