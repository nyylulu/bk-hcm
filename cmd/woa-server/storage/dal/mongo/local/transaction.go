/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package local

import (
	"context"
	"fmt"

	"hcm/cmd/woa-server/common"
	"hcm/cmd/woa-server/common/metadata"
	"hcm/cmd/woa-server/storage/dal/redis"
	"hcm/pkg/logs"
)

// CommitTransaction 提交事务
func (c *Mongo) CommitTransaction(ctx context.Context, cap *metadata.TxnCapable) error {
	rid := ctx.Value(common.ContextRequestIDField)

	// check if txn number exists, if not, then no db operation with transaction is executed, committing will return an
	// error: "(NoSuchTransaction) Given transaction number 1 does not match any in-progress transactions. The active
	// transaction number is -1.". So we will return directly in this situation.
	txnNumber, err := c.tm.GetTxnNumber(cap.SessionID)
	if err != nil {
		if redis.IsNilErr(err) {
			logs.Infof("commit transaction: %s but no transaction need to commit, *skip*, rid: %s", cap.SessionID, rid)
			return nil
		}
		return fmt.Errorf("get txn number failed, err: %v", err)
	}
	if txnNumber == 0 {
		logs.Infof("commit transaction: %s but no transaction to commit, **skip**, rid: %s", cap.SessionID, rid)
		return nil
	}

	reloadSession, err := c.tm.PrepareTransaction(cap, c.dbc)
	if err != nil {
		logs.Errorf("commit transaction, but prepare transaction failed, err: %v, rid: %v", err, rid)
		return err
	}
	// reset the transaction state, so that we can commit the transaction after start the
	// transaction immediately.
	if err := CmdbPrepareCommitOrAbort(reloadSession); err != nil {
		logs.Errorf("reset the commit transaction state failed, err: %v, rid: %v", err, rid)
		return err
	}

	// we commit the transaction with the session id
	err = reloadSession.CommitTransaction(ctx)
	if err != nil {
		return fmt.Errorf("commit transaction: %s failed, err: %v, rid: %v", cap.SessionID, err, rid)
	}

	err = c.tm.RemoveSessionKey(cap.SessionID)
	if err != nil {
		// this key has ttl, it's ok if we not delete it, cause this key has a ttl.
		logs.Errorf("commit transaction, but delete txn session: %s key failed, err: %v, rid: %v", cap.SessionID, err,
			rid)
		// do not return.
	}

	return nil
}

// AbortTransaction 取消事务
func (c *Mongo) AbortTransaction(ctx context.Context, cap *metadata.TxnCapable) (bool, error) {
	rid := ctx.Value(common.ContextRequestIDField)
	reloadSession, err := c.tm.PrepareTransaction(cap, c.dbc)
	if err != nil {
		logs.Errorf("abort transaction, but prepare transaction failed, err: %v, rid: %v", err, rid)
		return false, err
	}
	// reset the transaction state, so that we can abort the transaction after start the
	// transaction immediately.
	if err := CmdbPrepareCommitOrAbort(reloadSession); err != nil {
		logs.Errorf("reset abort transaction state failed, err: %v, rid: %v", err, rid)
		return false, err
	}

	// we abort the transaction with the session id
	err = reloadSession.AbortTransaction(ctx)
	if err != nil {
		return false, fmt.Errorf("abort transaction: %s failed, err: %v, rid: %v", cap.SessionID, err, rid)
	}

	err = c.tm.RemoveSessionKey(cap.SessionID)
	if err != nil {
		// this key has ttl, it's ok if we not delete it, cause this key has a ttl.
		logs.Errorf("abort transaction, but delete txn session: %s key failed, err: %v, rid: %v", cap.SessionID, err,
			rid)
		// do not return.
	}

	errorType := c.tm.GetTxnError(sessionKey(cap.SessionID))
	switch errorType {
	// retry when the transaction error type is write conflict, which means the transaction conflicts with another one
	case WriteConflictType:
		return true, nil
	}

	return false, nil
}
