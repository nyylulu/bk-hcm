import { Ref, defineComponent, onMounted, ref, computed } from 'vue';
import './index.scss';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import CommonCard from '@/components/CommonCard';
import { Button, Table, Timeline } from 'bkui-vue';
import http from '@/http';
import { useRoute } from 'vue-router';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import { Share } from 'bkui-vue/lib/icon';
import { useRequireTypes } from '@/views/ziyanScr/hooks/use-require-types';
import { timeFormatter } from '@/common/util';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
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
      {
        label: '操作',
        width: 120,
        render: () => {
          return (
            <Button text theme='primary' onClick={() => {}}>
              查看变更记录
            </Button>
          );
        },
      },
    ];
    const Machinecolumns = [
      ...physicalcolumns,
      {
        label: '操作',
        width: 120,
        render: () => {
          return (
            <Button text theme='primary' onClick={() => {}}>
              查看变更记录
            </Button>
          );
        },
      },
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

      return selections.value.map((item) => {
        batchCopyIps = batchCopyIps.concat(ips.value[item.suborderId]);
      });
    });
    // 获取需求子单
    const getdemandDetail = async (orderId: string) => {
      const { data } = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/task/findmany/apply`, {
        order_id: [+orderId],
        page: { start: 0, limit: 50 },
      });
      detail.value.info = data.info;
    };
    // 获取单据详情
    const getOrderDetail = async (orderId: string) => {
      const { data } = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/task/get/apply/ticket`, {
        order_id: +orderId,
      });
      detail.value = data;
    };
    // 获取单据审核记录
    const getOrderAuditRecords = async (orderId: string) => {
      const { data } = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/task/get/apply/ticket/audit`, {
        order_id: +orderId,
      });
      applyRecord.value = data;
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
            <Button class={'mr8'} v-clipboard={clipHostIp.value.join('\n')} disabled={selections.value.length === 0}>
              批量复制IP
            </Button>
            {detail.value.info?.some(({ resource_type }) => resource_type === 'QCLOUDCVM') && (
              <>
                <p class={'mt16 mb8'}>云主机</p>
                <Table
                  data={detail.value.info}
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
            {detail.value.suborders?.some(({ resource_type }) => resource_type === 'IDCDVM') && (
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
                  data={detail.value.suborders}
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
