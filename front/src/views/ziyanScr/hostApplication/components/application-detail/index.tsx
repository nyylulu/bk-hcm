import { Ref, defineComponent, onMounted, ref, computed } from 'vue';
import './index.scss';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import CommonCard from '@/components/CommonCard';
import { Button, Table, Timeline, Message } from 'bkui-vue';
import http from '@/http';
import { useRoute } from 'vue-router';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import { Copy, Share } from 'bkui-vue/lib/icon';
import { useRequireTypes } from '@/views/ziyanScr/hooks/use-require-types';
import { timeFormatter } from '@/common/util';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';

import { isEqual } from 'lodash';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;
export default defineComponent({
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
    const Hostcolumns = [
      ...cloudcolumns,
      // {
      //   label: '操作',
      //   width: 120,
      //   render: () => {
      //     return (
      //       <Button text theme='primary' onClick={() => {}}>
      //         查看变更记录
      //       </Button>
      //     );
      //   },
      // },
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
        render: ({ index }: any) => detail.value.info?.[index]?.total_num,
      },
      {
        label: '交付情况-待交付',
        field: 'pending_num',
        render: ({ index }: any) => detail.value.info?.[index]?.pending_num,
      },
      {
        label: '交付情况-已交付',
        field: 'success_num',
        render: ({ index }: any) => (
          <span class={'copy-wrapper'}>
            {detail.value.info?.[index]?.success_num}
            <Copy class={'copy-icon ml4'} v-clipboard:copy={detail.value.info?.index?.success_num} />
          </span>
        ),
      },
      ...physicalcolumns,
      {
        label: '状态',
        field: 'stage',
        width: 180,
        render: ({ index }: any) => detail.value.info?.[index]?.status,
      },
      // {
      //   label: '操作',
      //   width: 120,
      //   render: () => {
      //     return (
      //       <Button text theme='primary' onClick={() => {}}>
      //         查看变更记录
      //       </Button>
      //     );
      //   },
      // },
    ];
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
    // 获取需求子单
    const getdemandDetail = async (orderId: string) => {
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
    const getOrderAuditRecords = async (orderId: string) => {
      const { data } = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/task/get/apply/ticket/audit`, {
        order_id: +orderId,
      });
      applyRecord.value = data;
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
    onMounted(() => {
      getOrderDetail(route.params.id as string);
      getOrderAuditRecords(route.params.id as string);
      getdemandDetail(route.params.id as string);
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
            {detail.value.suborders?.some(({ resource_type }) => resource_type === 'QCLOUDCVM') && (
              <>
                <p class={'mt16 mb8'}>云主机</p>
                <Table
                  data={suborders.value}
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
            {detail.value.suborders?.some(({ resource_type }) => ['IDCPM', 'IDCDVM'].includes(resource_type)) && (
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
                  data={suborders.value}
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
      </div>
    );
  },
});
