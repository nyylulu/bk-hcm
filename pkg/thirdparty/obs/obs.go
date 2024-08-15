package obs

import (
	"errors"
	"fmt"

	"hcm/pkg/cc"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"
	"hcm/pkg/tools/ssl"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shopspring/decimal"
)

// OBSAccountType OBS侧账号类型标识
type OBSAccountType string

const (
	// AccountTypeGCP GCP
	AccountTypeGCP OBSAccountType = "GCP"
	// AccountTypeHuawei Huawei
	AccountTypeHuawei OBSAccountType = "Huawei"
	// AccountTypeAzure Azure
	AccountTypeAzure OBSAccountType = "Azure"
	// AccountTypeAzureCN Azure_CN 中国站
	AccountTypeAzureCN OBSAccountType = "Azure_CN"
	// AccountTypeAws Aws
	AccountTypeAws OBSAccountType = "AWS"
	// AccountTypeZenlayer Zenlayer
	AccountTypeZenlayer OBSAccountType = "Zenlayer"
)

// BaseRequest base request of obs
type BaseRequest[T any] struct {
	JSONRPC string `json:"jsonrpc"`
	ID      string `json:"id"`
	Method  string `json:"method"`
	Params  *T     `json:"params"`
}

// NotifyObsPullReq base request param
type NotifyObsPullReq struct {
	// 格式为202406
	YearMonth       int64         `json:"yearMonth" validate:"required"`
	AccountInfoList []AccountInfo `json:"accountInfoList" validate:"required,min=1,dive,required"`
}

// Validate ...
func (r *NotifyObsPullReq) Validate() error {
	return validator.Validate.Struct(r)
}

// AccountInfo obs account info
type AccountInfo struct {
	// 账号类型 枚举值【GCP，Huawei，Azure，Azure_CN，AWS，Zenlayer】
	AccountType OBSAccountType `json:"accountType" validate:"required"`
	// 数据总条数（用于拉数完成后对账）
	Total uint64 `json:"total" validate:"required"`
	// 对账字段。比如 【总成本】字段（用于拉数完成后对账）
	Column string `json:"column" validate:"required"`
	// sum(对账字段)，对账字段的累积和（用于拉数完成后对账）
	SumColValue decimal.Decimal `json:"sumColValue"`
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
	// NotifyRePull 通知obs重新拉取账单
	NotifyRePull(kt *kit.Kit, req *NotifyObsPullReq) error
}

// IEGObs obs client
type IEGObs struct {
	Option *cc.IEGObsOption
	Client rest.ClientInterface
}

// NewIEGObs create new ieg obs client
func NewIEGObs(opt *cc.IEGObsOption, reg prometheus.Registerer) (Client, error) {

	cli, err := client.NewClient(&ssl.TLSConfig{
		InsecureSkipVerify: opt.TLS.InsecureSkipVerify,
		CertFile:           opt.TLS.CertFile,
		KeyFile:            opt.TLS.KeyFile,
		CAFile:             opt.TLS.CAFile,
		Password:           opt.TLS.Password,
	})
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
	restCli := rest.NewClient(cap, "/jsonrpc")
	return &IEGObs{
		Option: opt,
		Client: restCli,
	}, nil
}

// NotifyRePull 通知obs重新拉取账单
func (io *IEGObs) NotifyRePull(kt *kit.Kit, req *NotifyObsPullReq) error {

	if req == nil {
		return errors.New("NotifyObsPullReq is required")
	}
	if err := req.Validate(); err != nil {
		return err
	}
	url := "/obs-api?api_key=%s"
	resp := new(BaseResponse)

	// OBS侧要求`accountInfoList`参数作为格式化后的字符串传参，因此这里创建一个临时结构体行参数转换
	type obsNotifyObsPullReq struct {
		YearMonth       int64  `json:"yearMonth" `
		AccountInfoList string `json:"accountInfoList"`
	}
	baseReq := BaseRequest[obsNotifyObsPullReq]{
		JSONRPC: "2.0",
		ID:      "0",
		Method:  "rePullIegAccountData",
		Params: &obsNotifyObsPullReq{
			YearMonth:       req.YearMonth,
			AccountInfoList: formatAccountInfo(req),
		},
	}

	err := io.Client.Verb(rest.POST).
		SubResourcef(url, io.Option.APIKey).
		WithContext(kt.Ctx).
		WithHeaders(kt.Header()).
		Body(baseReq).
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

func formatAccountInfo(req *NotifyObsPullReq) string {
	var accountInfoStr = "["
	for _, info := range req.AccountInfoList {
		accountInfoStr += fmt.Sprintf(`{"accountType": "%s","total": %d,"column": "%s","sumColValue": %s},`,
			info.AccountType, info.Total, info.Column, info.SumColValue.String())
	}
	if len(req.AccountInfoList) > 0 {
		accountInfoStr = accountInfoStr[:len(accountInfoStr)-1]
	}
	accountInfoStr += "]"
	return accountInfoStr
}
