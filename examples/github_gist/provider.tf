terraform {
  required_providers {
    goplugin = {
      #source = "slok/goplugin"
      source = "slok.dev/tf/goplugin"
    }
  }
}

provider "goplugin" {
  resource_plugins_v1 = {
    "github_gist" : {
      source_code   = { dir = "plugins/resource_gist" }
      configuration = jsonencode({}) // gh token loaded from `TF_GITHUB_TOKEN`
    }
  }
}
