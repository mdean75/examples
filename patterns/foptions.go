package patterns

import (
	"net"
	"net/http"
	"time"
)

// https://www.sohamkamani.com/golang/options-pattern/
/*
1. define struct
2. define constructor for default
3. create function type being a func that accepts a pointer of the type
4. define functional options that modify the instance
5. add functional options to constructor ie. range over options
*/

type ClientWrapper struct {
	Cl http.Client
}

type ClientOption func(wrapper *ClientWrapper)

func NewClientWrapper(opts ...ClientOption) *ClientWrapper {
	c := http.Client{
		Transport:     http.DefaultTransport,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       0,
	}

	cl := &ClientWrapper{
		Cl: c,
	}

	for _, opt := range opts {
		opt(cl)
	}

	return cl
}

func Timeout(t time.Duration) ClientOption {
	return func(c *ClientWrapper) {
		c.Cl.Timeout = t
	}
}

func Transport(tr *TransportWrapper) ClientOption {
	return func(c *ClientWrapper) {
		c.Cl.Transport = tr.Tr
	}
}

// transport options
type TransportWrapper struct {
	Tr *http.Transport
}

type TransportOption func(wrapper *TransportWrapper)

func NewTransportWrapper(opts ...TransportOption) *TransportWrapper {
	t := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	tr := &TransportWrapper{
		Tr: t,
	}

	for _, opt := range opts {
		opt(tr)
	}

	return tr
}

func MaxIdleCons(ic int) TransportOption {
	return func(t *TransportWrapper) {
		t.Tr.MaxIdleConns = ic
	}
}

func MaxIdleConsPerHost(ich int) TransportOption {
	return func(t *TransportWrapper) {
		t.Tr.MaxIdleConnsPerHost = ich
	}
}

func MaxConsPerHost(cph int) TransportOption {
	return func(t *TransportWrapper) {
		t.Tr.MaxConnsPerHost = cph
	}
}

func IdleConTimeout(ict time.Duration) TransportOption {
	return func(t *TransportWrapper) {
		t.Tr.IdleConnTimeout = ict
	}
}
