package yaegicustom

import (
	"reflect"
)

//go:generate yaegi extract --name yaegicustom  github.com/slok/terraform-provider-goplugin/pkg/api/v1

// Symbols variable stores the map of custom symbols per package.
var Symbols = map[string]map[string]reflect.Value{}
