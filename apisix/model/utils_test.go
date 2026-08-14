package model

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPreferPluginsValuePrefersFreshApiValue(t *testing.T) {
	fresh := types.StringValue(`{"limit-count":{"count":2000}}`)
	stale := types.StringValue(`{"limit-count":{"count":5}}`)

	if got := PreferPluginsValue(fresh, stale); got.ValueString() != fresh.ValueString() {
		t.Fatalf("expected API value to win when both values are set, got %s", got.ValueString())
	}
}

func TestPreferPluginsValueFallsBackToKnownState(t *testing.T) {
	fresh := types.StringNull()
	stale := types.StringValue(`{"limit-count":{"count":5}}`)

	if got := PreferPluginsValue(fresh, stale); got.ValueString() != stale.ValueString() {
		t.Fatalf("expected fallback state value when the API response is null, got %s", got.ValueString())
	}
}
