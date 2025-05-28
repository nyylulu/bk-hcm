import i18n from '@/language/i18n';
import { ModelPropertyColumn } from '@/model/typings';

import { CLOUD_HOST_STATUS } from '@/common/constant';
import { h, withDirectives } from 'vue';
import { bkTooltips, Tag } from 'bkui-vue';

const { t } = i18n.global;

export default [
  { id: 'bk_asset_id', name: t('设备固资号'), type: 'string' },
  { id: 'private_ip_address', name: t('内网IP'), type: 'string' },
  { id: 'public_ip_address', name: t('外网IP'), type: 'string' },
  { id: 'name', name: t('实例名称'), type: 'string' },
  { id: 'bk_host_name', name: t('OS主机名称'), type: 'string' },
  { id: 'bk_os_name', name: t('操作系统'), type: 'string' },
  { id: 'region', name: t('地域'), type: 'region' },
  { id: 'zone', name: t('可用区'), type: 'string' },
  {
    id: 'topo_module',
    name: t('所属模块'),
    type: 'string',
    render: ({ row }: any) => {
      const isIdle = row.operate_status !== 2;
      const theme = !isIdle ? 'danger' : '';
      if (isIdle) return h(Tag, t('空闲机'));
      return withDirectives(h(Tag, { theme }, t('业务模块')), [[bkTooltips, { content: row.topo_module }]]);
    },
  },
  {
    id: 'status',
    name: '主机状态',
    type: 'enum',
    option: CLOUD_HOST_STATUS,
    meta: {
      display: {
        appearance: 'cvm-status',
      },
    },
  },
  { id: 'device_type', name: t('机型'), type: 'string' },
] as ModelPropertyColumn[];
