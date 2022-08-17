terraform {
  required_providers {
    goplugin = {
      source = "slok/goplugin"
    }
  }
}

# Load all plugins.
provider goplugin { 
  resource_plugins_v1 = {
    "github_gist": {
      source_code = [for f in fileset("./", "plugins/resource_gist/*"): file(f)]
      configuration =  jsonencode({
        api_url = "https://api.github.com"
      })
    }
  }
}

# Use the plugin.
resource "goplugin_plugin_v1" "github_gist_test" {  
  plugin_id = "github_gist"

  resource_data = jsonencode({
    description = "Managed by terraform."
    public = true
    files = {
      "test-goplugin1.txt": "Hello"
      "test-goplugin2.txt": "World"
    }
  })
}
