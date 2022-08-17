# We have a plugin loaded that has been registered as `os_file`.
# This plugin manages files in the system.
resource "goplugin_plugin_v1" "os_file_test" {
  plugin_id = "os_file"

  resource_data = jsonencode({
    path = "/tmp/hello-world.txt"
    content = "Hello world!"
    mode = 644
  })
}
