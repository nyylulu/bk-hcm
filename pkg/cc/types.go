/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

// Package cc ...
package cc

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/logs"
	"hcm/pkg/tools/ssl"
	"hcm/pkg/version"

	etcd3 "go.etcd.io/etcd/client/v3"
)

// Recover 配置是否开启recover服务
type Recover struct {
	EnableApplyRecover   bool `yaml:"enableApplyRecover"`   // 开启申请订单恢复服务
	EnableRecycleRecover bool `yaml:"enableRecycleRecover"` // 开启回收订单恢复服务
}

// Service defines Setting related runtime.
type Service struct {
	Etcd Etcd `yaml:"etcd"`
}

// trySetDefault set the Setting default value if user not configured.
func (s *Service) trySetDefault() {
	s.Etcd.trySetDefault()
}

// validate Setting related runtime.
func (s Service) validate() error {
	if err := s.Etcd.validate(); err != nil {
		return err
	}

	return nil
}

// Etcd defines etcd related runtime
type Etcd struct {
	// Endpoints is a list of URLs.
	Endpoints []string `yaml:"endpoints"`
	// DialTimeoutMS is the timeout seconds for failing
	// to establish a connection.
	DialTimeoutMS uint `yaml:"dialTimeoutMS"`
	// Username is a user's name for authentication.
	Username string `yaml:"username"`
	// Password is a password for authentication.
	Password string    `yaml:"password"`
	TLS      TLSConfig `yaml:"tls"`
}

// trySetDefault set the etcd default value if user not configured.
func (es *Etcd) trySetDefault() {
	if len(es.Endpoints) == 0 {
		es.Endpoints = []string{"127.0.0.1:2379"}
	}

	if es.DialTimeoutMS == 0 {
		es.DialTimeoutMS = 200
	}
}

// ToConfig convert to etcd config.
func (es Etcd) ToConfig() (etcd3.Config, error) {
	var tlsC *tls.Config
	if es.TLS.Enable() {
		var err error
		tlsC, err = ssl.ClientTLSConfVerify(es.TLS.InsecureSkipVerify, es.TLS.CAFile, es.TLS.CertFile,
			es.TLS.KeyFile, es.TLS.Password)
		if err != nil {
			return etcd3.Config{}, fmt.Errorf("init etcd tls config failed, err: %v", err)
		}
	}

	c := etcd3.Config{
		Endpoints:            es.Endpoints,
		AutoSyncInterval:     0,
		DialTimeout:          time.Duration(es.DialTimeoutMS) * time.Millisecond,
		DialKeepAliveTime:    0,
		DialKeepAliveTimeout: 0,
		MaxCallSendMsgSize:   0,
		MaxCallRecvMsgSize:   0,
		TLS:                  tlsC,
		Username:             es.Username,
		Password:             es.Password,
		RejectOldCluster:     false,
		DialOptions:          nil,
		Context:              nil,
		LogConfig:            nil,
		PermitWithoutStream:  false,
	}

	return c, nil
}

// validate etcd runtime
func (es Etcd) validate() error {
	if len(es.Endpoints) == 0 {
		return errors.New("etcd endpoints is not set")
	}

	if err := es.TLS.validate(); err != nil {
		return fmt.Errorf("etcd tls, %v", err)
	}

	return nil
}

// Limiter defines the request limit options
type Limiter struct {
	// QPS should >=1
	QPS uint `yaml:"qps"`
	// Burst should >= 1;
	Burst uint `yaml:"burst"`
}

// validate if the limiter is valid or not.
func (lm Limiter) validate() error {
	if lm.QPS <= 0 {
		return errors.New("invalid QPS value, should >= 1")
	}

	if lm.Burst <= 0 {
		return errors.New("invalid Burst value, should >= 1")
	}

	return nil
}

// trySetDefault try set the default value of limiter
func (lm *Limiter) trySetDefault() {
	if lm.QPS == 0 {
		lm.QPS = 500
	}

	if lm.Burst == 0 {
		lm.Burst = 500
	}
}

// Async defines async relating.
type Async struct {
	Scheduler  Parser     `yaml:"scheduler"`
	Executor   Executor   `yaml:"executor"`
	Dispatcher Dispatcher `yaml:"dispatcher"`
	WatchDog   WatchDog   `yaml:"watchDog"`
}

// Validate Async
func (a Async) Validate() error {
	// 这里不进行校验，统一由异步任务框架进行校验
	return nil
}

// Parser 公共组件，负责获取分配给当前节点的任务流，并解析成任务树后，派发当前要执行的任务给executor执行
type Parser struct {
	WatchIntervalSec uint `yaml:"watchIntervalSec"`
	WorkerNumber     uint `yaml:"workerNumber"`
}

// Executor 公共组件，负责执行异步任务
type Executor struct {
	WorkerNumber       uint `yaml:"workerNumber"`
	TaskExecTimeoutSec uint `yaml:"taskExecTimeoutSec"`
}

// Dispatcher 主节点组件，负责派发任务
type Dispatcher struct {
	WatchIntervalSec uint `yaml:"watchIntervalSec"`
}

// WatchDog 主节点组件，负责异常任务修正（超时任务，任务处理节点已经挂掉的任务等）
type WatchDog struct {
	WatchIntervalSec uint `yaml:"watchIntervalSec"`
	TaskTimeoutSec   uint `yaml:"taskTimeoutSec"`
}

// DataBase defines database related runtime
type DataBase struct {
	Resource ResourceDB `yaml:"resource"`
	// MaxSlowLogLatencyMS defines the max tolerance in millisecond to execute
	// the database command, if the cost time of execute have >= the MaxSlowLogLatencyMS
	// then this request will be logged.
	MaxSlowLogLatencyMS uint `yaml:"maxSlowLogLatencyMS"`
	// Limiter defines request's to ORM's limitation for each sharding, and
	// each sharding have the independent request limitation.
	Limiter *Limiter `yaml:"limiter"`
}

// trySetDefault set the sharding default value if user not configured.
func (s *DataBase) trySetDefault() {
	s.Resource.trySetDefault()

	if s.MaxSlowLogLatencyMS == 0 {
		s.MaxSlowLogLatencyMS = 100
	}

	if s.Limiter == nil {
		s.Limiter = new(Limiter)
	}

	s.Limiter.trySetDefault()
}

// validate sharding runtime
func (s DataBase) validate() error {
	if err := s.Resource.validate(); err != nil {
		return err
	}

	if s.MaxSlowLogLatencyMS <= 0 {
		return errors.New("invalid maxSlowLogLatencyMS")
	}

	if s.Limiter != nil {
		if err := s.Limiter.validate(); err != nil {
			return fmt.Errorf("sharding.limiter is invalid, %v", err)
		}
	}

	return nil
}

// ResourceDB defines database related runtime.
type ResourceDB struct {
	// Endpoints is a seed list of host:port addresses of database nodes.
	Endpoints []string `yaml:"endpoints"`
	Database  string   `yaml:"database"`
	User      string   `yaml:"user"`
	Password  string   `yaml:"password"`
	// DialTimeoutSec is timeout in seconds to wait for a
	// response from the db server
	// all the timeout default value reference:
	// https://dev.mysql.com/doc/refman/8.0/en/server-system-variables.html
	DialTimeoutSec    uint      `yaml:"dialTimeoutSec"`
	ReadTimeoutSec    uint      `yaml:"readTimeoutSec"`
	WriteTimeoutSec   uint      `yaml:"writeTimeoutSec"`
	MaxIdleTimeoutMin uint      `yaml:"maxIdleTimeoutMin"`
	MaxOpenConn       uint      `yaml:"maxOpenConn"`
	MaxIdleConn       uint      `yaml:"maxIdleConn"`
	TLS               TLSConfig `yaml:"tls"`
}

// trySetDefault set the database's default value if user not configured.
func (ds *ResourceDB) trySetDefault() {
	if len(ds.Endpoints) == 0 {
		ds.Endpoints = []string{"127.0.0.1:3306"}
	}

	if ds.DialTimeoutSec == 0 {
		ds.DialTimeoutSec = 15
	}

	if ds.ReadTimeoutSec == 0 {
		ds.ReadTimeoutSec = 10
	}

	if ds.WriteTimeoutSec == 0 {
		ds.WriteTimeoutSec = 10
	}

	if ds.MaxOpenConn == 0 {
		ds.MaxOpenConn = 500
	}

	if ds.MaxIdleConn == 0 {
		ds.MaxIdleConn = 5
	}
}

// validate database runtime.
func (ds ResourceDB) validate() error {
	if len(ds.Endpoints) == 0 {
		return errors.New("database endpoints is not set")
	}

	if len(ds.Database) == 0 {
		return errors.New("database is not set")
	}

	if (ds.DialTimeoutSec > 0 && ds.DialTimeoutSec < 1) || ds.DialTimeoutSec > 60 {
		return errors.New("invalid database dialTimeoutMS, should be in [1:60]s")
	}

	if (ds.ReadTimeoutSec > 0 && ds.ReadTimeoutSec < 1) || ds.ReadTimeoutSec > 60 {
		return errors.New("invalid database readTimeoutMS, should be in [1:60]s")
	}

	if (ds.WriteTimeoutSec > 0 && ds.WriteTimeoutSec < 1) || ds.WriteTimeoutSec > 30 {
		return errors.New("invalid database writeTimeoutMS, should be in [1:30]s")
	}

	if err := ds.TLS.validate(); err != nil {
		return fmt.Errorf("database tls, %v", err)
	}

	return nil
}

// LogOption defines log's related configuration
type LogOption struct {
	LogDir           string `yaml:"logDir"`
	MaxPerFileSizeMB uint32 `yaml:"maxPerFileSizeMB"`
	MaxPerLineSizeKB uint32 `yaml:"maxPerLineSizeKB"`
	MaxFileNum       uint   `yaml:"maxFileNum"`
	LogAppend        bool   `yaml:"logAppend"`
	// log the log to std err only, it can not be used with AlsoToStdErr
	// at the same time.
	ToStdErr bool `yaml:"toStdErr"`
	// log the log to file and also to std err. it can not be used with ToStdErr
	// at the same time.
	AlsoToStdErr bool `yaml:"alsoToStdErr"`
	Verbosity    uint `yaml:"verbosity"`
}

// trySetDefault set the log's default value if user not configured.
func (log *LogOption) trySetDefault() {
	if len(log.LogDir) == 0 {
		log.LogDir = "./"
	}

	if log.MaxPerFileSizeMB == 0 {
		log.MaxPerFileSizeMB = 500
	}

	if log.MaxPerLineSizeKB == 0 {
		log.MaxPerLineSizeKB = 5
	}

	if log.MaxFileNum == 0 {
		log.MaxFileNum = 5
	}
}

// Logs convert it to logs.LogConfig.
func (log LogOption) Logs() logs.LogConfig {
	l := logs.LogConfig{
		LogDir:             log.LogDir,
		LogMaxSize:         log.MaxPerFileSizeMB,
		LogLineMaxSize:     log.MaxPerLineSizeKB,
		LogMaxNum:          log.MaxFileNum,
		RestartNoScrolling: log.LogAppend,
		ToStdErr:           log.ToStdErr,
		AlsoToStdErr:       log.AlsoToStdErr,
		Verbosity:          log.Verbosity,
	}

	return l
}

// Network defines all the network related options
type Network struct {
	// BindIP is ip where server working on
	BindIP string `yaml:"bindIP"`
	// Port is port where server listen to http port.
	Port uint      `yaml:"port"`
	TLS  TLSConfig `yaml:"tls"`
}

// trySetFlagBindIP try set flag bind ip, bindIP only can set by one of the flag or configuration file.
func (n *Network) trySetFlagBindIP(ip net.IP) error {
	if len(ip) != 0 {
		if len(n.BindIP) != 0 {
			return errors.New("bind ip only can set by one of the flags or configuration file")
		}

		n.BindIP = ip.String()
		return nil
	}

	return nil
}

// trySetDefault set the network's default value if user not configured.
func (n *Network) trySetDefault() {
	if len(n.BindIP) == 0 {
		n.BindIP = "127.0.0.1"
	}
}

// validate network options
func (n Network) validate() error {
	if len(n.BindIP) == 0 {
		return errors.New("network bindIP is not set")
	}

	if ip := net.ParseIP(n.BindIP); ip == nil {
		return errors.New("invalid network bindIP")
	}

	if err := n.TLS.validate(); err != nil {
		return fmt.Errorf("network tls, %v", err)
	}

	return nil
}

// TLSConfig defines tls related options.
type TLSConfig struct {
	// Server should be accessed without verifying the TLS certificate.
	// For testing only.
	InsecureSkipVerify bool `yaml:"insecureSkipVerify"`
	// Server requires TLS client certificate authentication
	CertFile string `yaml:"certFile"`
	// Server requires TLS client certificate authentication
	KeyFile string `yaml:"keyFile"`
	// Trusted root certificates for server
	CAFile string `yaml:"caFile"`
	// the password to decrypt the certificate
	Password string `yaml:"password"`
}

// Enable test tls if enable.
func (tls TLSConfig) Enable() bool {
	if len(tls.CertFile) == 0 &&
		len(tls.KeyFile) == 0 &&
		len(tls.CAFile) == 0 {
		return false
	}

	return true
}

// validate tls configs
func (tls TLSConfig) validate() error {
	if !tls.Enable() {
		return nil
	}

	// TODO: add tls config validate.

	return nil
}

// SysOption is the system's normal option, which is parsed from
// flag commandline.
type SysOption struct {
	ConfigFile string
	// BindIP Setting startup bind ip.
	BindIP net.IP
	// Versioned Setting if show current version info.
	Versioned bool
}

// CheckV check if show current version info.
func (s SysOption) CheckV() {
	if s.Versioned {
		version.ShowVersion()
		os.Exit(0)
	}
}

// IAM defines all the iam related runtime.
type IAM struct {
	// Endpoints is a seed list of host:port addresses of iam nodes.
	Endpoints []string `yaml:"endpoints"`
	// AppCode blueking belong to hcm's appcode.
	AppCode string `yaml:"appCode"`
	// AppSecret blueking belong to hcm app's secret.
	AppSecret string    `yaml:"appSecret"`
	TLS       TLSConfig `yaml:"tls"`
}

// validate iam runtime.
func (s IAM) validate() error {
	if len(s.Endpoints) == 0 {
		return errors.New("iam endpoints is not set")
	}

	if len(s.AppCode) == 0 {
		return errors.New("iam appcode is not set")
	}

	if len(s.AppSecret) == 0 {
		return errors.New("iam app secret is not set")
	}

	if err := s.TLS.validate(); err != nil {
		return fmt.Errorf("iam tls validate failed, err: %v", err)
	}

	return nil
}

// Web 服务依赖所需特有配置， 包括登录、静态文件等配置的定义
type Web struct {
	StaticFileDirPath string `yaml:"staticFileDirPath"`

	BkLoginCookieName      string `yaml:"bkLoginCookieName"`
	BkLoginUrl             string `yaml:"bkLoginUrl"`
	BkComponentApiUrl      string `yaml:"bkComponentApiUrl"`
	BkItsmUrl              string `yaml:"bkItsmUrl"`
	BkDomain               string `yaml:"bkDomain"`
	BkCmdbCreateBizUrl     string `yaml:"bkCmdbCreateBizUrl"`
	BkCmdbCreateBizDocsUrl string `yaml:"bkCmdbCreateBizDocsUrl"`
	EnableCloudSelection   bool   `yaml:"enableCloudSelection"`
	EnableAccountBill      bool   `yaml:"enableAccountBill"`
}

func (s Web) validate() error {
	if len(s.BkLoginUrl) == 0 {
		return errors.New("bk_login_url is not set")
	}

	if len(s.BkComponentApiUrl) == 0 {
		return errors.New("bk_component_api_url is not set")
	}

	if len(s.BkItsmUrl) == 0 {
		return errors.New("bk_itsm_url is not set")
	}

	if len(s.BkDomain) == 0 {
		return errors.New("bk_domain is not set")
	}

	return nil
}

// Esb defines the esb related runtime.
type Esb struct {
	// Endpoints is a seed list of host:port addresses of esb nodes.
	Endpoints []string `yaml:"endpoints"`
	// AppCode is the BlueKing app code of hcm to request esb.
	AppCode string `yaml:"appCode"`
	// AppSecret is the BlueKing app secret of hcm to request esb.
	AppSecret string `yaml:"appSecret"`
	// User is the BlueKing user of hcm to request esb.
	User string    `yaml:"user"`
	TLS  TLSConfig `yaml:"tls"`
}

// validate esb runtime.
func (s Esb) validate() error {
	if len(s.Endpoints) == 0 {
		return errors.New("esb endpoints is not set")
	}
	if len(s.AppCode) == 0 {
		return errors.New("esb app code is not set")
	}
	if len(s.AppSecret) == 0 {
		return errors.New("esb app secret is not set")
	}
	if len(s.User) == 0 {
		return errors.New("esb user is not set")
	}
	if err := s.TLS.validate(); err != nil {
		return fmt.Errorf("validate esb tls failed, err: %v", err)
	}
	return nil
}

// AesGcm Aes Gcm加密
type AesGcm struct {
	Key   string `yaml:"key"`
	Nonce string `yaml:"nonce"`
}

func (a AesGcm) validate() error {
	if len(a.Key) != 16 && len(a.Key) != 32 {
		return errors.New("invalid key, should be 16 or 32 bytes")
	}

	if len(a.Nonce) != 12 {
		return errors.New("invalid nonce, should be 12 bytes")
	}

	return nil
}

// Crypto 定义项目里需要用到的加密，包括选择的算法等
// TODO: 这里默认只支持AES Gcm算法，后续需要支持国密等的选择，可能还需要支持根据不同场景配置不同（比如不同场景，加密的密钥等都不一样）
type Crypto struct {
	AesGcm AesGcm `yaml:"aesGcm"`
}

func (c Crypto) validate() error {
	if err := c.AesGcm.validate(); err != nil {
		return err
	}

	return nil
}

// CloudResource 云资源配置
type CloudResource struct {
	Sync CloudResourceSync `yaml:"sync"`
}

func (c CloudResource) validate() error {
	if err := c.Sync.validate(); err != nil {
		return err
	}

	return nil
}

// CloudResourceSync 云资源同步配置
type CloudResourceSync struct {
	Enable                       bool   `yaml:"enable"`
	SyncIntervalMin              uint64 `yaml:"syncIntervalMin"`
	SyncFrequencyLimitingTimeMin uint64 `yaml:"syncFrequencyLimitingTimeMin"`
}

func (c CloudResourceSync) validate() error {
	if c.Enable {
		if c.SyncFrequencyLimitingTimeMin < 10 {
			return errors.New("syncFrequencyLimitingTimeMin must > 10")
		}
	}

	return nil
}

// Recycle configuration.
type Recycle struct {
	AutoDeleteTime uint `yaml:"autoDeleteTimeHour"`
}

func (a Recycle) validate() error {
	if a.AutoDeleteTime == 0 {
		return errors.New("autoDeleteTimeHour must > 0")
	}

	return nil
}

// BillConfig 账号账单配置
type BillConfig struct {
	Enable          bool   `yaml:"enable"`
	SyncIntervalMin uint64 `yaml:"syncIntervalMin"`
}

func (c BillConfig) validate() error {
	if c.Enable && c.SyncIntervalMin < 1 {
		return errors.New("BillConfig.SyncIntervalMin must >= 1")
	}

	return nil
}

// ApiGateway defines the api gateway config.
type ApiGateway struct {
	// Endpoints is a seed list of host:port addresses of api gateway.
	Endpoints []string `yaml:"endpoints"`
	// AppCode is the BlueKing app code of hcm to request api gateway.
	AppCode string `yaml:"appCode"`
	// AppSecret is the BlueKing app secret of hcm to request api gateway.
	AppSecret string `yaml:"appSecret"`
	// User is the BlueKing user of hcm to request api gateway.
	User string `yaml:"user"`
	// BkTicket is the BlueKing access ticket of hcm to request api gateway.
	BkTicket string `yaml:"bkTicket"`
	// BkToken is the BlueKing access token of hcm to request api gateway.
	BkToken         string    `yaml:"bkToken"`
	ServiceID       int64     `yaml:"serviceID"`
	ApplyLinkFormat string    `yaml:"applyLinkFormat"`
	TLS             TLSConfig `yaml:"tls"`
}

// validate hcm runtime.
func (gt ApiGateway) validate() error {
	if len(gt.Endpoints) == 0 {
		return errors.New("api gateway endpoints is not set")
	}
	if len(gt.AppCode) == 0 {
		return errors.New("app code is not set")
	}
	if len(gt.AppSecret) == 0 {
		return errors.New("app secret is not set")
	}

	if len(gt.BkToken) != 0 && len(gt.BkTicket) != 0 {
		return errors.New("bkToken or bkTicket only one is needed")
	}

	if err := gt.TLS.validate(); err != nil {
		return fmt.Errorf("validate tls failed, err: %v", err)
	}
	return nil
}

// GetAuthValue get auth value.
func (gt ApiGateway) GetAuthValue() string {

	if len(gt.BkTicket) != 0 {
		return fmt.Sprintf("{\"bk_app_code\": \"%s\", \"bk_app_secret\": \"%s\", \"bk_ticket\":\"%s\"}",
			gt.AppCode, gt.AppSecret, gt.BkTicket)
	}

	if len(gt.BkToken) != 0 {
		return fmt.Sprintf("{\"bk_app_code\": \"%s\", \"bk_app_secret\": \"%s\", \"access_token\":\"%s\"}",
			gt.AppCode, gt.AppSecret, gt.BkToken)
	}

	return fmt.Sprintf("{\"bk_app_code\": \"%s\", \"bk_app_secret\": \"%s\"}", gt.AppCode, gt.AppSecret)
}

// CloudSelection define cloud selection relation setting.
type CloudSelection struct {
	DefaultSampleOffset  int                       `yaml:"userDistributionSampleOffset"`
	AvgLatencySampleDays int                       `yaml:"avgLatencySampleDays"`
	CoverRate            float64                   `yaml:"coverRate"`
	CoverPingRanges      []ThreshHoldRanges        `yaml:"coverPingRanges"`
	IDCPriceRanges       []ThreshHoldRanges        `yaml:"idcPriceRanges"`
	AlgorithmPlugin      Plugin                    `yaml:"algorithmPlugin"`
	TableNames           CloudSelectionTableNames  `yaml:"tableNames"`
	DataSourceType       string                    `yaml:"dataSourceType"`
	BkBase               BkBase                    `yaml:"bkBase"`
	DefaultIdcPrice      map[enumor.Vendor]float64 `yaml:"defaultIdcPrice"`
}

// Plugin outside binary plugin
type Plugin struct {
	BinaryPath string   `yaml:"binaryPath"`
	Args       []string `yaml:"args"`
}

// BkBase define bkbase relation setting.
type BkBase struct {
	QueryLimit uint   `yaml:"queryLimit"`
	DataToken  string `yaml:"dataToken"`
	ApiGateway `yaml:"-,inline"`
}

// Validate ...
func (b BkBase) Validate() error {
	if err := b.ApiGateway.validate(); err != nil {
		return err
	}

	if len(b.DataToken) == 0 {
		return errors.New("data token is required")
	}

	return nil
}

// Validate define cloud selection relation setting.
func (c CloudSelection) Validate() error {
	switch c.DataSourceType {
	case "bk_base":
		if err := c.BkBase.validate(); err != nil {
			return err
		}

	default:
		return fmt.Errorf("data source: %s not support", c.DataSourceType)
	}

	return nil
}

// CloudSelectionTableNames ...
type CloudSelectionTableNames struct {
	LatencyPingProvinceIdc   string `yaml:"latencyPingProvinceIdc"`
	LatencyBizProvinceIdc    string `yaml:"latencyBizProvinceIdc"`
	UserCountryDistribution  string `yaml:"userCountryDistribution"`
	UserProvinceDistribution string `yaml:"userProvinceDistribution"`
	RecommendDataSource      string `yaml:"recommendDataSource"`
}

// ThreshHoldRanges 评分范围
type ThreshHoldRanges struct {
	Score int   `yaml:"score" json:"score"`
	Range []int `yaml:"range" json:"range"`
}

// MongoDB mongodb config
type MongoDB struct {
	Host                 string `yaml:"host"`
	Port                 string `yaml:"port"`
	Usr                  string `yaml:"usr"`
	Pwd                  string `yaml:"pwd"`
	Database             string `yaml:"database"`
	MaxOpenConns         uint64 `yaml:"maxOpenConns"`
	MaxIdleConns         uint64 `yaml:"maxIdleConns"`
	Mechanism            string `yaml:"mechanism"`
	RsName               string `yaml:"rsName"`
	SocketTimeoutSeconds int    `yaml:"socketTimeoutSeconds"`
}

// validate mongodb.
func (m MongoDB) validate() error {
	if len(m.Host) == 0 {
		return errors.New("mongodb host is not set")
	}
	if len(m.Usr) == 0 {
		return errors.New("mongodb usr is not set")
	}
	if len(m.Pwd) == 0 {
		return errors.New("mongodb pwd is not set")
	}
	if len(m.Database) == 0 {
		return errors.New("mongodb database is not set")
	}
	if len(m.RsName) == 0 {
		return errors.New("mongodb rsName is not set")
	}

	return nil
}

// Redis config
type Redis struct {
	Host         string `yaml:"host"`
	Pwd          string `yaml:"pwd"`
	SentinelPwd  string `yaml:"sentinelPwd"`
	Database     string `yaml:"database"`
	MaxOpenConns int    `yaml:"maxOpenConns"`
	MasterName   string `yaml:"masterName"`
}

// validate redis.
func (r Redis) validate() error {
	if len(r.Host) == 0 {
		return errors.New("redis host is not set")
	}
	if len(r.Pwd) == 0 {
		return errors.New("redis pwd is not set")
	}
	if len(r.Database) == 0 {
		return errors.New("redis database is not set")
	}

	return nil
}

// ClientConfig third-party api client config set
type ClientConfig struct {
	CvmOpt    CVMCliConf `yaml:"cvm"`
	TjjOpt    TjjCli     `yaml:"tjj"`
	XshipOpt  XshipCli   `yaml:"xship"`
	TCloudOpt TCloudCli  `yaml:"tencentcloud"`
	DvmOpt    DVMCli     `yaml:"dvm"`
	ErpOpt    ErpCli     `yaml:"erp"`
	TmpOpt    TmpCli     `yaml:"tmp"`
	Uwork     UworkCli   `yaml:"uwork"`
	GCS       GCSCli     `yaml:"gcs"`
	Tcaplus   TcaplusCli `yaml:"tcaplus"`
	TGW       TGWCli     `yaml:"tgw"`
	L5        L5Cli      `yaml:"l5"`
	Safety    SafetyCli  `yaml:"safety"`
	BkChat    BkChatCli  `yaml:"bkchat"`
	Sops      SopsCli    `yaml:"sops"`
	ITSM      ApiGateway `yaml:"itsm"`
	Ngate     NgateCli   `yaml:"ngate"`
}

func (c ClientConfig) validate() error {
	if err := c.CvmOpt.validate(); err != nil {
		return err
	}

	if err := c.TjjOpt.validate(); err != nil {
		return err
	}

	if err := c.XshipOpt.validate(); err != nil {
		return err
	}

	if err := c.TCloudOpt.validate(); err != nil {
		return err
	}

	if err := c.DvmOpt.validate(); err != nil {
		return err
	}

	if err := c.ErpOpt.validate(); err != nil {
		return err
	}

	if err := c.TmpOpt.validate(); err != nil {
		return err
	}

	if err := c.Uwork.validate(); err != nil {
		return err
	}

	if err := c.GCS.validate(); err != nil {
		return err
	}

	if err := c.Tcaplus.validate(); err != nil {
		return err
	}

	if err := c.TGW.validate(); err != nil {
		return err
	}

	if err := c.L5.validate(); err != nil {
		return err
	}

	if err := c.Safety.validate(); err != nil {
		return err
	}

	if err := c.BkChat.validate(); err != nil {
		return err
	}

	if err := c.Sops.validate(); err != nil {
		return err
	}

	return nil
}

// CVMCliConf yunti client config
type CVMCliConf struct {
	CvmApiAddr        string `yaml:"host"`
	CvmOldApiAddr     string `yaml:"old_host"`
	CvmLaunchPassword string `yaml:"launch_password"`
}

func (c CVMCliConf) validate() error {
	if len(c.CvmApiAddr) == 0 {
		return errors.New("cvm.host is not set")
	}

	if len(c.CvmOldApiAddr) == 0 {
		return errors.New("cvm.old_host is not set")
	}

	if len(c.CvmLaunchPassword) == 0 {
		return errors.New("cvm.launch_password is not set")
	}

	return nil
}

// TjjCli tjj client options
type TjjCli struct {
	// tjj api address
	TjjApiAddr string `yaml:"host"`
	SecretID   string `yaml:"secret_id"`
	SecretKey  string `yaml:"secret_key"`
	Operator   string `yaml:"operator"`
}

func (t TjjCli) validate() error {
	if len(t.TjjApiAddr) == 0 {
		return errors.New("tjj.host is not set")
	}

	if len(t.SecretID) == 0 {
		return errors.New("tjj.secret_id is not set")
	}

	if len(t.SecretKey) == 0 {
		return errors.New("tjj.secret_key is not set")
	}

	if len(t.Operator) == 0 {
		return errors.New("tjj.operator is not set")
	}

	return nil
}

// XshipCli xship client options
type XshipCli struct {
	// Xship api address
	XshipApiAddr string `yaml:"host"`
	ClientID     string `yaml:"client_id"`
	SecretKey    string `yaml:"secret_key"`
}

func (x XshipCli) validate() error {
	if len(x.XshipApiAddr) == 0 {
		return errors.New("xship.host is not set")
	}

	if len(x.ClientID) == 0 {
		return errors.New("xship.client_id is not set")
	}

	if len(x.SecretKey) == 0 {
		return errors.New("xship.secret_key is not set")
	}

	return nil
}

// TCloudCli  tencent cloud client options
type TCloudCli struct {
	Endpoints  TCloudEndpoints  `yaml:"endpoints"`
	Credential TCloudCredential `yaml:"credential"`
}

func (t TCloudCli) validate() error {
	if err := t.Endpoints.validate(); err != nil {
		return err
	}

	if err := t.Credential.validate(); err != nil {
		return err
	}

	return nil
}

// TCloudEndpoints tencent cloud endpoints
type TCloudEndpoints struct {
	Cvm string `yaml:"cvm"`
	Vpc string `yaml:"vpc"`
	Clb string `yaml:"clb"`
}

func (e TCloudEndpoints) validate() error {
	if len(e.Cvm) == 0 {
		return errors.New("tencentcloud.endpoints.cvm is not set")
	}

	if len(e.Vpc) == 0 {
		return errors.New("tencentcloud.endpoints.vpc is not set")
	}

	if len(e.Clb) == 0 {
		return errors.New("tencentcloud.endpoints.clb is not set")
	}

	return nil
}

// TCloudCredential tencent cloud credential
type TCloudCredential struct {
	ID  string `yaml:"id"`
	Key string `yaml:"key"`
}

func (e TCloudCredential) validate() error {
	if len(e.ID) == 0 {
		return errors.New("tencentcloud.credential.id is not set")
	}

	if len(e.Key) == 0 {
		return errors.New("tencentcloud.credential.key is not set")
	}

	return nil
}

// ErpCli erp client options
type ErpCli struct {
	// erp api address
	ErpApiAddr string `yaml:"host"`
}

func (c ErpCli) validate() error {
	if len(c.ErpApiAddr) == 0 {
		return errors.New("erp.host is not set")
	}

	return nil
}

// TmpCli tmp client options
type TmpCli struct {
	// tmp api address
	TMPApiAddr string `yaml:"host"`
}

func (c TmpCli) validate() error {
	if len(c.TMPApiAddr) == 0 {
		return errors.New("tmp.host is not set")
	}

	return nil
}

// UworkCli Uwork client options
type UworkCli struct {
	// Uwork api address
	UworkApiAddr string `yaml:"host"`
}

func (c UworkCli) validate() error {
	if len(c.UworkApiAddr) == 0 {
		return errors.New("uwork.host is not set")
	}

	return nil
}

// GCSCli gcs client options
type GCSCli struct {
	// gcs api address
	GcsApiAddr string `yaml:"host"`
	SecretID   string `yaml:"secret_id"`
	SecretKey  string `yaml:"secret_key"`
	Operator   string `yaml:"operator"`
}

func (c GCSCli) validate() error {
	if len(c.GcsApiAddr) == 0 {
		return errors.New("gcs.host is not set")
	}

	if len(c.SecretID) == 0 {
		return errors.New("gcs.secret_id is not set")
	}

	if len(c.SecretKey) == 0 {
		return errors.New("gcs.secret_key is not set")
	}

	if len(c.Operator) == 0 {
		return errors.New("gcs.operator is not set")
	}

	return nil
}

// TcaplusCli tcaplus client options
type TcaplusCli struct {
	// tcaplus api address
	TcaplusApiAddr string `yaml:"host"`
}

func (c TcaplusCli) validate() error {
	if len(c.TcaplusApiAddr) == 0 {
		return errors.New("tcaplus.host is not set")
	}

	return nil
}

// TGWCli tgw client options
type TGWCli struct {
	// tgw api address
	TgwApiAddr string `yaml:"host"`
}

func (c TGWCli) validate() error {
	if len(c.TgwApiAddr) == 0 {
		return errors.New("tgw.host is not set")
	}

	return nil
}

// L5Cli l5 client options
type L5Cli struct {
	// l5 api address
	L5ApiAddr string `yaml:"host"`
}

func (c L5Cli) validate() error {
	if len(c.L5ApiAddr) == 0 {
		return errors.New("l5.host is not set")
	}

	return nil
}

// SafetyCli Safety client options
type SafetyCli struct {
	// safety api address
	SafetyApiAddr string `yaml:"host"`
}

func (c SafetyCli) validate() error {
	if len(c.SafetyApiAddr) == 0 {
		return errors.New("safety.host is not set")
	}

	return nil
}

// DVMCli dvm client options
type DVMCli struct {
	DvmApiAddr string `yaml:"host"`
	SecretID   string `yaml:"secret_id"`
	SecretKey  string `yaml:"secret_key"`
	Operator   string `yaml:"operator"`
}

func (c DVMCli) validate() error {
	if len(c.DvmApiAddr) == 0 {
		return errors.New("dvm.host is not set")
	}

	if len(c.SecretID) == 0 {
		return errors.New("dvm.secret_id is not set")
	}

	if len(c.SecretKey) == 0 {
		return errors.New("dvm.secret_key is not set")
	}

	if len(c.Operator) == 0 {
		return errors.New("dvm.operator is not set")
	}

	return nil
}

// BkChatCli bkchat client options
type BkChatCli struct {
	BkChatApiAddr string `yaml:"host"`
	NoticeFmt     string `yaml:"notify_format"`
}

func (c BkChatCli) validate() error {
	if len(c.BkChatApiAddr) == 0 {
		return errors.New("bkchat.host is not set")
	}

	return nil
}

// SopsCli sops client options
type SopsCli struct {
	SopsApiAddr string `yaml:"host"`
	AppCode     string `yaml:"app_code"`
	AppSecret   string `yaml:"app_secret"`
	Operator    string `yaml:"operator"`
	DevnetIP    string `yaml:"devnet_download_ip"`
}

func (c SopsCli) validate() error {
	if len(c.SopsApiAddr) == 0 {
		return errors.New("sops.host is not set")
	}

	if len(c.AppCode) == 0 {
		return errors.New("sops.app_code is not set")
	}

	if len(c.AppSecret) == 0 {
		return errors.New("sops.app_secret is not set")
	}

	if len(c.DevnetIP) == 0 {
		return errors.New("sops.devnet_download_ip is not set")
	}

	return nil
}

// ItsmFlow defines the itsm flow related runtime.
type ItsmFlow struct {
	// ServiceName is the itsm service name.
	ServiceName string `yaml:"serviceName"`
	// ServiceID is the itsm service id.
	ServiceID int64 `yaml:"serviceID"`
	// StateNodes is the itsm state nodes.
	StateNodes []StateNode `yaml:"stateNodes"`
	// RedirectUrlTemplate is the itsm service redirect url template.
	RedirectUrlTemplate string `yaml:"redirectUrlTemplate"`
}

// validate ItsmFlow runtime.
func (i ItsmFlow) validate() error {
	if i.ServiceID == 0 {
		return errors.New("itsm service id is not set")
	}

	for _, stateNode := range i.StateNodes {
		if err := stateNode.validate(); err != nil {
			return err
		}
	}

	return nil
}

// ResourceDissolve resource dissolve config
type ResourceDissolve struct {
	OriginDate string `yaml:"originDate"`
}

func (r ResourceDissolve) validate() error {
	if len(r.OriginDate) == 0 {
		return errors.New("resourceDissolve.originDate is not set")
	}

	return nil
}

// StateNode defines the itsm state node related runtime.
type StateNode struct {
	ID          int64  `yaml:"id"`
	NodeName    string `yaml:"nodeName"`
	Approver    string `json:"approver"`
	ApprovalKey string `yaml:"approvalKey"`
	RemarkKey   string `yaml:"remarkKey"`
}

// validate StateNode runtime.
func (i StateNode) validate() error {
	if i.ID == 0 {
		return errors.New("state node id is not set")
	}

	if len(i.NodeName) == 0 {
		return errors.New("state node name is not set")
	}

	if len(i.ApprovalKey) == 0 {
		return errors.New("state node approval key is not set")
	}

	if len(i.RemarkKey) == 0 {
		return errors.New("state node remark key is not set")
	}
	return nil
}

// ObjectStore object store config
type ObjectStore struct {
	Type              string `yaml:"type"`
	ObjectStoreTCloud `yaml:",inline"`
}

// ObjectStoreTCloud tencent cloud cos config
type ObjectStoreTCloud struct {
	UIN             string `yaml:"uin"`
	COSPrefix       string `yaml:"prefix"`
	COSSecretID     string `yaml:"secretId"`
	COSSecretKey    string `yaml:"secretKey"`
	COSBucketURL    string `yaml:"bucketUrl"`
	CosBucketName   string `yaml:"bucketName"`
	CosBucketRegion string `yaml:"bucketRegion"`
	COSIsDebug      bool   `yaml:"isDebug"`
}

// Validate do validate
func (ost ObjectStoreTCloud) Validate() error {
	if len(ost.COSSecretID) == 0 {
		return errors.New("cos secret_id cannot be empty")
	}
	if len(ost.COSSecretKey) == 0 {
		return errors.New("cos secret_key cannot be empty")
	}
	if len(ost.COSBucketURL) == 0 {
		return errors.New("cos bucket_url cannot be empty")
	}
	if len(ost.CosBucketName) == 0 {
		return errors.New("cos bucket_name cannot be empty")
	}
	if len(ost.CosBucketRegion) == 0 {
		return errors.New("cos bucket_region cannot be empty")
	}
	if len(ost.UIN) == 0 {
		return errors.New("cos uin cannot be empty")
	}
	return nil
}

// Es elasticsearch config
type Es struct {
	Url      string    `json:"url"`
	User     string    `json:"user"`
	Password string    `json:"password"`
	TLS      TLSConfig `yaml:"tls"`
}

func (e Es) validate() error {
	if len(e.Url) == 0 {
		return errors.New("elasticsearch.url is not set")
	}

	if len(e.User) == 0 {
		return errors.New("elasticsearch.user is not set")
	}

	if len(e.Password) == 0 {
		return errors.New("elasticsearch.password is not set")
	}

	if err := e.TLS.validate(); err != nil {
		return fmt.Errorf("validate tls failed, err: %v", err)
	}

	return nil
}

var (
	defaultControllerSyncDuration         = 30 * time.Second
	defaultMainAccountSummarySyncDuration = 10 * time.Minute
	defaultRootAccountSummarySyncDuration = 10 * time.Minute
	defaultDailySummarySyncDuration       = 30 * time.Second
)

// BillControllerOption bill controller option
type BillControllerOption struct {
	// 是否关闭整个账单同步，默认为不关闭
	Disable                        bool           `yaml:"disable"`
	ControllerSyncDuration         *time.Duration `yaml:"controllerSyncDuration,omitempty"`
	MainAccountSummarySyncDuration *time.Duration `yaml:"mainAccountSummarySyncDuration,omitempty"`
	RootAccountSummarySyncDuration *time.Duration `yaml:"rootAccountSummarySyncDuration,omitempty"`
	DailySummarySyncDuration       *time.Duration `yaml:"dailySummarySyncDuration,omitempty"`
}

func (bco *BillControllerOption) trySetDefault() {
	if bco.ControllerSyncDuration == nil {
		bco.ControllerSyncDuration = &defaultControllerSyncDuration
	}
	if bco.MainAccountSummarySyncDuration == nil {
		bco.MainAccountSummarySyncDuration = &defaultMainAccountSummarySyncDuration
	}
	if bco.RootAccountSummarySyncDuration == nil {
		bco.RootAccountSummarySyncDuration = &defaultRootAccountSummarySyncDuration
	}
	if bco.DailySummarySyncDuration == nil {
		bco.DailySummarySyncDuration = &defaultDailySummarySyncDuration
	}
}

// CMSI cmsi config
type CMSI struct {
	CC         []string `yaml:"cc"`
	Sender     string   `yaml:"sender"`
	ApiGateway `yaml:"-,inline"`
}

// Validate do validate
func (c *CMSI) validate() error {
	if err := c.ApiGateway.validate(); err != nil {
		return err
	}

	if c.CC == nil || len(c.CC) == 0 {
		c.CC = make([]string, 0)
	}

	if len(c.Sender) == 0 {
		return errors.New("sender cannot be empty")
	}

	return nil
}

// Jarvis 财务侧api配置
type Jarvis struct {
	AppID     string    `yaml:"appID"`
	AppKey    string    `yaml:"appKey"`
	Endpoints []string  `yaml:"endpoints"`
	TLS       TLSConfig `yaml:"tls"`
}

// validate do validate
func (c *Jarvis) validate() error {
	if len(c.Endpoints) == 0 {
		return errors.New("jarvis endpoints is empty")
	}
	return nil
}

// ExchangeRate 汇率自动拉取配置
type ExchangeRate struct {
	EnablePull      bool                  `yaml:"enablePull"`
	PullIntervalMin uint64                `yaml:"pullIntervalMin"`
	ToCurrency      []enumor.CurrencyCode `yaml:"toCurrency"`
	FromCurrency    []enumor.CurrencyCode `yaml:"fromCurrency"`
}

func (r *ExchangeRate) trySetDefault() {
	if r.PullIntervalMin == 0 {
		r.PullIntervalMin = 120
	}
	if len(r.ToCurrency) == 0 {
		r.ToCurrency = []enumor.CurrencyCode{enumor.CurrencyCNY}
	}
}

// IEGObsOption OBS 账单拉取配置
type IEGObsOption struct {
	Endpoints []string  `yaml:"endpoints"`
	APIKey    string    `yaml:"apiKey"`
	TLS       TLSConfig `yaml:"tls"`
}

// validate hcm runtime.
func (gt IEGObsOption) validate() error {
	if len(gt.Endpoints) == 0 {
		return errors.New("obs endpoints is not set")
	}
	if len(gt.APIKey) == 0 {
		return errors.New("obs api key is not set")
	}

	if err := gt.TLS.validate(); err != nil {
		return fmt.Errorf("validate obs tls failed, err: %v", err)
	}
	return nil
}

// AwsSavingsPlansOption savings plans allocation option
type AwsSavingsPlansOption struct {
	// RootAccountCloudID which root account these savings plans belongs to
	RootAccountCloudID string `yaml:"rootAccountCloudID" validate:"required"`
	// SpArnPrefix arn prefix to match savings plans, empty for no filter
	SpArnPrefix string `yaml:"spArnPrefix" validate:"omitempty"`
	// SpPurchaseAccountCloudID which account purchase this saving plans,
	// the cost of savings plans will be added to this account as income
	SpPurchaseAccountCloudID string `yaml:"SpPurchaseAccountCloudID" validate:"required"`
}

func (opt *AwsSavingsPlansOption) validate() error {
	if opt.RootAccountCloudID == "" {
		return errors.New("root account cloud id cannot be empty for aws savings plans")
	}

	if opt.SpPurchaseAccountCloudID == "" {
		return errors.New("sp purchase account cloud id cannot be empty for aws savings plans")
	}
	return nil
}

// BillCommonExpense ...
type BillCommonExpense struct {
	ExcludeAccountCloudIDs []string `yaml:"excludeAccountCloudIDs" validate:"dive,required"`
}

// CreditReturn ...
type CreditReturn struct {
	CreditID string `yaml:"creditId" validate:"required"`
	// which account this credit will return to
	AccountCloudID string `yaml:"accountCloudID" validate:"required"`
	CreditName     string `yaml:"creditName" `
}

// Validate ...
func (r CreditReturn) Validate() error {
	if r.CreditID == "" {
		return errors.New("credit id cannot be empty")
	}
	if r.AccountCloudID == "" {
		return errors.New("account cloud id cannot be empty")
	}
	return nil
}

// GcpCreditConfig ...
type GcpCreditConfig struct {
	// RootAccountCloudID which root account these savings plans belongs to
	RootAccountCloudID string         `yaml:"rootAccountCloudID" validate:"required"`
	ReturnConfigs      []CreditReturn `yaml:"returnConfigs" validate:"required,dive,required"`
}

// Validate ...
func (opt *GcpCreditConfig) Validate() error {
	if opt.RootAccountCloudID == "" {
		return errors.New("root account cloud id cannot be empty for gcp credits config")
	}
	if len(opt.ReturnConfigs) == 0 {
		return errors.New("return configs cannot be empty for gcp credits config")
	}
	for i := range opt.ReturnConfigs {
		if err := opt.ReturnConfigs[i].Validate(); err != nil {
			return errors.New(fmt.Sprintf("gcp credit return config index %d validation failed, %v", i, err))
		}
	}
	return nil
}

// BillAllocationOption ...
type BillAllocationOption struct {
	AwsSavingsPlans  []AwsSavingsPlansOption `yaml:"awsSavingsPlans"`
	AwsCommonExpense BillCommonExpense       `yaml:"awsCommonExpense"`
	GcpCredits       []GcpCreditConfig       `yaml:"gcpCredits"`
	GcpCommonExpense BillCommonExpense       `yaml:"gcpCommonExpense"`
}

func (opt *BillAllocationOption) validate() error {
	for i := range opt.AwsSavingsPlans {
		if err := opt.AwsSavingsPlans[i].validate(); err != nil {
			return errors.New(fmt.Sprintf("aws savings plans index %d validation failed, %v", i, err))
		}
	}
	return nil
}

// Notice ...
type Notice struct {
	Enable     bool `yaml:"enable"`
	ApiGateway `yaml:"-,inline"`
}

// Validate do validate
func (c *Notice) validate() error {
	if !c.Enable {
		return nil
	}
	if err := c.ApiGateway.validate(); err != nil {
		return err
	}

	return nil
}

// NgateCli sops client options
type NgateCli struct {
	Host      string `yaml:"host"`
	AppCode   string `yaml:"app_code"`
	AppSecret string `yaml:"app_secret"`
}

func (c NgateCli) validate() error {
	if len(c.Host) == 0 {
		return errors.New("ngate.host is not set")
	}

	if len(c.AppCode) == 0 {
		return errors.New("ngate.app_code is not set")
	}

	if len(c.AppSecret) == 0 {
		return errors.New("ngate.app_secret is not set")
	}

	return nil
}

// RollingServer 滚服相关配置
type RollingServer struct {
	SyncBill bool `yaml:"syncBill"`
}

// MOA MOA api配置
type MOA struct {
	PaasID    string    `yaml:"paasID"`
	Token     string    `yaml:"token"`
	Endpoints []string  `yaml:"endpoints"`
	TLS       TLSConfig `yaml:"tls"`
}

// validate do validate
func (c *MOA) validate() error {
	if len(c.Endpoints) == 0 {
		return errors.New("moa endpoints is empty")
	}
	return nil
}
