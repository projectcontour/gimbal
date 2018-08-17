package wrk

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type LatencyDistribution struct {
	P99  string
	P999 string
}

type Report struct {
	Threads           int
	Connections       int
	TargetURL         string
	Time              time.Duration
	Latency           LatencyDistribution
	Non200Responses   int
	RequestPerSecond  float64
	TransferPerSecond string
	TotalRequests     int
}

func BuildReport(logs io.Reader) (*Report, error) {

	r := &Report{}
	s := bufio.NewScanner(logs)
	for s.Scan() {
		t := strings.TrimSpace(s.Text())
		if r.TargetURL == "" {
			m := regexp.MustCompile("^Running (.*) test @ (.*)$").FindStringSubmatch(t)
			if len(m) == 3 {
				r.TargetURL = m[2]
				reportTime, err := time.ParseDuration(m[1])
				if err != nil {
					return nil, err
				}
				r.Time = reportTime
				continue
			}
		}
		if r.Threads == 0 {
			m := regexp.MustCompile("^([0-9]+) threads and ([0-9]+) connections$").FindStringSubmatch(t)
			if len(m) == 3 {
				var err error
				r.Threads, err = strconv.Atoi(m[1])
				if err != nil {
					return nil, err
				}
				r.Connections, err = strconv.Atoi(m[2])
				if err != nil {
					return nil, err
				}
				continue
			}
		}
		if r.Latency.P99 == "" {
			m := regexp.MustCompile("^99.000%\\s+(.*)$").FindStringSubmatch(t)
			if len(m) == 2 {
				r.Latency.P99 = m[1]
				continue
			}
		}
		if r.Latency.P999 == "" {
			m := regexp.MustCompile("^99.900%\\s+(.*)$").FindStringSubmatch(t)
			if len(m) == 2 {
				r.Latency.P999 = m[1]
				continue
			}
		}
		if r.TotalRequests == 0 {
			m := regexp.MustCompile("^([0-9]+) requests in").FindStringSubmatch(t)
			if len(m) == 2 {
				var err error
				r.TotalRequests, err = strconv.Atoi(m[1])
				if err != nil {
					return nil, err
				}
				continue
			}
		}
		if r.Non200Responses == 0 {
			m := regexp.MustCompile("^Non-2xx or 3xx responses: ([0-9]+)$").FindStringSubmatch(t)
			if len(m) == 2 {
				var err error
				r.Non200Responses, err = strconv.Atoi(m[1])
				if err != nil {
					return nil, err
				}
				continue
			}
		}
		if r.RequestPerSecond == 0 {
			m := regexp.MustCompile("^Requests/sec:\\s+([0-9.]+)$").FindStringSubmatch(t)
			if len(m) == 2 {
				var err error
				r.RequestPerSecond, err = strconv.ParseFloat(m[1], 64)
				if err != nil {
					return nil, err
				}
				continue
			}
		}
		if r.TransferPerSecond == "" {
			m := regexp.MustCompile("^Transfer/sec:\\s+([A-Za-z0-9.]+)$").FindStringSubmatch(t)
			if len(m) == 2 {
				r.TransferPerSecond = m[1]
				continue
			}
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return r, nil
}
