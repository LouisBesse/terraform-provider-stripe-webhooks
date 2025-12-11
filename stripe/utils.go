package stripe

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stripe/stripe-go/v84"
)

func ExtractString(d *schema.ResourceData, key string) string {
	return ToString(d.Get(key))
}

func ToString(value interface{}) string {
	switch value.(type) {
	case string:
		return value.(string)
	case *string:
		return *(value.(*string))
	default:
		return ""
	}
}

func ToSlice(value interface{}) []interface{} {
	switch value.(type) {
	case []interface{}:
		return value.([]interface{})
	default:
		return []interface{}{}
	}
}

func ExtractStringSlice(d *schema.ResourceData, key string) []string {
	return ToStringSlice(d.Get(key))
}

func ToStringSlice(value interface{}) []string {
	slice := ToSlice(value)
	stringSlice := make([]string, len(slice), len(slice))
	for i := range slice {
		stringSlice[i] = ToString(slice[i])
	}
	return stringSlice
}

func ExtractBool(d *schema.ResourceData, key string) bool {
	return ToBool(d.Get(key))
}

func ToBool(value interface{}) bool {
	switch value.(type) {
	case bool:
		return value.(bool)
	case *bool:
		return *(value.(*bool))
	default:
		return false
	}
}

func ToMap(value interface{}) map[string]interface{} {
	switch value.(type) {
	case map[string]interface{}:
		return value.(map[string]interface{})
	case []interface{}:
		sl := value.([]interface{})
		if len(sl) > 0 {
			return sl[0].(map[string]interface{})
		}
	}
	return map[string]interface{}{}
}

func CallSet(err ...error) (d diag.Diagnostics) {
	for _, e := range err {
		if e != nil {
			d = append(d, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  e.Error(),
			})
		}
	}
	return d
}

type MetadataAdder interface {
	AddMetadata(key, value string)
}

func UpdateMetadata(d *schema.ResourceData, adder MetadataAdder) {
	oldMeta, newMeta := d.GetChange("metadata")
	oldMetaMap := ToMap(oldMeta)
	newMetaMap := ToMap(newMeta)
	for k := range newMetaMap {
		if _, set := oldMetaMap[k]; set {
			delete(oldMetaMap, k)
		}
	}

	for k, v := range newMetaMap {
		adder.AddMetadata(k, ToString(v))
	}
	for k := range oldMetaMap {
		adder.AddMetadata(k, "")
	}
}

func isRateLimitErr(e error) bool {
	err, ok := e.(*stripe.Error)
	return ok && err.HTTPStatusCode == 429
}

func retryWithBackOff(call func() error) error {
	var backOff time.Duration = 1
	for {
		err := call()
		switch {
		case err == nil:
			return nil
		case isRateLimitErr(err):
			time.Sleep(time.Second * backOff)
			backOff *= 2
		default:
			return err
		}
	}
}
