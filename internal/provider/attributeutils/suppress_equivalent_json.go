package attributeutils

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SuppressEquivalentJSON Will modify the plan to avoid changes when 2 json are lexically different
// but semantically equivalent (e.g map keys order, json format...).
const SuppressEquivalentJSON = suppressEquivalentJSON(false)

type suppressEquivalentJSON bool

func (s suppressEquivalentJSON) Description(ctx context.Context) string { return "" }

func (s suppressEquivalentJSON) MarkdownDescription(ctx context.Context) string {
	return s.Description(ctx)
}

func (s suppressEquivalentJSON) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	if req.AttributeConfig.IsNull() || req.AttributeState.IsNull() || req.AttributeState.IsNull() {
		return
	}

	config := req.AttributeConfig.(types.String)
	state := req.AttributeState.(types.String)

	if s.isEqualJSON(config.ValueString(), state.ValueString()) {
		resp.AttributePlan = req.AttributeState
	}
}

func (s suppressEquivalentJSON) isEqualJSON(old, new string) bool {
	ob := bytes.NewBufferString("")
	if err := json.Compact(ob, []byte(old)); err != nil {
		return false
	}

	nb := bytes.NewBufferString("")
	if err := json.Compact(nb, []byte(new)); err != nil {
		return false
	}

	var o1 any
	if err := json.Unmarshal(ob.Bytes(), &o1); err != nil {
		return false
	}

	var o2 any
	if err := json.Unmarshal(nb.Bytes(), &o2); err != nil {
		return false
	}

	return reflect.DeepEqual(o1, o2)
}
