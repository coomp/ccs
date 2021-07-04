package comm

import (
	"net/url"
	"strings"
)

// ParsePortalMessage parses a portal message
func ParsePortalMessage(msg string) (url.Values, error) {
	values := make(url.Values)
	err := parseQuery(values, msg)
	return values, err
}

// parseQuery copied from bossWaterDistribution 标准库的url.ParseQuery把";"也作为分隔符
func parseQuery(m url.Values, query string) (err error) {
	for query != "" {
		key := query
		if i := strings.IndexAny(key, "&"); i >= 0 {
			key, query = key[:i], key[i+1:]
		} else {
			query = ""
		}
		if key == "" {
			continue
		}
		value := ""
		if i := strings.Index(key, "="); i >= 0 {
			key, value = key[:i], key[i+1:]
		}
		key, err1 := url.QueryUnescape(key)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		value, err1 = url.QueryUnescape(value)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		m[key] = append(m[key], value)
	}
	return err
}
