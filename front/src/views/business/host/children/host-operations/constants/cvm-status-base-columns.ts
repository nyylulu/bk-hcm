import { h } from 'vue';
import i18n from '@/language/i18n';
import { ModelPropertyColumn } from '@/model/typings';

import StatusAbnormal from '@/assets/image/Status-abnormal.png';
import StatusNormal from '@/assets/image/Status-normal.png';
import StatusUnknown from '@/assets/image/Status-unknown.png';
import {
  HOST_RUNNING_STATUS,
  HOST_SHUTDOWN_STATUS,
} from '@/views/resource/resource-manage/common/table/HostOperations';
import { CLOUD_HOST_STATUS } from '@/common/constant';

const { t } = i18n.global;

export default [
  { id: 'bk_asset_id', name: t('设备固资号'), type: 'string' },
  { id: 'private_ip_address', name: t('内网IP'), type: 'string' },
  { id: 'public_ip_address', name: t('外网IP'), type: 'string' },
  { id: 'bk_host_name', name: t('主机名称'), type: 'string' },
  { id: 'region', name: t('地域'), type: 'region' },
  { id: 'zone', name: t('可用区'), type: 'string' },
  {
    id: 'status',
    name: t('主机状态'),
    type: 'string',
    render: ({ cell }: any) => {
      // eslint-disable-next-line no-nested-ternary
      const src = HOST_SHUTDOWN_STATUS.includes(cell)
        ? cell.toLowerCase() === 'stopped'
          ? StatusUnknown
          : StatusAbnormal
        : HOST_RUNNING_STATUS.includes(cell)
        ? StatusNormal
        : StatusUnknown;

      return h('div', { class: 'flex-row align-items-center' }, [
        h('img', { class: 'mr6', src, width: 14, height: 14 }),
        h('span', null, CLOUD_HOST_STATUS[cell] || cell || t('未获取')),
      ]);
    },
  },
  { id: 'device_type', name: t('机型'), type: 'string' },
] as ModelPropertyColumn[];
