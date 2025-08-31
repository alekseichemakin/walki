package mux

import (
	"net/url"
	"strings"
)

type Values map[string]string

// Поддерживает:
//   - "city:Москва"      -> {"a":"city","id":"Москва"}
//   - "route:123"        -> {"a":"route","id":"123"}
//   - "buy:route=5&c=ABC"-> {"a":"buy","route":"5","c":"ABC"}
//   - "plain"            -> {"id":"plain"} (если вдруг без двоеточия)
func Parse(data string) Values {
	v := Values{}
	if data == "" {
		return v
	}

	parts := strings.SplitN(data, ":", 2)
	// без префикса
	if len(parts) == 1 {
		rest := parts[0]
		if strings.ContainsAny(rest, "=&") {
			q, _ := url.ParseQuery(rest)
			for k, vals := range q {
				if len(vals) > 0 {
					v[k] = vals[0]
				}
			}
		} else {
			// просто значение
			if dec, err := url.QueryUnescape(rest); err == nil {
				rest = dec
			}
			v["id"] = rest
		}
		return v
	}

	a, rest := parts[0], parts[1]
	v["a"] = a // действие/префикс (опционально; убери строку, если не нужно)

	if rest == "" {
		return v
	}

	if strings.ContainsAny(rest, "=&") {
		q, _ := url.ParseQuery(rest)
		for k, vals := range q {
			if len(vals) > 0 {
				v[k] = vals[0]
			}
		}
	} else {
		if dec, err := url.QueryUnescape(rest); err == nil {
			rest = dec
		}
		v["id"] = rest
	}

	return v
}
