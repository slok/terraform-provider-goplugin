terraform {
  required_providers {
    goplugin = {
      source = "slok/goplugin"
    }
  }
}

provider "goplugin" {
  data_source_plugins_v1 = {
    "murmur3" : {
      source_code   = { dir = "./plugins/murmur3" }
      configuration = jsonencode({})
    }

    "ulid" : {
      source_code   = { dir = "./plugins/ulid" }
      configuration = jsonencode({})
    }
  }
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

