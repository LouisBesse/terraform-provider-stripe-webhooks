package stripe

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/client"
)

func resourceStripeWebhookEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceStripeWebhookEndpointCreate,
		ReadContext:   resourceStripeWebhookEndpointRead,
		UpdateContext: resourceStripeWebhookEndpointUpdate,
		DeleteContext: resourceStripeWebhookEndpointDelete,
		Schema: map[string]*schema.Schema{
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The URL of the webhook endpoint.",
			},
			"enabled_events": {
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The list of events to enable for this endpoint. Use ['*'] to enable all events.",
			},
			"connect": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether the endpoint receives events from connected accounts.",
			},
			"api_version": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The API version to use for events sent to this endpoint.",
			},
			"secret": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The endpoint’s secret, used to generate webhook signatures. Returned only upon creation.",
			},
		},
	}
}

func resourceStripeWebhookEndpointRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.API)
	var webhookEndpoint *stripe.WebhookEndpoint
	var err error

	err = retryWithBackOff(func() error {
		webhookEndpoint, err = c.WebhookEndpoints.Get(d.Id(), nil)
		return err
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return CallSet(
		d.Set("enabled_events", webhookEndpoint.EnabledEvents),
		d.Set("url", webhookEndpoint.URL),
		d.Set("description", webhookEndpoint.Description),
		d.Set("disabled", webhookEndpoint.Status != "enabled"),
		d.Set("connect", webhookEndpoint.Application != ""),
		d.Set("api_version", webhookEndpoint.APIVersion),
		d.Set("application", webhookEndpoint.Application),
		d.Set("metadata", webhookEndpoint.Metadata),
	)
}

func resourceStripeWebhookEndpointCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.API)
	var webhookEndpoint *stripe.WebhookEndpoint
	var err error

	params := &stripe.WebhookEndpointParams{
		URL:           stripe.String(ExtractString(d, "url")),
		EnabledEvents: stripe.StringSlice(ExtractStringSlice(d, "enabled_events")),
	}
	if description, set := d.GetOk("description"); set {
		params.Description = stripe.String(ToString(description))
	}
	if connect, set := d.GetOk("connect"); set {
		params.Connect = stripe.Bool(ToBool(connect))
	}
	if APIVersion, set := d.GetOk("api_version"); set {
		params.APIVersion = stripe.String(ToString(APIVersion))
	}
	if meta, set := d.GetOk("metadata"); set {
		for k, v := range ToMap(meta) {
			params.AddMetadata(k, ToString(v))
		}
	}

	err = retryWithBackOff(func() error {
		webhookEndpoint, err = c.WebhookEndpoints.New(params)
		return err
	})
	if err != nil {
		return diag.FromErr(err)
	}

	dg := CallSet(
		d.Set("secret", webhookEndpoint.Secret),
	)
	if len(dg) > 0 {
		return dg
	}

	d.SetId(webhookEndpoint.ID)
	return resourceStripeWebhookEndpointRead(ctx, d, m)
}

func resourceStripeWebhookEndpointUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.API)
	var err error

	params := &stripe.WebhookEndpointParams{}

	if d.HasChange("enabled_events") {
		params.EnabledEvents = stripe.StringSlice(ExtractStringSlice(d, "enabled_events"))
	}
	if d.HasChange("url") {
		params.URL = stripe.String(ExtractString(d, "url"))
	}
	if d.HasChange("description") {
		params.Description = stripe.String(ExtractString(d, "description"))
	}
	if d.HasChange("disabled") {
		params.Disabled = stripe.Bool(ExtractBool(d, "disabled"))
	}
	if d.HasChange("metadata") {
		params.Metadata = nil
		UpdateMetadata(d, params)
	}

	err = retryWithBackOff(func() error {
		_, err = c.WebhookEndpoints.Update(d.Id(), params)
		return err
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceStripeWebhookEndpointRead(ctx, d, m)
}

func resourceStripeWebhookEndpointDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.API)
	var err error

	err = retryWithBackOff(func() error {
		_, err = c.WebhookEndpoints.Del(d.Id(), nil)
		return err
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
