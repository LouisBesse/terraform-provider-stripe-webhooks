package stripe

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stripe/stripe-go/v83"
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
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "An optional description of what the webhook is used for.",
			},
			"disabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Disable the webhook endpoint if set to true.",
			},
			"connect": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "Whether the endpoint receives events from connected accounts.",
			},
			"api_version": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The API version to use for events sent to this endpoint.",
			},
			"secret": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The endpoint’s secret, used to generate webhook signatures. Returned only upon creation.",
			},
			"application": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the associated Connect application.",
			},
			"metadata": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Description: "Set of key-value pairs that you can attach to an object. " +
					"This can be useful for storing additional information about the object in a structured format.",
			},
		},
	}
}

func resourceStripeWebhookEndpointRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*stripe.Client)
	var webhookEndpoint *stripe.WebhookEndpoint
	var err error

	err = retryWithBackOff(func() error {
		webhookEndpoint, err = c.V1WebhookEndpoints.Retrieve(ctx, d.Id(), nil)
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
	c := m.(*stripe.Client)
	var webhookEndpoint *stripe.WebhookEndpoint
	var err error

	params := &stripe.WebhookEndpointCreateParams{
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
		webhookEndpoint, err = c.V1WebhookEndpoints.Create(ctx, params)
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
	c := m.(*stripe.Client)
	var err error

	params := &stripe.WebhookEndpointUpdateParams{}

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
		_, err = c.V1WebhookEndpoints.Update(ctx, d.Id(), params)
		return err
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceStripeWebhookEndpointRead(ctx, d, m)
}

func resourceStripeWebhookEndpointDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*stripe.Client)
	var err error

	err = retryWithBackOff(func() error {
		_, err = c.V1WebhookEndpoints.Delete(ctx, d.Id(), nil)
		return err
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
