# terraform-provider-goplugin

[![CI](https://github.com/slok/terraform-provider-goplugin/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/slok/terraform-provider-goplugin/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/slok/terraform-provider-goplugin)](https://goreportcard.com/report/github.com/slok/terraform-provider-goplugin)
[![Apache 2 licensed](https://img.shields.io/badge/license-Apache2-blue.svg)](https://raw.githubusercontent.com/slok/terraform-provider-goplugin/master/LICENSE)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/slok/terraform-provider-goplugin)](https://github.com/slok/terraform-provider-goplugin/releases/latest)
[![Terraform regsitry](https://img.shields.io/badge/Terraform-Registry-color=green?logo=Terraform&style=flat&color=5C4EE5&logoColor=white)](https://registry.terraform.io/providers/slok/goplugin/latest/docs)

A Terraform provider to create terraform providers 🤯, but easier and faster!

Terraform go plugin provider is a Terraform provider that will let you execute Go plugins (using [yaegi]) in terraform by implementing a very simple and small Go API.

## Features

- Implement Terraform providers using small Go plugins.
- Go Plugin code doesn't require compilation.
- Supports plugin 3rd party dependencies using Go `vendor` dir.
- compatible with [Terraform cloud](https://app.terraform.io/).

## Why

Sometimes I want to manage resources in Terraform that don't have a provider, however, creating a Terraform provider takes time and a lot of effort, including understanding low level concepts. So this poor resource will not end in Terraform.

Unless... in the cases where we don't need to manage tons of resources or its a simple API, A small go plugin would be enough to manage them in Terraform.

## Use cases

### When to use it

- A terraform provider doesn't have support of the resource you need (e.g [Github provider][gh-provider] doesn't have [gist] support).
- You want to manage private/internal APIs with Terraform.
- You don't want/need to understand low level Terraform concepts.
- A simple plugin that communicates with an API and marshal/unmarshal JSON is enough for you.
- Prototyping, MVPs and exploring ideas around terraform provider development.
- Implement private terraform providers for your company/organization

### When NOT to use it

- You need performance, interpreted code will be less efficient and slower.
- Your provider is complex and with tons of resources.
- You need to provide official Terraform support for a product.

## Examples

The best way of how to use it is by checking the [examples](./examples).

## Plugins `v1`

- API [Go docs][godoc-v1].
- [Examples][examples].

### Resource

- You will need to implement [`NewResourcePlugin`][resource-apiv1-factory-method-godoc] method.
- You will need to implement interface: [`ResourcePlugin`][resource-apiv1-interface-godoc] interface.
- You may use [`NewTestResourcePlugin`][apiv1-testing-godoc] for writing tests of the plugin.

Example of a NOOP resource plugin:

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

### Data source

- You will need to implement [`NewDataSourcePlugin`][data-source-apiv1-factory-method-godoc] method.
- You will need to implement interface: [`DataSourcePlugin`][data-source-apiv1-interface-godoc] interface.
- You may use [`NewTestDataSourcePlugin`][apiv1-testing-godoc] for writing tests of the plugin.

Example of a NOOP data source plugin:

```go
package terraform

import (
 "context"

 apiv1 "github.com/slok/terraform-provider-goplugin/pkg/api/v1"
)

func NewDataSourcePlugin(opts string) (apiv1.DataSourcePlugin, error) {
 return plugin{}, nil
}

type plugin struct{}

func (p plugin) ReadDataSource(ctx context.Context, r apiv1.ReadDataSourceRequest) (*apiv1.ReadDataSourceResponse, error) {
 return &apiv1.ReadDataSourceResponse{}, nil
}
```

## Important concepts

### IDs

Resources will have 2 ids:

The common Terraform ID that it's used internally by terraform to identify
the terraform resource, to refer to that tf resource in the HCL code and to import the resource into terraform.

And the resource ID itself, the one that identifies teh resource Id outside terraform (e.g a User ID
in a rest API). Normally this ID is the one you want to use to get information of the resource by using it
in a datasource.

> **Warning**
> `plugin_id` it's part of the Terraform identifier, this attribute should not change, if it changes, resource will be recreated.

### Plugin design and limitations

Plugin have some limitations, some imposed by the engine itself, [Yaegi], and other ones imposed by this provider
design in favor of UX, simplicity and portability:

- Small and simple API: Less features, more reliable and easy to maintain.
- Plugin must point to the source code root.
- Source code root must be a valid go module (`go.mod`).
- If 3rd party dependencies are used, they must be on `vendor` package (use `go mod vendor`).
- Plugin factory can be customized to have multiple plugins on the same go module codebase (e.g `NewPlugin1`, `NewPlugin2`...).
- Plugin factory must be on the root of the go module.

### JSON input/output

Instead of using `interface{}`/`any` for the data that is being passed and returned in the plugins, we decided to treat the plugins as another remote API, and use a common way that its an standard on communication, JSON.

This although less performant and a bit more verbose, benefits the plugin reliability and portability making them less brittle to changes and unknown side effects of magical auto encode/decode. Apart from this:

- Go standard library has native support and is well tested.
- Terraform has native support and by using `jsonencode`/`jsondecode` to use it in HCL code and see changes on plans.

### No computed data from plugins

[Computed] attributes are static attributes that are generated at the creation or the import phase of a resource, this data once generated can't change.

Giving the ability the user to return this data from the plugins, could make the plugins return different data on each run, making Terraform break.

So, to ease the user plugin development and usage, we decided to avoid computed data on plugins, and instead add support for plugin data sources in case users need to get extra data from a resource.

This is less performant, as a data source will fetch data every time, but its more reliable and less brittle, avoiding shoot ourselves in the foot.

## Requirements

- Terraform `>=1.x`.

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
terraform init
terraform plan
```

[yaegi]: https://github.com/traefik/yaegi
[gh-provider]: https://registry.terraform.io/providers/integrations/github/latest/docs
[gist]: https://gist.github.com/
[computed]: https://www.terraform.io/plugin/sdkv2/schemas/schema-behaviors#computed
[godoc-v1]: https://pkg.go.dev/github.com/slok/terraform-provider-goplugin/pkg/api/v1
[resource-apiv1-factory-method-godoc]: https://pkg.go.dev/github.com/slok/terraform-provider-goplugin/pkg/api/v1#NewResourcePlugin
[resource-apiv1-interface-godoc]: https://pkg.go.dev/github.com/slok/terraform-provider-goplugin/pkg/api/v1#ResourcePlugin
[data-source-apiv1-factory-method-godoc]: https://pkg.go.dev/github.com/slok/terraform-provider-goplugin/pkg/api/v1#NewDataSourcePlugin
[data-source-apiv1-interface-godoc]: https://pkg.go.dev/github.com/slok/terraform-provider-goplugin/pkg/api/v1#DataSourcePlugin
[apiv1-testing-godoc]: https://pkg.go.dev/github.com/slok/terraform-provider-goplugin/pkg/api/v1/testing
[examples]: https://github.com/slok/terraform-provider-goplugin/tree/main/examples
