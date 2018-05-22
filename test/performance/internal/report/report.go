package report

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/heptio/gimbal/test/performance/internal/gimbalbench"
	"github.com/heptio/gimbal/test/performance/internal/wrk"
)

type Latency struct {
	P99  string
	P999 string
}

type ConcurrentConnectionsTestResult struct {
	Connections int
	Latency
}

type BackendServicesCountTestResult struct {
	BackendServices int
	Latency
}

type BackendEndpointsOnSingleServiceTestResult struct {
	Endpoints int
	Latency
}

type IngressCountTestResult struct {
	IngressResources int
	Latency
}

type IngressRouteCountTestResult struct {
	IngressRoutes int
	Latency
}

type TestConfiguration struct {
	StartTime      time.Time
	WrkRequestRate int
	WrkHostNetwork bool
	WrkNodeCount   int
	NginxNodeCount int
}

type Report struct {
	Generated                       time.Time
	TestConfiguration               TestConfiguration
	ConcurrentConnections           []ConcurrentConnectionsTestResult
	BackendServices                 []BackendServicesCountTestResult
	BackendEndpointsOnSingleService []BackendEndpointsOnSingleServiceTestResult
	IngressCount                    []IngressCountTestResult
	IngressRoutes                   []IngressRouteCountTestResult
	KubernetesDiscoveryTime         []gimbalbench.KubernetesDiscoveryTestResult
}

func PrettyPrint(w io.Writer, report Report) error {
	t, err := template.New("prettyreport").Parse(prettyReportTmpl)
	if err != nil {
		return err
	}
	return t.Execute(w, report)
}

func Build(logsDir string) (*Report, error) {
	r := &Report{
		Generated: time.Now(),
	}
	tcFile, err := os.Open(filepath.Join(logsDir, "testconfig.json"))
	if err != nil {
		return nil, err
	}
	tcBytes, err := ioutil.ReadAll(tcFile)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(tcBytes, &r.TestConfiguration); err != nil {
		return nil, err
	}
	dir, err := ioutil.ReadDir(logsDir)
	if err != nil {
		return nil, err
	}
	for _, d := range dir {
		if d.IsDir() {
			err := parseTestLogs(r, filepath.Join(logsDir, d.Name()))
			if err != nil {
				return nil, err
			}
		}
	}
	return r, nil
}

func parseTestLogs(r *Report, testLogs string) (err error) {
	switch filepath.Base(testLogs) {
	default:
		return fmt.Errorf("found test results for test that we are not familiar with: %q", testLogs)
	case "test-concurrent-connections":
		r.ConcurrentConnections, err = parseConcurrentConnectionsLogs(testLogs)
		if err != nil {
			return err
		}
	case "test-num-endpoints":
		r.BackendEndpointsOnSingleService, err = parseBackendEndpointsOnSingleServiceLogs(testLogs)
		if err != nil {
			return err
		}
	case "test-num-ingress":
		r.IngressCount, err = parseIngressCountTestLogs(testLogs)
		if err != nil {
			return err
		}
	case "test-num-ingressroutes":
		r.IngressRoutes, err = parseIngressRoutesTestLogs(testLogs)
		if err != nil {
			return err
		}

	case "test-num-services":
		r.BackendServices, err = parseBackendServicesTestLogs(testLogs)
		if err != nil {
			return err
		}
	case "test-kubernetes-service-discovery-time":
		r.KubernetesDiscoveryTime, err = parseKubernetesDiscoveryTimeLogs(testLogs)
		if err != nil {
			return err
		}
	}
	return nil
}

func parseConcurrentConnectionsLogs(logsDir string) ([]ConcurrentConnectionsTestResult, error) {
	dir, err := ioutil.ReadDir(logsDir)
	if err != nil {
		return nil, err
	}
	var res []ConcurrentConnectionsTestResult
	for _, f := range dir {
		if f.IsDir() {
			return nil, fmt.Errorf("unknown directory %q found inside %q", f.Name(), logsDir)
		}
		file, err := os.Open(filepath.Join(logsDir, f.Name()))
		if err != nil {
			return nil, err
		}
		report, err := wrk.BuildReport(file)
		if err != nil {
			return nil, err
		}
		res = append(res, ConcurrentConnectionsTestResult{Connections: report.Connections, Latency: Latency{P99: report.Latency.P99, P999: report.Latency.P999}})
	}
	return res, nil
}

func parseBackendEndpointsOnSingleServiceLogs(logsDir string) ([]BackendEndpointsOnSingleServiceTestResult, error) {
	dir, err := ioutil.ReadDir(logsDir)
	if err != nil {
		return nil, err
	}
	var res []BackendEndpointsOnSingleServiceTestResult
	for _, f := range dir {
		if f.IsDir() {
			return nil, fmt.Errorf("unknown directory %q found inside %q", f.Name(), logsDir)
		}
		file, err := os.Open(filepath.Join(logsDir, f.Name()))
		if err != nil {
			return nil, err
		}
		report, err := wrk.BuildReport(file)
		if err != nil {
			return nil, err
		}
		matches := regexp.MustCompile("wrk2-test-num-backend-endpoints-(\\d+)-.*\\.log").FindStringSubmatch(f.Name())
		if len(matches) != 2 {
			return nil, fmt.Errorf("could not determine the number of backend endpoints from filename %q", f.Name())
		}
		endpoints, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, err
		}
		res = append(res, BackendEndpointsOnSingleServiceTestResult{Endpoints: endpoints, Latency: Latency{P99: report.Latency.P99, P999: report.Latency.P999}})
	}
	return res, nil
}

func parseIngressCountTestLogs(logsDir string) ([]IngressCountTestResult, error) {
	dir, err := ioutil.ReadDir(logsDir)
	if err != nil {
		return nil, err
	}
	var res []IngressCountTestResult
	for _, f := range dir {
		if f.IsDir() {
			return nil, fmt.Errorf("unknown directory %q found inside %q", f.Name(), logsDir)
		}
		file, err := os.Open(filepath.Join(logsDir, f.Name()))
		if err != nil {
			return nil, err
		}
		report, err := wrk.BuildReport(file)
		if err != nil {
			return nil, err
		}
		matches := regexp.MustCompile("wrk2-test-num-ingresses-(\\d+)-.*\\.log").FindStringSubmatch(f.Name())
		if len(matches) != 2 {
			return nil, fmt.Errorf("could not determine the number of backend endpoints from filename %q", f.Name())
		}
		ingressCount, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, err
		}
		res = append(res, IngressCountTestResult{IngressResources: ingressCount, Latency: Latency{P99: report.Latency.P99, P999: report.Latency.P999}})
	}
	return res, nil
}

func parseIngressRoutesTestLogs(logsDir string) ([]IngressRouteCountTestResult, error) {
	dir, err := ioutil.ReadDir(logsDir)
	if err != nil {
		return nil, err
	}
	var res []IngressRouteCountTestResult
	for _, f := range dir {
		if f.IsDir() {
			return nil, fmt.Errorf("unknown directory %q found inside %q", f.Name(), logsDir)
		}
		file, err := os.Open(filepath.Join(logsDir, f.Name()))
		if err != nil {
			return nil, err
		}
		report, err := wrk.BuildReport(file)
		if err != nil {
			return nil, err
		}
		matches := regexp.MustCompile("wrk2-test-num-ingressroutes-(\\d+)-.*\\.log").FindStringSubmatch(f.Name())
		if len(matches) != 2 {
			return nil, fmt.Errorf("could not determine the number of backend endpoints from filename %q", f.Name())
		}
		ingressRoutes, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, err
		}
		res = append(res, IngressRouteCountTestResult{IngressRoutes: ingressRoutes, Latency: Latency{P99: report.Latency.P99, P999: report.Latency.P999}})
	}
	return res, nil
}

func parseBackendServicesTestLogs(logsDir string) ([]BackendServicesCountTestResult, error) {
	dir, err := ioutil.ReadDir(logsDir)
	if err != nil {
		return nil, err
	}
	var res []BackendServicesCountTestResult
	for _, f := range dir {
		if f.IsDir() {
			return nil, fmt.Errorf("unknown directory %q found inside %q", f.Name(), logsDir)
		}
		file, err := os.Open(filepath.Join(logsDir, f.Name()))
		if err != nil {
			return nil, err
		}
		report, err := wrk.BuildReport(file)
		if err != nil {
			return nil, err
		}
		matches := regexp.MustCompile("wrk2-test-num-backends-(\\d+)-.*\\.log").FindStringSubmatch(f.Name())
		if len(matches) != 2 {
			return nil, fmt.Errorf("could not determine the number of backend endpoints from filename %q", f.Name())
		}
		backendServices, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, err
		}
		res = append(res, BackendServicesCountTestResult{BackendServices: backendServices, Latency: Latency{P99: report.Latency.P99, P999: report.Latency.P999}})
	}
	return res, nil
}

func parseKubernetesDiscoveryTimeLogs(logsDir string) ([]gimbalbench.KubernetesDiscoveryTestResult, error) {
	dir, err := ioutil.ReadDir(logsDir)
	if err != nil {
		return nil, err
	}
	var res []gimbalbench.KubernetesDiscoveryTestResult
	for _, f := range dir {
		if f.IsDir() {
			return nil, fmt.Errorf("unknown directory %q found inside %q", f.Name(), logsDir)
		}
		file, err := os.Open(filepath.Join(logsDir, f.Name()))
		if err != nil {
			return nil, err
		}
		var tr gimbalbench.KubernetesDiscoveryTestResult
		if err := json.NewDecoder(file).Decode(&tr); err != nil {
			return nil, err
		}
		res = append(res, tr)
		file.Close()
	}
	return res, nil
}
