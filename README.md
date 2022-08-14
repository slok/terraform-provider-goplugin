# terraform-provider-goplugin

[![CI](https://github.com/slok/terraform-provider-goplugin/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/slok/terraform-provider-goplugin/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/slok/terraform-provider-goplugin)](https://goreportcard.com/report/github.com/slok/terraform-provider-goplugin)
[![Apache 2 licensed](https://img.shields.io/badge/license-Apache2-blue.svg)](https://raw.githubusercontent.com/slok/terraform-provider-goplugin/master/LICENSE)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/slok/terraform-provider-goplugin)](https://github.com/slok/terraform-provider-goplugin/releases/latest)
[![Terraform regsitry](https://img.shields.io/badge/Terraform-Registry-color=green?logo=Terraform&style=flat&color=5C4EE5&logoColor=white)](https://registry.terraform.io/providers/slok/goplugin/latest/docs)

TODO.

## Use cases

TODO.

## Requirements

- Terraform `>=1.x`.

## Terraform cloud

TODO.

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
