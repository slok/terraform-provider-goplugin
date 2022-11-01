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
    "os_file" : {
      source_code = {
        dir = "../os_file/plugins/resource_os_file/"
      }
      configuration = jsonencode({})
    },

    "murmur3" : {
      source_code = {
        dir = "../with_dependencies/plugins/murmur3"
      }
      configuration = jsonencode({})
    }

    "ulid" : {
      source_code = {
        dir = "../with_dependencies/plugins/ulid"
      }
      configuration = jsonencode({})
    }
  }
}

locals {
  files = {
    "/tmp/tf-local-dev1.txt" : "test-1"
    "/tmp/tf-local-dev2.txt" : "test-2"
    "/tmp/tf-local-dev3.txt" : "test-3"
  }
}

resource "goplugin_plugin_v1" "os_file_test" {
  for_each = local.files

  plugin_id = "os_file"
  attributes = jsonencode({
    path    = each.key
    content = each.value
    mode    = 644
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


data "goplugin_plugin_v1" "murmur_test" {
  plugin_id = "murmur3"
  attributes = jsonencode({
    value = "testing the hashes"
  })
}

output "test_murmur" {
  value = jsondecode(data.goplugin_plugin_v1.murmur_test.result)
}


data "goplugin_plugin_v1" "ulid_test" {
  plugin_id = "ulid"
  attributes = jsonencode({
    value = "testing the hashes"
  })
}

output "test_ulid" {
  value = jsondecode(data.goplugin_plugin_v1.ulid_test.result)
}

