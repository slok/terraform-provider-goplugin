# We have a datasource plugin loaded that has
# been registered as `os_file`.
# This plugin gets file information from the sytesm.
data "goplugin_plugin_v1" "os_file_test" {
  for_each = goplugin_plugin_v1.os_file_test
  
  plugin_id = "os_file"
  arguments = jsonencode({
    path = "/tmp/hello-world.txt"
  })
}

output "test" {
   value = data.goplugin_plugin_v1.os_file_test.result
}
