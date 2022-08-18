
locals {
    files = {
        "tf-test1.txt": { content: "Hello world: 'test-1'", public: true}
        "tf-test2.txt": { content: "Hello world: 'test-2'", public: true}
        "tf-test3.txt": { content: "Hello world: 'test-3'", public: true}
        "tf-custom.txt":   { content: "Hello world: 'custom'", public: true}
    }
}

resource "goplugin_plugin_v1" "gist_test" {  
  for_each = local.files

  plugin_id = "github_gist"
  resource_data = jsonencode({
    description = "Managed by terraform."
    public = each.value.public
    files = {
      "${each.key}": each.value.content
    }
  })
}
