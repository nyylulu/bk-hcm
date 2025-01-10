/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package plan

import (
	"encoding/csv"
	"errors"
	"io"
	"strconv"

	ptypes "hcm/cmd/woa-server/types/plan"
	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// ListDemandClass lists demand class.
func (s *service) ListDemandClass(_ *rest.Contexts) (interface{}, error) {
	return &core.ListResultT[enumor.DemandClass]{Details: enumor.GetDemandClassMembers()}, nil
}

// ListResMode lists resource mode.
func (s *service) ListResMode(_ *rest.Contexts) (interface{}, error) {
	return &core.ListResultT[enumor.ResMode]{Details: enumor.GetResModeMembers()}, nil
}

// ListDemandSource lists demand source.
func (s *service) ListDemandSource(_ *rest.Contexts) (interface{}, error) {
	return &core.ListResultT[enumor.DemandSource]{Details: enumor.GetDemandSourceMembers()}, nil
}

// ListRPTicketStatus lists resource plan ticket status.
func (s *service) ListRPTicketStatus(_ *rest.Contexts) (interface{}, error) {
	// get resource plan ticket status members.
	statuses := enumor.GetRPTicketStatusMembers()
	// convert to ptypes.RPTicketStatusItem slice.
	details := make([]ptypes.RPTicketStatusItem, 0, len(statuses))
	for _, status := range statuses {
		details = append(details, ptypes.RPTicketStatusItem{
			Status:     status,
			StatusName: status.Name(),
		})
	}
	return &core.ListResultT[ptypes.RPTicketStatusItem]{Details: details}, nil
}

// GetDemandAvailableTime gets resource plan demand available time according to expect time.
// docs: docs/api-docs/web-server/docs/scr/resource-plan/get_demand_available_time.md
func (s *service) GetDemandAvailableTime(cts *rest.Contexts) (interface{}, error) {
	req := new(ptypes.DemandAvailTimeReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("failed to get demand available time, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	date, err := req.Validate()
	if err != nil {
		logs.Errorf("failed to validate get demand available time parameter, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return s.planController.GetDemandAvailableTime(cts.Kit, date)
}

// ImportDemandWeek imports demand week from csv file.
func (s *service) ImportDemandWeek(cts *rest.Contexts) (interface{}, error) {
	file, _, err := cts.Request.Request.FormFile("file")
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	defer file.Close()

	createReq, err := parseDemandWeekFromCSV(cts.Kit, file)
	if err != nil {
		logs.Errorf("failed to parse demand week from csv, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return s.planController.CreateDemandWeek(cts.Kit, createReq)
}

func parseDemandWeekFromCSV(kt *kit.Kit, reader io.Reader) ([]rpproto.ResPlanWeekCreateReq, error) {
	csvR := csv.NewReader(reader)
	headers, err := csvR.Read()
	if err != nil {
		logs.Errorf("failed to read csv file, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 确认表头
	yearIdx, monthIdx, weekIdx, startIdx, endIdx, isHolidayIdx, err := parseDemandWeekCSVHeaders(kt, headers)
	if err != nil {
		logs.Errorf("failed to parse csv file headers, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	var records []rpproto.ResPlanWeekCreateReq
	line := 0
	for {
		line += 1
		record, err := csvR.Read()
		if err != nil {
			break
		}

		year, err := strconv.Atoi(record[yearIdx])
		if err != nil {
			logs.Errorf("failed to parse year from csv, err: %v, line: %d, year: %s, rid: %s", err, line,
				record[yearIdx], kt.Rid)
			return nil, err
		}

		month, err := strconv.Atoi(record[monthIdx])
		if err != nil {
			logs.Errorf("failed to parse month from csv, err: %v, line: %d, month: %s, rid: %s", err, line,
				record[monthIdx], kt.Rid)
			return nil, err
		}

		week, err := strconv.Atoi(record[weekIdx])
		if err != nil {
			logs.Errorf("failed to parse week from csv, err: %v, line: %d, week: %s, rid: %s", err, line,
				record[weekIdx], kt.Rid)
			return nil, err
		}

		start, err := strconv.Atoi(record[startIdx])
		if err != nil {
			logs.Errorf("failed to parse start from csv, err: %v, line: %d, start: %s, rid: %s", err, line,
				record[startIdx], kt.Rid)
			return nil, err
		}

		end, err := strconv.Atoi(record[endIdx])
		if err != nil {
			logs.Errorf("failed to parse end from csv, err: %v, line: %d, end: %s, rid: %s", err, line,
				record[endIdx], kt.Rid)
			return nil, err
		}

		isHolidayInt, err := strconv.Atoi(record[isHolidayIdx])
		if err != nil {
			logs.Errorf("failed to parse is_holiday from csv, err: %v, line: %d, is_holiday: %s, rid: %s", err,
				line, record[isHolidayIdx], kt.Rid)
			return nil, err
		}
		isHoliday := enumor.ResPlanWeekHolidayStatus(isHolidayInt)

		records = append(records, rpproto.ResPlanWeekCreateReq{
			Year:      year,
			Month:     month,
			YearWeek:  week,
			Start:     start,
			End:       end,
			IsHoliday: &isHoliday,
		})
	}

	return records, nil
}

func parseDemandWeekCSVHeaders(kt *kit.Kit, headers []string) (yearIdx int, monthIdx int, weekIdx int,
	startIdx int, endIdx int, isHolidayIdx int, err error) {

	yearIdx = -1
	monthIdx = -1
	weekIdx = -1
	startIdx = -1
	endIdx = -1
	isHolidayIdx = -1
	for i, header := range headers {
		switch header {
		case "year":
			yearIdx = i
		case "month":
			monthIdx = i
		case "week":
			weekIdx = i
		case "start":
			startIdx = i
		case "end":
			endIdx = i
		case "is_holiday":
			isHolidayIdx = i
		default:
			continue
		}
	}

	if yearIdx == -1 || monthIdx == -1 || weekIdx == -1 || startIdx == -1 || endIdx == -1 || isHolidayIdx == -1 {
		logs.Errorf("failed to find csv headers, need headers: year/month/week/start/end/is_holiday, rid: %s",
			kt.Rid)
		return yearIdx, monthIdx, weekIdx, startIdx, endIdx, isHolidayIdx, errors.New("failed to find csv headers")
	}

	return yearIdx, monthIdx, weekIdx, startIdx, endIdx, isHolidayIdx, nil
}
