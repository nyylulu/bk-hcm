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

// Package dal provices ...
package dal

import (
	"fmt"

	"hcm/cmd/woa-server/storage/driver/mongodb"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"go.mongodb.org/mongo-driver/mongo"
)

// RunTransaction runs a function in a transaction.
func RunTransaction(kt *kit.Kit, logicFunc func(mongo.SessionContext) error) error {
	session, err := mongodb.Client().GetDBClient().StartSession()
	if err != nil {
		logs.Errorf("create start session failed, err: %v, rid: %s", err, kt.Rid)
		return fmt.Errorf("create start session failed, err: %v, rid: %s", err, kt.Rid)
	}
	defer session.EndSession(kt.Ctx)

	txnErr := mongo.WithSession(kt.Ctx, session, func(sc mongo.SessionContext) error {
		if err = session.StartTransaction(); err != nil {
			logs.Errorf("start transaction failed,  err: %v, rid: %s", err, kt.Rid)
			return err
		}

		err = logicFunc(sc)
		if err != nil {
			return err
		}

		if err = session.CommitTransaction(sc); err != nil {
			logs.Errorf("commit transaction failed, err: %v, rid: %s", err, kt.Rid)
			return fmt.Errorf("commit transaction failed, err: %v", err)
		}

		return nil
	})

	if txnErr != nil {
		logs.Errorf("transaction failed, err: %v, rid: %s", txnErr, kt.Rid)
		return txnErr
	}
	return nil
}

// RunTransactionKit runs a function in a transaction, sub kit version
func RunTransactionKit(kt *kit.Kit, logicFunc func(kt *kit.Kit) error) error {
	session, err := mongodb.Client().GetDBClient().StartSession()
	if err != nil {
		logs.Errorf("create start session failed, err: %v, rid: %s", err, kt.Rid)
		return fmt.Errorf("create start session failed, err: %v, rid: %s", err, kt.Rid)
	}
	subKit := kt.NewSubKit()
	defer session.EndSession(subKit.Ctx)
	sc := mongo.NewSessionContext(subKit.Ctx, session)
	subKit.Ctx = sc

	if err = session.StartTransaction(); err != nil {
		logs.Errorf("start transaction failed,  err: %v, rid: %s", err, kt.Rid)
		return err
	}

	err = logicFunc(subKit)
	if err != nil {
		return err
	}

	if err = session.CommitTransaction(sc); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, kt.Rid)
		return fmt.Errorf("commit transaction failed, err: %v", err)
	}

	return nil
}
