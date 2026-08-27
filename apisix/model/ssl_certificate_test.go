package model

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	api_client "github.com/holubovskyi/apisix-client-go"
)

func TestSSLCertificateFromAPIToTerraform_NilClient(t *testing.T) {
	ctx := context.Background()
	status := int64(1)

	cert, key := "cert-content", "key-content"

	apiData := &api_client.SSLCertificate{
		Status:      &status,
		Certificate: &cert,
		PrivateKey:  &key,
		Client:      nil,
	}

	result := SSLCertificateFromAPIToTerraform(ctx, apiData)

	if !result.Client.IsNull() {
		t.Fatalf("Expected Client to be null, got: %v", result.Client)
	}

	attTypes := result.Client.AttributeTypes(ctx)

	if len(attTypes) == 0 {
		t.Fatalf("expected Client null object to carry its attribute types (schema shape), got empty map")
	}

	if _, ok := attTypes["ca"]; !ok {
		t.Fatalf("expected attribute type %q on the null Client object", "ca")
	}
}

func TestSSLCertificateFromAPIToTerraform_WithClient(t *testing.T) {
	ctx := context.Background()
	ca := "-----BEGIN CERTIFICATE-----\nMIIDDzCCAfegAwIBAgIUQNUaZeybQ/SzMq/aqCtoIV/TFLYwDQYJKoZIhvcNAQEL\nBQAwFzEVMBMGA1UEAwwMVGVzdCBtVExTIENBMB4XDTI2MDgyNDA3MzgwOVoXDTM2\nMDgyMTA3MzgwOVowFzEVMBMGA1UEAwwMVGVzdCBtVExTIENBMIIBIjANBgkqhkiG\n9w0BAQEFAAOCAQ8AMIIBCgKCAQEA5bvvlJh7jRMo20vFGHxob9iGtL4fflJKJaTy\ntTs4NFx3UZyOtl+WFn0erNJrk+Y7pJnAfcHfOjJ1vCLIdI0QYSOReOFt2dVbmYPM\n11lWOaaiodt5FutMuJ3vbaFLfaMrxJAGuBaOkXoGNBjHbDwJ5fgFgIiUEr4AWL0O\nU/dmG500MWsemlo3vEU9Wi4ajkJIFD2guDWyRxQv9YdbiLpC6xLQ+f/AQ3jlWWY0\nKrxgtFDz8gE0sZIE9HI/DMmG+eDuyxFTGoSRa0cX2u/G3Q2eQJ0fTmT1k628dpE6\nFlakxDevM+54HtqQnQwHFInfAblJ6y7f3Et29pjkryWml9YT+wIDAQABo1MwUTAd\nBgNVHQ4EFgQU2mciX2ZEmeou4CdeTn54N77DQUswHwYDVR0jBBgwFoAU2mciX2ZE\nmeou4CdeTn54N77DQUswDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOC\nAQEAvPOyk+e431rKvEFp/07Wi8K8+38w5xLCexZZejs0RJW2Ak6p8l9W8TFVgi0h\nPviwFwEf9iCKxlQoyNl33UPs8mmwZVbG5Z8xCigExxKK5OSFFaIIYNn1pmhoGuLd\nKJ9TCwhZqS2dbs6gUJ7jtPeJO3ffB81TVVVhzpfvIfAmuqi3bztGJh8mu0gZr6Bs\nsvHx3U6MY+kJoeds6K+05Ntj0szeDseJtVcMWJslt86eV8AQdhwSP9tLoSOiXs4h\nhFBBxxSTqZRKW2FVh8muIfbAoG8/ltGTvzH6nMwTatZAHwDRT5ZOTpBHSlHiw6xw\n6qtGtlL1+nu0bGriFL62sp66ng==\n-----END CERTIFICATE-----\n"
	depth := int64(3)

	apiData := &api_client.SSLCertificate{
		Client: &api_client.SSLClient{
			CA:    &ca,
			Depth: &depth,
		},
	}

	result := SSLCertificateFromAPIToTerraform(ctx, apiData)

	if result.Client.IsNull() {
		t.Fatalf("expected Client to be non-nul")
	}

	var extract struct {
		CA               types.String `tfsdk:"ca"`
		Depth            types.Int64  `tfsdk:"depth"`
		SkipMtlsUriRegex types.List   `tfsdk:"skip_mtls_uri_regex"`
	}

	if diags := result.Client.As(ctx, &extract, basetypes.ObjectAsOptions{}); diags.HasError() {
		t.Fatalf("failed to extract Client object: %v", diags)
	}

	if extract.CA.ValueString() != ca {
		t.Errorf("expected CA %q, got %q", ca, extract.CA.ValueString())
	}

	if extract.Depth.ValueInt64() != depth {
		t.Errorf("expected Depth %d, got %d", depth, extract.Depth.ValueInt64())
	}
}

func TestSSLCertificateFromTerraformToAPI_WithClient(t *testing.T) {
	ctx := context.Background()

	clientObj, diags := types.ObjectValue(
		map[string]attr.Type{
			"ca":                  types.StringType,
			"depth":               types.Int64Type,
			"skip_mtls_uri_regex": types.ListType{ElemType: types.StringType},
		},
		map[string]attr.Value{
			"ca":                  types.StringValue("test-ca-pem"),
			"depth":               types.Int64Value(5),
			"skip_mtls_uri_regex": types.ListNull(types.StringType),
		},
	)

	if diags.HasError() {
		t.Fatalf("failed to build test Client object: %v", diags)
	}

	plan := &SSLCertificateResourceModel{
		Client: clientObj,
		Snis:   types.ListNull(types.StringType),
		Labels: types.MapNull(types.StringType),
	}

	apiDataModel := SSLCertificateFromTerraformToAPI(ctx, plan)

	if apiDataModel.Client == nil {
		t.Fatalf("expected Client to be non-nil")
	}
	if apiDataModel.Client.CA == nil || *apiDataModel.Client.CA != "test-ca-pem" {
		t.Errorf("expected CA 'test-ca-pem', got %v", apiDataModel.Client.CA)
	}
	if apiDataModel.Client.Depth == nil || *apiDataModel.Client.Depth != 5 {
		t.Errorf("expected Depth 5, got %v", apiDataModel.Client.Depth)
	}
}
