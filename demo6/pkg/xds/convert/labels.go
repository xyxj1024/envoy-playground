package convert

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	valid "github.com/asaskevich/govalidator"

	types "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
)

type ServiceEndpoint struct {
	RequestTimeout time.Duration
	Protocol       types.SocketAddress_Protocol
	Port           types.SocketAddress_PortValue
}

type ServiceRoute struct {
	Domain       string
	ExtraDomains []string
	PathPrefix   string
}

type ServiceLabel struct {
	Endpoint ServiceEndpoint
	Route    ServiceRoute
}

func NewServiceLabel() ServiceLabel {
	return ServiceLabel{
		ServiceEndpoint{
			RequestTimeout: 15 * time.Second,
			Protocol:       types.SocketAddress_TCP,
			Port:           types.SocketAddress_PortValue{PortValue: 0},
		},
		ServiceRoute{
			ExtraDomains: []string{},
			PathPrefix:   "/",
		},
	}
}

var serviceLabelRegex = regexp.MustCompile(`(?Uim)envoy\.(?P<type>\S+)\.(?P<property>\S+$)`)

func ParseServiceLabels(labels map[string]string) *ServiceLabel {
	s := NewServiceLabel()
	for key, value := range labels {
		if !serviceLabelRegex.MatchString(key) {
			continue
		}

		matches := serviceLabelRegex.FindStringSubmatch(key)
		switch strings.ToLower(matches[1]) {
		case "endpoint":
			s.setEndpointProperty(matches[2], value)
		case "route":
			s.setRouteProperty(matches[2], value)
		}
	}

	return &s
}

func (l *ServiceLabel) setEndpointProperty(property, value string) {
	switch strings.ToLower(property) {
	case "timeout":
		if timeout, err := time.ParseDuration(value); err != nil {
			l.Endpoint.RequestTimeout = timeout
		}
	case "protocol":
		p := types.SocketAddress_TCP
		if strings.EqualFold(value, "udp") {
			p = types.SocketAddress_UDP
		}

		l.Endpoint.Protocol = p
	case "port":
		v, _ := strconv.ParseUint(value, 10, 32)
		l.Endpoint.Port = types.SocketAddress_PortValue{
			PortValue: uint32(v),
		}
	}
}

func (l *ServiceLabel) setRouteProperty(property, value string) {
	switch strings.ToLower(property) {
	case "path":
		l.Route.PathPrefix = fmt.Sprintf("/%s", strings.TrimPrefix(value, "/"))
	case "domain":
		l.Route.Domain = value
	case "extra-domains":
		l.Route.ExtraDomains = strings.Split(value, ",")
	}
}

func (l ServiceLabel) Validate() error {
	if l.Endpoint.Port.PortValue <= 0 {
		return errors.New("there is no endpoint.port label specified")
	}

	if l.Route.Domain == "" {
		return errors.New("there is no route.domain label specified")
	}

	if l.Endpoint.RequestTimeout.Seconds() < 0 {
		return errors.New("the endpoint.timeout can't be a negative number")
	}

	if !valid.IsDNSName(l.Route.Domain) {
		return errors.New("the route.domain is not a valid DNS name")
	}

	for i := range l.Route.ExtraDomains {
		if !valid.IsDNSName(l.Route.ExtraDomains[i]) {
			return errors.New("the route.extra-domains contains an invalid DNS name")
		}
	}

	return nil
}
