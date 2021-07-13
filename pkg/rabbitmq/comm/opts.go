package comm

import (
	"errors"
	"strings"
)

// Options TODO
type Options map[string]string

// AddOption TODO
func (o Options) AddOption(k, v string) {
	o[k] = v
}

// GetOption TODO
func (o Options) GetOption(k, dv string) string {
	if v, ok := o[k]; ok {
		return v
	}
	return dv
}

// HasOption TODO
func (o Options) HasOption(k string) bool {
	_, ok := o[k]
	return ok
}

// Encode TODO
func (o Options) Encode() string {
	b := strings.Builder{}
	i := 0
	for k, v := range o {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(k)
		b.WriteByte(':')
		b.WriteString(v)
		i++
	}
	return b.String()
}

// ParseFromString TODO
func (o Options) ParseFromString(str string) error {
	ps := strings.Split(str, ",")
	for _, p := range ps {
		kv := strings.Split(p, ":")
		if len(kv) != 2 {
			return errors.New("bad kv")
		}
		o[kv[0]] = kv[1]
	}
	return nil
}
