package tcpx

import "strings"

type Route struct {
	URLPattern string
	MessageId  int
	Whereis    []string
}

func (r *Route) Merge(r2 Route) Route{
	if r.URLPattern == r2.URLPattern || r.MessageId == r2.MessageId {
		r.Whereis = append(r.Whereis, r2.Whereis...)
	}
	return *r
}
func (r Route) Location() string {
	return strings.Join(r.Whereis, "\n")
}
