package wrk

import (
	"strings"
	"testing"
	"time"
)

func TestBuildReport(t *testing.T) {
	expected := Report{
		Threads:     40,
		Connections: 250,
		TargetURL:   "http://10.0.0.1",
		Time:        1 * time.Minute,
		Latency: LatencyDistribution{
			P99:  "3.81ms",
			P999: "23.60ms",
		},
		Non200Responses:   77,
		RequestPerSecond:  997.95,
		TransferPerSecond: "830.47KB",
		TotalRequests:     59915,
	}
	got, err := BuildReport(strings.NewReader(wrkSampleResult))
	if err != nil {
		t.Fatal(err)
	}
	if expected != *got {
		t.Errorf("expected %#v, got %#v", expected, got)
	}
}

const wrkSampleResult string = `
Running 1m test @ http://10.0.0.1
  40 threads and 250 connections
  Thread calibration: mean lat.: 14.833ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 24.653ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 14.849ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 15.908ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 1.948ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 14.729ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 3.432ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 3.853ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 12.240ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 11.265ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 25.806ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 15.553ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 2.744ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 13.311ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 7.299ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 2.459ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 5.105ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 14.420ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 11.239ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 3.908ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 1.780ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 2.581ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 2.781ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 1.836ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 5.387ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 11.254ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 12.212ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 3.218ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 3.855ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 9.888ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 1.640ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 19.696ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 8.303ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 17.072ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 4.765ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 4.991ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 7.556ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 26.871ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 16.193ms, rate sampling interval: 10ms
  Thread calibration: mean lat.: 10.657ms, rate sampling interval: 10ms
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.67ms    1.13ms  33.66ms   98.16%
    Req/Sec    26.67     77.15   666.00     88.67%
  Latency Distribution (HdrHistogram - Recorded Latency)
 50.000%    1.59ms
 75.000%    1.88ms
 90.000%    2.19ms
 99.000%    3.81ms
 99.900%   23.60ms
 99.990%   31.49ms
 99.999%   33.69ms
100.000%   33.69ms

  Detailed Percentile spectrum:
       Value   Percentile   TotalCount 1/(1-Percentile)

       0.550     0.000000            1         1.00
       1.037     0.100000         4994         1.11
       1.201     0.200000         9989         1.25
       1.338     0.300000        14999         1.43
       1.467     0.400000        19968         1.67
       1.590     0.500000        24975         2.00
       1.647     0.550000        27476         2.22
       1.702     0.600000        29957         2.50
       1.759     0.650000        32452         2.86
       1.818     0.700000        34950         3.33
       1.883     0.750000        37451         4.00
       1.919     0.775000        38718         4.44
       1.959     0.800000        39939         5.00
       2.004     0.825000        41211         5.71
       2.055     0.850000        42442         6.67
       2.115     0.875000        43707         8.00
       2.151     0.887500        44315         8.89
       2.191     0.900000        44946        10.00
       2.235     0.912500        45574        11.43
       2.285     0.925000        46197        13.33
       2.337     0.937500        46802        16.00
       2.373     0.943750        47118        17.78
       2.413     0.950000        47435        20.00
       2.455     0.956250        47738        22.86
       2.509     0.962500        48053        26.67
       2.565     0.968750        48364        32.00
       2.601     0.971875        48520        35.56
       2.655     0.975000        48680        40.00
       2.709     0.978125        48830        45.71
       2.785     0.981250        48986        53.33
       2.945     0.984375        49140        64.00
       3.087     0.985938        49218        71.11
       3.289     0.987500        49296        80.00
       3.593     0.989062        49374        91.43
       3.967     0.990625        49452       106.67
       4.435     0.992188        49531       128.00
       4.715     0.992969        49569       142.22
       4.963     0.993750        49609       160.00
       5.143     0.994531        49647       182.86
       5.347     0.995313        49686       213.33
       5.699     0.996094        49726       256.00
       5.827     0.996484        49745       284.44
       6.007     0.996875        49764       320.00
       6.327     0.997266        49784       365.71
       6.739     0.997656        49803       426.67
       8.871     0.998047        49823       512.00
      13.703     0.998242        49833       568.89
      18.063     0.998437        49842       640.00
      19.647     0.998633        49852       731.43
      21.615     0.998828        49862       853.33
      24.287     0.999023        49873      1024.00
      24.815     0.999121        49877      1137.78
      25.391     0.999219        49881      1280.00
      25.999     0.999316        49886      1462.86
      28.575     0.999414        49891      1706.67
      28.831     0.999512        49896      2048.00
      29.103     0.999561        49899      2275.56
      29.551     0.999609        49901      2560.00
      30.287     0.999658        49903      2925.71
      30.351     0.999707        49906      3413.33
      30.399     0.999756        49908      4096.00
      30.463     0.999780        49910      4551.11
      30.511     0.999805        49911      5120.00
      30.527     0.999829        49912      5851.43
      31.151     0.999854        49913      6826.67
      31.311     0.999878        49914      8192.00
      31.487     0.999890        49915      9102.22
      31.631     0.999902        49916     10240.00
      31.631     0.999915        49916     11702.86
      31.791     0.999927        49917     13653.33
      31.791     0.999939        49917     16384.00
      32.303     0.999945        49918     18204.44
      32.303     0.999951        49918     20480.00
      32.303     0.999957        49918     23405.71
      33.663     0.999963        49919     27306.67
      33.663     0.999969        49919     32768.00
      33.663     0.999973        49919     36408.89
      33.663     0.999976        49919     40960.00
      33.663     0.999979        49919     46811.43
      33.695     0.999982        49920     54613.33
      33.695     1.000000        49920          inf
#[Mean    =        1.668, StdDeviation   =        1.132]
#[Max     =       33.664, Total count    =        49920]
#[Buckets =           27, SubBuckets     =         2048]
----------------------------------------------------------
  59915 requests in 1.00m, 48.69MB read
  Non-2xx or 3xx responses: 77
Requests/sec:    997.95
Transfer/sec:    830.47KB
`
