package ip

import (
	"net"
	"time"
)

// MaxTries is the maximum amount of tries to attempt to one service.
const MaxTries = 3

// Timeout after which result will be collected
var Timeout = 2 * time.Second

// APIURIs is list of services used to get IP address
var APIURIs = []string{
	"https://api.ipify.org",
	"http://myexternalip.com/raw",
	"http://ipinfo.io/ip",
	"http://ipecho.net/plain",
	"http://icanhazip.com",
	"http://ifconfig.me/ip",
	"http://ident.me",
	"http://checkip.amazonaws.com",
	"http://bot.whatismyipaddress.com",
	"http://whatismyip.akamai.com",
	"http://wgetip.com",
	"http://ip.tyk.nu",
}
// Get returns public IP
func Get() (net.IP, error) {
	var results []net.IP
	resultCh := make(chan net.IP, len(APIURIs))
	var errs []error
	errCh := make(chan error, len(APIURIs))

	for _, d := range APIURIs {
		go worker(d, resultCh, errCh)
	}
	for {
		select {
		case err := <-errCh:
			errs = append(errs, err)
		case r := <-resultCh:
			results = append(results, r)
		case <-time.After(Timeout):
			r, err := validate(results)
			if err != nil {
				return nil, detailErr(err, errs)
			}
			return r, nil
		}
	}
}
