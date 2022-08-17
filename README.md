# terraform-provider-goplugin

[![CI](https://github.com/slok/terraform-provider-goplugin/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/slok/terraform-provider-goplugin/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/slok/terraform-provider-goplugin)](https://goreportcard.com/report/github.com/slok/terraform-provider-goplugin)
[![Apache 2 licensed](https://img.shields.io/badge/license-Apache2-blue.svg)](https://raw.githubusercontent.com/slok/terraform-provider-goplugin/master/LICENSE)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/slok/terraform-provider-goplugin)](https://github.com/slok/terraform-provider-goplugin/releases/latest)
[![Terraform regsitry](https://img.shields.io/badge/Terraform-Registry-color=green?logo=Terraform&style=flat&color=5C4EE5&logoColor=white)](https://registry.terraform.io/providers/slok/goplugin/latest/docs)

A Terraform provider to create terraform providers 🤯, but easier and faster!

Terraform go plugin provider is a Terraform provider that will let you execute Go plugins (using [yaegi]) in terraform by implemeting a very simple and small Go API.

## Why

Sometimes I want to manage resources in Terraform that don't have a provider, however, creating a Terraform provider takes time and a lot of effort, including understanding low level concepts. So this poor resource will not end in Terraform.

Unless... in the cases where we don't need to manage tons of resources or its a simple API, A small go plugin would be enough to manage them in Terraform.

## Use cases

### When to use it

- A terraform provider doesn't have support of the resource you need (e.g [Github provider][gh-provider] doesn't have [gist] support).
- You want ot manage private/internal APIs with Terraform.
- You don't want/need to understand low level Terraform concepts.
- A simple plugin that communicates with an API and marshal/unmarshal JSON is enough for you.

### When NOT to use it

- You need performance, interpreted code will be less efficient and slower.
- Your provider is complex and with tons of resources.
- You need Go third party libraries (this smells like a complex use case).
- You need to provide official Terraform support for a product.

## Plugins `v1`

Here are the API [Go docs][godoc-v1]

### Resource

- You will need to implement [`NewResourcePlugin`][apiv1-factory-method-godoc] method.
- You will need to implement interface: [`ResourcePlugin`][apiv1-interface-godoc] interface.
- You may use [`NewTestResourcePlugin`][apiv1-testing-method-godoc] for writing tests of the plugin.

Example of a NOOP plugin:

```go
package terraform

import (
 "context"

 apiv1 "github.com/slok/terraform-provider-goplugin/pkg/api/v1"
)

func NewResourcePlugin(opts string) (apiv1.ResourcePlugin, error) {
 return plugin{}, nil
}

type plugin struct{}

func (p plugin) CreateResource(ctx context.Context, r apiv1.CreateResourceRequest) (*apiv1.CreateResourceResponse, error) {
 return &apiv1.CreateResourceResponse{}, nil
}

func (p plugin) ReadResource(ctx context.Context, r apiv1.ReadResourceRequest) (*apiv1.ReadResourceResponse, error) {
 return &apiv1.ReadResourceResponse{}, nil
}

func (p plugin) DeleteResource(ctx context.Context, r apiv1.DeleteResourceRequest) (*apiv1.DeleteResourceResponse, error) {
 return &apiv1.DeleteResourceResponse{}, nil
}

func (p plugin) UpdateResource(ctx context.Context, r apiv1.UpdateResourceRequest) (*apiv1.UpdateResourceResponse, error) {
 return &apiv1.UpdateResourceResponse{}, nil
}
```

## Requirements

- Terraform `>=1.x`.

## Design decisions

- String JSON as input data
  - The same we we communicate with APIs.
  - JSON is a first class citizen in Go and Terraform out of the box.
  - JSON is very dynamic, perfect for plugins that can receive/return anything as resource data/configuration.
- Avoided using result data ([computed]).
  - Computed data is very static, goes against the dynamism of the Go plugins
  - Once stored, it is hard to change.
  - Is the root of lots of provider bugs and problems, so we avoid users creating those mistakes
  - Embrace data sources instead, less efficient (get real data instead of getting from state), but more reliable and flexible.
- No optimizations (we are already using interpreted code).
  - JSON marshall/unmarshall of input/output data
  - Use data-sources instead of computed fields
  - If optimization needed, create a regular Terraform provider.
- No support for external libraries.
  - More portable plugins.
  - Less Plugin engine bugs.
- Allow splitting code into multiple files
  - Easier to read and maintain.
  - Share helper utils between plugins.
- Plugins created at provider loading time.
- Simplicity and reliability by reducing features and setting conventions.
  - Small and simple API.
- Terraform cloud support.
  - When Plugins are loaded/created, these could download a binary if required.
- Ignore plugin tests (package `_test`) on plugin load
  - Avoid the user the need to ignore specific go files in a directory.
- A Resource Plugin ID is not mutable.
  - Changing requires resource recreation.
  - Used as part of the Terraform resource ID `{PLUGIN_ID}/{RESOURCE_ID}`.
  - Avoid pitfalls of reusing the same data with different plugins and end in a corrupt state.

## Terraform cloud

This provider supports terraform cloud.

## Development

To install your plugin locally you can do `make install`, it will build and install in your `${HOME}/.terraform/plugins/...`

Note: The installation is ready for `OS_ARCH=linux_amd64`, so you make need to change the [`Makefile`](./Makefile) if using other OS.

Example:

```bash
cd ./examples/local
rm -rf ./.terraform ./.terraform.lock.hcl
cd -
make install
cd -
terraform plan
```

[yaegi]: https://github.com/traefik/yaegi
[gh-provider]: https://registry.terraform.io/providers/integrations/github/latest/docs
[gist]: https://gist.github.com/
[computed]: https://www.terraform.io/plugin/sdkv2/schemas/schema-behaviors#computed
[godoc-v1]: https://pkg.go.dev/github.com/slok/terraform-provider-goplugin/pkg/api/v1
[apiv1-factory-method-godoc]: https://pkg.go.dev/github.com/slok/terraform-provider-goplugin/pkg/api/v1#NewResourcePlugin
[apiv1-interface-godoc]: https://pkg.go.dev/github.com/slok/terraform-provider-goplugin/pkg/api/v1#ResourcePlugin
[apiv1-testing-method-godoc]: https://pkg.go.dev/github.com/slok/terraform-provider-goplugin/pkg/api/v1/testing
