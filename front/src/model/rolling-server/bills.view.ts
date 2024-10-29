import { QueryRuleOPEnum } from '@/typings';
import { ModelProperty } from '@/model/typings';

export default [
  {
    id: 'date',
    name: '核算日期',
    type: 'datetime',
    meta: {
      search: {
        filterRules(value: Date[]) {
          const start = new Date(value[0]);
          const end = new Date(value[1]);
          return {
            op: QueryRuleOPEnum.AND,
            rules: [
              {
                field: 'year',
                op: QueryRuleOPEnum.GTE,
                value: start.getFullYear(),
              },
              {
                field: 'month',
                op: QueryRuleOPEnum.GTE,
                value: start.getMonth() + 1,
              },
              {
                field: 'day',
                op: QueryRuleOPEnum.GTE,
                value: start.getDate(),
              },
              {
                field: 'year',
                op: QueryRuleOPEnum.LTE,
                value: end.getFullYear(),
              },
              {
                field: 'month',
                op: QueryRuleOPEnum.LTE,
                value: end.getMonth() + 1,
              },
              {
                field: 'day',
                op: QueryRuleOPEnum.LTE,
                value: end.getDate(),
              },
            ],
          };
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
