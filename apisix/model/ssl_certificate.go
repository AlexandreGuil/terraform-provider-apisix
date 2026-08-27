package model

import (
	"context"

	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	api_client "github.com/holubovskyi/apisix-client-go"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func toTerraformStrings(slice []string) []attr.Value {
	vals := make([]attr.Value, len(slice))
	for i, s := range slice {
		vals[i] = types.StringValue(s)
	}
	return vals
}

// SSLCertificateResourceModel maps the resource schema data.
type SSLCertificateResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Status      types.Int64  `tfsdk:"status"`
	Certificate types.String `tfsdk:"certificate"`
	PrivateKey  types.String `tfsdk:"private_key"`
	Client      types.Object `tfsdk:"client"`
	Snis        types.List   `tfsdk:"snis"`
	Type        types.String `tfsdk:"type"`
	Labels      types.Map    `tfsdk:"labels"`
}

var SSLCertificateSchema = schema.Schema{
	Description: "Manages APISIX SSL certificates.",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Identifier of the certificate.",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"certificate": schema.StringAttribute{
			Description: "HTTPS certificate.",
			Required:    true,
		},
		"private_key": schema.StringAttribute{
			Description: "HTTPS private key.",
			Required:    true,
			Sensitive:   true,
		},
		"client": schema.SingleNestedAttribute{
			Description: "Mutual TLS client certificate configuration. Set to require and verify a client certificate signed by the given CA for requests matching this SSL entry's SNI",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"ca": schema.StringAttribute{
					Description: "Certificate Authority that verifies the client",
					Required:    true,
				},
				"depth": schema.Int64Attribute{
					Description: "Verification depth in client certificate chains. Default is 1.",
					Optional:    true,
					Computed:    true,
				},
				"skip_mtls_uri_regex": schema.ListAttribute{
					Description: "URIs (regex) that bypass client certificate checking",
					Optional:    true,
					ElementType: types.StringType,
				},
			},
		},
		"snis": schema.ListAttribute{
			MarkdownDescription: "A non-empty array of HTTPS SNI. Required if `type` is `server`.",
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
		},
		"type": schema.StringAttribute{
			MarkdownDescription: "Identifies the type of certificate, default `server`.\n" +
				"`client` Indicates that the certificate is a client certificate, which is used when APISIX accesses the upstream; " +
				"`server` Indicates that the certificate is a server-side certificate, which is used by APISIX when verifying client requests.",
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString("server"),
			Validators: []validator.String{
				// Validate string value must be "server" or "client"
				stringvalidator.OneOf([]string{"server", "client"}...),
			},
		},
		"labels": schema.MapAttribute{
			MarkdownDescription: "Attributes of the resource specified as key-value pairs. An individual pair cannot be deleted using APISIX API. " +
				"In order to delete an individual pair, you can delete all labels and reapply the resource with the desired labels map",
			Optional:    true,
			ElementType: types.StringType,
		},
		"status": schema.Int64Attribute{
			MarkdownDescription: "Enables the current SSL. Set to `1` (enabled) by default. `1` to enable, `0` to disable",
			Optional:            true,
			Computed:            true,
			Default:             int64default.StaticInt64(1),
			Validators: []validator.Int64{
				// Validate integer value must be 0 or 1
				int64validator.OneOf([]int64{0, 1}...),
			},
		},
	},
}

func SSLCertificateFromTerraformToAPI(ctx context.Context, terraformDataModel *SSLCertificateResourceModel) (apiDataModel api_client.SSLCertificate) {
	apiDataModel.Status = terraformDataModel.Status.ValueInt64Pointer()
	apiDataModel.Certificate = terraformDataModel.Certificate.ValueStringPointer()
	apiDataModel.PrivateKey = terraformDataModel.PrivateKey.ValueStringPointer()
	apiDataModel.Type = terraformDataModel.Type.ValueStringPointer()

	terraformDataModel.Snis.ElementsAs(ctx, &apiDataModel.SNIs, false)
	terraformDataModel.Labels.ElementsAs(ctx, &apiDataModel.Labels, false)

	if !terraformDataModel.Client.IsNull() {
		var clientModel struct {
			CA               types.String `tfsdk:"ca"`
			Depth            types.Int64  `tfsdk:"depth"`
			SkipMtlsUriRegex types.List   `tfsdk:"skip_mtls_uri_regex"`
		}
		_ = terraformDataModel.Client.As(ctx, &clientModel, basetypes.ObjectAsOptions{})

		var client api_client.SSLClient
		if !clientModel.CA.IsNull() {
			ca := clientModel.CA.ValueString()
			client.CA = &ca
		}
		if !clientModel.Depth.IsNull() {
			depth := clientModel.Depth.ValueInt64()
			client.Depth = &depth
		}
		if !clientModel.SkipMtlsUriRegex.IsNull() {
			var skip []string
			_ = clientModel.SkipMtlsUriRegex.ElementsAs(ctx, &skip, false)
			client.SkipMtlsUriRegex = &skip
		}
		apiDataModel.Client = &client
	}

	tflog.Debug(ctx, "Result of the SSLCertificateFromTerraformToAPI", map[string]any{
		"Values": apiDataModel,
	})

	return apiDataModel
}

func SSLCertificateFromAPIToTerraform(ctx context.Context, apiDataModel *api_client.SSLCertificate) (terraformDataModel SSLCertificateResourceModel) {
	terraformDataModel.ID = types.StringPointerValue(apiDataModel.ID)
	terraformDataModel.Status = types.Int64PointerValue(apiDataModel.Status)
	terraformDataModel.Certificate = types.StringPointerValue(apiDataModel.Certificate)
	// APISIX API returns the private key in base64 form
	//terraformDataModel.PrivateKey = types.StringValue(apiDataModel.PrivateKey)
	terraformDataModel.Type = types.StringPointerValue(apiDataModel.Type)

	terraformDataModel.Snis, _ = types.ListValueFrom(ctx, types.StringType, apiDataModel.SNIs)
	terraformDataModel.Labels, _ = types.MapValueFrom(ctx, types.StringType, apiDataModel.Labels)

	if apiDataModel.Client != nil {
		clientMap := map[string]attr.Value{
			"ca":                  types.StringNull(),
			"depth":               types.Int64Null(),
			"skip_mtls_uri_regex": types.ListNull(types.StringType),
		}
		if apiDataModel.Client.CA != nil {
			clientMap["ca"] = types.StringValue(*apiDataModel.Client.CA)
		}
		if apiDataModel.Client.Depth != nil {
			clientMap["depth"] = types.Int64Value(*apiDataModel.Client.Depth)
		}
		if apiDataModel.Client.SkipMtlsUriRegex != nil {
			clientMap["skip_mtls_uri_regex"] = types.ListValueMust(types.StringType, toTerraformStrings(*apiDataModel.Client.SkipMtlsUriRegex))
		}
		obj, _ := types.ObjectValue(map[string]attr.Type{
			"ca":                  types.StringType,
			"depth":               types.Int64Type,
			"skip_mtls_uri_regex": types.ListType{ElemType: types.StringType},
		}, clientMap)
		terraformDataModel.Client = obj
	} else {
		terraformDataModel.Client = types.ObjectNull(map[string]attr.Type{
			"ca":                  types.StringType,
			"depth":               types.Int64Type,
			"skip_mtls_uri_regex": types.ListType{ElemType: types.StringType},
		})
	}

	tflog.Debug(ctx, "Result of the SSLCertificateFromAPIToTerraform", map[string]any{
		"Values": terraformDataModel,
	})

	return terraformDataModel
}

// Get SNIS list from the certificate
func CertSNIS(crt string, key string) ([]string, error) {
	certDERBlock, _ := pem.Decode([]byte(crt))
	if certDERBlock == nil {
		return []string{}, nil
	}

	_, err := tls.X509KeyPair([]byte(crt), []byte(key))
	if err != nil {
		return []string{}, err
	}

	x509Cert, err := x509.ParseCertificate(certDERBlock.Bytes)

	if err != nil {
		return []string{}, err
	}

	var snis []string
	if len(x509Cert.DNSNames) > 0 {
		snis = x509Cert.DNSNames
	} else if len(x509Cert.IPAddresses) > 0 {
		for _, ip := range x509Cert.IPAddresses {
			snis = append(snis, ip.String())
		}
	} else {
		if len(x509Cert.Subject.Names) > 1 {
			var attributeTypeNames = map[string]string{
				"2.5.4.6":  "C",
				"2.5.4.10": "O",
				"2.5.4.11": "OU",
				"2.5.4.3":  "CN",
				"2.5.4.5":  "SERIALNUMBER",
				"2.5.4.7":  "L",
				"2.5.4.8":  "ST",
				"2.5.4.9":  "STREET",
				"2.5.4.17": "POSTALCODE",
			}
			for _, tv := range x509Cert.Subject.Names {
				oidString := tv.Type.String()
				typeName, ok := attributeTypeNames[oidString]
				if ok && typeName == "CN" {
					valueString := fmt.Sprint(tv.Value)
					snis = append(snis, valueString)
				}
			}
		}
	}

	return snis, nil
}
