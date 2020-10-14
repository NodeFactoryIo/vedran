package ip

import (
	"errors"
	"fmt"
	"github.com/cenkalti/backoff"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func getIPBy(dest string) (net.IP, error) {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 100 * time.Millisecond
	b.MaxElapsedTime = 10 * time.Second
	b.Multiplier = 2

	client := &http.Client{}

	req, err := http.NewRequest("GET", dest, nil)
	if err != nil {
		return nil, err
	}

	for tries := 0; tries < MaxTries; tries++ {
		resp, err := client.Do(req)
		if err != nil {
			d := b.NextBackOff()
			time.Sleep(d)
			continue
		}

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != 200 {
			return nil, errors.New(dest + " status code " + strconv.Itoa(resp.StatusCode) + ", body: " + string(body))
		}

		tb := strings.TrimSpace(string(body))
		ip := net.ParseIP(tb)
		if ip == nil {
			return nil, errors.New("IP address not valid: " + tb)
		}
		return ip, nil
	}

	return nil, errors.New("Failed to reach " + dest)
}

func detailErr(err error, errs []error) error {
	errStrs := []string{err.Error()}
	for _, e := range errs {
		errStrs = append(errStrs, e.Error())
	}
	j := strings.Join(errStrs, "\n")
	return errors.New(j)
}

func validate(rs []net.IP) (net.IP, error) {
	if rs == nil {
		return nil, fmt.Errorf("Failed to get any result from %d APIs", len(APIURIs))
	}
	if len(rs) < 3 {
		return nil, fmt.Errorf("Less than %d results from %d APIs", 3, len(APIURIs))
	}
	first := rs[0]
	for i := 1; i < len(rs); i++ {
		if !reflect.DeepEqual(first, rs[i]) { //first != rs[i] {
			return nil, fmt.Errorf("Results are not identical: %s", rs)
		}
	}
	return first, nil
}

func worker(d string, r chan<- net.IP, e chan<- error) {
	ip, err := getIPBy(d)
	if err != nil {
		e <- err
		return
	}
	r <- ip
}
