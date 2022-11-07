# Changelog

## [Unreleased]

## [v0.5.1] - 2022-11-07

### Fixed

- Git source code plugins using `dir` option now will load correctly the module instead of failing always.

## [v0.5.0] - 2022-11-02

## Breaking

- File based plugins have been removed in favor of go modules.
- Plugins must be directories and be a valid go module (`go.mod`).

### Added

- 3rd party dependencies support using `vendor` directory.

## [v0.4.0] - 2022-08-24

### Breaking

- `NewTestResourcePlugin` now receives a config struct as parameters with and will set safe defaults.
- `NewTestDataSourcePlugin` now receives a config struct as parameters with and will set safe defaults.

### Added

- Git authentication using basic auth on plugin's source code.

## [v0.3.0] - 2022-08-23

### Beaking

- On `plugin_v1` resources, `resource_data` renamed to `attributes`.
- On `plugin_v1` data source, `arguments` renamed to `attributes`.

### Added

- Plugins can customize the name of the plugin factory that the plugin engine will use to make plugin instances.

## [v0.2.0] - 2022-08-21

### Added

- `resource_id` attribute (The one without the plugin ID).
- `plugin_v1` data source.
- V1 data source plugin engine.
- Data source plugin v1 lib including testing utils.

## [v0.1.1] - 2022-08-18

### Fixed

- Terraform import docs.

## [v0.1.0] - 2022-08-18

### Added

- The porovider itself.
- `plugin_v1` resource.
- V1 Resource Plugin engine.
- Resource plugin v1 lib including testing utils.
- Allow loading source code from Raw data.
- Allow loading source code from Git repository.
- File example.
- Remote plugin (using Git) example.
- Github gist example.
- Documentation.

[unreleased]: https://github.com/slok/terraform-provider-goplugin/compare/v0.5.1...HEAD
[v0.5.1]: https://github.com/slok/terraform-provider-goplugin/compare/v0.5.0...v0.5.1
[v0.5.0]: https://github.com/slok/terraform-provider-goplugin/compare/v0.4.0...v0.5.0
[v0.4.0]: https://github.com/slok/terraform-provider-goplugin/compare/v0.3.0...v0.4.0
[v0.3.0]: https://github.com/slok/terraform-provider-goplugin/compare/v0.2.0...v0.3.0
[v0.2.0]: https://github.com/slok/terraform-provider-goplugin/compare/v0.1.1...v0.2.0
[v0.1.1]: https://github.com/slok/terraform-provider-goplugin/compare/v0.1.0...v0.1.1
[v0.1.0]: https://github.com/slok/terraform-provider-goplugin/releases/tag/v0.1.0
