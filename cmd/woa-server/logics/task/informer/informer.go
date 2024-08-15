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

// Package informer define informer interface
package informer

import (
	"hcm/cmd/woa-server/logics/task/informer/apply"
	"hcm/cmd/woa-server/logics/task/informer/generate"
	"hcm/cmd/woa-server/logics/task/informer/notice"
	"hcm/cmd/woa-server/logics/task/informer/recycle"
	"hcm/cmd/woa-server/logics/task/informer/ticket"
	"hcm/cmd/woa-server/storage/dal"
	"hcm/cmd/woa-server/storage/stream"
)

// Interface informer interface
type Interface interface {
	// Ticket apply ticket informer interface
	Ticket() ticket.Interface
	// Apply apply informer interface
	Apply() apply.Interface
	// Recycle recycle informer interface
	Recycle() recycle.Interface
	// Event event informer interface
	Event() notice.Interface
	// Generate generate informer interface
	Generate() generate.Interface
}

type informer struct {
	ticket   ticket.Interface
	apply    apply.Interface
	recycle  recycle.Interface
	event    notice.Interface
	generate generate.Interface
}

// New create a informer
func New(loopWatch stream.LoopInterface, watchDB dal.DB) (*informer, error) {
	ticketIf, err := ticket.New(loopWatch, watchDB)
	if err != nil {
		return nil, err
	}

	applyIf, err := apply.New(loopWatch, watchDB)
	if err != nil {
		return nil, err
	}

	eventIf, err := notice.New(loopWatch, watchDB)
	if err != nil {
		return nil, err
	}

	generateIf, err := generate.New(loopWatch, watchDB)
	if err != nil {
		return nil, err
	}

	informer := &informer{
		ticket:   ticketIf,
		apply:    applyIf,
		event:    eventIf,
		generate: generateIf,
	}

	return informer, nil
}

// Ticket apply ticket informer interface
func (i *informer) Ticket() ticket.Interface {
	return i.ticket
}

// Apply apply informer interface
func (i *informer) Apply() apply.Interface {
	return i.apply
}

// Recycle recycle informer interface
func (i *informer) Recycle() recycle.Interface {
	return i.recycle
}

// Event event informer interface
func (i *informer) Event() notice.Interface {
	return i.event
}

// Generate generate informer interface
func (i *informer) Generate() generate.Interface {
	return i.generate
}
