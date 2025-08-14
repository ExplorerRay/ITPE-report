package input

import (
	"time"

	"github.com/explorerray/itpe-report/internal/client/promclient"
)

type KeplerPowerMetrics struct {
	NodePlatformJ float64
	NodeGPUJ      float64
	NodePackageJ  float64
	NodeDRAMJ     float64
	NodeOtherJ    float64
	PodGPUJ       float64
	PodDRAMJ      float64
	PodPackageJ   float64
	PodPlatformJ  float64
	PodOtherJ     float64
}

func GetPowerMetrics(exp Experiment) KeplerPowerMetrics {
	reqs := exp.Requests
	lastReq := reqs[len(reqs)-1]
	// for Power metrics collected by Kepler
	expBegin := reqs[0].Timestamp
	expEnd := lastReq.ResponseTimestamps[len(lastReq.ResponseTimestamps)-1]

	expBeginQuery := promclient.MakeKeplerQueryInfo(time.Unix(0, expBegin), "ollama")
	expEndQuery := promclient.MakeKeplerQueryInfo(time.Unix(0, expEnd), "ollama")

	beginResp := promclient.MultiQuery(expBeginQuery)
	endResp := promclient.MultiQuery(expEndQuery)

	// Sum up if there are several results
	return KeplerPowerMetrics{
		NodePlatformJ: promclient.SumResults(endResp[0].Results) - promclient.SumResults(beginResp[0].Results),
		NodeGPUJ:      promclient.SumResults(endResp[1].Results) - promclient.SumResults(beginResp[1].Results),
		NodePackageJ:  promclient.SumResults(endResp[2].Results) - promclient.SumResults(beginResp[2].Results),
		NodeDRAMJ:     promclient.SumResults(endResp[3].Results) - promclient.SumResults(beginResp[3].Results),
		NodeOtherJ:    promclient.SumResults(endResp[4].Results) - promclient.SumResults(beginResp[4].Results),
		PodGPUJ:       promclient.SumResults(endResp[5].Results) - promclient.SumResults(beginResp[5].Results),
		PodDRAMJ:      promclient.SumResults(endResp[6].Results) - promclient.SumResults(beginResp[6].Results),
		PodPackageJ:   promclient.SumResults(endResp[7].Results) - promclient.SumResults(beginResp[7].Results),
		PodPlatformJ:  promclient.SumResults(endResp[8].Results) - promclient.SumResults(beginResp[8].Results),
		PodOtherJ:     promclient.SumResults(endResp[9].Results) - promclient.SumResults(beginResp[9].Results),
	}
}
