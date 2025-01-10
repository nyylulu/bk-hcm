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

// Package service ...
package service

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"hcm/cmd/woa-server/logics/biz"
	disLogics "hcm/cmd/woa-server/logics/dissolve"
	gclogics "hcm/cmd/woa-server/logics/green-channel"
	planctrl "hcm/cmd/woa-server/logics/plan"
	ressynclogics "hcm/cmd/woa-server/logics/res-sync"
	rslogics "hcm/cmd/woa-server/logics/rolling-server"
	"hcm/cmd/woa-server/logics/task/informer"
	"hcm/cmd/woa-server/logics/task/operation"
	"hcm/cmd/woa-server/logics/task/recoverer"
	"hcm/cmd/woa-server/logics/task/recycler"
	"hcm/cmd/woa-server/logics/task/scheduler"
	"hcm/cmd/woa-server/service/capability"
	"hcm/cmd/woa-server/service/config"
	"hcm/cmd/woa-server/service/cvm"
	"hcm/cmd/woa-server/service/dissolve"
	greenchannel "hcm/cmd/woa-server/service/green-channel"
	"hcm/cmd/woa-server/service/meta"
	"hcm/cmd/woa-server/service/plan"
	"hcm/cmd/woa-server/service/pool"
	ressync "hcm/cmd/woa-server/service/res-sync"
	rollingserver "hcm/cmd/woa-server/service/rolling-server"
	"hcm/cmd/woa-server/service/task"
	"hcm/cmd/woa-server/storage/dal/mongo"
	"hcm/cmd/woa-server/storage/dal/mongo/local"
	"hcm/cmd/woa-server/storage/dal/redis"
	"hcm/cmd/woa-server/storage/driver/mongodb"
	redisCli "hcm/cmd/woa-server/storage/driver/redis"
	"hcm/cmd/woa-server/storage/stream"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/handler"
	"hcm/pkg/iam/auth"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/metrics"
	"hcm/pkg/rest"
	restcli "hcm/pkg/rest/client"
	"hcm/pkg/runtime/shutdown"
	"hcm/pkg/serviced"
	"hcm/pkg/thirdparty"
	"hcm/pkg/thirdparty/api-gateway/itsm"
	"hcm/pkg/thirdparty/es"
	"hcm/pkg/thirdparty/esb"
	"hcm/pkg/tools/ssl"

	"github.com/emicklei/go-restful/v3"
)

// Service do all the woa server's work
type Service struct {
	client         *client.ClientSet
	dao            dao.Set
	planController *planctrl.Controller
	// EsbClient 调用接入ESB的第三方系统API集合
	esbClient esb.Client
	itsmCli   itsm.Client
	// authorizer 鉴权所需接口集合
	authorizer    auth.Authorizer
	thirdCli      *thirdparty.Client
	clientConf    cc.WoaServerSetting
	schedulerIf   scheduler.Interface
	informerIf    informer.Interface
	recyclerIf    recycler.Interface
	operationIf   operation.Interface
	esCli         *es.EsCli
	rsLogic       rslogics.Logics
	gcLogic       gclogics.Logics
	bizLogic      biz.Logics
	dissolveLogic disLogics.Logics
	resSyncLogic  ressynclogics.Logics
}

// NewService create a service instance.
func NewService(dis serviced.ServiceDiscover, sd serviced.State) (*Service, error) {
	tls := cc.WoaServer().Network.TLS

	var tlsConfig *ssl.TLSConfig
	if tls.Enable() {
		tlsConfig = &ssl.TLSConfig{
			InsecureSkipVerify: tls.InsecureSkipVerify,
			CertFile:           tls.CertFile,
			KeyFile:            tls.KeyFile,
			CAFile:             tls.CAFile,
			Password:           tls.Password,
		}
	}

	// initiate system api client set.
	restCli, err := restcli.NewClient(tlsConfig)
	if err != nil {
		return nil, err
	}
	apiClientSet := client.NewClientSet(restCli, dis)

	// init db client
	daoSet, err := dao.NewDaoSet(cc.WoaServer().Database)
	if err != nil {
		return nil, err
	}

	// 创建ESB Client
	esbConfig := cc.WoaServer().Esb
	esbClient, err := esb.NewClient(&esbConfig, metrics.Register())
	if err != nil {
		return nil, err
	}

	itsmCfg := cc.WoaServer().ITSM
	itsmCli, err := itsm.NewClient(&itsmCfg, metrics.Register())
	if err != nil {
		return nil, err
	}

	// 创建调用第三方平台Client
	thirdCli, err := thirdparty.NewClient(cc.WoaServer().ClientConfig, metrics.Register())
	if err != nil {
		return nil, err
	}

	// create authorizer
	authorizer, err := auth.NewAuthorizer(dis, tls)
	if err != nil {
		return nil, err
	}

	// init redis client
	rConf := cc.WoaServer().Redis
	redisConf, err := redis.NewConf(&rConf)
	if err != nil {
		return nil, err
	}
	if err = redisCli.InitClient("redis", redisConf); err != nil {
		return nil, err
	}

	rsLogics, err := rslogics.New(sd, apiClientSet, esbClient, thirdCli)
	if err != nil {
		logs.Errorf("new rolling server logics failed, err: %v", err)
		return nil, err
	}

	gcLogics, err := gclogics.New(apiClientSet, thirdCli)
	if err != nil {
		logs.Errorf("new green channel logics failed, err: %v", err)
		return nil, err
	}

	bizLogic, err := biz.New(esbClient, authorizer)
	if err != nil {
		logs.Errorf("new biz logic failed, err: %v", err)
		return nil, err
	}

	planCtrl, err := planctrl.New(sd, apiClientSet, daoSet, itsmCli, thirdCli.CVM, esbClient, bizLogic)
	if err != nil {
		logs.Errorf("new plan controller failed, err: %v", err)
		return nil, err
	}

	esCli, err := es.NewEsClient(cc.WoaServer().Es, cc.WoaServer().Blacklist)
	if err != nil {
		return nil, err
	}

	dissolveLogics := disLogics.New(daoSet, esbClient, esCli, thirdCli, cc.WoaServer())

	kt := kit.New()
	// Mongo开关打开才生成Client链接
	var informerIf informer.Interface
	var schedulerIf scheduler.Interface

	// Mongo开关打开才进行Init检测
	if cc.WoaServer().UseMongo {
		loopW, watchDB, err := initMongoDB(kt, dis)
		if err != nil {
			return nil, err
		}

		informerIf, err = informer.New(loopW, watchDB)
		if err != nil {
			logs.Errorf("new informer failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		schedulerIf, err = scheduler.New(kt.Ctx, rsLogics, gcLogics, thirdCli, esbClient, informerIf,
			cc.WoaServer().ClientConfig, planCtrl)
		if err != nil {
			logs.Errorf("new scheduler failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	service := &Service{
		client:         apiClientSet,
		dao:            daoSet,
		esbClient:      esbClient,
		authorizer:     authorizer,
		thirdCli:       thirdCli,
		clientConf:     cc.WoaServer(),
		informerIf:     informerIf,
		schedulerIf:    schedulerIf,
		esCli:          esCli,
		rsLogic:        rsLogics,
		gcLogic:        gcLogics,
		planController: planCtrl,
		bizLogic:       bizLogic,
		dissolveLogic:  dissolveLogics,
	}
	return newOtherClient(kt, service, itsmCli, sd)
}

// initMongoDB init mongodb client and watch client
func initMongoDB(kt *kit.Kit, dis serviced.ServiceDiscover) (stream.LoopInterface, *local.Mongo, error) {
	// init mongodb client
	mConf := cc.WoaServer().MongoDB
	mongoConf, err := mongo.NewConf(&mConf)
	if err != nil {
		return nil, nil, err
	}
	if err = mongodb.InitClient("", mongoConf); err != nil {
		return nil, nil, err
	}

	wConf := cc.WoaServer().Watch
	watchConf, err := mongo.NewConf(&wConf)
	if err != nil {
		return nil, nil, err
	}

	if err = mongodb.InitClient("", watchConf); err != nil {
		return nil, nil, err
	}

	// init task service logics
	loopW, err := stream.NewLoopStream(mongoConf.GetMongoConf(), dis)
	if err != nil {
		logs.Errorf("new loop stream failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, err
	}

	watchDB, err := local.NewMgo(watchConf.GetMongoConf(), time.Minute)
	if err != nil {
		logs.Errorf("new watch mongo client failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, err
	}
	return loopW, watchDB, err
}

func newOtherClient(kt *kit.Kit, service *Service, itsmCli itsm.Client, sd serviced.State) (*Service, error) {
	recyclerIf, err := recycler.New(kt.Ctx, service.thirdCli, service.esbClient, service.authorizer, service.rsLogic,
		service.dissolveLogic)
	if err != nil {
		logs.Errorf("new recycler failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// init recoverer client
	recoverConf := cc.WoaServer().Recover
	if err := recoverer.New(&recoverConf, kt, itsmCli, recyclerIf, service.schedulerIf, service.esbClient.Cmdb(),
		service.thirdCli.Sops); err != nil {
		logs.Errorf("new recoverer failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	operationIf, err := operation.New(kt.Ctx)
	if err != nil {
		logs.Errorf("new operation failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resSyncLogics, err := ressynclogics.New(sd, service.client, service.thirdCli)
	if err != nil {
		logs.Errorf("new resource sync logics failed, err: %v", err)
		return nil, err
	}

	service.clientConf = cc.WoaServer()
	service.recyclerIf = recyclerIf
	service.operationIf = operationIf
	service.resSyncLogic = resSyncLogics
	return service, nil
}

// ListenAndServeRest listen and serve the restful server
func (s *Service) ListenAndServeRest() error {
	root := http.NewServeMux()
	root.HandleFunc("/", s.apiSet().ServeHTTP)
	root.HandleFunc("/healthz", s.Healthz)
	handler.SetCommonHandler(root)

	network := cc.WoaServer().Network
	server := &http.Server{
		Addr:    net.JoinHostPort(network.BindIP, strconv.FormatUint(uint64(network.Port), 10)),
		Handler: root,
	}

	if network.TLS.Enable() {
		tls := network.TLS
		tlsC, err := ssl.ClientTLSConfVerify(tls.InsecureSkipVerify, tls.CAFile, tls.CertFile, tls.KeyFile,
			tls.Password)
		if err != nil {
			return fmt.Errorf("init restful tls config failed, err: %v", err)
		}

		server.TLSConfig = tlsC
	}

	logs.Infof("listen restful server on %s with secure(%v) now.", server.Addr, network.TLS.Enable())

	go func() {
		notifier := shutdown.AddNotifier()
		select {
		case <-notifier.Signal:
			defer notifier.Done()

			logs.Infof("start shutdown restful server gracefully...")

			ctx, cancel := context.WithTimeout(context.TODO(), 20*time.Second)
			defer cancel()
			if err := server.Shutdown(ctx); err != nil {
				logs.Errorf("shutdown restful server failed, err: %v", err)
				return
			}

			logs.Infof("shutdown restful server success...")
		}
	}()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logs.Errorf("serve restful server failed, err: %v", err)
			shutdown.SignalShutdownGracefully()
		}
	}()

	return nil
}

func (s *Service) apiSet() *restful.Container {
	ws := new(restful.WebService)
	ws.Path("/api/v1/woa")
	ws.Produces(restful.MIME_JSON)

	c := &capability.Capability{
		Dao:            s.dao,
		WebService:     ws,
		Authorizer:     s.authorizer,
		PlanController: s.planController,
		EsbClient:      s.esbClient,
		ThirdCli:       s.thirdCli,
		Conf:           s.clientConf,
		SchedulerIf:    s.schedulerIf,
		InformerIf:     s.informerIf,
		RecyclerIf:     s.recyclerIf,
		OperationIf:    s.operationIf,
		EsCli:          s.esCli,
		RsLogic:        s.rsLogic,
		Client:         s.client,
		GcLogic:        s.gcLogic,
		BizLogic:       s.bizLogic,
		DissolveLogic:  s.dissolveLogic,
		ResSyncLogic:   s.resSyncLogic,
	}

	config.InitService(c)
	pool.InitService(c)
	cvm.InitService(c)
	task.InitService(c)
	meta.InitService(c)
	plan.InitService(c)
	dissolve.InitService(c)
	rollingserver.InitService(c)
	greenchannel.InitService(c)
	ressync.InitService(c)

	return restful.NewContainer().Add(c.WebService)
}

// Healthz service health check.
func (s *Service) Healthz(w http.ResponseWriter, r *http.Request) {
	if shutdown.IsShuttingDown() {
		logs.Errorf("service healthz check failed, current service is shutting down")
		w.WriteHeader(http.StatusServiceUnavailable)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealthy, "current service is shutting down"))
		return
	}

	if err := serviced.Healthz(r.Context(), cc.WoaServer().Service); err != nil {
		logs.Errorf("serviced healthz check failed, err: %v", err)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealthy, "serviced healthz error, "+err.Error()))
		return
	}

	rest.WriteResp(w, rest.NewBaseResp(errf.OK, "healthy"))
	return
}
