terraform {
  required_providers {
    goplugin = {
      source = "slok/goplugin"
    }
  }
}

provider goplugin { 
  resource_plugins_v1 = {
    "github_gist": {
      source_code = {
        data = [for f in fileset("./", "plugins/resource_gist/*"): file(f)]
      }
      configuration =  jsonencode({
        // gh token loaded from `TF_GITHUB_TOKEN`
      })
    }
  }
}
