# Import schema.
terraform import goplugin_plugin_v1.myresource ${PLUGIN_ID}/${RESOURCE_ID}

# Example:
# - We have a plugin registered as `github_gist`.
# - We have a resource managed by this plugin with the ID `04fb1aafb3074664089990ea793cb245`.
terraform import goplugin_plugin_v1.myresource github_gist/04fb1aafb3074664089990ea793cb245

