/* eslint-disable no-nested-ternary */
// table 字段相关信息
import { useAccountStore } from '@/store';
import { Button } from 'bkui-vue';
import i18n from '@/language/i18n';
import { type Settings } from 'bkui-vue/lib/table/props';
import { ref } from 'vue';
import type { Ref } from 'vue';
import { CloudType } from '@/typings';
import { RouteLocationRaw, useRoute, useRouter } from 'vue-router';
import { CLOUD_HOST_STATUS, VendorEnum } from '@/common/constant';
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
import cssModule from './index.module.scss';
import { defaults } from 'lodash';
import { timeFormatter } from '@/common/util';
import { capacityLevel } from '@/utils/scr';
import { getResourceTypeName, getReturnPlanName } from '@/utils';
import {
  getRecycleTaskStatusLabel,
  getBusinessNameById,
  dateTimeTransform,
  getPrecheckStatusLabel,
} from '@/views/ziyanScr/host-recycle/field-dictionary';
import { getRegionCn, getZoneCn } from '@/views/ziyanScr/cvm-web/transform';
import { getCvmProduceStatus, getTypeCn } from '@/views/ziyanScr/cvm-produce/transform';
import { getDiskTypesName, getImageName } from '@/components/property-list/transform';
import { useApplyStages } from '@/views/ziyanScr/hooks/use-apply-stages';
import { transformAntiAffinityLevels } from '@/views/ziyanScr/hostApplication/components/transform';
import { Spinner, Share } from 'bkui-vue/lib/icon';
import WName from '@/components/w-name';
import { SCR_POOL_PHASE_MAP, SCR_RECALL_DETAIL_STATUS_MAP } from '@/constants';
import CopyToClipboard from '@/components/copy-to-clipboard/index.vue';

interface LinkFieldOptions {
  type: string; // 资源类型
  label?: string; // 显示文本
  field?: string; // 字段
  idFiled?: string; // id字段
  onlyShowOnList?: boolean; // 只在列表中显示
  onLinkInBusiness?: boolean; // 只在业务下可链接
  render?: (data: any) => any; // 自定义渲染内容
  renderSuffix?: (data: any) => any; // 自定义后缀渲染内容
  contentClass?: string; // 内容class
  sort?: boolean; // 是否支持排序
}

export default (type: string, isSimpleShow = false) => {
  const router = useRouter();
  const route = useRoute();
  const { t } = i18n.global;
  const accountStore = useAccountStore();
  const { getRegionName } = useRegionsStore();
  const { whereAmI } = useWhereAmI();
  const businessMapStore = useBusinessMapStore();
  const cloudAreaStore = useCloudAreaStore();
  const { transformApplyStages } = useApplyStages();
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

    const { type, label, field, idFiled, onlyShowOnList, onLinkInBusiness, render, renderSuffix, contentClass, sort } =
      options;

    return {
      label,
      field,
      sort,
      width: label === 'ID' ? '120' : 'auto',
      onlyShowOnList,
      isDefaultShow: true,
      render({ data }: { cell: string; data: any }) {
        if (data[idFiled] < 0 || !data[idFiled]) return '--';
        const onlyLinkInBusiness = onLinkInBusiness && whereAmI.value !== Senarios.business;
        const isZiyan = data.vendor === VendorEnum.ZIYAN;
        if (onlyLinkInBusiness || isZiyan)
          return (
            <div class={cssModule[`${contentClass}`]}>
              {render ? render(data) : data[field] || '--'}
              {renderSuffix?.(data)}
            </div>
          );
        return (
          <div class={cssModule[`${contentClass}`]}>
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
            {renderSuffix?.(data)}
          </div>
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

  const hIColumns = [
    {
      label: '需求类型',
      field: 'require_type',
      render: ({ row }: any) => getTypeCn(row.require_type),
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
      render: ({ row }: any) => getRegionCn(row.region),
    },
    {
      label: '园区',
      field: 'zone',
      render: ({ row }: any) => getZoneCn(row.zone),
    },
    {
      label: '库存情况',
      field: 'capacity_flag',
      sort: { value: 'desc' },
      render({ cell }: { cell: string }) {
        const { class: theClass, text } = capacityLevel(cell);
        return <span class={cssModule[`${theClass}`]}>{text}</span>;
      },
    },
  ];
  const CRSOcolumns = [
    { type: 'selection', width: 30, minWidth: 30, isDefaultShow: true },
    {
      label: '机型',
      field: 'spec.device_type',
      width: 140,
    },
    {
      label: '状态',
      field: 'stage',
      render: ({ row }: any) => transformApplyStages(row.stage),
    },
    {
      label: '总数',
      field: 'total_num',
      width: 70,
      minWidth: 70,
    },
    {
      label: '待交付',
      field: 'pending_num',
      width: 70,
      minWidth: 70,
    },
    {
      label: '地域',
      field: 'spec.region',
      width: 160,
      render: ({ row }: any) => getRegionCn(row.spec.region),
    },
    {
      label: '园区',
      field: 'spec.zone',
      width: 160,
      render: ({ row }: any) => getZoneCn(row.spec.zone),
    },
    {
      label: '反亲和性',
      field: 'anti_affinity_level',
      width: 90,
      render: ({ row }: any) => transformAntiAffinityLevels(row.anti_affinity_level),
    },
    {
      label: '镜像',
      field: 'spec.image_id',
      render: ({ row }: any) => getImageName(row.spec.image_id),
    },
    {
      label: 'VPC',
      field: 'spec.vpc',
    },
    {
      label: '子网',
      field: 'spec.subnet',
    },
    {
      label: '数据盘大小',
      field: 'spec.disk_size',
      width: 100,
    },
    {
      label: '数据盘类型',
      field: 'spec.disk_type',
      width: 100,
      render: ({ row }: any) => getDiskTypesName(row.spec.disk_type),
    },
    {
      label: '备注',
      field: 'remark',
      render: ({ cell }: { cell: string }) => cell || '--',
    },
  ];
  const PRSOcolumns = [
    {
      label: '地域',
      field: 'spec.region',
      render: ({ row }: any) => getRegionCn(row.spec.region),
    },
    {
      label: '园区',
      field: 'spec.zone',
      render: ({ row }: any) => getZoneCn(row.spec.zone),
    },
    {
      label: '反亲和性',
      field: 'anti_affinity_level',
      width: 90,
      render: ({ row }: any) => transformAntiAffinityLevels(row.anti_affinity_level),
    },
    {
      label: '操作系统',
      field: 'spec.os_type',
    },
    {
      label: '数据盘大小',
      field: 'spec.disk_size',
      width: 100,
    },
    {
      label: 'RAID类型',
      field: 'spec.raid_type',
      width: 100,
    },
    {
      label: '备注',
      field: 'remark',
      render: ({ cell }: { cell: string }) => cell || '--',
    },
    {
      label: '状态',
      field: 'stage',
      render: ({ row }: any) => transformApplyStages(row.stage),
    },
  ];
  const CHColumns = [
    {
      label: '机型',
      field: 'spec.device_type',
      width: 150,
    },
    {
      label: '需求数量',
      field: 'replicas',
      width: 90,
    },
    {
      label: '地域',
      field: 'spec.region',
      width: 150,
      render: ({ cell }: { cell: string }) => getRegionName(VendorEnum.TCLOUD, cell) || '--',
    },
    {
      label: '园区',
      field: 'spec.zone',
      width: 150,
      render: ({ row }: any) => getZoneCn(row.spec.zone),
    },
    {
      label: '镜像',
      field: 'spec.image_id',
      render: ({ row }: any) => getImageName(row.spec.image_id),
    },
    {
      label: '数据盘大小',
      field: 'spec.disk_size',
      width: 95,
    },
    {
      label: '数据盘类型',
      field: 'spec.disk_type',
      width: 95,
      render: ({ row }: any) => getDiskTypesName(row.spec.disk_type),
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
      width: 90,
    },
    {
      label: '地域',
      field: 'spec.region',
      width: 150,
      render: ({ cell }: { cell: string }) => getRegionName(VendorEnum.TCLOUD, cell) || '--',
    },
    {
      label: '园区',
      field: 'spec.zone',
      width: 150,
    },
    {
      label: 'RAID 类型',
      field: 'spec.raid_type',
      width: 110,
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
    { type: 'selection', width: 30, minWidth: 30, isDefaultShow: true },
    {
      label: '状态',
      field: 'recyclable',
      render: ({ cell, data }: any) => (
        <span
          class={cssModule[cell ? 'c-success' : 'c-danger']}
          v-bk-tooltips={{ content: data.message, disabled: cell, placement: 'right', theme: 'light' }}>
          {cell ? t('可回收') : t('不可回收')}
        </span>
      ),
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
    { type: 'selection', width: 30, minWidth: 30, isDefaultShow: true },
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
    { type: 'selection', width: 30, minWidth: 30, isDefaultShow: true },
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
      render: ({ data, cell }: any) => {
        return (
          <Button
            text
            theme='primary'
            onClick={() => {
              router.push({
                name: 'host-application-detail',
                params: {
                  id: data.order_id,
                },
              });
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
      render: ({ row }: any) => getTypeCn(row.require_type),
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
      field: 'zone_name',
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
      render: ({ row }: any) => getTypeCn(row.require_type),
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
      render: ({ cell }: { cell: string }) => getRegionCn(cell) || '--',
    },
    {
      label: '园区',
      field: 'zone',
      render: ({ row }: any) => getZoneCn(row.zone),
    },
    {
      label: '库存情况',
      field: 'capacity_flag',
      sort: {
        value: 'desc',
      },
      render({ cell }: { cell: string }) {
        const { class: theClass, text } = capacityLevel(cell);
        return <span class={cssModule[`${theClass}`]}>{text}</span>;
      },
    },
  ];
  // 预检详情状态 render
  const getPrecheckStatusView = (value: string) => {
    const label = getPrecheckStatusLabel(value);
    if (value === 'SUCCESS') {
      return <span class={cssModule['c-success']}>{label}</span>;
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
          <div class={[cssModule['c-danger'], cssModule['fail-flex']]}>
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
      width: 300,
    },
    {
      label: '状态',
      field: 'status',
      width: 80,
      render: ({ row }: any) => {
        return getPrecheckStatusView(row.status);
      },
    },
    {
      label: '状态描述',
      field: 'message',
    },
    {
      label: '开始时间',
      field: 'create_at',
      width: 180,
      render: ({ row }: any) => {
        return <span>{dateTimeTransform(row.create_at)}</span>;
      },
    },
    {
      label: '结束时间',
      field: 'end_at',
      width: 180,
      render: ({ row }: any) => {
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
      render: ({ row }: any) => {
        return getPrecheckStatusView(row.status);
      },
      exportFormatter: (row: any) => getPrecheckStatusLabel(row.status),
    },
    {
      label: '已执行/总数',
      field: 'mem',
      render: ({ row }: any) => {
        return (
          <div>
            <span class={cssModule[row.success_num > 0 ? 'c-success' : '']}>{row.success_num}</span>
            <span>/</span>
            <span>{row.total_num}</span>
          </div>
        );
      },
      exportFormatter: (row: any) => {
        return `${row.success_num}/${row.total_num}`;
      },
    },
    {
      label: '更新时间',
      field: 'update_at',
      render: ({ row }: any) => {
        return <span>{dateTimeTransform(row.update_at)}</span>;
      },
      formatter: ({ update_at }: any) => {
        return dateTimeTransform(update_at);
      },
    },
    {
      label: '创建时间',
      field: 'create_at',
      render: ({ row }: any) => {
        return <span>{dateTimeTransform(row.create_at)}</span>;
      },
      formatter: ({ create_at }: any) => {
        return dateTimeTransform(create_at);
      },
    },
  ];
  const getRecycleTaskStatusView = (value: string) => {
    const label = getRecycleTaskStatusLabel(value);
    if (value === 'DONE') {
      return (
        <>
          <span class={cssModule['c-success']}>{label}</span>
        </>
      );
    }
    if (value.includes('ING')) {
      return (
        <>
          <span>{label}</span>
          <Spinner />
        </>
      );
    }
    if (value === 'DETECT_FAILED') {
      return (
        <bk-badge
          class={cssModule['c-danger']}
          v-bk-tooltips={{ content: '请到“预检详情”查看失败原因，或者点击“去除预检失败IP提交”' }}
          dot>
          {label}
        </bk-badge>
      );
    }
    if (value.includes('FAILED')) {
      return <span class={cssModule['c-danger']}>{label}</span>;
    }
    return <span>{label}</span>;
  };
  // 资源 - 主机回收列表
  const recycleOrderColumns = [
    { type: 'selection', width: 30, minWidth: 30, isDefaultShow: true },
    {
      label: '业务',
      field: 'bk_biz_id',
      render: ({ row }: any) => {
        return getBusinessNameById(row.bk_biz_id);
      },
      formatter: ({ bk_biz_id }: any) => {
        return getBusinessNameById(bk_biz_id);
      },
    },
    {
      label: '资源类型',
      field: 'resource_type',
      width: 120,
      render: ({ row }: any) => {
        return <span>{getResourceTypeName(row.resource_type)}</span>;
      },
      formatter: ({ resource_type }: any) => {
        return getResourceTypeName(resource_type);
      },
    },
    {
      label: '回收类型',
      field: 'return_plan',
      render: ({ row }: any) => {
        return <span>{getReturnPlanName(row.return_plan, row.resource_type)}</span>;
      },
      formatter: ({ return_plan, resource_type }: any) => {
        return getReturnPlanName(return_plan, resource_type);
      },
    },
    {
      label: '回收成本',
      field: 'cost_concerned',
      render: ({ row }: any) => {
        return <span>{row.cost_concerned ? '涉及' : '不涉及'}</span>;
      },
      formatter: ({ cost_concerned }: any) => {
        return cost_concerned ? '涉及' : '不涉及';
      },
    },
    {
      label: '状态',
      field: 'status',
      width: 100,
      render: ({ row }: any) => {
        return getRecycleTaskStatusView(row.status);
      },
      exportFormatter: (row: any) => getRecycleTaskStatusLabel(row.status),
    },
    {
      label: '当前处理人',
      field: 'handler',
      width: 100,
      render: ({ row }: any) => {
        return row.handler !== 'AUTO' ? (
          <Button
            text
            theme='primary'
            onClick={() => {
              window.open(`wxwork://message?username=${row.handler}`);
            }}>
            {row.handler}
          </Button>
        ) : (
          <span class={cssModule['cell-font-color']}>{row.handler}</span>
        );
      },
    },
    {
      label: '总数/成功/失败',
      width: 120,
      render: ({ row }: any) => {
        return (
          <div>
            <span>{row.total_num}</span>
            <span>/</span>
            <span class={cssModule[row.success_num > 0 ? 'c-success' : '']}>{row.success_num}</span>
            <span>/</span>
            <span class={cssModule[row.failed_num > 0 ? 'c-danger' : '']}>{row.failed_num}</span>
          </div>
        );
      },
      exportFormatter: (row: any) => {
        return `${row.success_num}/${row.failed_num}/${row.total_num}`;
      },
    },
    {
      label: '回收人',
      field: 'bk_username',
      render: ({ row }: any) => {
        return (
          <Button
            text
            onClick={() => {
              window.open(`wxwork://message?username=${row.bk_username}`);
            }}>
            {row.bk_username}
          </Button>
        );
      },
    },
    {
      label: '回收时间',
      field: 'create_at',
      render: ({ row }: any) => {
        return <span>{dateTimeTransform(row.create_at)}</span>;
      },
      formatter: ({ create_at }: any) => {
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
      formatter: ({ bk_biz_id }: any) => {
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
      render: ({ row }: any) => {
        return row.return_cost_rate ? `${Math.ceil(row.return_cost_rate * 100)}%` : '-';
      },
    },
    {
      label: '状态',
      field: 'status',
      render: ({ row }: any) => getRecycleTaskStatusView(row.status),
      exportFormatter: (row: any) => getRecycleTaskStatusLabel(row.status),
    },
    {
      label: '回收人',
      field: 'bk_username',
      render: ({ row }: any) => {
        return (
          <Button
            text
            theme='primary'
            onClick={() => {
              window.open(`wxwork://message?username=${row.bk_username}`);
            }}>
            {row.bk_username}
          </Button>
        );
      },
    },
    {
      label: '创建时间',
      field: 'create_at',
      render: ({ row }: any) => {
        return <span>{dateTimeTransform(row.create_at)}</span>;
      },
      formatter: ({ create_at }: any) => {
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
    { type: 'selection', width: 30, minWidth: 30, isDefaultShow: true },
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
      render: ({ row }: any) => {
        return (
          <Button
            text
            theme='primary'
            onClick={() => {
              window.open(`wxwork://message?username=${row.operator}`);
            }}>
            {row.operator}
          </Button>
        );
      },
    },
    {
      label: '备份维护人',
      field: 'bak_operator',
      render: ({ row }: any) => {
        return (
          <Button
            text
            theme='primary'
            onClick={() => {
              window.open(`wxwork://message?username=${row.bak_operator}`);
            }}>
            {row.bak_operator}
          </Button>
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
      render: ({ row }: any) => {
        return row.return_cost_rate ? `${Math.ceil(row.return_cost_rate * 100)}%` : '-';
      },
    },
    {
      label: '校验结果',
      field: 'return_plan_msg',
      showOverflowTooltip: true,
      render: ({ row }: any) => {
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
      render: ({ row }: any) => {
        return <span>{dateTimeTransform(row.input_time)}</span>;
      },
      formatter: ({ input_time }: any) => {
        return dateTimeTransform(input_time);
      },
    },
    {
      label: '销毁时间',
      field: 'return_time',
      render: ({ row }: any) => {
        return <span>{dateTimeTransform(row.return_time)}</span>;
      },
      formatter: ({ return_time }: any) => {
        return dateTimeTransform(return_time);
      },
    },
    {
      label: '回收单号',
      field: 'return_id',
      render: ({ row }: any) => {
        return (
          <Button
            text
            theme='primary'
            onClick={() => {
              window.open(row.return_link);
            }}>
            {row.return_id}
          </Button>
        );
      },
    },
    {
      label: '状态',
      field: 'status',
      render: ({ row }: any) => getRecycleTaskStatusView(row.status),
      exportFormatter: (row: any) => getRecycleTaskStatusLabel(row.status),
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

        if (phase === 'INIT') return <span class={cssModule['c-info']}>{desc}</span>;
        if (phase === 'RUNNING')
          return (
            <span class={cssModule['status-column-cell']}>
              <img class={[cssModule['status-icon'], cssModule['spin-icon']]} src={StatusLoading} alt='' />
              {desc}
            </span>
          );
        if (phase === 'SUCCESS') return <span class={cssModule['c-success']}>{desc}</span>;
        if (phase === 'FAILED') return <span class={cssModule['c-danger']}>{desc}</span>;

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

        if (phase === 'INIT') return <span class={cssModule['c-info']}>{desc}</span>;
        if (phase === 'RUNNING')
          return (
            <span class={cssModule['status-column-cell']}>
              <img class={[cssModule['status-icon'], cssModule['spin-icon']]} src={StatusLoading} alt='' />
              {desc}
            </span>
          );
        if (phase === 'SUCCESS') return <span class={cssModule['c-success']}>{desc}</span>;
        if (phase === 'FAILED') return <span class={cssModule['c-danger']}>{desc}</span>;

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

        if (cell === 'INIT') return <span class={cssModule['c-info']}>{desc}</span>;
        if (cell === 'RUNNING')
          return (
            <span class={cssModule['status-column-cell']}>
              <img class={[cssModule['status-icon'], cssModule['spin-icon']]} src={StatusLoading} alt='' />
              {desc}
            </span>
          );
        if (cell === 'SUCCESS') return <span class={cssModule['c-success']}>{desc}</span>;
        if (cell === 'FAILED') return <span class={cssModule['c-danger']}>{desc}</span>;

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
        <Button
          text
          theme='primary'
          onClick={() => {
            window.open(data.reinstall_link, '_blank');
          }}>
          {data.reinstall_id}
        </Button>
      ),
      exportFormatter: ({ reinstall_id }: any) => reinstall_id,
    },
    {
      label: '配置检查任务',
      field: 'conf_check_link',
      render: ({ data }: any) => (
        <Button
          text
          theme='primary'
          onClick={() => {
            window.open(data.conf_check_link, '_blank');
          }}>
          {data.conf_check_id}
        </Button>
      ),
      exportFormatter: ({ conf_check_id }: any) => conf_check_id,
    },
    {
      label: '状态',
      field: 'status',
      render: ({ cell }: any) => {
        const desc = SCR_RECALL_DETAIL_STATUS_MAP[cell];

        if (cell === 'TERMINATE') return <span class={cssModule['c-info']}>{desc}</span>;
        if (cell === 'REINSTALLING') {
          return (
            <span class={cssModule['status-column-cell']}>
              <img class={[cssModule['status-icon'], cssModule['spin-icon']]} src={StatusLoading} alt='' />
              {desc}
            </span>
          );
        }
        if (cell === 'RETURNED' || cell === 'DONE') return <span class={cssModule['c-success']}>{desc}</span>;
        if (cell === 'REINSTALL_FAILED') return <span class={cssModule['c-danger']}>{desc}</span>;

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
    { type: 'selection', width: 30, minWidth: 30, isDefaultShow: true },
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
    { type: 'selection', width: 30, minWidth: 30, isDefaultShow: true },
    {
      label: 'VPC',
      field: 'vpc_name',
      render: ({ row }: any) => {
        return (
          <div class={cssModule['cvm-cell-height']}>
            <div>{row.vpc_name}</div>
            <div>{row.vpc_id}</div>
          </div>
        );
      },
    },
    {
      label: 'Subnet',
      field: 'subnet_name',
      render: ({ row }: any) => {
        return (
          <div class={cssModule['cvm-cell-height']}>
            <div>{row.subnet_name}</div>
            <div>{row.subnet_id}</div>
          </div>
        );
      },
    },
    {
      label: '地域',
      field: 'region',
      render: ({ row }: any) => {
        return (
          <div class={cssModule['cvm-cell-height']}>
            <div> {getRegionCn(row.region)}</div>
            <div>{row.region}</div>
          </div>
        );
      },
    },
    {
      label: '园区',
      field: 'zone',
      render: ({ row }: any) => {
        return (
          <div class={cssModule['cvm-cell-height']}>
            <div> {getZoneCn(row.zone)}</div>
            <div>{row.zone}</div>
          </div>
        );
      },
    },
  ];

  const ApplicationListColumns = [
    {
      label: '申请时间',
      field: 'create_at',
      width: 120,
      render: ({ data }: any) => timeFormatter(data.create_at, 'YYYY-MM-DD'),
    },
    {
      label: '期望交付时间',
      field: 'expect_time',
      width: 120,
      render: ({ data }: any) => timeFormatter(data.expect_time, 'YYYY-MM-DD'),
    },
    {
      label: '备注信息',
      field: 'remark',
      width: 300,
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
    { type: 'selection', width: 30, minWidth: 30, isDefaultShow: true },
    {
      label: '机型',
      field: 'device_type',
      width: 200,
    },
    {
      label: '需求类型',
      field: 'require_type',
      render: ({ row }: any) => getTypeCn(row.require_type),
    },
    {
      label: '地域',
      field: 'region',
      render: ({ row }: any) => getRegionCn(row.region),
    },
    {
      label: '园区',
      field: 'zone',
      render: ({ row }: any) => getZoneCn(row.zone),
    },
    {
      label: '实例族',
      field: 'label.device_group',
    },
    {
      label: 'CPU(核)',
      field: 'cpu',
      width: 50,
    },
    {
      label: '内存(G)',
      field: 'mem',
      width: 50,
    },
    {
      label: '其他信息',
      field: 'remark',
      render: ({ cell }: any) => cell || '--',
    },
    {
      label: '可查询容量',
      field: 'enable_capacity',
      render: ({ cell }: any) => (cell ? '是' : '否'),
    },
    {
      label: '可申请',
      field: 'enable_apply',
      render: ({ cell }: any) => (cell ? <span style={'color:#67c23a'}>是</span> : <span>否</span>),
    },
    {
      label: '推荐分数',
      field: 'score',
    },
    {
      label: '备注',
      field: 'comment',
      render: ({ cell }: any) => cell || '--',
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
          <div class={cssModule['cvm-status-container']}>
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
      field: 'status',
      label: '状态',
      render: ({ row }: any) => {
        if (row.status === -1) return <span class={cssModule['c-disabled']}>未执行</span>;
        if (row.status === 0) return <span class={cssModule['c-success']}>成功</span>;
        if (row.status === 1)
          return (
            <span>
              <Spinner />
              执行中
            </span>
          );
        return <span class={cssModule['c-danger']}>失败</span>;
      },
    },
    {
      label: '成功台数/总台数',
      render: ({ row }: any) => {
        return (
          <div>
            <span class={cssModule['c-success']}>{row.success_num}</span>
            <span>/</span>
            <span>{row.total_num}</span>
          </div>
        );
      },
    },
    {
      field: 'message',
      label: '状态说明',
      showOverflowTooltip: true,
    },
    {
      field: 'start_at',
      label: '开始时间',
      render: ({ data }: any) => (data.status === -1 ? '-' : timeFormatter(data.start_at)),
    },
    {
      field: 'end_at',
      label: '结束时间',
      render: ({ data }: any) => (![0, 2].includes(data.status) ? '-' : timeFormatter(data.end_at)),
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
      render: ({ data }: any) => {
        if (data.status === -1) return <span class={cssModule['c-disabled']}>未执行</span>;
        if (data.status === 0) return <span class={cssModule['c-success']}>成功</span>;
        if (data.status === 1)
          return (
            <span>
              <i class='el-icon-loading mr-2'></i>执行中
            </span>
          );
        return <span class={cssModule['c-danger']}>失败</span>;
      },
    },
    {
      field: 'message',
      label: '状态说明',
      showOverflowTooltip: true,
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
      render: ({ data }: any) => (![0, 2].includes(data.status) ? '-' : timeFormatter(data.end_at)),
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
      render: ({ data }: any) => {
        if (data.status === -1) return <span class={cssModule['c-disabled']}>未执行</span>;
        if (data.status === 0) return <span class={cssModule['c-success']}>成功</span>;
        if (data.status === 1)
          return (
            <span>
              <i class='el-icon-loading mr-2'></i>执行中
            </span>
          );
        return <span class={cssModule['c-danger']}>失败</span>;
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
      render: ({ data }: any) => (![0, 2].includes(data.status) ? '-' : timeFormatter(data.end_at)),
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
          <Button
            text
            theme='primary'
            onClick={() => {
              window.open(`https://tmp.woa.com/host/home/${cell}`, '_blank');
            }}>
            {cell}
          </Button>
        );
      },
    },
    {
      label: '公网IP',
      field: 'outer_ip',
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
      field: 'raid_name',
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

  // CVM虚拟机 - CVM生产
  const cvmProduceColumns = [
    {
      type: 'expand',
    },
    {
      label: '单号',
      field: 'order_id',
      width: 60,
    },
    {
      label: '云梯单号',
      field: 'task_id',
      showOverflowTooltip: () => ({
        theme: 'light',
      }),
      render: ({ row }: any) => {
        if (row.task_id)
          return (
            <Button
              text
              theme='primary'
              onClick={() => {
                window.open(row.task_link, '_blank');
              }}>
              {row.task_id}
            </Button>
          );
        return '-';
      },
    },
    {
      label: '需求类型',
      field: 'require_type',
      width: 100,
      render: ({ row }: any) => getTypeCn(row.require_type),
    },
    {
      label: '状态',
      field: 'status',
      width: 80,
      render: ({ row }: any) => {
        const desc = getCvmProduceStatus(row.status);

        if (row.status === 'INIT') return <span class={cssModule['c-info']}>{desc}</span>;
        if (row.status === 'RUNNING')
          return (
            <span>
              <i class='el-icon-loading mr-2'></i>
              {desc}
            </span>
          );
        if (row.status === 'SUCCESS') return <span class={cssModule['c-success']}>{desc}</span>;
        if (row.status === 'FAILED') return <span class={cssModule['c-danger']}>{desc}</span>;

        return row.status;
      },
    },
    {
      label: '状态描述',
      field: 'message',
      showOverflowTooltip: true,
    },
    {
      label: '地域',
      field: 'spec.region',
      render: ({ row }: any) => getRegionCn(row.spec.region),
    },
    {
      label: '园区',
      field: 'spec.zone',
      render: ({ row }: any) => getZoneCn(row.spec.zone),
    },
    {
      label: '机型',
      field: 'spec.device_type',
    },
    {
      label: '生产情况-待交付',
      field: 'pending_num',
      width: 150,
    },
    {
      label: '生产情况-失败',
      field: 'failed_num',
      width: 150,
      render: ({ row }: any) => <span class={cssModule['c-danger']}>{row.failed_num}</span>,
    },
    {
      label: '生产情况-总数',
      field: 'total_num',
      width: 150,
    },
    {
      label: '创建人',
      field: 'bk_username',
      width: 100,
      showOverflowTooltip: true,
    },
    {
      label: '创建时间',
      field: 'create_at',
      sort: {
        value: 'desc',
      },
      width: 99,
      render: ({ row }: any) => dateTimeTransform(row.create_at),
    },
    {
      label: '结束时间',
      field: 'update_at',
      width: 90,
      render: ({ row }: any) => dateTimeTransform(row.update_at),
    },
    {
      label: '备注',
      field: 'remark',
      showOverflowTooltip: true,
    },
  ];

  // CVM虚拟机 - CVM生产 - 快速生产
  const cvmFastProduceColumns = [
    {
      field: 'require_type',
      label: '需求类型',
      width: 100,
      render: ({ row }: any) => getTypeCn(row.require_type),
    },
    {
      field: 'label.device_group',
      label: '实例族',
      width: 120,
    },
    {
      field: 'device_type',
      label: '机型',
    },
    {
      field: 'cpu',
      label: 'CPU(核)',
      sort: true,
      width: 100,
    },
    {
      field: 'mem',
      label: '内存(G)',
      sort: true,
      width: 100,
    },
    {
      field: 'region',
      label: '地域',
      render: ({ row }: any) => getRegionCn(row.region),
    },
    {
      field: 'zone',
      label: '园区',
      render: ({ row }: any) => getZoneCn(row.zone),
    },
    {
      field: 'capacity_flag',
      label: '库存情况',
      width: 140,
      sort: { value: 'desc' },
      render: ({ row }: any) => {
        const { class: theClass, text } = capacityLevel(row.capacity_flag);
        return <span class={cssModule[`${theClass}`]}>{text}</span>;
      },
    },
  ];

  // CVM虚拟机 - CVM生产 - 详情
  const cvmProduceDetailColumns = [
    {
      label: '固资号',
      field: 'asset_id',
    },
    {
      label: '内网 IP',
      field: 'ip',
    },
    {
      label: '实例 ID',
      field: 'cvm_inst_id',
    },
    {
      label: '生产时间',
      field: 'update_at',
      render: ({ row }: any) => dateTimeTransform(row.update_at),
    },
  ];

  const billsRootAccountSummaryColumns = [
    {
      label: '一级账号ID',
      field: 'root_account_id',
    },
    {
      label: '一级账号名称',
      field: 'root_account_name',
    },
    {
      label: '账号状态',
      field: 'state',
    },
    {
      label: '账单同步（人民币-元）当月',
      field: 'current_month_rmb_cost_synced',
    },
    {
      label: '账单同步（人民币-元）上月',
      field: 'last_month_rmb_cost_synced',
    },
    {
      label: '账单同步（美金-美元）当月',
      field: 'current_month_cost_synced',
    },
    {
      label: '账单同步（美金-美元）上月',
      field: 'last_month_cost_synced',
    },
    {
      label: '账单同步环比',
      field: 'month_on_month_value',
    },
    {
      label: '当前账单人民币（元）',
      field: 'current_month_rmb_cost',
    },
    {
      label: '当前账单美金（美元）',
      field: 'current_month_cost',
    },
    {
      label: '调账人民币（元）',
      field: 'adjustment_cost',
    },
    {
      label: '调账美金（美元）',
      field: 'adjustment_cost',
    },
  ];

  const billsMainAccountSummaryColumns = [
    {
      label: '二级账号ID',
      field: 'main_account_id',
    },
    {
      label: '二级账号名称',
      field: 'main_account_name',
    },
    {
      label: '运营产品',
      field: 'product_name',
    },
    {
      label: '已确认账单人民币（元）',
      field: 'current_month_rmb_cost_synced',
    },
    {
      label: '已确认账单美金（美元）',
      field: 'current_month_cost_synced',
    },
    {
      label: '当前账单人民币（元）',
      field: 'current_month_rmb_cost',
    },
    {
      label: '当前账单美金（美元）',
      field: 'current_month_cost',
    },
  ];

  const billsSummaryOperationRecordColumns = [
    {
      label: '操作时间',
      field: 'operationTime',
    },
    {
      label: '状态',
      field: 'status',
    },
    {
      label: '账单月份',
      field: 'billingMonth',
    },
    {
      label: '云厂商',
      field: 'cloudVendor',
    },
    {
      label: '一级账号ID',
      field: 'primaryAccountId',
    },
    {
      label: '操作人',
      field: 'operator',
    },
    {
      label: '人民币（元）',
      field: 'rmbAmount',
    },
    {
      label: '美金（美元）',
      field: 'usdAmount',
    },
  ];

  const businessHostColumns = [
    { type: 'selection', width: 30, minWidth: 30, onlyShowOnList: true },
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
      renderSuffix: (data) => {
        const ips = [...(data.private_ipv4_addresses || []), ...(data.private_ipv6_addresses || [])].join(',') || '--';
        return <CopyToClipboard content={ips} class={[cssModule['copy-icon'], 'ml4']} />;
      },
      contentClass: 'cell-private-ip',
      sort: false,
    }),
    {
      label: '公网IP',
      field: 'public_ipv4_addresses',
      isDefaultShow: true,
      onlyShowOnList: true,
      render: ({ data }: any) => {
        const ips = [...(data.public_ipv4_addresses || []), ...(data.public_ipv6_addresses || [])].join(',') || '--';
        return (
          <div class={'cell-public-ip'}>
            <span>{ips}</span>
            <CopyToClipboard content={ips} class={[cssModule['copy-icon'], 'ml4']} />
          </div>
        );
      },
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
      render: ({ data }: any) => <span>{CloudType[data.vendor]}</span>,
    },
    {
      label: '地域',
      onlyShowOnList: true,
      field: 'region',
      sort: true,
      isDefaultShow: true,
      render: ({ cell, row }: { cell: string; row: { vendor: VendorEnum } }) => getRegionName(row.vendor, cell) || '--',
    },
    {
      label: '主机名称',
      field: 'name',
      sort: true,
      isDefaultShow: true,
      render: ({ cell }: { cell: string }) => cell || '--',
    },
    {
      label: '主机状态',
      field: 'status',
      sort: true,
      isDefaultShow: true,
      render({ data }: any) {
        return (
          <div class={cssModule['cvm-status-container']}>
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
            <span>{CLOUD_HOST_STATUS[data.status] || data.status || t('未获取')}</span>
          </div>
        );
      },
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
      label: '操作系统',
      field: 'os_name',
      render: ({ data }: any) => <span>{data.os_name || '--'}</span>,
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

  const columnsMap = {
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
    cvmProduceQuery: cvmProduceColumns,
    cvmFastProduceQuery: cvmFastProduceColumns,
    cvmProduceDetailQuery: cvmProduceDetailColumns,
    billsRootAccountSummary: billsRootAccountSummaryColumns,
    billsMainAccountSummary: billsMainAccountSummaryColumns,
    billsSummaryOperationRecord: billsSummaryOperationRecordColumns,
    businessHostColumns,
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
