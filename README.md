# Stripe Webhooks Terraform Provider

The Stripe Terraform Webhooks provider uses the official Stripe SDK based on Golang. On top of that, the provider is
developed
around the official Stripe API documentation [website](https://stripe.com/docs/api).

This release changes the pinned API version to `2025-10-29.clover`

The Stripe Webhooks Terraform Provider documentation can be found on the Terraform Provider documentation

## Usage:

```
terraform {
  required_providers {
    stripe-webhooks = {
      source = "louisbesse/stripe-webhooks"
      version = "5.1.0"
    }
  }
}

provider "stripe-webhooks" {
  api_key="<api_secret_key>"
}
```

### Environmental variable support

The parameter `api_key` can be omitted when the `STRIPE_API_KEY` environmental variable is present.