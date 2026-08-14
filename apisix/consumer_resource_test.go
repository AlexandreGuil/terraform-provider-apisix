package apisix

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestConsumerResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "apisix_consumer" "test" {
	username = "example"
	desc = "Example of the consumer resource"
	plugins = jsonencode(
		{
			"limit-count" = {
				count = 5
				time_window = 1
				key = "consumer_name"
				key_type = "var"
				policy = "local"
			}
		}
	)
	labels = {
		version = "v1"
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// NOTE: State value checking is only necessary for Computed attributes,
					//       as the testing framework will automatically return test failures
					//       for configured attributes that mismatch the saved state.
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("apisix_consumer.test", "username"),
				),
			},
			// ImportState testing
			{
				ResourceName:                     "apisix_consumer.test",
				ImportState:                      true,
				ImportStateId:                    "example",
				ImportStateVerify:                true,
				ImportStateVerifyIdentifierAttribute: "username",
				// Ignore plugins value during import
				//ImportStateVerifyIgnore: []string{"plugins"},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "apisix_consumer" "test" {
	username = "example"
	desc = "Example of the consumer resource"
	plugins = jsonencode(
		{
			"limit-count" = {
				count = 10
				time_window = 1
				key = "consumer_name"
				key_type = "var"
				policy = "local"
			}
		}
	)
	labels = {
		version = "v2"
		env = "stage"
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("apisix_consumer.test", "username"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
