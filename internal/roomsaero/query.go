package roomsaero

import (
	"net/url"
	"strconv"
)

func setIfNotEmpty(q url.Values, key, val string) {
	if val != "" {
		q.Set(key, val)
	}
}

func setBoolPtr(q url.Values, key string, val *bool) {
	if val != nil {
		q.Set(key, strconv.FormatBool(*val))
	}
}

func setIntPtr(q url.Values, key string, val *int) {
	if val != nil {
		q.Set(key, strconv.Itoa(*val))
	}
}

func setFloatPtr(q url.Values, key string, val *float64) {
	if val != nil {
		q.Set(key, strconv.FormatFloat(*val, 'f', -1, 64))
	}
}
