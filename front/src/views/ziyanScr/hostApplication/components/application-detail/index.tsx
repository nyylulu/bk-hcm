import { Ref, defineComponent, ref, computed, onUnmounted, reactive, onBeforeMount } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import './index.scss';

import { isEqual } from 'lodash';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { useRequireTypes } from '@/views/ziyanScr/hooks/use-require-types';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import { useUserStore } from '@/store';
import { timeFormatter } from '@/common/util';
import { getBusinessNameById } from '@/views/ziyanScr/host-recycle/field-dictionary';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import { MENU_SERVICE_HOST_APPLICATION } from '@/constants/menu-symbol';
import http from '@/http';

import { Button, Table, Message, PopConfirm } from 'bkui-vue';
import { Copy } from 'bkui-vue/lib/icon';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import Panel from '@/components/panel';
import WName from '@/components/w-name';
import ModifyRecord from './modify-record';
import ItsmTicketAudit, { type IItsmTicketAudit } from './itsm-ticket-audit.vue';
import type { IQueryResData } from '@/typings';
import ApprovalStatus from './approval-status.vue';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default defineComponent({
  components: {
    WName,
    ModifyRecord,
  },
  setup() {
    const route = useRoute();
    const router = useRouter();
    const userStore = useUserStore();
    const { whereAmI, getBusinessApiPath, getBizsId } = useWhereAmI();

    const backRoute = computed(() => {
      if (whereAmI.value === Senarios.business) {
        return {
          name: 'ApplicationsManage',
          query: { [GLOBAL_BIZS_KEY]: detail.value?.bk_biz_id, type: 'host_apply' },
        };
      }
      return { name: MENU_SERVICE_HOST_APPLICATION };
    });

    const ips = ref<{ [key: string]: any }>({});
    const detail: Ref<{
      info: any;
      [key: string]: any;
    }> = ref({ info: [] });
    const { transformRequireTypes } = useRequireTypes();
    const { columns: cloudColumns } = useColumns('cloudRequirementSubOrder');
    const { columns: physicalColumns } = useColumns('physicalRequirementSubOrder');
    const { selections, handleSelectionChange } = useSelection();

    // 需求子单相关num字段
    const numColumns = [
      {
        label: '总数',
        field: 'total_num',
        width: 80,
        render: ({ row, cell }: any) => (detail.value?.stage === 'AUDIT' ? row.replicas : cell),
      },
      {
        label: '待交付',
        field: 'pending_num',
        width: 80,
        render: ({ row, cell }: any) => (detail.value?.stage === 'AUDIT' ? row.replicas : cell),
      },
      {
        label: '已交付',
        field: 'success_num',
        width: 80,
        render: ({ row, cell }: any) => {
          if (detail.value?.stage === 'AUDIT') return 0;
          return (
            <span class={'copy-wrapper'}>
              {cell}
              {cell > 0 ? (
                <Button text theme='primary'>
                  <Copy class={'copy-icon'} v-clipboard:copy={(ips.value[row.suborder_id] || []).join('\n')} />
                </Button>
              ) : null}
            </span>
          );
        },
      },
    ];

    // 给云主机添加num字段
    cloudColumns.splice(3, 0, ...numColumns);

    const hostColumns = [
      ...cloudColumns,
      {
        label: '操作',
        width: 120,
        fixed: 'right',
        render: ({ row }: any) => {
          return (
            <Button text theme='primary' onClick={() => showRecord(row)}>
              查看变更记录
            </Button>
          );
        },
      },
    ];

    const machineColumns = [
      { type: 'selection', width: 30, minWidth: 30, align: 'center' },
      {
        label: '机型',
        field: 'spec.device_type',
        width: 140,
      },
      // 给物理机添加num字段
      ...numColumns,
      ...physicalColumns,
      {
        label: '操作',
        width: 120,
        render: ({ row }: any) => {
          return (
            <Button text theme='primary' onClick={() => showRecord(row)}>
              查看变更记录
            </Button>
          );
        },
      },
    ];
    const showRecordSlider = ref(false);
    const recordParams = ref({});
    const showRecord = (row: any) => {
      showRecordSlider.value = true;
      recordParams.value = {
        suborderId: row.suborder_id,
        bkBizId: row.bk_biz_id,
      };
    };

    const clipHostIp = computed(() => {
      let batchCopyIps: any[] = [];
      selections.value.forEach((item) => {
        batchCopyIps = batchCopyIps.concat(ips.value?.[item.suborder_id] || []);
      });
      return batchCopyIps;
    });
    const batchMessage = () => {
      const message = !clipHostIp.value.length ? '仅复制已交付IP' : '已复制';
      Message({ message, theme: 'success', duration: 1500 });
    };
    const suborders = ref([]);
    const cloudMachineList = computed(() => {
      return suborders.value.filter((item) => {
        return item.resource_type === 'QCLOUDCVM';
      });
    });
    const physicMachineList = computed(() => {
      return suborders.value.filter((item) => {
        return ['IDCPM', 'IDCDVM'].includes(item.resource_type);
      });
    });

    const demandDetailTimer: any = { id: null, count: 0 };
    // 获取需求子单
    const getDemandDetail = async () => {
      if (detail.value.stage === 'AUDIT') return;
      const orderId = route.params.id;
      const { data } = await http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/${getBusinessApiPath()}task/findmany/apply`,
        {
          order_id: [+orderId],
          bk_biz_id: [detail.value.bk_biz_id],
          page: { start: 0, limit: 50 },
        },
      );
      detail.value.info = data.info;
      const list = data?.info || [];
      list.forEach((item: any) => {
        if (item.enableDiskCheck) item.spec.enableDiskCheck = '是';
      });
      if (!isEqual(suborders.value, list)) {
        suborders.value = list;
        suborders.value.forEach(async (item) => {
          const { suborder_id } = item;
          if (suborder_id) ips.value[suborder_id] = await getDeliveredHostField(suborder_id);
        });
      }
      // 如果需求子单中存在待交付云主机, 创建定时任务(30s刷新一次, 最多刷新60次)
      if (demandDetailTimer.count < 60 && list.some((item: any) => item.pending_num !== 0)) {
        demandDetailTimer.count += 1;
        demandDetailTimer.id = setTimeout(() => {
          getDemandDetail();
        }, 30000);
      }
    };

    // 获取单据详情
    const getOrderDetail = async (orderId: string) => {
      const { data } = await http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/${getBusinessApiPath()}task/get/apply/ticket`,
        { order_id: +orderId },
      );
      detail.value = data;
      suborders.value = data?.suborders || [];
    };

    const getDeliveredHostField = async (suborderId: string) => {
      const params = {
        filter: {
          condition: 'AND',
          rules: [
            { field: 'suborder_id', operator: 'equal', value: suborderId },
            { field: 'bk_biz_id', operator: 'in', value: [detail.value.bk_biz_id] },
          ],
        },
      };
      const { data } = await http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/${getBusinessApiPath()}task/findmany/apply/device`,
        params,
      );
      return Promise.resolve().then(() => {
        const value = data?.info?.map((item: any) => item.ip);
        return value;
      });
    };

    // 获取审批流信息
    const itsmTicketAuditOptions = reactive({ data: null, isLoading: false });
    const getItsmTicketAudit = async () => {
      itsmTicketAuditOptions.isLoading = true;
      try {
        const order_id = Number(route.params.id);
        const bk_biz_id = Number(route.query.bkBizId);
        const res: IQueryResData<IItsmTicketAudit> = await http.post(
          `/api/v1/woa/${getBusinessApiPath()}task/get/apply/ticket/audit`,
          { order_id, bk_biz_id },
        );
        itsmTicketAuditOptions.data = res.data;
      } catch (error) {
        console.error(error);
        itsmTicketAuditOptions.data = null;
      } finally {
        itsmTicketAuditOptions.isLoading = false;
      }
    };

    // 撤单
    const hasCancelBtn = computed(
      () =>
        route.query?.creator === userStore.username &&
        ['管理员审批', 'leader审批'].includes(itsmTicketAuditOptions.data?.current_steps[0]?.name),
    );
    const isCancelItsmTicketLoading = ref(false);
    const cancelItsmTicket = async () => {
      const { order_id } = itsmTicketAuditOptions.data;
      isCancelItsmTicketLoading.value = true;
      try {
        await http.post(`/api/v1/woa/${getBusinessApiPath()}task/apply/ticket/itsm_audit/cancel`, { order_id });
        Message({ theme: 'success', message: '撤单成功' });
        getItsmTicketAudit();
      } catch (error) {
        console.error(error);
      } finally {
        isCancelItsmTicketLoading.value = false;
      }
    };

    const validateBizId = () => {
      const globalBizId = getBizsId();
      const orderBizId = Number(route.query.bkBizId);
      return globalBizId === orderBizId;
    };

    onBeforeMount(async () => {
      if (whereAmI.value === Senarios.business && !validateBizId()) {
        router.replace(backRoute.value);
        return;
      }
      getItsmTicketAudit();
      await getOrderDetail(route.params.id as string);
      await getDemandDetail();
    });

    onUnmounted(() => {
      // 清除定时任务
      clearTimeout(demandDetailTimer.id);
    });

    return () => (
      <div class={'application-detail-container'}>
        <DetailHeader useRouterAction>
          {{
            default: () => '单据详情',
            right: () =>
              hasCancelBtn.value ? (
                <PopConfirm
                  trigger='click'
                  placement='top-end'
                  title='撤销单据'
                  content='撤销单据后，将取消本次的资源申请！'
                  onConfirm={cancelItsmTicket}>
                  <Button class='host-application-cancel-btn' loading={isCancelItsmTicketLoading.value}>
                    撤单
                  </Button>
                </PopConfirm>
              ) : null,
          }}
        </DetailHeader>
        <div class={'detail-wrapper'}>
          <ApprovalStatus class='mb24' ticketAuditDetail={itsmTicketAuditOptions.data} />

          <Panel title='审批信息'>
            <ItsmTicketAudit
              data={itsmTicketAuditOptions.data}
              isLoading={itsmTicketAuditOptions.isLoading}
              refreshApi={getItsmTicketAudit}
            />
          </Panel>

          <Panel title='基本信息' class='mt24'>
            <DetailInfo
              detail={detail.value}
              fields={[
                { name: '单据 ID', prop: 'order_id' },
                {
                  name: '创建时间',
                  prop: 'create_at',
                  render: () => timeFormatter(detail.value.create_at, 'YYYY-MM-DD'),
                },
                { name: '提单人', prop: 'bk_username' },
                {
                  name: '期望交付时间',
                  prop: 'expect_time',
                  render: () => timeFormatter(detail.value.expect_time, 'YYYY-MM-DD'),
                },
                {
                  name: '业务',
                  prop: 'bk_biz_id',
                  render: () => getBusinessNameById(detail.value.bk_biz_id),
                },
                { name: '关注人', prop: 'follower' },
                {
                  name: '需求类型',
                  prop: 'require_type',
                  render: () => transformRequireTypes(detail.value.require_type),
                },
                { name: '备注', prop: 'remark' },
              ]}
            />
          </Panel>

          <Panel title='需求子单' class='mt24'>
            <Button
              class={'mr8'}
              v-clipboard:copy={clipHostIp.value.join('\n')}
              v-clipboard:success={batchMessage}
              disabled={selections.value.length === 0}>
              批量复制IP
            </Button>
            {cloudMachineList.value.length > 0 && (
              <>
                <p class={'mt16 mb8'}>云主机</p>
                <Table
                  showOverflowTooltip
                  data={cloudMachineList.value}
                  columns={hostColumns}
                  {...{
                    onSelect: (selections: any) => {
                      handleSelectionChange(selections, () => true, false);
                    },
                    onSelectAll: (selections: any) => {
                      handleSelectionChange(selections, () => true, true);
                    },
                  }}
                />
              </>
            )}
            {physicMachineList.value.length > 0 && (
              <>
                <p class={'mt16 mb8'}>物理机</p>
                <Table
                  showOverflowTooltip
                  {...{
                    onSelect: (selections: any) => {
                      handleSelectionChange(selections, () => true, false);
                    },
                    onSelectAll: (selections: any) => {
                      handleSelectionChange(selections, () => true, true);
                    },
                  }}
                  data={physicMachineList.value}
                  columns={machineColumns}
                />
              </>
            )}
          </Panel>
        </div>
        <ModifyRecord v-model={showRecordSlider.value} showObj={recordParams.value} />
      </div>
    );
  },
});
