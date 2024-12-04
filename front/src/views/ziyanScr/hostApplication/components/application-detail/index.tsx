import { Ref, defineComponent, onMounted, ref, computed, onUnmounted } from 'vue';
import { useRoute } from 'vue-router';
import './index.scss';

import { isEqual } from 'lodash';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { useRequireTypes } from '@/views/ziyanScr/hooks/use-require-types';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import { timeFormatter } from '@/common/util';
import { getBusinessNameById } from '@/views/ziyanScr/host-recycle/field-dictionary';
import http from '@/http';

import { Button, Table, Message } from 'bkui-vue';
import { Copy } from 'bkui-vue/lib/icon';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import Panel from '@/components/panel';
import WName from '@/components/w-name';
import ModifyRecord from './modify-record';
import ItsmTicketAudit from './itsm-ticket-audit.vue';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default defineComponent({
  components: {
    WName,
    ModifyRecord,
  },
  setup() {
    const route = useRoute();
    const { getBusinessApiPath } = useWhereAmI();
    const ips = ref({});
    const detail: Ref<{
      info: any;
      [key: string]: any;
    }> = ref({ info: [] });
    const { transformRequireTypes } = useRequireTypes();
    const { columns: cloudcolumns } = useColumns('cloudRequirementSubOrder');
    const { columns: physicalcolumns } = useColumns('physicalRequirementSubOrder');
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
    cloudcolumns.splice(3, 0, ...numColumns);

    const Hostcolumns = [
      ...cloudcolumns,
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

    const Machinecolumns = [
      { type: 'selection', width: 30, minWidth: 30, align: 'center' },
      {
        label: '机型',
        field: 'spec.device_type',
        width: 140,
      },
      // 给物理机添加num字段
      ...numColumns,
      ...physicalcolumns,
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
    const showRecord = (row) => {
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
      Message({
        message,
        theme: 'success',
        duration: 1500,
      });
    };
    const suborders = ref([]);
    const cloundMachineList = computed(() => {
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
    const getdemandDetail = async () => {
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
      list.forEach((item) => {
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
          getdemandDetail();
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

    const getDeliveredHostField = async (suborderId) => {
      const params = {
        filter: {
          condition: 'AND',
          rules: [
            {
              field: 'suborder_id',
              operator: 'equal',
              value: suborderId,
            },
            {
              field: 'bk_biz_id',
              operator: 'in',
              value: [detail.value.bk_biz_id],
            },
          ],
        },
      };
      const { data } = await http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/${getBusinessApiPath()}task/findmany/apply/device`,
        params,
      );
      return Promise.resolve().then(() => {
        const value = data?.info?.map((item) => item.ip);
        return value;
      });
    };
    onMounted(async () => {
      await getOrderDetail(route.params.id as string);
      await getdemandDetail();
    });

    onUnmounted(() => {
      // 清除定时任务
      clearTimeout(demandDetailTimer.id);
    });

    return () => (
      <div class={'application-detail-container'}>
        <DetailHeader>单据详情</DetailHeader>
        <div class={'detail-wrapper'}>
          <Panel title='基本信息'>
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
            {cloundMachineList.value.length > 0 && (
              <>
                <p class={'mt16 mb8'}>云主机</p>
                <Table
                  showOverflowTooltip
                  data={cloundMachineList.value}
                  columns={Hostcolumns}
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
                  columns={Machinecolumns}
                />
              </>
            )}
          </Panel>

          <Panel title='审批流程' class='mt24'>
            <ItsmTicketAudit
              orderId={+route.params.id}
              creator={route.query.creator as string}
              bkBizId={Number(route.query.bkBizId)}
            />
          </Panel>
        </div>
        <ModifyRecord v-model={showRecordSlider.value} showObj={recordParams.value} />
      </div>
    );
  },
});
