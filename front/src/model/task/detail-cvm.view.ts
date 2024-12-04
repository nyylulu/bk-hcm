/* eslint-disable @typescript-eslint/no-unused-vars */
import { Model, Column } from '@/decorator';
import { VendorEnum, VendorMap } from '@/common/constant';
import { TaskDetailStatus } from '@/views/task/typings';
import { TASK_DETAIL_STATUS_NAME } from '@/views/task/constants';
import { getPrivateIPs } from '@/utils/common';

@Model('task/detail-cvm.view')
export class DetailCvmView {
  @Column('string', { name: '任务ID' })
  task_management_id: string;

  @Column('datetime', { name: '开始时间', sort: true })
  created_at: string;

  @Column('datetime', { name: '结束时间', sort: true })
  updated_at: string;

  @Column('enum', {
    name: '任务状态',
    option: TASK_DETAIL_STATUS_NAME,
    sort: true,
    meta: {
      display: { appearance: 'status' },
    },
  })
  state: TaskDetailStatus;

  @Column('string', { name: '失败原因' })
  reason: string;

  @Column('array', {
    name: '内网IP',
    render: ({ data }) => getPrivateIPs(data.param),
  })
  'param.private_ipv4_addresses': string[];

  @Column('string', { name: '固资号' })
  'param.extension.bk_asset_id': string;

  @Column('enum', { name: '云厂商', option: VendorMap })
  'param.vendor': VendorEnum;

  @Column('account', { name: '云账号' })
  'param.account_id': string;

  @Column('region', { name: '地域' })
  'param.region': string;

  @Column('string', { name: '园区' })
  'param.zone': string;

  @Column('string', { name: '所属VPC' })
  'param.cloud_vpc_ids': string;

  @Column('string', { name: '所属子网' })
  'param.cloud_subnet_id': string;

  @Column('string', { name: '原镜像名称' })
  'param.image_name_old': string;

  @Column('string', { name: '新镜像名称' })
  'param.image_name': string;
}
