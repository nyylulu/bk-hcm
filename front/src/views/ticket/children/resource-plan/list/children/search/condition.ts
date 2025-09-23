import dayjs from 'dayjs';
import { useResourcePlanStore } from '@/store/resource-plan';
import { useResourcePlanTicketStore } from '@/store/ticket/resource-plan';
import { ModelPropertySearch } from '@/model/typings';
import { toArray } from '@/common/util';

const resourcePlanStore = useResourcePlanStore();
const resourcePlanTicketStore = useResourcePlanTicketStore();

export const properties: ModelPropertySearch[] = [
  {
    id: 'bk_biz_ids',
    name: '业务',
    type: 'business',
    props: {
      showAll: true,
    },
  },
  {
    id: 'op_product_ids',
    name: '运营产品',
    type: 'list',
    list: async () => await resourcePlanStore.getOpProductList(),
    format: (value) => {
      return toArray(value).map((val: string) => Number(val));
    },
    props: {
      idKey: 'op_product_id',
      displayKey: 'op_product_name',
    },
  },
  {
    id: 'plan_product_ids',
    name: '规划产品',
    type: 'list',
    list: async () => await resourcePlanStore.getPlanProductList(),
    props: {
      idKey: 'plan_product_id',
      displayKey: 'plan_product_name',
    },
  },
  {
    id: 'ticket_types',
    name: '类型',
    type: 'list',
    list: async () => await resourcePlanTicketStore.getTicketTypeList(),
    props: {
      idKey: 'ticket_type',
      displayKey: 'ticket_type_name',
    },
  },
  {
    id: 'ticket_ids',
    name: '预测单号',
    type: 'string',
    props: {
      multiple: false,
      placeholder: '请输入预测单号，多个预测单号可使用分号分隔',
      separator: ';',
    },
    format: (value: string | string[]) => {
      if (Array.isArray(value)) {
        return value;
      }
      return value.split(';').filter((v) => v);
    },
  },
  {
    id: 'statuses',
    name: '单据状态',
    type: 'list',
    list: async () => await resourcePlanTicketStore.getTicketStatusList(),
    props: {
      idKey: 'status',
      displayKey: 'status_name',
    },
  },
  {
    id: 'applicants',
    name: '提单人',
    type: 'user',
  },
  {
    id: 'submit_time_range',
    name: '提单时间',
    type: 'datetime',
    props: {
      type: 'daterange',
      format: 'yyyy-MM-dd',
    },
    converter: (value) => {
      return {
        submit_time_range: {
          start: dayjs(value[0]).format('YYYY-MM-DD'),
          end: dayjs(value[1]).format('YYYY-MM-DD'),
        },
      };
    },
  },
];
