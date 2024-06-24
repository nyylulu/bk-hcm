import { Ref, defineComponent, onMounted, ref, computed, onUnmounted } from 'vue';
import { useUserStore } from '@/store';
import './index.scss';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import CommonCard from '@/components/CommonCard';
import { Button, Table, Timeline, Message, Input } from 'bkui-vue';
import http from '@/http';
import { useRoute } from 'vue-router';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import { Copy, Share } from 'bkui-vue/lib/icon';
import { useRequireTypes } from '@/views/ziyanScr/hooks/use-require-types';
import { timeFormatter } from '@/common/util';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import WName from '@/components/w-name';
import ModifyRecord from './modify-record';
import { getBusinessNameById } from '@/views/ziyanScr/host-recycle/field-dictionary';
import { isEqual } from 'lodash';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;
export default defineComponent({
  components: {
    WName,
    ModifyRecord,
  },
  setup() {
    const route = useRoute();
    const ips = ref({});
    const detail: Ref<{
      info: any;
    }> = ref({});
    const { transformRequireTypes } = useRequireTypes();
    const { columns: cloudcolumns } = useColumns('cloudRequirementSubOrder');
    const { columns: physicalcolumns } = useColumns('physicalRequirementSubOrder');
    const { selections, handleSelectionChange } = useSelection();
    cloudcolumns.splice(4, 0, {
      label: '交付情况-已支付',
      field: 'success_num',
      render: ({ row }: any) => (
        <span class={'copy-wrapper'}>
          {row.success_num}
          {row.success_num > 0 ? (
            <Button text theme='primary'>
              <Copy class={'copy-icon'} v-clipboard:copy={(ips.value[row.suborder_id] || []).join('\n')} />
            </Button>
          ) : null}
        </span>
      ),
    });
    const Hostcolumns = [
      ...cloudcolumns,
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
    const Machinecolumns = [
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
        render: ({ row }: any) => (
          <span class={'copy-wrapper'}>
            {row.success_num}
            {row.success_num > 0 ? (
              <Button text theme='primary'>
                <Copy class={'copy-icon'} v-clipboard:copy={(ips.value[row.suborder_id] || []).join('\n')} />
              </Button>
            ) : null}
          </span>
        ),
      },
      ...physicalcolumns,
      {
        label: '状态',
        field: 'stage',
      },
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
    const applyRecord = ref({
      order_id: 0,
      itsm_ticket_id: '',
      itsm_ticket_link: '',
      status: '',
      current_steps: [],
      logs: [
        {
          operator: '',
          operate_at: '',
          message: '',
          source: '',
        },
      ],
    });
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
    // 获取需求子单
    const getdemandDetail = async () => {
      if (detail.value.stage === 'AUDIT') return;
      const orderId = route.params.id;
      const { data } = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/task/findmany/apply`, {
        order_id: [+orderId],
        page: { start: 0, limit: 50 },
      });
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
    };
    // 获取单据详情
    const getOrderDetail = async (orderId: string) => {
      const { data } = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/task/get/apply/ticket`, {
        order_id: +orderId,
      });
      detail.value = data;
      suborders.value = data?.suborders || [];
    };
    // 获取单据审核记录
    const getOrderAuditRecords = async () => {
      const orderId = route.params.id;
      const { data } = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/task/get/apply/ticket/audit`, {
        order_id: +orderId,
      });
      applyRecord.value = data;
    };
    const userStore = useUserStore();
    const currentAuditStep = computed(() => {
      return {
        name: applyRecord.value.currentSteps?.[0]?.name || '',
        processors: applyRecord.value.currentSteps?.[0]?.processors || '',
        stateId: applyRecord.value.currentSteps?.[0]?.stateId || '',
      };
    });
    const auditRemark = ref('');
    const approvalOrder = (params: Object) =>
      http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/task/audit/apply/ticket`, params);
    const approval = (resolve) => {
      const { itsmTicketId } = applyRecord.value;
      approvalOrder({
        orderId: +route.params.id,
        itsmTicketId,
        stateId: +currentAuditStep.value.stateId,
        operator: userStore.username,
        approval: resolve,
        remark: auditRemark.value,
      }).then(() => {
        getOrderAuditRecords();
        auditRemark.value = '';
      });
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
              operator: 'equal',
              value: Number(route?.query?.bk_biz_id),
            },
          ],
        },
      };
      const { data } = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/task/findmany/apply/device`, params);
      return Promise.resolve().then(() => {
        const value = data?.info?.map((item) => item.ip);
        return value;
      });
    };
    const refreshTimer = ref(null);
    onMounted(async () => {
      await getOrderDetail(route.params.id as string);
      await getdemandDetail();
      getOrderAuditRecords();
      if (refreshTimer.value) clearInterval(refreshTimer.value);
      refreshTimer.value = setInterval(() => {
        getOrderAuditRecords();
        getdemandDetail();
      }, 5000);
    });
    onUnmounted(() => {
      clearInterval(refreshTimer.value);
    });
    return () => (
      <div class={'application-detail-container'}>
        <DetailHeader>单据详情</DetailHeader>
        <div class={'detail-wrapper'}>
          <CommonCard title={() => '基本信息'}>
            <DetailInfo
              detail={detail.value}
              fields={[
                {
                  name: '单据 ID',
                  prop: 'order_id',
                },
                {
                  name: '提单人',
                  prop: 'bk_username',
                },
                {
                  name: '创建时间',
                  prop: 'create_at',
                  render() {
                    return timeFormatter(detail.value.create_at, 'YYYY-MM-DD');
                  },
                },
              ]}
            />
          </CommonCard>
          <CommonCard title={() => '主单信息'} class={'mt24'}>
            <DetailInfo
              detail={detail.value}
              fields={[
                {
                  name: '业务',
                  prop: 'bk_biz_id',
                  render() {
                    return getBusinessNameById(detail.value.bk_biz_id);
                  },
                },
                {
                  name: '需求类型',
                  prop: 'require_type',
                  render() {
                    return transformRequireTypes(detail.value.requireType);
                  },
                },
                {
                  name: '期望交付时间',
                  prop: 'expect_time',
                  render() {
                    return timeFormatter(detail.value.expect_time, 'YYYY-MM-DD');
                  },
                },
                {
                  name: '关注人',
                  prop: 'follower',
                },
                {
                  name: '备注',
                  prop: 'remark',
                },
              ]}
            />
          </CommonCard>

          <CommonCard title={() => '需求子单'} class={'mt24'}>
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
          </CommonCard>
          <CommonCard title={() => '审批流程'} class={'mt24'}>
            <Button
              theme='primary'
              text
              onClick={() => {
                window.open(applyRecord.value.itsm_ticket_link, '_blank');
              }}>
              <Share width={12} height={12} class={'mr4'} fill='#3A84FF' />
              跳转到 ITSM 查看审批详情
            </Button>
            {currentAuditStep.value.name ? (
              <div class='apply-human'>
                当前审批步骤：{currentAuditStep.value.name} 审核人：
                <WName name={currentAuditStep.value.processors} />
              </div>
            ) : null}
            {currentAuditStep.value.processors.includes(userStore.username) ? (
              <div>
                <Input v-model={auditRemark.value} type='textarea' placeholder='请输入审核意见' />
                <div class='apply-operate'>
                  <Button theme='primary' onClick={() => approval(true)}>
                    审核通过
                  </Button>
                  <Button theme='danger' onClick={() => approval(false)}>
                    驳回
                  </Button>
                </div>
              </div>
            ) : null}
            <div class={'timeline-container'}>
              <Timeline
                list={applyRecord.value.logs.map(({ message, operate_at }) => ({
                  tag: message,
                  content: <span class={'timeline-content-txt'}>{operate_at}</span>,
                  nodeType: 'vnode',
                }))}
              />
            </div>
          </CommonCard>
        </div>
        <ModifyRecord v-model={showRecordSlider.value} showObj={recordParams.value} />
      </div>
    );
  },
});
