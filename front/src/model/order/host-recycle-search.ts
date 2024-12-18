/* eslint-disable @typescript-eslint/no-unused-vars */
import { Model, Column } from '@/decorator';
import dayjs from 'dayjs';
import { toArray } from '@/common/util';

@Model('order/host-recycle-search')
export class HostRecycleSearch {
  @Column('req-type', { name: '需求类型', index: 1 })
  require_type: number;

  @Column('string', {
    name: '单号',
    meta: {
      display: { appearance: 'tag-input' },
      search: {
        format: (value) => {
          const values = toArray(value).map((val: string) => Number(val));
          return values.filter((val: number) => !isNaN(val));
        },
      },
    },
    index: 1,
  })
  order_id: number;

  @Column('string', {
    name: '子单号',
    meta: {
      display: { appearance: 'tag-input' },
      search: {
        format(value) {
          return toArray(value);
        },
      },
    },
    index: 1,
  })
  suborder_id: string;

  @Column('enum', {
    name: '资源类型',
    option: {
      QCLOUDCVM: '腾讯云虚拟机',
      IDCPM: 'IDC物理机',
      OTHERS: '其他',
    },
    index: 1,
  })
  resource_type: string;

  @Column('enum', {
    name: '回收类型',
    option: {
      IMMEDIATE: '立即销毁',
      DELAY: '延迟销毁',
    },
    index: 1,
  })
  return_plan: string;

  @Column('enum', {
    name: '状态',
    index: 1,
  })
  stage: string;

  @Column('user', { name: '回收人', index: 1 })
  bk_username: string;

  @Column('datetime', {
    name: '回收时间',
    meta: {
      search: {
        converter(value) {
          const start = new Date(value[0]);
          const end = new Date(value[1]);
          return {
            start: dayjs(start).format('YYYY-MM-DD'),
            end: dayjs(end).format('YYYY-MM-DD'),
          };
        },
      },
    },
    index: 1,
  })
  create_at: [Date, Date];
}

@Model('order/host-recycle-search')
export class HostRecycleSearchNonBusiness extends HostRecycleSearch {
  @Column('business', {
    name: '业务',
    meta: {
      search: {
        format(value) {
          if (value === '0' || !value?.length) {
            return [0];
          }
          if (!Array.isArray(value)) {
            return [Number(value)];
          }
          return value.map((val: string) => Number(val));
        },
      },
    },
    index: 0,
  })
  bk_biz_id: number;
}
