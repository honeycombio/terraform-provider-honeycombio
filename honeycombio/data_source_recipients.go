package honeycombio

import (
	"context"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/honeycombio/internal/hashcode"
)

func dataSourceHoneycombioRecipients() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHoneycombioRecipientsRead,

		Description: `
'honeycombio_recipients' data source provides recipient IDs of recipients matching a set of criteria.
`,

		Schema: map[string]*schema.Schema{
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The type of recipients.",
				ValidateFunc: validation.StringInSlice([]string{"email", "pagerduty", "slack", "webhook"}, false),
			},
			"detail_filter": {
				Type:         schema.TypeList,
				Optional:     true,
				MinItems:     1,
				MaxItems:     1,
				Description:  "Attributes to filter the recipients with. `type` must be set when providing a filter.",
				RequiredWith: []string{"type"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The name of the detail field to filter by",
							ValidateFunc: validation.StringInSlice([]string{"address", "channel", "name", "integration_name", "url"}, false),
						},
						"value": {
							Type:          schema.TypeString,
							Optional:      true,
							Description:   "The value of the detail field to match on.",
							ConflictsWith: []string{"detail_filter.0.value_regex"},
							ValidateFunc:  validation.NoZeroValues,
						},
						"value_regex": {
							Type:          schema.TypeString,
							Optional:      true,
							Description:   "A regular expression string to apply to the value of the detail field to match on.",
							ConflictsWith: []string{"detail_filter.0.value"},
							ValidateFunc:  validation.StringIsValidRegExp,
						},
					},
				},
			},
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: false,
				Required: false,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceHoneycombioRecipientsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	rcpts, err := client.Recipients.List(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	var matchType honeycombio.RecipientType
	var rcptFilter *recipientFilter
	if v, ok := d.GetOk("type"); ok {
		matchType = honeycombio.RecipientType(v.(string))
		if matchType != "" {
			// type has been provided, create a type-only filter which may be 'upgraded'
			// to a `detail_filter`
			rcptFilter = &recipientFilter{Type: matchType}
		}
	}
	if v, ok := d.GetOk("detail_filter"); ok {
		rcptFilter = expandRecipientFilter(v.([]interface{}))
		if rcptFilter.Type != matchType {
			return diag.Errorf("provided type doesn't match filter type")
		}
	}

	var rcptIDs []string
	for _, r := range rcpts {
		if rcptFilter.IsMatch(r) {
			rcptIDs = append(rcptIDs, r.ID)
		}
	}

	d.SetId(strconv.Itoa(hashcode.String(strings.Join(rcptIDs, ","))))
	d.Set("ids", rcptIDs)

	return nil
}
