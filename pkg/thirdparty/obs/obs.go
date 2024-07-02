package obs

import (
	"fmt"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"

	"github.com/prometheus/client_golang/prometheus"
)

// BaseRequest base request of obs
type BaseRequest struct {
	JSONRPC string             `json:"jsonrpc"`
	ID      string             `json:"id"`
	Method  string             `json:"method"`
	Params  []BaseRequestParam `json:"param"`
}

// BaseRequestParam base request param
type BaseRequestParam struct {
	YearMonth       int64  `json:"yearMonth"`
	AccountInfoList string `json:"accountInfoList"`
}

// AccountInfo obs account info
type AccountInfo struct {
	AccountType string `json:"accountType"`
	Total       uint64 `json:"total"`
	Column      string `json:"string"`
	SumColValue int    `json:"sumColValue"`
}

// BaseResponse base response of obs
type BaseResponse struct {
	JSONRPC string              `json:"jsonrpc"`
	ID      string              `json:"id"`
	Result  *BaseResponseResult `json:"result,omitempty"`
}

// BaseResponseResult base response result
type BaseResponseResult struct {
	Data    string `json:"data"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

// ObsDiscover implements Discovery
type ObsDiscover struct {
	Servers []string
}

// GetServers get server addresses
func (od *ObsDiscover) GetServers() ([]string, error) {
	return od.Servers, nil
}

// Client obs interface
type Client interface {
}

// IEGObsOption option of ieg obs option
type IEGObsOption struct {
	Endpoints []string
	APIKey    string
}

// IEGObs obs client
type IEGObs struct {
	Option *IEGObsOption
	Client rest.ClientInterface
}

// NewIEGObs create new ieg obs client
func NewIEGObs(opt *IEGObsOption, reg prometheus.Registerer) (*IEGObs, error) {
	cli, err := client.NewClient(nil)
	if err != nil {
		return nil, err
	}
	dis := &ObsDiscover{}
	dis.Servers = append(dis.Servers, opt.Endpoints...)
	cap := &client.Capability{
		Client:     cli,
		Discover:   dis,
		MetricOpts: client.MetricOption{Register: reg},
	}
	restCli := rest.NewClient(cap, "/")
	return &IEGObs{
		Option: opt,
		Client: restCli,
	}, nil
}

// NotifyObsPullIegBill
func (io *IEGObs) NotifyObsPullIegBill(kt *kit.Kit, req *BaseRequest) error {
	url := "/obs-api?api_key=%s"
	resp := new(BaseResponse)
	err := io.Client.Verb(rest.GET).
		SubResourcef(url, io.Option.APIKey).
		WithContext(kt.Ctx).
		WithHeaders(kt.Header()).
		Body(req).
		Do().Into(resp)

	if err != nil {
		logs.Errorf("fail to call obs api, err: %v, url: %s, rid: %s", err, url, kt.Rid)
		return err
	}

	if resp.Result == nil || resp.Result.Status != 0 {
		err := fmt.Errorf("failed to call obs api, resp %v", resp)
		logs.Errorf("obs api returns error, url: %s, err: %v, rid: %s", url, err, kt.Rid)
		return err
	}
	return nil
}
