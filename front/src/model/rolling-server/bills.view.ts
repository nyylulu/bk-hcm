import dayjs from 'dayjs';
import { QueryRuleOPEnum } from '@/typings';
import { ModelProperty } from '@/model/typings';

export default [
  {
    id: 'roll_date',
    name: '核算日期',
    type: 'datetime',
    meta: {
      search: {
        format(value: Date[] | Date | string) {
          // TODO: 数组时不转换，当qs获取时传入的是数组
          if (Array.isArray(value)) {
            return value;
          }
          const date = dayjs(value);
          return Number(date.format('YYYYMMDD'));
        },
      },
    },
  },
  {
    id: 'bk_biz_id',
    name: '业务',
    type: 'business',
    meta: {
      search: {
        op: QueryRuleOPEnum.EQ,
      },
    },
  },
  {
    id: 'not_returned_core',
    name: '当天未退还',
    type: 'string',
    unit: '核',
  },
] as ModelProperty[];
