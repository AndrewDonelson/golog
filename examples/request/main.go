package main

import (
	"net/http"
	"net/http/httptrace"
	"time"

	"github.com/AndrewDonelson/golog"
	"github.com/AndrewDonelson/golog/handlers/text"
)

// Isolate the concern of tracing in a single layer/component.
// http.RoundTripper is like the client-side http.Handler,
// and a good place to wire in decorators like this.

type tracingRoundTripper struct {
	next http.RoundTripper
	dest *golog.Logger
}

func (rt *tracingRoundTripper) RoundTrip(r *http.Request) (resp *http.Response, err error) {
	var (
		start     = time.Now()
		dnsStart  time.Duration
		firstByte time.Duration
	)

	rt.dest.HandlerLog(nil, r)

	defer func() {
		f := golog.Fields{
			"dns_start_ms":  dnsStart.Milliseconds(),
			"first_byte_ms": firstByte.Milliseconds(),
			"total_ms":      time.Since(start).Milliseconds(),
			"url":           r.URL.String(),
		}
		switch {
		case err == nil:
			f["response_code"] = resp.StatusCode
		case err != nil:
			f["error"] = err.Error()
		}
		rt.dest.WithFields(f).Trace("Client request")
	}()

	tr := &httptrace.ClientTrace{
		DNSStart:             func(httptrace.DNSStartInfo) { dnsStart = time.Since(start) },
		GotFirstResponseByte: func() { firstByte = time.Since(start) },
	}

	ctx := httptrace.WithClientTrace(r.Context(), tr)

	return rt.next.RoundTrip(r.WithContext(ctx))
}

func main() {
	// Make a logger.
	logger := golog.NewLogger(golog.NewDefaultOptions())
	logger.SetModuleName("request")
	logger.SetEnvironment(golog.EnvDevelopment)
	logger.Handler = text.Default

	// Create a client with a decorated DefaultTransport.
	client := &http.Client{
		Transport: &tracingRoundTripper{
			next: http.DefaultTransport,
			dest: logger,
		},
	}

	// Now it's just a regular request...
	req, err := http.NewRequest("GET", "http://ip.jsontest.com/", nil)
	if err != nil {
		logger.FatalE(err)
	}

	// ...sent through that client, but otherwise nothing special.
	resp, err := client.Do(req)
	if err != nil {
		logger.FatalE(err)
	}

	defer resp.Body.Close()

	logger.Successf("%d", resp.StatusCode)
}
