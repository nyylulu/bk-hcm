/* eslint-disable no-nested-ternary */
// table 字段相关信息
import i18n from '@/language/i18n';
import { CloudType, SecurityRuleEnum, HuaweiSecurityRuleEnum, AzureSecurityRuleEnum } from '@/typings';
import { useAccountStore, useLoadBalancerStore } from '@/store';
import { Button } from 'bkui-vue';
import { type Settings } from 'bkui-vue/lib/table/props';
import { h, ref } from 'vue';
import type { Ref } from 'vue';
import { RouteLocationRaw, useRoute, useRouter } from 'vue-router';
import { CLB_BINDING_STATUS, CLOUD_HOST_STATUS, VendorEnum, VendorMap } from '@/common/constant';
import { useRegionsStore } from '@/store/useRegionsStore';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { useCloudAreaStore } from '@/store/useCloudAreaStore';
import StatusAbnormal from '@/assets/image/Status-abnormal.png';
import StatusNormal from '@/assets/image/Status-normal.png';
import StatusUnknown from '@/assets/image/Status-unknown.png';
import StatusSuccess from '@/assets/image/success-account.png';
import StatusLoading from '@/assets/image/status_loading.png';
import StatusFailure from '@/assets/image/failed-account.png';

import { HOST_RUNNING_STATUS, HOST_SHUTDOWN_STATUS } from '../common/table/HostOperations';
import './use-columns.scss';
import { defaults } from 'lodash';
import { timeFormatter } from '@/common/util';
import { capacityLevel } from '@/utils/scr';
import { IP_VERSION_MAP, LBRouteName, LB_NETWORK_TYPE_MAP, SCHEDULER_MAP } from '@/constants/clb';
import { getInstVip, getResourceTypeName, getReturnPlanName } from '@/utils';
import {
  getRecycleTaskStatusLabel,
  getBusinessNameById,
  dateTimeTransform,
  getPrecheckStatusLabel,
} from '@/views/ziyanScr/host-recycle/field-dictionary';
import { getRegionCn, getZoneCn } from '@/views/ziyanScr/cvm-web/transform';
import { Spinner, Share, Copy, DataShape } from 'bkui-vue/lib/icon';
import dayjs from 'dayjs';
import WName from '@/components/w-name';

interface LinkFieldOptions {
  type: string; // 资源类型
  label?: string; // 显示文本
  field?: string; // 字段
  idFiled?: string; // id字段
  onlyShowOnList?: boolean; // 只在列表中显示
  onLinkInBusiness?: boolean; // 只在业务下可链接
  render?: (data: any) => any; // 自定义渲染内容
  sort?: boolean; // 是否支持排序
}

export default (type: string, isSimpleShow = false, vendor?: string) => {
  const router = useRouter();
  const route = useRoute();
  const accountStore = useAccountStore();
  const loadBalancerStore = useLoadBalancerStore();
  const { t } = i18n.global;
  const { getRegionName } = useRegionsStore();
  const { whereAmI } = useWhereAmI();
  const businessMapStore = useBusinessMapStore();
  const cloudAreaStore = useCloudAreaStore();

  const getLinkField = (options: LinkFieldOptions) => {
    // 设置options的默认值
    defaults(options, {
      label: 'ID',
      field: 'id',
      idFiled: 'id',
      onlyShowOnList: true,
      onLinkInBusiness: false,
      render: undefined,
      sort: true,
    });

    const { type, label, field, idFiled, onlyShowOnList, onLinkInBusiness, render, sort } = options;

    return {
      label,
      field,
      sort,
      width: label === 'ID' ? '120' : 'auto',
      onlyShowOnList,
      isDefaultShow: true,
      render({ data }: { cell: string; data: any }) {
        if (data[idFiled] < 0 || !data[idFiled]) return '--';
        // 如果设置了onLinkInBusiness=true, 则只在业务下可以链接至指定路由
        if (onLinkInBusiness && whereAmI.value !== Senarios.business) return data[field] || '--';
        return (
          <Button
            text
            theme='primary'
            onClick={() => {
              const routeInfo: any = {
                query: {
                  ...route.query,
                  id: data[idFiled],
                  type: data.vendor,
                },
              };
              // 业务下
              if (route.path.includes('business')) {
                routeInfo.query.bizs = accountStore.bizs;
                Object.assign(routeInfo, {
                  name: `${type}BusinessDetail`,
                });
              } else {
                Object.assign(routeInfo, {
                  name: 'resourceDetail',
                  params: {
                    type,
                  },
                });
              }
              router.push(routeInfo);
            }}>
            {render ? render(data) : data[field] || '--'}
          </Button>
        );
      },
    };
  };

  /**
   * 自定义 render field 的 push 导航
   * @param to 目标路由信息
   */
  const renderFieldPushState = (to: RouteLocationRaw, cb?: (...args: any) => any) => {
    return (e: Event) => {
      // 阻止事件冒泡
      e.stopPropagation();
      // 导航
      router.push(to);
      // 执行回调
      typeof cb === 'function' && cb();
    };
  };

  const vpcColumns = [
    {
      type: 'selection',
      width: 32,
      minWidth: 32,
      onlyShowOnList: true,
      align: 'right',
    },
    getLinkField({ type: 'vpc', label: 'VPC ID', field: 'cloud_id' }),
    // {
    //   label: '资源 ID',
    //   field: 'cloud_id',
    //   sort: true,
    //   isDefaultShow: true,
    //   render({ cell }: { cell: string }) {
    //     return h('span', [cell || '--']);
    //   },
    // },
    {
      label: 'VPC 名称',
      field: 'name',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [cell || '--']);
      },
    },
    {
      label: '云厂商',
      field: 'vendor',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [CloudType[cell] || '--']);
      },
    },
    {
      label: '地域',
      field: 'region',
      sort: true,
      isDefaultShow: true,
      render: ({ cell, row }: { cell: string; row: { vendor: VendorEnum } }) => getRegionName(row.vendor, cell) || '--',
    },
    {
      label: '是否分配',
      field: 'bk_biz_id',
      sort: true,
      isOnlyShowInResource: true,
      isDefaultShow: true,
      render: ({ data, cell }: { data: { bk_biz_id: number }; cell: number }) => (
        <bk-tag
          v-bk-tooltips={{
            content: businessMapStore.businessMap.get(cell),
            disabled: !cell || cell === -1,
            theme: 'light',
          }}
          theme={data.bk_biz_id === -1 ? false : 'success'}>
          {data.bk_biz_id === -1 ? '未分配' : '已分配'}
        </bk-tag>
      ),
    },
    {
      label: '所属业务',
      field: 'bk_biz_id2',
      isOnlyShowInResource: true,
      render: ({ data }: any) => businessMapStore.businessMap.get(data.bk_biz_id) || '未分配',
    },
    {
      label: '管控区域',
      field: 'bk_cloud_id',
      isDefaultShow: true,
      render({ cell }: { cell: number }) {
        if (cell !== -1) {
          return `[${cell}] ${cloudAreaStore.getNameFromCloudAreaMap(cell)}`;
        }
        return '--';
      },
    },
    {
      label: '创建时间',
      field: 'created_at',
      sort: true,
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
    },
    {
      label: '更新时间',
      field: 'updated_at',
      sort: true,
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
    },
  ];

  const subnetColumns = [
    {
      type: 'selection',
      width: 32,
      minWidth: 32,
      onlyShowOnList: true,
      align: 'right',
    },
    getLinkField({ type: 'subnet', label: '子网 ID', field: 'cloud_id', idFiled: 'id', onlyShowOnList: false }),
    // {
    //   label: '资源 ID',
    //   field: 'cloud_id',
    //   sort: true,
    //   render({ cell }: { cell: string }) {
    //     const index =          cell.lastIndexOf('/') <= 0 ? 0 : cell.lastIndexOf('/') + 1;
    //     const value = cell.slice(index);
    //     return h('span', [value || '--']);
    //   },
    // },
    {
      label: '子网名称',
      field: 'name',
      sort: true,
      render({ cell }: { cell: string }) {
        return h('span', [cell || '--']);
      },
    },
    {
      label: '云厂商',
      field: 'vendor',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [CloudType[cell] || '--']);
      },
    },
    {
      label: '地域',
      field: 'region',
      sort: true,
      isDefaultShow: true,
      render: ({ cell, row }: { cell: string; row: { vendor: VendorEnum } }) => getRegionName(row.vendor, cell),
    },
    {
      label: '可用区',
      field: 'zone',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [cell || '--']);
      },
    },
    getLinkField({ type: 'vpc', label: '所属 VPC', field: 'cloud_vpc_id', idFiled: 'vpc_id', onlyShowOnList: false }),
    {
      label: 'IPv4 CIDR',
      field: 'ipv4_cidr',
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [cell || '--']);
      },
    },
    // getLinkField(
    //   'route',
    //   '关联路由表',
    //   'route_table_id',
    //   'route_table_id',
    //   false,
    // ),
    // {
    //   label: '可用IPv4地址数',
    //   field: 'count_of_ipv4_cidr',
    //   isDefaultShow: true,
    //   render({ data }: any) {
    //     return data.ipv4_cidr.length;
    //   },
    // },
    {
      label: '是否分配',
      field: 'bk_biz_id',
      sort: true,
      isOnlyShowInResource: true,
      isDefaultShow: true,
      render: ({ data, cell }: { data: { bk_biz_id: number }; cell: number }) => (
        <bk-tag
          v-bk-tooltips={{
            content: businessMapStore.businessMap.get(cell),
            disabled: !cell || cell === -1,
            theme: 'light',
          }}
          theme={data.bk_biz_id === -1 ? false : 'success'}>
          {data.bk_biz_id === -1 ? '未分配' : '已分配'}
        </bk-tag>
      ),
    },
    {
      label: '所属业务',
      field: 'bk_biz_id2',
      isOnlyShowInResource: true,
      render: ({ data }: any) => businessMapStore.businessMap.get(data.bk_biz_id) || '未分配',
    },
    {
      label: '创建时间',
      field: 'created_at',
      sort: true,
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
    },
    {
      label: '更新时间',
      field: 'updated_at',
      sort: true,
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
    },
  ];

  const groupColumns = [
    {
      type: 'selection',
      width: 32,
      minWidth: 32,
      onlyShowOnList: true,
      align: 'right',
    },
    getLinkField({ type: 'subnet' }),
    {
      label: '资源 ID',
      field: 'account_id',
      sort: true,
    },
    {
      label: '名称',
      field: 'name',
      sort: true,
    },
    {
      label: t('云厂商'),
      render({ data }: any) {
        return h('span', {}, [CloudType[data.vendor]]);
      },
    },
    {
      label: '地域',
      field: 'region',
      render: ({ cell, row }: { cell: string; row: { vendor: VendorEnum } }) => getRegionName(row.vendor, cell),
    },
    {
      label: '描述',
      field: 'memo',
    },
  ];

  const gcpColumns = [
    {
      type: 'selection',
      width: 32,
      minWidth: 32,
      onlyShowOnList: true,
      align: 'right',
    },
    getLinkField({ type: 'subnet' }),
    {
      label: '资源 ID',
      field: 'account_id',
      sort: true,
    },
    {
      label: '名称',
      field: 'name',
      sort: true,
    },
    // {
    //   label: '业务',
    //   render({ cell }: any) {
    //     return h(
    //       'span',
    //       {},
    //       [
    //         cell,
    //       ],
    //     );
    //   },
    // },
    // {
    //   label: '业务拓扑',
    //   field: 'zone',
    // },
    {
      label: 'VPC',
      field: 'vpc_id',
    },
    {
      label: '描述',
      field: 'memo',
    },
  ];

  const driveColumns: any[] = [
    {
      type: 'selection',
      width: 32,
      minWidth: 32,
      onlyShowOnList: true,
      align: 'right',
    },
    getLinkField({ type: 'drive', label: '云硬盘ID', field: 'cloud_id' }),
    // {
    //   label: '资源 ID',
    //   field: 'cloud_id',
    //   sort: true,
    // },
    {
      label: '云硬盘名称',
      field: 'name',
      sort: true,
      isDefaultShow: true,
      render: ({ cell }: any) => cell || '--',
    },
    {
      label: '云厂商',
      field: 'vendor',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [CloudType[cell] || '--']);
      },
    },
    {
      label: '地域',
      field: 'region',
      sort: true,
      isDefaultShow: true,
      render: ({ cell, row }: { cell: string; row: { vendor: VendorEnum } }) => getRegionName(row.vendor, cell),
    },
    {
      label: '可用区',
      field: 'zone',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [cell || '--']);
      },
    },
    {
      label: '云硬盘状态',
      field: 'status',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [cell || '--']);
      },
    },
    {
      label: '硬盘分类',
      field: 'is_system_disk',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: boolean }) {
        return h('span', [cell ? '系统盘' : '数据盘']);
      },
    },
    {
      label: '类型',
      field: 'disk_type',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [cell || '--']);
      },
    },
    {
      label: '容量(GB)',
      field: 'disk_size',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [cell || '--']);
      },
    },
    getLinkField({ type: 'host', label: '挂载的主机', field: 'instance_id', idFiled: 'instance_id' }),
    {
      label: '是否分配',
      field: 'bk_biz_id',
      sort: true,
      isOnlyShowInResource: true,
      isDefaultShow: true,
      render: ({ data, cell }: { data: { bk_biz_id: number }; cell: number }) => (
        <bk-tag
          v-bk-tooltips={{
            content: businessMapStore.businessMap.get(cell),
            disabled: !cell || cell === -1,
            theme: 'light',
          }}
          theme={data.bk_biz_id === -1 ? false : 'success'}>
          {data.bk_biz_id === -1 ? '未分配' : '已分配'}
        </bk-tag>
      ),
    },
    {
      label: '创建时间',
      field: 'created_at',
      sort: true,
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
    },
    {
      label: '更新时间',
      field: 'updated_at',
      sort: true,
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
    },
  ];

  const imageColumns = [
    getLinkField({ type: 'image', label: '镜像ID', field: 'cloud_id', idFiled: 'id' }),
    // {
    //   label: '资源 ID',
    //   field: 'cloud_id',
    //   sort: true,
    // },
    {
      label: '镜像名称',
      field: 'name',
      sort: true,
      isDefaultShow: true,
    },
    {
      label: '云厂商',
      field: 'vendor',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [CloudType[cell] || '--']);
      },
    },
    {
      label: '操作系统类型',
      field: 'platform',
      sort: true,
      isDefaultShow: true,
    },
    {
      label: '架构',
      field: 'architecture',
      sort: true,
      isDefaultShow: true,
    },
    {
      label: '状态',
      field: 'state',
      sort: true,
      isDefaultShow: true,
    },
    {
      label: '类型',
      field: 'type',
      sort: true,
      isDefaultShow: true,
    },
    {
      label: '创建时间',
      field: 'created_at',
      sort: true,
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
    },
    {
      label: '更新时间',
      field: 'updated_at',
      sort: true,
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
    },
  ];

  const networkInterfaceColumns = [
    getLinkField({ type: 'network-interface', label: '接口 ID', field: 'cloud_id', idFiled: 'id' }),
    {
      label: '接口名称',
      field: 'name',
      sort: true,
      isDefaultShow: true,
    },
    {
      label: '云厂商',
      field: 'vendor',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [CloudType[cell] || '--']);
      },
    },
    {
      label: '地域',
      field: 'region',
      sort: true,
      isDefaultShow: true,
      render: ({ cell, row }: { cell: string; row: { vendor: VendorEnum } }) => getRegionName(row.vendor, cell),
    },
    {
      label: '可用区',
      field: 'zone',
      render({ cell }: { cell: string }) {
        return h('span', [cell || '--']);
      },
    },
    {
      label: '所属VPC',
      field: 'cloud_vpc_id',
      sort: true,
      isDefaultShow: true,
      showOverflowTooltip: true,
      render({ cell }: { cell: string }) {
        return h('span', [cell || '--']);
      },
    },
    {
      label: '所属子网',
      showOverflowTooltip: true,
      field: 'cloud_subnet_id',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [cell || '--']);
      },
    },
    // {
    //   label: '关联的实例',
    //   field: 'instance_id',
    //   showOverflowTooltip: true,
    //   render({ cell }: { cell: string }) {
    //     return h('span', [cell || '--']);
    //   },
    // },
    {
      label: '内网IP',
      field: 'private_ipv4_or_ipv6',
      isDefaultShow: true,
      render({ data }: any) {
        return [h('span', {}, [data?.private_ipv4.join(',') || data?.private_ipv6.join(',') || '--'])];
      },
    },
    {
      label: '关联的公网IP地址',
      field: 'public_ip',
      // 目前公网IP地址不支持排序
      // sort: true,
      isDefaultShow: true,
      render({ data }: any) {
        return [h('span', {}, [data?.public_ipv4.join(',') || data?.public_ipv6.join(',') || '--'])];
      },
    },
    {
      label: '所属业务',
      field: 'bk_biz_id',
      sort: true,
      isOnlyShowInResource: true,
      render: ({ data }: any) => businessMapStore.businessMap.get(data.bk_biz_id) || '未分配',
    },
    {
      label: '创建时间',
      field: 'created_at',
      sort: true,
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
    },
    {
      label: '更新时间',
      field: 'updated_at',
      sort: true,
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
    },
  ];

  const routeColumns = [
    getLinkField({ type: 'route', label: '路由表ID', field: 'cloud_id', idFiled: 'id' }),
    // {
    //   label: '资源 ID',
    //   field: 'cloud_id',
    //   sort: true,
    // },
    {
      label: '路由表名称',
      field: 'name',
      sort: true,
      isDefaultShow: true,
      render: ({ cell }: any) => cell || '--',
    },
    {
      label: '云厂商',
      field: 'vendor',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [CloudType[cell] || '--']);
      },
    },
    {
      label: '地域',
      field: 'region',
      sort: true,
      isDefaultShow: true,
      render: ({ cell, row }: { cell: string; row: { vendor: VendorEnum } }) => getRegionName(row.vendor, cell),
    },
    getLinkField({ type: 'vpc', label: '所属网络(VPC)', field: 'vpc_id', idFiled: 'vpc_id' }),
    // {
    //   label: '关联子网',
    //   field: '',
    //   sort: true,
    // },
    {
      label: '所属业务',
      field: 'bk_biz_id',
      isOnlyShowInResource: true,
      sort: true,
      render: ({ data }: any) => businessMapStore.businessMap.get(data.bk_biz_id) || '未分配',
    },
    {
      label: '创建时间',
      field: 'created_at',
      sort: true,
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
    },
    {
      label: '更新时间',
      field: 'updated_at',
      sort: true,
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
    },
  ];

  const cvmsColumns = [
    {
      type: 'selection',
      width: 32,
      minWidth: 32,
      onlyShowOnList: true,
      align: 'right',
    },
    //   移除 ID 搜索条件
    // {
    //   label: 'ID',
    //   field: 'id',
    //   isDefaultShow: false,
    //   onlyShowOnList: true,
    // },
    {
      label: '主机ID',
      field: 'cloud_id',
      isDefaultShow: false,
      onlyShowOnList: true,
    },
    getLinkField({
      type: 'host',
      label: '内网IP',
      field: 'private_ipv4_addresses',
      idFiled: 'id',
      onlyShowOnList: false,
      render: (data) =>
        [...(data.private_ipv4_addresses || []), ...(data.private_ipv6_addresses || [])].join(',') || '--',
      sort: false,
    }),
    {
      label: '公网IP',
      field: 'vendor',
      isDefaultShow: true,
      onlyShowOnList: true,
      render: ({ data }: any) =>
        [...(data.public_ipv4_addresses || []), ...(data.public_ipv6_addresses || [])].join(',') || '--',
    },
    {
      label: '所属VPC',
      field: 'cloud_vpc_ids',
      isDefaultShow: true,
      onlyShowOnList: true,
      render: ({ data }: any) => data.cloud_vpc_ids?.join(',') || '--',
    },
    {
      label: '云厂商',
      field: 'vendor',
      sort: true,
      onlyShowOnList: true,
      isDefaultShow: true,
      render({ data }: any) {
        return h('span', {}, [CloudType[data.vendor]]);
      },
    },
    {
      label: '地域',
      onlyShowOnList: true,
      field: 'region',
      sort: true,
      isDefaultShow: true,
      render: ({ cell, row }: { cell: string; row: { vendor: VendorEnum } }) => getRegionName(row.vendor, cell),
    },
    {
      label: '主机名称',
      field: 'name',
      sort: true,
      isDefaultShow: true,
    },
    {
      label: '主机状态',
      field: 'status',
      sort: true,
      isDefaultShow: true,
      render({ data }: any) {
        // return h('span', {}, [CLOUD_HOST_STATUS[data.status] || data.status]);
        return (
          <div class={'cvm-status-container'}>
            {HOST_SHUTDOWN_STATUS.includes(data.status) ? (
              data.status.toLowerCase() === 'stopped' ? (
                <img src={StatusUnknown} class={'mr6'} width={14} height={14}></img>
              ) : (
                <img src={StatusAbnormal} class={'mr6'} width={14} height={14}></img>
              )
            ) : HOST_RUNNING_STATUS.includes(data.status) ? (
              <img src={StatusNormal} class={'mr6'} width={14} height={14}></img>
            ) : (
              <img src={StatusUnknown} class={'mr6'} width={14} height={14}></img>
            )}
            <span>{CLOUD_HOST_STATUS[data.status] || data.status}</span>
          </div>
        );
      },
    },
    {
      label: '是否分配',
      field: 'bk_biz_id',
      sort: true,
      isOnlyShowInResource: true,
      isDefaultShow: true,
      render: ({ data, cell }: { data: { bk_biz_id: number }; cell: number }) => (
        <bk-tag
          v-bk-tooltips={{
            content: businessMapStore.businessMap.get(cell),
            disabled: !cell || cell === -1,
            theme: 'light',
          }}
          theme={data.bk_biz_id === -1 ? false : 'success'}>
          {data.bk_biz_id === -1 ? '未分配' : '已分配'}
        </bk-tag>
      ),
    },

    {
      label: '所属业务',
      field: 'bk_biz_id2',
      isOnlyShowInResource: true,
      render: ({ data }: any) => businessMapStore.businessMap.get(data.bk_biz_id) || '未分配',
    },
    {
      label: '管控区域',
      field: 'bk_cloud_id',
      sort: true,
      render({ cell }: { cell: number }) {
        if (cell !== -1) {
          return `[${cell}] ${cloudAreaStore.getNameFromCloudAreaMap(cell)}`;
        }
        return '--';
      },
    },
    {
      label: '实例规格',
      field: 'machine_type',
      sort: true,
      isOnlyShowInResource: true,
    },
    {
      label: '操作系统',
      field: 'os_name',
      render({ data }: any) {
        return h('span', {}, [data.os_name || '--']);
      },
    },
    {
      label: '创建时间',
      field: 'created_at',
      sort: true,
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
    },
    {
      label: '更新时间',
      field: 'updated_at',
      sort: true,
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
    },
  ];

  const securityCommonColumns = [
    {
      label: t('来源'),
      field: 'resource',
      render({ data }: any) {
        return h('span', {}, [
          data.cloud_address_group_id ||
            data.cloud_address_id ||
            data.cloud_service_group_id ||
            data.cloud_service_id ||
            data.cloud_target_security_group_id ||
            data.ipv4_cidr ||
            data.ipv6_cidr ||
            data.cloud_remote_group_id ||
            data.remote_ip_prefix ||
            (data.source_address_prefix === '*' ? t('任何') : data.source_address_prefix) ||
            data.source_address_prefixes ||
            data.cloud_source_security_group_ids ||
            (data.destination_address_prefix === '*' ? t('任何') : data.destination_address_prefix) ||
            data.destination_address_prefixes ||
            data.cloud_destination_security_group_ids ||
            '--',
        ]);
      },
    },
    {
      label: '协议端口',
      render({ data }: any) {
        return h('span', {}, [
          // eslint-disable-next-line no-nested-ternary
          vendor === 'aws' && data.protocol === '-1' && data.to_port === -1
            ? t('全部')
            : vendor === 'huawei' && !data.protocol && !data.port
            ? t('全部')
            : vendor === 'azure' && data.protocol === '*' && data.destination_port_range === '*'
            ? t('全部')
            : `${data.protocol}:${data.port || data.to_port || data.destination_port_range || '--'}`,
        ]);
      },
    },
    {
      label: t('策略'),
      render({ data }: any) {
        return h('span', {}, [
          // eslint-disable-next-line no-nested-ternary
          vendor === 'huawei'
            ? HuaweiSecurityRuleEnum[data.action]
            : vendor === 'azure'
            ? AzureSecurityRuleEnum[data.access]
            : vendor === 'aws'
            ? t('允许')
            : SecurityRuleEnum[data.action] || '--',
        ]);
      },
    },
    {
      label: '备注',
      field: 'memo',
      render({ data }: any) {
        return h('span', {}, [data.memo || '--']);
      },
    },
    {
      label: t('修改时间'),
      field: 'updated_at',
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
    },
  ];

  const eipColumns = [
    {
      type: 'selection',
      width: 32,
      minWidth: 32,
      onlyShowOnList: true,
      align: 'right',
    },
    getLinkField({ type: 'eips', label: 'IP资源ID', field: 'cloud_id', idFiled: 'id' }),
    // {
    //   label: '资源 ID',
    //   field: 'cloud_id',
    //   sort: true,
    // },
    {
      label: 'IP名称',
      field: 'name',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [cell || '--']);
      },
    },
    {
      label: '云厂商',
      field: 'vendor',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [CloudType[cell] || '--']);
      },
    },
    {
      label: '地域',
      field: 'region',
      sort: true,
      isDefaultShow: true,
      render: ({ cell, row }: { cell: string; row: { vendor: VendorEnum } }) => getRegionName(row.vendor, cell),
    },
    {
      label: '公网 IP',
      field: 'public_ip',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [cell || '--']);
      },
    },
    // {
    //   label: '状态',
    //   field: 'status',
    //   render({ cell }: { cell: string }) {
    //     return h('span', [cell || '--']);
    //   },
    // },
    getLinkField({
      type: 'host',
      label: '绑定的资源实例',
      field: 'cvm_id',
      idFiled: 'cvm_id',
      onlyShowOnList: false,
      render: (data) => data.host,
      sort: false,
    }),
    {
      label: '绑定的资源类型',
      field: 'instance_type',
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [cell || '--']);
      },
    },
    {
      label: '是否分配',
      field: 'bk_biz_id',
      sort: true,
      isOnlyShowInResource: true,
      isDefaultShow: true,
      render: ({ data, cell }: { data: { bk_biz_id: number }; cell: number }) => (
        <bk-tag
          v-bk-tooltips={{
            content: businessMapStore.businessMap.get(cell),
            disabled: !cell || cell === -1,
            theme: 'light',
          }}
          theme={data.bk_biz_id === -1 ? false : 'success'}>
          {data.bk_biz_id === -1 ? '未分配' : '已分配'}
        </bk-tag>
      ),
    },
    {
      label: '所属业务',
      field: 'bk_biz_id2',
      isOnlyShowInResource: true,
      render: ({ data }: any) => businessMapStore.businessMap.get(data.bk_biz_id) || '未分配',
    },
    {
      label: '创建时间',
      field: 'created_at',
      sort: true,
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
    },
    {
      label: '更新时间',
      field: 'updated_at',
      sort: true,
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
    },
  ];

  const operationRecordColumns = [
    {
      label: '操作时间',
      field: 'created_at',
      isDefaultShow: true,
      sort: true,
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
    },
    {
      label: '资源类型',
      field: 'res_type',
    },
    {
      label: '资源名称',
      field: 'res_name',
      isDefaultShow: true,
    },
    // {
    //   label: '云资源ID',
    //   field: 'cloud_res_id',
    // },
    {
      label: '操作方式',
      field: 'action',
      isDefaultShow: true,
      filter: true,
    },
    {
      label: '操作来源',
      field: 'source',
      isDefaultShow: true,
      filter: true,
    },
    {
      label: '所属业务',
      field: 'bk_biz_id',
      isOnlyShowInResource: true,
      render: ({ cell }: { cell: number }) => businessMapStore.businessMap.get(cell) || '未分配',
    },
    // {
    //   label: '云厂商',
    //   field: 'vendor',
    // },
    {
      label: '云账号',
      field: 'account_id',
    },
    {
      label: '操作人',
      field: 'operator',
      isDefaultShow: true,
    },
  ];

  const lbColumns = [
    {
      type: 'selection',
      width: 32,
      minWidth: 32,
      onlyShowOnList: true,
      align: 'right',
    },
    getLinkField({
      type: 'lb',
      label: '负载均衡名称',
      field: 'name',
      onLinkInBusiness: true,
      render: (data) => (
        <Button
          text
          theme='primary'
          onClick={renderFieldPushState(
            {
              name: LBRouteName.lb,
              params: { id: data.id },
              query: { ...route.query, type: 'detail' },
            },
            () => {
              loadBalancerStore.setLbTreeSearchTarget({ ...data, searchK: 'lb_name', searchV: data.name, type: 'lb' });
            },
          )}>
          {data.name || '--'}
        </Button>
      ),
    }),
    {
      label: () => (
        <span v-bk-tooltips={{ content: '用户通过该域名访问负载均衡流量', placement: 'top' }}>负载均衡域名</span>
      ),
      field: 'domain',
      isDefaultShow: true,
      render: ({ cell }: { cell: string }) => cell || '--',
    },
    {
      label: '负载均衡VIP',
      field: 'vip',
      isDefaultShow: true,
      render: ({ data }: any) => {
        return getInstVip(data);
      },
    },
    {
      label: '网络类型',
      field: 'lb_type',
      isDefaultShow: true,
      sort: true,
      filter: {
        list: [
          { text: LB_NETWORK_TYPE_MAP.OPEN, value: LB_NETWORK_TYPE_MAP.OPEN },
          { text: LB_NETWORK_TYPE_MAP.INTERNAL, value: LB_NETWORK_TYPE_MAP.INTERNAL },
        ],
      },
    },
    {
      label: '监听器数量',
      field: 'listenerNum',
      isDefaultShow: true,
      render: ({ cell }: { cell: number }) => cell || '0',
    },
    {
      label: '分配状态',
      field: 'bk_biz_id',
      isDefaultShow: true,
      isOnlyShowInResource: true,
      render: ({ cell }: { cell: number }) => (
        <bk-tag
          v-bk-tooltips={{
            content: businessMapStore.businessMap.get(cell),
            disabled: !cell || cell === -1,
          }}
          theme={cell === -1 ? false : 'success'}>
          {cell === -1 ? '未分配' : '已分配'}
        </bk-tag>
      ),
    },
    {
      label: '删除保护',
      field: 'delete_protect',
      isDefaultShow: true,
      render: ({ cell }: { cell: boolean }) => (cell ? <bk-tag theme='success'>开启</bk-tag> : <bk-tag>关闭</bk-tag>),
      filter: {
        list: [
          { text: '开启', value: true },
          { text: '关闭', value: false },
        ],
      },
    },
    {
      label: 'IP版本',
      field: 'ip_version',
      isDefaultShow: true,
      render: ({ cell }: { cell: string }) => IP_VERSION_MAP[cell],
      sort: true,
    },
    {
      label: '云厂商',
      field: 'vendor',
      render({ cell }: { cell: string }) {
        return h('span', [CloudType[cell] || '--']);
      },
      sort: true,
    },
    {
      label: '地域',
      field: 'region',
      render: ({ cell, row }: { cell: string; row: { vendor: VendorEnum } }) => getRegionName(row.vendor, cell) || '--',
      sort: true,
    },
    {
      label: '可用区域',
      field: 'zones',
      render: ({ cell }: { cell: string[] }) => cell?.join(','),
      sort: true,
    },
    {
      label: '状态',
      field: 'status',
      sort: true,
      render: ({ cell }: { cell: string }) => {
        let icon = StatusSuccess;
        switch (cell) {
          case '创建中':
            icon = StatusLoading;
            break;
          case '正常运行':
            icon = StatusSuccess;
            break;
        }
        return cell ? (
          <div class='status-column-cell'>
            <img class={`status-icon${cell === 'binding' ? ' spin-icon' : ''}`} src={icon} alt='' />
            <span>{cell}</span>
          </div>
        ) : (
          '--'
        );
      },
    },
    {
      label: '所属vpc',
      field: 'cloud_vpc_id',
      sort: true,
    },
  ];

  const listenerColumns = [
    getLinkField({
      type: 'listener',
      label: '监听器名称',
      field: 'name',
      render: (data) => (
        <Button
          text
          theme='primary'
          onClick={renderFieldPushState(
            {
              name: LBRouteName.listener,
              params: { id: data.id },
              query: { ...route.query, type: 'detail', protocol: data.protocol },
            },
            () => {
              loadBalancerStore.setLbTreeSearchTarget({
                ...data,
                searchK: 'listener_name',
                searchV: data.name,
                type: 'listener',
              });
            },
          )}>
          {data.name || '--'}
        </Button>
      ),
    }),
    {
      label: '监听器ID',
      field: 'cloud_id',
    },
    {
      label: '协议',
      field: 'protocol',
      isDefaultShow: true,
    },
    {
      label: '端口',
      field: 'port',
      isDefaultShow: true,
      render: ({ data, cell }: any) => `${cell}${data.end_port ? `-${data.end_port}` : ''}`,
    },
    {
      label: '均衡方式',
      field: 'scheduler',
      isDefaultShow: true,
      render: ({ cell }: { cell: string }) => SCHEDULER_MAP[cell] || '--',
    },
    {
      label: '域名数量',
      field: 'domain_num',
      isDefaultShow: true,
    },
    {
      label: 'URL数量',
      field: 'url_num',
      isDefaultShow: true,
    },
    {
      label: '同步状态',
      field: 'binding_status',
      isDefaultShow: true,
      render: ({ cell }: { cell: string }) => {
        let icon = StatusSuccess;
        switch (cell) {
          case 'binding':
            icon = StatusLoading;
            break;
          case 'success':
            icon = StatusSuccess;
            break;
        }
        return cell ? (
          <div class='status-column-cell'>
            <img class={`status-icon${cell === 'binding' ? ' spin-icon' : ''}`} src={icon} alt='' />
            <span>{CLB_BINDING_STATUS[cell]}</span>
          </div>
        ) : (
          '--'
        );
      },
    },
  ];

  const targetGroupColumns = [
    {
      type: 'selection',
      width: 32,
      minWidth: 32,
      onlyShowOnList: true,
      align: 'right',
    },
    getLinkField({
      type: 'name',
      label: '目标组名称',
      field: 'name',
      idFiled: 'name',
      onlyShowOnList: false,
      render: ({ id, name }) => (
        <Button
          text
          theme='primary'
          onClick={renderFieldPushState(
            {
              name: LBRouteName.tg,
              params: { id },
              query: { ...route.query, type: 'detail' },
            },
            () => {
              loadBalancerStore.setTgSearchTarget(name);
            },
          )}>
          {name}
        </Button>
      ),
    }),
    {
      label: '关联的负载均衡',
      field: 'lb_name',
      isDefaultShow: true,
      render({ cell }: any) {
        return cell?.trim() || '--';
      },
    },
    {
      label: '绑定监听器数量',
      field: 'listener_num',
      isDefaultShow: true,
    },
    {
      label: '协议',
      field: 'protocol',
      render({ cell }: any) {
        return cell?.trim() || '--';
      },
      isDefaultShow: true,
      sort: true,
      filter: {
        list: [
          { value: 'TCP', text: 'TCP' },
          { value: 'UDP', text: 'UDP' },
          { value: 'HTTP', text: 'HTTP' },
          { value: 'HTTPS', text: 'HTTPS' },
        ],
      },
    },
    {
      label: '端口',
      field: 'port',
      isDefaultShow: true,
      sort: true,
    },
    {
      label: '健康检查',
      field: 'health_check.health_switch',
      isDefaultShow: true,
      filter: {
        list: [
          { value: 1, text: '已开启' },
          { value: 0, text: '未开启' },
        ],
      },
      render({ cell }: { cell: Number }) {
        return cell ? <bk-tag theme='success'>已开启</bk-tag> : <bk-tag>未开启</bk-tag>;
      },
    },
    {
      label: '云厂商',
      field: 'vendor',
      render({ cell }: { cell: string }) {
        return h('span', [CloudType[cell] || '--']);
      },
      sort: true,
      filter: {
        list: [{ value: VendorEnum.TCLOUD, text: VendorMap[VendorEnum.TCLOUD] }],
      },
    },
    {
      label: '地域',
      field: 'region',
      render: ({ cell, row }: { cell: string; row: { vendor: VendorEnum } }) => getRegionName(row.vendor, cell) || '--',
      sort: true,
    },
    {
      label: '所属VPC',
      field: 'cloud_vpc_id',
      sort: true,
    },
    {
      label: '健康检查端口',
      field: 'health_check',
      render: ({ cell }: any) => {
        const { health_num, un_health_num } = cell;
        const total = health_num + un_health_num;
        if (!health_num || !un_health_num) return '--';
        return (
          <div class='port-status-col'>
            <span class={un_health_num ? 'un-health' : total ? 'health' : 'special-health'}>{un_health_num}</span>/
            <span>{health_num + un_health_num}</span>
          </div>
        );
      },
    },
  ];

  const rsConfigColumns = [
    {
      label: '内网IP',
      field: 'private_ip_address',
      isDefaultShow: true,
      render: ({ data }: any) => {
        return [
          ...(data.private_ipv4_addresses || []),
          ...(data.private_ipv6_addresses || []),
          // 更新目标组detail中的rs字段
          ...(data.private_ip_address || []),
        ].join(',');
      },
    },
    {
      label: '公网IP',
      field: 'public_ip_address',
      render: ({ data }: any) => {
        return (
          [
            ...(data.public_ipv4_addresses || []),
            ...(data.public_ipv6_addresses || []),
            // 更新目标组detail中的rs字段
            ...(data.public_ip_address || []),
          ].join(',') || '--'
        );
      },
    },
    {
      label: '名称',
      field: 'name',
      isDefaultShow: true,
      render: ({ data }: any) => {
        return data.name || data.inst_name;
      },
    },
    {
      label: '地域',
      field: 'region',
      render: ({ cell }: { cell: string }) => getRegionName(VendorEnum.TCLOUD, cell) || '--',
    },
    {
      label: '资源类型',
      field: 'inst_type',
      render: ({ data }: any) => {
        return data.machine_type || data.inst_type;
      },
    },
    {
      label: '所属VPC',
      field: 'cloud_vpc_ids',
      isDefaultShow: true,
      render: ({ cell }: { cell: string[] }) => cell?.join(','),
    },
  ];

  const domainColumns = [
    {
      label: 'URL数量',
      field: 'url_count',
      isDefaultShow: true,
      sort: true,
    },
    {
      label: '同步状态',
      field: 'sync_status',
      isDefaultShow: true,
      render: ({ cell }: { cell: string }) => {
        let icon = StatusSuccess;
        switch (cell) {
          case 'binding':
            icon = StatusLoading;
            break;
          case 'success':
            icon = StatusSuccess;
            break;
        }
        return cell ? (
          <div class='status-column-cell'>
            <img class={`status-icon${cell === 'binding' ? ' spin-icon' : ''}`} src={icon} alt='' />
            <span>{CLB_BINDING_STATUS[cell]}</span>
          </div>
        ) : (
          '--'
        );
      },
    },
  ];

  const targetGroupListenerColumns = [
    getLinkField({
      type: 'targetGroup',
      label: '绑定的监听器',
      field: 'lbl_name',
      render: ({ lbl_id, lbl_name, protocol }: any) => (
        <Button
          text
          theme='primary'
          onClick={renderFieldPushState({
            name: LBRouteName.listener,
            params: { id: lbl_id },
            query: {
              ...route.query,
              type: 'detail',
              protocol,
            },
          })}>
          {lbl_name}
        </Button>
      ),
    }),
    {
      label: '关联的负载均衡',
      field: 'lb_name',
      isDefaultShow: true,
      width: 300,
      render: ({ data }: any) => {
        const vip = getInstVip(data);
        const { lb_name } = data;
        return `${lb_name}（${vip}）`;
      },
    },
    {
      label: '关联的URL',
      field: 'url',
      isDefaultShow: true,
      render: ({ cell }: { cell: string }) => cell || '--',
    },
    {
      label: '协议',
      field: 'protocol',
      isDefaultShow: true,
      filter: {
        list: [
          { value: 'TCP', text: 'TCP' },
          { value: 'UDP', text: 'UDP' },
          { value: 'HTTP', text: 'HTTP' },
          { value: 'HTTPS', text: 'HTTPS' },
        ],
      },
    },
    {
      label: '端口',
      field: 'port',
      isDefaultShow: true,
      render: ({ data, cell }: any) => `${cell}${data.end_port ? `-${data.end_port}` : ''}`,
    },
    {
      label: '异常端口数',
      field: 'healthCheck',
      isDefaultShow: true,
      render: ({ cell }: any) => {
        if (!cell) return '--';
        const { health_num, un_health_num } = cell;
        return (
          <div class='port-status-col'>
            <span class={un_health_num ? 'un-health' : 'health'}>{un_health_num}</span>/
            <span>{health_num + un_health_num}</span>
          </div>
        );
      },
    },
  ];

  const urlColumns = [
    {
      type: 'selection',
      width: 32,
      minWidth: 32,
      onlyShowOnList: true,
      align: 'right',
    },
    {
      label: 'URL路径',
      field: 'url',
      isDefaultShow: true,
      sort: true,
    },
    {
      label: '轮询方式',
      field: 'scheduler',
      isDefaultShow: true,
      render: ({ cell }: { cell: string }) => SCHEDULER_MAP[cell] || '--',
      sort: true,
    },
  ];

  const certColumns = [
    {
      label: '证书名称',
      field: 'name',
    },
    {
      label: '资源ID',
      field: 'cloud_id',
    },
    {
      label: '云厂商',
      field: 'vendor',
      render({ cell }: { cell: string }) {
        return h('span', [CloudType[cell] || '--']);
      },
    },
    {
      label: '证书类型',
      field: 'cert_type',
      filter: {
        list: [
          {
            text: '服务器证书',
            value: '服务器证书',
          },
          {
            text: '客户端CA证书',
            value: '客户端CA证书',
          },
        ],
      },
    },
    {
      label: '域名',
      field: 'domain',
      render: ({ cell }: { cell: string[] }) => {
        return cell?.join(';') || '--';
      },
    },
    {
      label: '上传时间',
      field: 'cloud_created_time',
      sort: true,
      render: ({ cell }: { cell: string }) => {
        // 由于云上返回的是(UTC+8)时间, 所以先转零时区
        const utcTime = dayjs(cell).subtract(8, 'hour');
        return timeFormatter(utcTime);
      },
    },
    {
      label: '过期时间',
      field: 'cloud_expired_time',
      sort: true,
      render: ({ cell }: { cell: string }) => {
        // 由于云上返回的是(UTC+8)时间, 所以先转零时区
        const utcTime = dayjs(cell).subtract(8, 'hour');
        return timeFormatter(utcTime);
      },
    },
    {
      label: '证书状态',
      field: 'cert_status',
      filter: {
        list: [
          {
            text: '正常',
            value: '正常',
          },
          {
            text: '已过期',
            value: '已过期',
          },
        ],
      },
      render: ({ cell }: { cell: string }) => {
        let icon;
        switch (cell) {
          case '正常':
            icon = StatusNormal;
            break;
          case '已过期':
            icon = StatusAbnormal;
            break;
        }
        return (
          <div class='status-column-cell'>
            <img class='status-icon' src={icon} alt='' />
            <span>{cell}</span>
          </div>
        );
      },
    },
    {
      label: '分配状态',
      field: 'bk_biz_id',
      isOnlyShowInResource: true,
      isDefaultShow: true,
      render: ({ data, cell }: { data: { bk_biz_id: number }; cell: number }) => (
        <bk-tag
          v-bk-tooltips={{
            content: businessMapStore.businessMap.get(cell),
            disabled: !cell || cell === -1,
          }}
          theme={data.bk_biz_id === -1 ? false : 'success'}>
          {data.bk_biz_id === -1 ? '未分配' : '已分配'}
        </bk-tag>
      ),
    },
  ];
  const hIColumns = [
    {
      label: '需求类型',
      field: 'require_type',
    },
    {
      label: '实例族',
      field: 'label.device_group',
    },
    {
      label: '机型',
      field: 'device_type',
    },
    {
      label: 'CPU(核)',
      field: 'cpu',
      sort: true,
    },
    {
      label: '内存(G)',
      field: 'mem',
      sort: true,
    },
    {
      label: '地域',
      field: 'region',
      render: ({ cell }: { cell: string }) => getRegionName(VendorEnum.TCLOUD, cell) || '--',
    },
    {
      label: '园区',
      field: 'zone',
    },
    {
      label: '库存情况',
      field: 'capacity_flag',
      render({ cell }: { cell: string }) {
        const { class: theClass, text } = capacityLevel(cell);
        return <span class={theClass}>{text}</span>;
      },
    },
  ];
  const CRSOcolumns = [
    {
      type: 'selection',
      width: 32,
      minWidth: 32,
      onlyShowOnList: true,
    },
    {
      label: '机型',
      field: 'spec.device_type',
      width: 180,
    },
    {
      label: '交付情况总数',
      field: 'total_num',
    },
    {
      label: '交付情况待支付',
      field: 'pending_num',
    },
    {
      label: '交付情况已支付',
      field: 'success_num',
    },
    {
      label: '地域',
      field: 'spec.region',
      render: ({ cell }: { cell: string }) => getRegionName(VendorEnum.TCLOUD, cell) || '--',
    },
    {
      label: '园区',
      field: 'spec.zone',
    },
    {
      label: '反亲和性',
      field: 'anti_affinity_level',
      render: ({ cell }: { cell: string }) => cell || '无要求',
    },
    {
      label: '镜像',
      field: 'spec.image_id',
    },
    {
      label: '数据盘大小',
      field: 'spec.disk_size',
    },
    {
      label: '数据盘类型',
      field: 'spec.disk_type',
    },
    {
      label: '网络类型',
      field: 'spec.network_type',
    },
    {
      label: '备注',
      field: 'remark',
      render: ({ cell }: { cell: string }) => cell || '--',
    },
    {
      label: '状态',
      field: 'stage',
      width: 180,
    },
  ];
  const PRSOcolumns = [
    {
      type: 'selection',
      width: 32,
      minWidth: 32,
      onlyShowOnList: true,
    },
    {
      label: '机型',
      field: 'spec.device_type',
      width: 180,
    },
    {
      label: '交付情况总数',
      field: 'total_num',
    },
    {
      label: '交付情况待支付',
      field: 'pending_num',
    },
    {
      label: '交付情况已支付',
      field: 'success_num',
    },
    {
      label: '地域',
      field: 'spec.region',
      render: ({ cell }: { cell: string }) => getRegionName(VendorEnum.TCLOUD, cell) || '--',
    },
    {
      label: '园区',
      field: 'spec.zone',
    },
    {
      label: '反亲和性',
      field: 'anti_affinity_level',
    },
    {
      label: '操作系统',
      field: 'spec.os_type',
    },
    {
      label: '数据盘大小',
      field: 'spec.zone',
    },
    {
      label: 'RAID类型',
      field: 'spec.raid_type',
    },
    {
      label: '备注',
      field: 'remark',
      render: ({ cell }: { cell: string }) => cell || '--',
    },
    {
      label: '状态',
      field: 'stage',
      width: 180,
    },
  ];
  const CHColumns = [
    {
      label: '机型',
      field: 'spec.device_type',
      width: 180,
    },
    {
      label: '需求数量',
      field: 'replicas',
    },
    {
      label: '地域',
      field: 'spec.region',
      render: ({ cell }: { cell: string }) => getRegionName(VendorEnum.TCLOUD, cell) || '--',
    },
    {
      label: '园区',
      field: 'spec.zone',
    },
    {
      label: '镜像',
      field: 'spec.image_id',
    },
    {
      label: '数据盘大小',
      field: 'spec.disk_size',
      width: 180,
    },
    {
      label: '数据盘类型',
      field: 'spec.disk_type',
    },
    {
      label: '私有网络',
      field: 'spec.vpc',
    },
    {
      label: '私有子网',
      field: 'spec.subnet',
    },
    {
      label: '网络类型',
      field: 'spec.network_type',
    },
    {
      label: '备注',
      field: 'remark',
    },
  ];
  const PMColumns = [
    {
      label: '机型',
      field: 'spec.device_type',
      width: 150,
    },
    {
      label: '需求数量',
      field: 'replicas',
    },
    {
      label: '地域',
      field: 'spec.region',
      render: ({ cell }: { cell: string }) => getRegionName(VendorEnum.TCLOUD, cell) || '--',
    },
    {
      label: '园区',
      field: 'spec.zone',
    },
    {
      label: 'RAID 类型',
      field: 'spec.raid_type',
    },
    {
      label: '操作系统',
      field: 'spec.os_type',
    },
    {
      label: '备注',
      field: 'remark',
    },
  ];
  const RRColumns = [
    {
      type: 'selection',
      width: 32,
      minWidth: 32,
      onlyShowOnList: true,
      align: 'right',
    },
    {
      label: '状态',
      field: 'recyclable',
    },
    {
      label: '固资号',
      field: 'asset_id',
    },
    {
      label: '内网IP',
      field: 'ip',
    },
    {
      label: '所属业务',
      field: 'bk_biz_name',
    },
    {
      label: '所属模块',
      field: 'topo_module',
    },
    {
      label: '维护人',
      field: 'operator',
    },
    {
      label: '备份维护人',
      field: 'bak_operator',
    },
    {
      label: '机型',
      field: 'device_type',
    },
    {
      label: '主机状态',
      field: 'state',
    },
    {
      label: '入库时间',
      field: 'input_time',
    },
  ];
  const BSAColumns = [
    {
      type: 'selection',
      width: 32,
      minWidth: 32,
      onlyShowOnList: true,
    },
    {
      label: '固资号',
      field: 'asset_id',
    },
    {
      label: '内网IP',
      field: 'ip',
    },
    {
      label: '机型',
      field: 'device_type',
    },
    {
      label: '园区',
      field: 'sub_zone',
    },
  ];
  const RTColumns = [
    {
      label: '固资号',
      field: 'asset_id',
    },
    {
      label: '内网IP',
      field: 'ip',
    },
    {
      label: '机型',
      field: 'device_type',
    },
    {
      label: '园区',
      field: 'sub_zone',
    },
    {
      label: '维护人',
      field: 'operator',
    },
    {
      label: '备份维护人',
      field: 'bak_operator',
    },
  ];
  const DQcolumns = [
    {
      type: 'selection',
      width: 32,
      minWidth: 32,
      onlyShowOnList: true,
      align: 'center',
    },
    {
      label: '业务',
      field: 'bk_biz_id',
      render({ cell }: any) {
        return businessMapStore.getNameFromBusinessMap(cell);
      },
    },
    {
      label: '单号',
      field: 'order_id',
      render: ({ cell }: any) => {
        return (
          <Button
            text
            theme='primary'
            onClick={() => {
              // 跳转到单据申请详情页
            }}>
            {cell}
          </Button>
        );
      },
    },
    {
      label: '子单号',
      field: 'suborder_id',
    },
    {
      label: '需求类型',
      field: 'require_type',
    },
    {
      label: '申请人',
      field: 'bk_username',
      render({ cell }: any) {
        return <WName name={cell} />;
      },
    },
    {
      label: '内网IP',
      field: 'ip',
    },
    {
      label: '固资号',
      field: 'asset_id',
    },
    {
      label: '资源类型',
      field: 'resource_type',
    },
    {
      label: '机型',
      field: 'device_type',
    },
    {
      label: '园区',
      field: 'zone_id',
    },
    {
      label: '交付时间',
      field: 'update_at',
      render: ({ cell }: any) => timeFormatter(cell),
    },
    {
      label: '申请时间',
      field: 'create_at',
      render: ({ cell }: any) => timeFormatter(cell),
    },
    {
      label: '备注信息',
      field: 'remark',
      render({ data }: any) {
        return `${data.description}${data.description && data.remark && '/'}${data.remark}` || '--';
      },
    },
  ];
  const CAcolumns = [
    {
      label: '需求类型',
      field: 'require_type',
    },
    {
      label: '实例族',
      field: 'label.device_group',
    },
    {
      label: '机型',
      field: 'device_type',
    },
    {
      label: 'CPU核',
      field: 'cpu',
    },
    {
      label: '内存(G)',
      field: 'mem',
    },
    {
      label: '地域',
      field: 'region',
      render: ({ cell }: { cell: string }) => getRegionName(VendorEnum.TCLOUD, cell) || '--',
    },
    {
      label: '园区',
      field: 'zone',
    },
    {
      label: '库存情况',
      field: 'capacity_flag',
      render({ cell }: { cell: string }) {
        const { class: theClass, text } = capacityLevel(cell);
        return <span class={theClass}>{text}</span>;
      },
    },
  ];
  // 预检详情状态 render
  const getPrecheckStatusView = (value: string) => {
    const label = getPrecheckStatusLabel(value);
    if (value === 'SUCCESS') {
      return <span class='c-success'>{label}</span>;
    }
    if (value === 'RUNNING') {
      return (
        <>
          <Spinner />
          <span>{label}</span>
        </>
      );
    }
    if (value === 'FAILED') {
      return (
        <div>
          <div class='c-danger fail-flex'>
            <span>{label}</span>
            <a
              target='_blank'
              href='https://iwiki.woa.com/pages/viewpage.action?pageId=349178371'
              v-bk-tooltips={{
                content:
                  '请查看解决方案，根据提示处理完成后，请返回“回收单据列表”或者跳转到“单据详情”进行“重试”或者“去除预检失败IP提交”',
              }}>
              <Share />
            </a>
          </div>
        </div>
      );
    }
    return <span>{label}</span>;
  };
  const ERcolumns = [
    {
      label: '步骤',
      field: 'step_desc',
    },
    {
      label: '状态',
      field: 'status',
      render: ({ row }) => {
        return getPrecheckStatusView(row.status);
      },
    },
    {
      label: '开始时间',
      field: 'create_at',
      render: ({ row }) => {
        return <span>{dateTimeTransform(row.create_at)}</span>;
      },
    },
    {
      label: '结束时间',
      field: 'end_at',
      render: ({ row }) => {
        return <span>{dateTimeTransform(row.end_at)}</span>;
      },
    },
    {
      label: '执行日志',
      field: 'log',
    },
  ];
  const PDcolumns = [
    {
      label: '单号',
      field: 'order_id',
      width: 80,
    },
    {
      label: '子单号',
      field: 'suborder_id',
      width: 80,
    },
    {
      label: '状态',
      field: 'status',
      render: ({ row }) => {
        return getPrecheckStatusView(row.status);
      },
      exportFormatter: (row) => getPrecheckStatusLabel(row.status),
    },
    {
      label: '已执行/总数',
      field: 'mem',
      render: ({ row }) => {
        return (
          <div>
            <span class={row.success_num > 0 ? 'c-success' : ''}>{row.success_num}</span>
            <span>/</span>
            <span>{row.total_num}</span>
          </div>
        );
      },
      exportFormatter: (row) => {
        return `${row.success_num}/${row.total_num}`;
      },
    },
    {
      label: '更新时间',
      field: 'update_at',
      render: ({ row }) => {
        return <span>{dateTimeTransform(row.update_at)}</span>;
      },
      formatter: ({ update_at }) => {
        return dateTimeTransform(update_at);
      },
    },
    {
      label: '创建时间',
      field: 'create_at',
      render: ({ row }) => {
        return <span>{dateTimeTransform(row.create_at)}</span>;
      },
      formatter: ({ create_at }) => {
        return dateTimeTransform(create_at);
      },
    },
  ];
  const getRecycleTaskStatusView = (value: string) => {
    const label = getRecycleTaskStatusLabel(value);
    if (value === 'DONE') {
      return <span class='c-success'>{label}</span>;
    }
    if (value.includes('ING')) {
      return (
        <>
          <Spinner />
          <span>{label}</span>
        </>
      );
    }
    if (value === 'DETECT_FAILED') {
      return (
        <bk-badge
          class='c-danger'
          v-bk-tooltips={{ content: '请到“预检详情”查看失败原因，或者点击“去除预检失败IP提交”' }}
          dot>
          {label}
        </bk-badge>
      );
    }
    if (value.includes('FAILED')) {
      return <span class='c-danger'>{label}</span>;
    }
    return <span>{label}</span>;
  };
  // 资源 - 主机回收列表
  const recycleOrderColumns = [
    {
      type: 'selection',
    },
    {
      label: '单号',
      field: 'order_id',
      width: 80,
    },
    {
      label: '业务',
      field: 'bk_biz_id',
      formatter: ({ bk_biz_id }) => {
        return getBusinessNameById(bk_biz_id);
      },
    },
    {
      label: '资源类型',
      field: 'resource_type',
      width: 120,
      render: ({ row }) => {
        return <span>{getResourceTypeName(row.resource_type)}</span>;
      },
      formatter: ({ resource_type }) => {
        return getResourceTypeName(resource_type);
      },
    },
    {
      label: '回收类型',
      field: 'return_plan',
      render: ({ row }) => {
        return <span>{getReturnPlanName(row.return_plan, row.resource_type)}</span>;
      },
      formatter: ({ return_plan, resource_type }) => {
        return getReturnPlanName(return_plan, resource_type);
      },
    },
    {
      label: '回收成本',
      field: 'cost_concerned',
      render: ({ row }) => {
        return <span>{row.cost_concerned ? '涉及' : '不涉及'}</span>;
      },
      formatter: ({ cost_concerned }) => {
        return cost_concerned ? '涉及' : '不涉及';
      },
    },
    {
      label: '状态',
      field: 'status',
      width: 100,
      render: ({ row }) => {
        return getRecycleTaskStatusView(row.status);
      },
      exportFormatter: (row) => getRecycleTaskStatusLabel(row.status),
    },
    {
      label: '当前处理人',
      field: 'handler',
      width: 100,
      render: ({ row }) => {
        return row.handler !== 'AUTO' ? (
          <a href={`wxwork://message?username=${row.handler}`} class='username'>
            {row.handler}
          </a>
        ) : (
          <span class='username'>{row.handler}</span>
        );
      },
    },
    {
      label: '总数/成功/失败',
      width: 120,
      render: ({ row }) => {
        return (
          <div>
            <span>{row.total_num}</span>
            <span>/</span>
            <span class={row.success_num > 0 ? 'c-success' : ''}>{row.success_num}</span>
            <span>/</span>
            <span class={row.failed_num > 0 ? 'c-danger' : ''}>{row.failed_num}</span>
          </div>
        );
      },
      exportFormatter: (row) => {
        return `${row.success_num}/${row.failed_num}/${row.total_num}`;
      },
    },
    {
      label: '回收人',
      field: 'bk_username',
      render: ({ row }) => {
        return (
          <a href={`wxwork://message?username=${row.bk_username}`} class='username'>
            {row.bk_username}
          </a>
        );
      },
    },
    {
      label: '回收时间',
      field: 'create_at',
      render: ({ row }) => {
        return <span>{dateTimeTransform(row.create_at)}</span>;
      },
      formatter: ({ create_at }) => {
        return dateTimeTransform(create_at);
      },
    },
    {
      label: '描述',
      field: 'remark',
      showOverflowTooltip: true,
    },
    {
      label: 'OBS项目类型',
      field: 'recycle_type',
      width: 120,
    },
  ];
  // 资源- 设备查询
  const deviceQueryColumns = [
    {
      label: '单号',
      field: 'order_id',
      width: 80,
    },
    {
      label: '固资号',
      field: 'asset_id',
    },
    {
      label: '机型',
      field: 'device_type',
    },
    {
      label: '内网IP',
      field: 'ip',
    },
    {
      label: '回收业务',
      field: 'bk_biz_id',
      formatter: ({ bk_biz_id }) => {
        return getBusinessNameById(bk_biz_id);
      },
    },
    {
      label: '地域',
      field: 'bk_zone_name',
    },
    {
      label: '园区',
      field: 'sub_zone',
    },
    {
      label: 'Module名称',
      field: 'module_name',
    },
    {
      label: '标记',
      field: 'return_tag',
    },
    {
      label: '成本分摊比例',
      field: 'return_cost_rate',
      render: ({ row }) => {
        return row.return_cost_rate ? `${Math.ceil(row.return_cost_rate * 100)}%` : '-';
      },
    },
    {
      label: '状态',
      field: 'status',
      render: ({ row }) => getRecycleTaskStatusView(row.status),
      exportFormatter: (row) => getRecycleTaskStatusLabel(row.status),
    },
    {
      label: '回收人',
      field: 'bk_username',
      render: ({ row }) => {
        return (
          <a href={`wxwork://message?username=${row.bk_username}`} class='username'>
            {row.bk_username}
          </a>
        );
      },
    },
    {
      label: '创建时间',
      field: 'create_at',
      render: ({ row }) => {
        return <span>{dateTimeTransform(row.create_at)}</span>;
      },
      formatter: ({ create_at }) => {
        return dateTimeTransform(create_at);
      },
    },
    {
      label: '完成时间',
      field: 'return_time',
    },
    {
      label: '备注',
      field: 'remark',
    },
  ];
  // 资源 - 主机回收 - 单据详情 设备销毁列表
  const deviceDestroyColumns = [
    {
      type: 'selection',
    },
    {
      label: '固资号',
      field: 'asset_id',
    },
    {
      label: '实例ID',
      field: 'instance_id',
    },
    {
      label: '机型',
      field: 'device_type',
    },
    {
      label: '园区',
      field: 'sub_zone',
    },
    {
      label: 'Module名称',
      field: 'module_name',
    },
    {
      label: '维护人',
      field: 'operator',
      render: ({ row }) => {
        return (
          <a href={`wxwork://message?username=${row.operator}`} class='username'>
            {row.operator}
          </a>
        );
      },
    },
    {
      label: '备份维护人',
      field: 'bak_operator',
      render: ({ row }) => {
        return (
          <a href={`wxwork://message?username=${row.bak_operator}`} class='username'>
            {row.bak_operator}
          </a>
        );
      },
    },
    {
      label: '标记',
      field: 'return_tag',
    },
    {
      label: '成本分摊比例',
      field: 'return_cost_rate',
      render: ({ row }) => {
        return row.return_cost_rate ? `${Math.ceil(row.return_cost_rate * 100)}%` : '-';
      },
    },
    {
      label: '校验结果',
      field: 'return_plan_msg',
      showOverflowTooltip: true,
      render: ({ row }) => {
        return (
          <bk-link type='info' v-clipboard={row.return_plan_msg} underline={false}>
            {row.return_plan_msg}
          </bk-link>
        );
      },
    },
    {
      label: '上架时间',
      field: 'input_time',
      render: ({ row }) => {
        return <span>{dateTimeTransform(row.input_time)}</span>;
      },
      formatter: ({ input_time }) => {
        return dateTimeTransform(input_time);
      },
    },
    {
      label: '销毁时间',
      field: 'return_time',
      render: ({ row }) => {
        return <span>{dateTimeTransform(row.return_time)}</span>;
      },
      formatter: ({ return_time }) => {
        return dateTimeTransform(return_time);
      },
    },
    {
      label: '回收单号',
      field: 'return_id',
      render: ({ row }) => {
        return (
          <bk-link type='primary' underline={false} href={row.return_link} target='_blank'>
            {row.return_id}
          </bk-link>
        );
      },
    },
    {
      label: '状态',
      field: 'status',
      render: ({ row }) => getRecycleTaskStatusView(row.status),
      exportFormatter: (row) => getRecycleTaskStatusLabel(row.status),
    },
  ];

  const scrResourceOnlineColumns = [
    getLinkField({
      type: 'scrResourceOnlineTask',
      label: '单号',
      field: 'id',
      render: ({ id }) => (
        <Button
          text
          theme='primary'
          onClick={renderFieldPushState({
            name: 'scrResourceManageDetail',
            params: { id },
            query: { type: 'online' },
          })}>
          {id}
        </Button>
      ),
    }),
    {
      label: '创建人',
      field: 'bk_username',
      render: ({ cell }: any) => <WName name={cell} />,
    },
    {
      label: '创建时间',
      field: 'create_at',
      render: ({ cell }: any) => timeFormatter(cell),
    },
    {
      label: '上架数量',
      field: 'total_num',
      render: ({ data }: any) => data?.status?.total_num || 0,
    },
    {
      label: '完成数量',
      field: 'success_num',
      render: ({ data }: any) => data?.status?.success_num || 0,
    },
    {
      label: '单据状态',
      field: 'phase',
      render: ({ data }: any) => {
        const phase = data?.status?.phase;
        const desc = SCR_POOL_PHASE_MAP[phase];

        if (phase === 'INIT') return <span class='c-info'>{desc}</span>;
        if (phase === 'RUNNING')
          return (
            <span class='status-column-cell'>
              <img class='status-icon spin-icon' src={StatusLoading} alt='' />
              {desc}
            </span>
          );
        if (phase === 'SUCCESS') return <span class='c-success'>{desc}</span>;
        if (phase === 'FAILED') return <span class='c-danger'>{desc}</span>;

        return phase;
      },
    },
  ];

  const scrResourceOfflineColumns = [
    getLinkField({
      type: 'scrResourceOnlineTask',
      label: '单号',
      field: 'id',
      render: ({ id }) => (
        <Button
          text
          theme='primary'
          onClick={renderFieldPushState({
            name: 'scrResourceManageDetail',
            params: { id },
            query: { type: 'offline' },
          })}>
          {id}
        </Button>
      ),
    }),
    {
      label: '创建人',
      field: 'bk_username',
      render: ({ cell }: any) => <WName name={cell} />,
    },
    {
      label: '创建时间',
      field: 'create_at',
      render: ({ cell }: any) => timeFormatter(cell),
    },
    {
      label: '下架数量',
      field: 'total_num',
      render: ({ data }: any) => data?.status?.total_num || 0,
    },
    {
      label: '完成数量',
      field: 'success_num',
      render: ({ data }: any) => data?.status?.success_num || 0,
    },
    {
      label: '单据状态',
      field: 'phase',
      render: ({ data }: any) => {
        const phase = data?.status?.phase;
        const desc = SCR_POOL_PHASE_MAP[phase];

        if (phase === 'INIT') return <span class='c-info'>{desc}</span>;
        if (phase === 'RUNNING')
          return (
            <span class='status-column-cell'>
              <img class='status-icon spin-icon' src={StatusLoading} alt='' />
              {desc}
            </span>
          );
        if (phase === 'SUCCESS') return <span class='c-success'>{desc}</span>;
        if (phase === 'FAILED') return <span class='c-danger'>{desc}</span>;

        return phase;
      },
    },
  ];

  const scrResourceOnlineHostColumns = [
    {
      label: '内网IP',
      field: 'ip',
      render: ({ data }: any) => data?.labels?.ip,
      exportFormatter: (data: any) => data?.labels?.ip,
    },
    {
      label: '固资号',
      field: 'bk_asset_id',
      render: ({ data }: any) => data?.labels?.bk_asset_id,
      exportFormatter: (data: any) => data?.labels?.bk_asset_id,
    },
    {
      label: '设备类型',
      field: 'device_type',
      render: ({ data }: any) => data?.labels?.device_type,
      exportFormatter: (data: any) => data?.labels?.device_type,
    },
    {
      label: '状态',
      field: 'phase',
      render: ({ cell }: any) => {
        const desc = SCR_POOL_PHASE_MAP[cell];

        if (cell === 'INIT') return <span class='c-info'>{desc}</span>;
        if (cell === 'RUNNING')
          return (
            <span class='status-column-cell'>
              <img class='status-icon spin-icon' src={StatusLoading} alt='' />
              {desc}
            </span>
          );
        if (cell === 'SUCCESS') return <span class='c-success'>{desc}</span>;
        if (cell === 'FAILED') return <span class='c-danger'>{desc}</span>;

        return cell;
      },
      exportFormatter: ({ phase }: any) => SCR_POOL_PHASE_MAP[phase],
    },
    {
      label: '开始时间',
      field: 'create_at',
      render: ({ cell }: any) => timeFormatter(cell),
      exportFormatter: ({ create_at }: any) => timeFormatter(create_at),
    },
    {
      label: '结束时间',
      field: 'update_at',
      render: ({ cell }: any) => timeFormatter(cell),
      exportFormatter: ({ update_at }: any) => timeFormatter(update_at),
    },
    {
      label: '信息',
      field: 'message',
      render: ({ cell }: any) => cell || '--',
    },
  ];

  const scrResourceOfflineHostColumns = [
    {
      label: '内网IP',
      field: 'ip',
      render: ({ data }: any) => data?.labels?.ip,
      exportFormatter: (data: any) => data?.labels?.ip,
    },
    {
      label: '固资号',
      field: 'bk_asset_id',
      render: ({ data }: any) => data?.labels?.bk_asset_id,
      exportFormatter: (data: any) => data?.labels?.bk_asset_id,
    },
    {
      label: '系统重装任务',
      field: 'reinstall_link',
      render: ({ data }: any) => (
        <a class='link-type' href={data.reinstall_link} target='_blank'>
          {data.reinstall_id}
        </a>
      ),
      exportFormatter: ({ reinstall_id }: any) => reinstall_id,
    },
    {
      label: '配置检查任务',
      field: 'conf_check_link',
      render: ({ data }: any) => (
        <a class='link-type' href={data.conf_check_link} target='_blank'>
          {data.conf_check_id}
        </a>
      ),
      exportFormatter: ({ conf_check_id }: any) => conf_check_id,
    },
    {
      label: '状态',
      field: 'status',
      render: ({ cell }: any) => {
        const desc = SCR_RECALL_DETAIL_STATUS_MAP[cell];

        if (cell === 'TERMINATE') return <span class='c-info'>{desc}</span>;
        if (cell === 'REINSTALLING') {
          return (
            <span class='status-column-cell'>
              <img class='status-icon spin-icon' src={StatusLoading} alt='' />
              {desc}
            </span>
          );
        }
        if (cell === 'RETURNED' || cell === 'DONE') return <span class='c-success'>{desc}</span>;
        if (cell === 'REINSTALL_FAILED') return <span class='c-danger'>{desc}</span>;

        return cell;
      },
      exportFormatter: ({ status }: any) => SCR_RECALL_DETAIL_STATUS_MAP[status],
    },
    {
      label: '开始时间',
      field: 'create_at',
      render: ({ cell }: any) => timeFormatter(cell),
      exportFormatter: ({ create_at }: any) => timeFormatter(create_at),
    },
    {
      label: '结束时间',
      field: 'update_at',
      render: ({ cell }: any) => timeFormatter(cell),
      exportFormatter: ({ update_at }: any) => timeFormatter(update_at),
    },
    {
      label: '信息',
      field: 'message',
      render: ({ cell }: any) => cell || '--',
    },
  ];

  const scrResourceOnlineCreateColumns = [
    {
      type: 'selection',
      width: 32,
      minWidth: 32,
      onlyShowOnList: true,
    },
    {
      label: '固资号',
      field: 'asset_id',
    },
    {
      label: '内网 IP',
      field: 'ip',
    },
    {
      label: '机型',
      field: 'device_type',
    },
    {
      label: '操作系统',
      field: 'os_type',
    },
    {
      label: '机架号',
      field: 'equipment',
    },
    {
      label: '园区',
      field: 'zone',
    },
    {
      label: '模块',
      field: 'module',
    },
    {
      label: 'IDC 单元',
      field: 'idc_unit',
    },
    {
      label: '逻辑区域',
      field: 'idc_logic_area',
    },
    {
      label: '入库时间',
      field: 'input_time',
      render: ({ cell }: any) => timeFormatter(cell),
    },
  ];

  const scrResourceOfflineCreateColumns = [
    {
      label: '机型',
      field: 'device_type',
    },
    {
      label: '地域',
      field: 'region',
    },
    {
      label: '园区',
      field: 'zone',
    },
    {
      label: '数量',
      field: 'amount',
    },
  ];
  // 资源配置管理-CVM子网
  const cvmWebColumns = [
    {
      type: 'selection',
    },
    {
      label: 'VPC',
      field: 'vpc_name',
      render: ({ row }) => {
        return (
          <div class='cvm-cell-height'>
            <div>{row.vpc_name}</div>
            <div>{row.vpc_id}</div>
          </div>
        );
      },
    },
    {
      label: 'Subnet',
      field: 'subnet_name',
      render: ({ row }) => {
        return (
          <div class='cvm-cell-height'>
            <div>{row.subnet_name}</div>
            <div>{row.subnet_id}</div>
          </div>
        );
      },
    },
    {
      label: '地域',
      field: 'region',
      render: ({ row }) => {
        return (
          <div class='cvm-cell-height'>
            <div> {getRegionCn(row.region)}</div>
            <div>{row.region}</div>
          </div>
        );
      },
    },
    {
      label: '园区',
      field: 'zone',
      render: ({ row }) => {
        return (
          <div class='cvm-cell-height'>
            <div> {getZoneCn(row.zone)}</div>
            <div>{row.zone}</div>
          </div>
        );
      },
    },
  ];

  const ApplicationListColumns = [
    {
      label: '申请人',
      render: ({ data }: any) => {
        return <WName name={data.bk_username}></WName>;
      },
    },
    {
      label: '交付情况-总数',
      field: 'total_num',
    },
    {
      label: '交付情况-待交付',
      field: 'pending_num',
    },
    {
      label: '交付情况-已交付',
      field: 'success_num',
      width: 180,
      render: ({ data }: any) => {
        if (data.success_num > 0) {
          const ips: any[] = [];
          const assetIds: any[] = [];
          const goToCmdb = (ips: string[]) => {
            window.open(`http://bkcc.oa.com/#/business/${data.bkBizId}/index?ip=text=${ips.join(',')}`);
          };

          return (
            <div class={'flex-row align-item-center'}>
              {data.success_num}
              <Button
                text
                theme={'primary'}
                class='ml8 mr8'
                v-clipboard:copy={ips.join('\n')}
                v-bk-tooltips={{
                  content: '复制 IP',
                }}>
                <Copy />
              </Button>
              <Button
                text
                theme={'primary'}
                class='mr8'
                v-clipboard:copy={assetIds.join('\n')}
                v-bk-tooltips={{
                  content: '复制固资号',
                }}>
                <Copy />
              </Button>
              <Button
                text
                theme={'primary'}
                onClick={() => goToCmdb(ips)}
                v-bk-tooltips={{
                  content: '去蓝鲸配置平台管理资源',
                }}>
                <DataShape />
              </Button>
            </div>
          );
        }

        return <span>{data.success_num}</span>;
      },
    },
    {
      label: '申请时间',
      field: 'create_at',
      render: ({ data }: any) => timeFormatter(data.create_at, 'YYYY-MM-DD'),
    },
    {
      label: '期望交付时间',
      field: 'expect_time',
      render: ({ data }: any) => timeFormatter(data.expect_time, 'YYYY-MM-DD'),
    },
    {
      label: '备注信息',
      field: 'remark',
      width: '150',
      render: ({ data }: any) => (
        <div>
          {data.description}
          {data.description && data.remark && '/'}
          {data.remark}
        </div>
      ),
    },
  ];
  const cvmModelColumns = [
    {
      type: 'selection',
    },
    {
      label: '机型',
      field: 'device_type',
    },
    {
      label: '需求类型',
      field: 'require_type',
    },
    {
      label: '地域',
      field: 'region',
      render: ({ cell }: { cell: string }) => getRegionName(VendorEnum.TCLOUD, cell) || '--',
    },
    {
      label: '园区',
      field: 'zone',
    },
    {
      label: '实例族',
      field: 'label.device_group',
    },
    {
      label: 'CPU(核)',
      field: 'cpu',
    },
    {
      label: '内存(G)',
      field: 'mem',
    },
    {
      label: '其他信息',
      field: 'remark',
      render: ({ cell }) => cell || '--',
    },
    {
      label: '可查询容量',
      field: 'enable_capacity',
      render: ({ cell }) => (cell ? '是' : '否'),
    },
    {
      label: '可申请',
      field: 'enable_apply',
      render: ({ cell }) => (cell ? '是' : '否'),
    },
    {
      label: '推荐分数',
      field: 'score',
    },
    {
      label: '备注',
      field: 'comment',
      render: ({ comment }) => comment || '--',
    },
  ];
  const firstAccountColumns = [
    {
      label: '一级帐号ID',
      field: 'primaryAccountId',
    },
    {
      label: '云厂商',
      field: 'cloudProvider',
    },
    {
      label: '帐号邮箱',
      field: 'accountEmail',
    },
    {
      label: '主负责人',
      field: 'mainResponsiblePerson',
    },
    {
      label: '组织架构',
      field: 'organizationalStructure',
    },
    {
      label: '二级帐号个数',
      field: 'secondaryAccountCount',
    },
  ];

  const secondaryAccountColumns = [
    {
      label: '二级帐号ID',
      field: 'secondaryAccountId',
    },
    {
      label: '所属一级帐号',
      field: 'parentPrimaryAccount',
    },
    {
      label: '云厂商',
      field: 'cloudProvider',
    },
    {
      label: '站点类型',
      field: 'siteType',
    },
    {
      label: '帐号邮箱',
      field: 'accountEmail',
    },
    {
      label: '主负责人',
      field: 'mainResponsiblePerson',
    },
    {
      label: '运营产品',
      field: 'operatingProduct',
    },
  ];

  const myApplyColumns = [
    // {
    //   label: '申请ID',
    //   field: 'id',
    // },
    // {
    //   label: '来源',
    //   field: 'source',
    // },
    {
      label: '申请类型',
      field: 'type',
    },
    {
      label: '单据状态',
      field: 'status',
      render({ data }: any) {
        let icon = StatusAbnormal;
        let txt = '审批拒绝';
        switch (data.status) {
          case 'pending':
          case 'delivering':
            icon = StatusLoading;
            txt = '审批中';
            break;
          case 'pass':
          case 'completed':
          case 'deliver_partial':
            icon = StatusSuccess;
            txt = '审批通过';
            break;
          case 'rejected':
          case 'cancelled':
          case 'deliver_error':
            icon = StatusFailure;
            txt = '审批拒绝';
            break;
        }
        return (
          <div class={'cvm-status-container'}>
            {txt === '审批中' ? (
              <Spinner fill='#3A84FF' class={'mr6'} width={14} height={14} />
            ) : (
              <img src={icon} class={'mr6'} width={14} height={14} />
            )}

            {txt}
          </div>
        );
      },
    },
    {
      label: '申请人',
      field: 'applicant',
    },
    {
      label: '创建时间',
      field: 'created_at',
      render({ cell }: any) {
        return timeFormatter(cell);
      },
    },
    {
      label: '更新时间',
      field: 'updated_at',
      render({ cell }: any) {
        return timeFormatter(cell);
      },
    },
    {
      label: '备注',
      field: 'memo',
      render({ cell }: any) {
        return cell || '--';
      },
    },
  ];

  // 服务请求 - 资源预测
  const forecastDemandColumns = [
    {
      label: '业务',
      field: 'bk_biz_name',
      isDefaultShow: true,
    },
    {
      label: '单据状态',
      field: 'status_name',
      isDefaultShow: true,
    },
    {
      label: '运营产品',
      field: 'bk_product_name',
    },
    {
      label: '规划产品',
      field: 'plan_product_name',
    },
    {
      label: 'CPU总核心数',
      field: 'cpu_core',
      isDefaultShow: true,
    },
    {
      label: '内存总量(GB)',
      field: 'memory',
      isDefaultShow: true,
    },
    {
      label: '云硬盘总量(GB)',
      field: 'disk_size',
      isDefaultShow: true,
    },
    {
      label: '提单人',
      field: 'applicant',
      isDefaultShow: true,
    },
    {
      label: '备注',
      field: 'remark',
    },
    {
      label: '创建时间',
      field: 'created_at',
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
    },
    {
      label: '提单时间',
      field: 'submitted_at',
      isDefaultShow: true,
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
    },
  ];

  // 资源预测详情
  const forecastDemandDetailColums = [
    {
      label: '机型规格',
      field: 'cvm.device_type',
      isDefaultShow: true,
    },
    {
      label: '总CPU核数',
      field: 'cvm.cpu_core',
      isDefaultShow: true,
    },
    {
      label: '总内存(G)',
      field: 'cvm.memory',
      isDefaultShow: true,
    },
    {
      label: '总云盘大小(G)',
      field: 'cbs.disk_size',
      isDefaultShow: true,
    },
    {
      label: '项目类型',
      field: 'obs_project',
      isDefaultShow: true,
    },
    {
      label: '地域',
      field: 'area_name',
      isDefaultShow: true,
    },
    {
      label: '城市',
      field: 'region_name',
      isDefaultShow: true,
    },
    {
      label: '可用区',
      field: 'zone_name',
      isDefaultShow: true,
    },
    {
      label: '资源模式',
      field: 'cvm.res_mode',
      isDefaultShow: true,
    },
    {
      label: '期望到货时间',
      field: 'expect_time',
      isDefaultShow: true,
    },
    {
      label: '机型族',
      field: 'cvm.device_family',
    },
    {
      label: '机型类型',
      field: 'cvm.device_class',
    },
    {
      label: '资源池',
      field: 'cvm.res_pool',
    },
    {
      label: '核心类型',
      field: 'cvm.core_type',
    },
    {
      label: '实例数',
      field: 'cvm.os',
    },
    {
      label: '单例磁盘IO(MB/s)',
      field: 'cbs.disk_io',
      isDefaultShow: true,
    },
    {
      label: '云磁盘类型',
      field: 'cbs.disk_type_name',
      isDefaultShow: true,
    },
    {
      label: '备注',
      field: 'remark',
    },
  ];

  // 预测清单
  const forecastListColums = [
    {
      label: '机型规格',
      field: 'cvm.device_type',
      isDefaultShow: true,
    },
    {
      label: 'CPU总核数',
      field: 'cvm.cpu_core',
      isDefaultShow: true,
    },
    {
      label: '内存总量(G)',
      field: 'cvm.memory',
      isDefaultShow: true,
    },
    {
      label: '云盘总量(G)',
      field: 'cbs.disk_size',
      isDefaultShow: true,
    },
    {
      label: '项目类型',
      field: 'obs_project',
      isDefaultShow: true,
    },
    {
      label: '期望到货时间',
      field: 'expect_time',
      isDefaultShow: true,
    },
    {
      label: '城市',
      field: 'region_name',
      isDefaultShow: true,
    },
    {
      label: '可用区',
      field: 'zone_name',
      isDefaultShow: true,
    },
    {
      label: '资源模式',
      field: 'cvm.res_mode',
      isDefaultShow: true,
    },
    {
      label: '机型类型',
      field: 'cvm.device_class',
    },
    {
      label: '单实例磁盘IO(MB/s)',
      field: 'cbs.disk_io',
      isDefaultShow: true,
    },
    {
      label: '云磁盘类型',
      field: 'cbs.disk_type_name',
      isDefaultShow: true,
    },
  ];

  // 单据管理 - 账号
  const accountColums = [
    {
      label: '单号',
      field: 'order_number',
      isDefaultShow: true,
    },
    {
      label: '资源类型',
      field: 'resource_type',
      isDefaultShow: true,
    },
    {
      label: '单据状态',
      field: 'document_status',
      isDefaultShow: true,
    },
    {
      label: '申请人',
      field: 'applicant',
      isDefaultShow: true,
    },
    {
      label: '申请时间',
      field: 'application_time',
    },
    {
      label: ' 结束时间',
      field: 'end_time',
    },
    {
      label: '备注',
      field: 'remarks',
    },
  ];

  const producingColumns = [
    {
      field: 'task_id',
      label: '任务ID',
      render: ({ data }: any) => {
        return (
          <Button
            theme='primary'
            text
            onClick={() => {
              window.open(data.task_link, '_blank');
            }}>
            {data.generate_id}
          </Button>
        );
      },
    },
    {
      field: 'message',
      label: '状态说明',
    },
    {
      field: 'start_at',
      label: '开始时间',
      render: ({ data }: any) => (data.status === -1 ? '-' : timeFormatter(data.start_at)),
    },
    {
      field: 'end_at',
      label: '结束时间',
      formatter: ({ data }: any) => (![0, 2].includes(data.status) ? '-' : timeFormatter(data.end_at)),
    },
  ];

  const initialColumns = [
    {
      field: 'ip',
      label: '内网 IP',
    },
    {
      field: 'status',
      label: '状态',
      width: 80,
      render: ({ data }: any) => {
        if (data.status === -1) return <span class='c-disabled'>未执行</span>;
        if (data.status === 0) return <span class='c-success'>成功</span>;
        if (data.status === 1)
          return (
            <span>
              <i class='el-icon-loading mr-2'></i>执行中
            </span>
          );
        return <span class='c-danger'>失败</span>;
      },
    },
    {
      field: 'message',
      label: '状态说明',
    },
    {
      field: 'task_id',
      label: '关联初始化单',
      render: ({ data }: any) => {
        return (
          <Button
            theme='primary'
            text
            onClick={() => {
              window.open(data.task_link, '_blank');
            }}>
            {data.task_id}
          </Button>
        );
      },
    },
    {
      field: 'start_at',
      label: '开始时间',
      render: ({ data }: any) => (data.status === -1 ? '-' : timeFormatter(data.start_at)),
    },
    {
      field: 'end_at',
      label: '结束时间',
      formatter: ({ data }: any) => (![0, 2].includes(data.status) ? '-' : timeFormatter(data.end_at)),
    },
  ];

  const deliveryColumns = [
    {
      field: 'ip',
      label: '内网 IP',
    },
    {
      field: 'asset_id',
      label: '固资号',
    },
    {
      field: 'status',
      label: '状态',
      width: 80,
      render: ({ data }: any) => {
        if (data.status === -1) return <span class='c-disabled'>未执行</span>;
        if (data.status === 0) return <span class='c-success'>成功</span>;
        if (data.status === 1)
          return (
            <span>
              <i class='el-icon-loading mr-2'></i>执行中
            </span>
          );
        return <span class='c-danger'>失败</span>;
      },
    },
    {
      field: 'message',
      label: '状态说明',
    },
    {
      field: 'deliverer',
      label: '匹配人',
      render: ({ data }: any) => <WName name={data.deliverer}></WName>,
    },
    {
      field: 'generate_task_id',
      label: '关联生产单',
      render: ({ data }: any) => {
        return (
          <Button
            theme='primary'
            text
            onClick={() => {
              window.open(data.generate_task_link, '_blank');
            }}>
            {data.generate_task_id}
          </Button>
        );
      },
    },
    {
      field: 'init_task_id',
      label: '关联初始化单',
      render: ({ data }: any) => {
        return (
          <Button
            theme='primary'
            text
            onClick={() => {
              window.open(data.init_task_link, '_blank');
            }}>
            {data.init_task_id}
          </Button>
        );
      },
    },
    {
      field: 'start_at',
      label: '开始时间',
      render: ({ data }: any) => (data.status === -1 ? '-' : timeFormatter(data.start_at)),
    },
    {
      field: 'end_at',
      label: '结束时间',
      formatter: ({ data }: any) => (![0, 2].includes(data.status) ? '-' : timeFormatter(data.end_at)),
    },
  ];

  const decommissionDetailsColumns = [
    {
      label: '固资号',
      field: 'server_asset_id',
      isDefaultShow: true,
    },
    {
      label: '内网IP',
      field: 'ip',
      isDefaultShow: true,
      render({ cell }: any) {
        return (
          <Button text theme='primary'>
            {cell}
          </Button>
        );
      },
    },
    {
      label: '公网IP',
      field: 'bk_host_outerip',
      isDefaultShow: true,
    },
    {
      label: '业务名称',
      field: 'app_name',
      isDefaultShow: true,
    },
    {
      label: '业务模块',
      field: 'module',
      isDefaultShow: true,
    },
    {
      label: 'SCM设备类型',
      field: 'device_type',
      isDefaultShow: true,
    },
    {
      label: '裁撤模块名称',
      field: 'module_name',
      isDefaultShow: true,
    },
    {
      label: '存放机房管理单元',
      field: 'idc_unit_name',
      isDefaultShow: true,
    },
    {
      label: '操作系统',
      field: 'sfw_name_version',
      isDefaultShow: true,
    },
    {
      label: '上架时间',
      field: 'go_up_date',
      isDefaultShow: true,
    },
    {
      label: 'RAID结构',
      field: 'raid_id',
      isDefaultShow: true,
    },
    {
      label: '逻辑区域',
      field: 'logic_area',
      isDefaultShow: true,
    },
    {
      label: '维护人',
      field: 'server_operator',
      isDefaultShow: true,
    },
    {
      label: '备份维护人',
      field: 'server_bak_operator',
    },
    {
      label: '设备技术分类',
      field: 'device_layer',
    },
    {
      label: 'CPU得分',
      field: 'cpu_score',
    },
    {
      label: '内存得分',
      field: 'mem_score',
    },
    {
      label: '内网流量得分',
      field: 'inner_net_traffic_score',
    },
    {
      label: '磁盘IO得分',
      field: 'disk_io_score',
    },
    {
      label: '磁盘IO使用率得分',
      field: 'disk_util_score',
    },
    {
      label: '是否达标',
      field: 'is_pass',
    },
    {
      label: '内存使用量(G)',
      field: 'mem4linux',
    },
    {
      label: '内网流量(Mb/s)',
      field: 'inner_net_traffic',
    },
    {
      label: '外网流量(Mb/s)',
      field: 'outer_net_traffic',
    },
    {
      label: '磁盘IO(Blocks/s)',
      field: 'disk_io',
    },
    {
      label: '磁盘IO使用率',
      field: 'disk_util',
    },
    {
      label: '磁盘总量(G)',
      field: 'disk_total',
    },
    {
      label: 'CPU核数',
      field: 'max_cpu_core_amount',
    },
    {
      label: '运维小组',
      field: 'group_name',
    },
    {
      label: '业务中心',
      field: 'center',
    },
  ];

  const firstAccountColumns = [
    {
      label: '一级帐号ID',
      field: 'primaryAccountId',
    },
    {
      label: '云厂商',
      field: 'cloudProvider',
    },
    {
      label: '帐号邮箱',
      field: 'accountEmail',
    },
    {
      label: '主负责人',
      field: 'mainResponsiblePerson',
    },
    {
      label: '组织架构',
      field: 'organizationalStructure',
    },
    {
      label: '二级帐号个数',
      field: 'secondaryAccountCount',
    },
  ];

  const secondaryAccountColumns = [
    {
      label: '二级帐号ID',
      field: 'secondaryAccountId',
    },
    {
      label: '所属一级帐号',
      field: 'parentPrimaryAccount',
    },
    {
      label: '云厂商',
      field: 'cloudProvider',
    },
    {
      label: '站点类型',
      field: 'siteType',
    },
    {
      label: '帐号邮箱',
      field: 'accountEmail',
    },
    {
      label: '主负责人',
      field: 'mainResponsiblePerson',
    },
    {
      label: '运营产品',
      field: 'operatingProduct',
    },
  ];

  const myApplyColumns = [
    // {
    //   label: '申请ID',
    //   field: 'id',
    // },
    // {
    //   label: '来源',
    //   field: 'source',
    // },
    {
      label: '申请类型',
      field: 'type',
    },
    {
      label: '单据状态',
      field: 'status',
      render({ data }: any) {
        let icon = StatusAbnormal;
        let txt = '审批拒绝';
        switch (data.status) {
          case 'pending':
          case 'delivering':
            icon = StatusLoading;
            txt = '审批中';
            break;
          case 'pass':
          case 'completed':
          case 'deliver_partial':
            icon = StatusSuccess;
            txt = '审批通过';
            break;
          case 'rejected':
          case 'cancelled':
          case 'deliver_error':
            icon = StatusFailure;
            txt = '审批拒绝';
            break;
        }
        return (
          <div class={'cvm-status-container'}>
            {txt === '审批中' ? (
              <Spinner fill='#3A84FF' class={'mr6'} width={14} height={14} />
            ) : (
              <img src={icon} class={'mr6'} width={14} height={14} />
            )}

            {txt}
          </div>
        );
      },
    },
    {
      label: '申请人',
      field: 'applicant',
    },
    {
      label: '创建时间',
      field: 'created_at',
      render({ cell }: any) {
        return timeFormatter(cell);
      },
    },
    {
      label: '更新时间',
      field: 'updated_at',
      render({ cell }: any) {
        return timeFormatter(cell);
      },
    },
    {
      label: '备注',
      field: 'memo',
      render({ cell }: any) {
        return cell || '--';
      },
    },
  ];

  const columnsMap = {
    vpc: vpcColumns,
    subnet: subnetColumns,
    group: groupColumns,
    gcp: gcpColumns,
    drive: driveColumns,
    image: imageColumns,
    networkInterface: networkInterfaceColumns,
    route: routeColumns,
    cvms: cvmsColumns,
    securityCommon: securityCommonColumns,
    eips: eipColumns,
    operationRecord: operationRecordColumns,
    lb: lbColumns,
    listener: listenerColumns,
    targetGroup: targetGroupColumns,
    rsConfig: rsConfigColumns,
    domain: domainColumns,
    url: urlColumns,
    targetGroupListener: targetGroupListenerColumns,
    cert: certColumns,
    hostInventor: hIColumns,
    CloudHost: CHColumns,
    cloudRequirementSubOrder: CRSOcolumns,
    physicalRequirementSubOrder: PRSOcolumns,
    PhysicalMachine: PMColumns,
    RecyclingResources: RRColumns,
    BusinessSelection: BSAColumns,
    ResourcesTotal: RTColumns,
    hostRecycle: recycleOrderColumns,
    deviceQuery: deviceQueryColumns,
    deviceDestroy: deviceDestroyColumns,
    DeviceQuerycolumns: DQcolumns,
    pdExecutecolumns: PDcolumns,
    ExecutionRecords: ERcolumns,
    scrResourceOnline: scrResourceOnlineColumns,
    scrResourceOffline: scrResourceOfflineColumns,
    forecastDemand: forecastDemandColumns,
    forecastDemandDetail: forecastDemandDetailColums,
    forecastList: forecastListColums,
    account: accountColums,
    CVMApplication: CAcolumns,
    scrResourceOnlineHost: scrResourceOnlineHostColumns,
    scrResourceOfflineHost: scrResourceOfflineHostColumns,
    scrResourceOnlineCreate: scrResourceOnlineCreateColumns,
    scrResourceOfflineCreate: scrResourceOfflineCreateColumns,
    cvmModel: cvmModelColumns,
    cvmWebQuery: cvmWebColumns,
    applicationList: ApplicationListColumns,
    firstAccount: firstAccountColumns,
    secondaryAccount: secondaryAccountColumns,
    myApply: myApplyColumns,
    scrProduction: producingColumns,
    scrInitial: initialColumns,
    scrDelivery: deliveryColumns,
    decommissionDetails: decommissionDetailsColumns,
    firstAccount: firstAccountColumns,
    secondaryAccount: secondaryAccountColumns,
    myApply: myApplyColumns,
  };

  let columns = (columnsMap[type] || []).filter((column: any) => !isSimpleShow || !column.onlyShowOnList);
  if (whereAmI.value !== Senarios.resource) columns = columns.filter((column: any) => !column.isOnlyShowInResource);

  type ColumnsType = typeof columns;
  const generateColumnsSettings = (columns: ColumnsType) => {
    let fields = [];
    for (const column of columns) {
      if (column.field && column.label) {
        fields.push({
          label: column.label,
          field: column.field,
          disabled: type !== 'cvms' && column.field === 'id',
          isDefaultShow: !!column.isDefaultShow,
          isOnlyShowInResource: !!column.isOnlyShowInResource,
        });
      }
    }
    if (whereAmI.value !== Senarios.resource) {
      fields = fields.filter((field) => !field.isOnlyShowInResource);
    }
    const settings: Ref<Settings> = ref({
      fields,
      checked: fields.filter((field) => field.isDefaultShow).map((field) => field.field),
    });

    return settings;
  };

  const settings = generateColumnsSettings(columns);

  return {
    columns,
    settings,
    generateColumnsSettings,
  };
};
