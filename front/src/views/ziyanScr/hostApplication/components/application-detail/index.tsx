import { Ref, defineComponent, onMounted, ref } from 'vue';
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
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default defineComponent({
  setup() {
    const route = useRoute();
    const detail: Ref<{
      suborders: [
        {
          resource_type: string;
        },
      ];
    }> = ref({});
    const { transformRequireTypes } = useRequireTypes();
    const { columns: CloudHostcolumns } = useColumns('CloudHost');
    const { columns: PhysicalMachinecolumns } = useColumns('PhysicalMachine');
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

    // 获取单据详情
    const getOrderDetail = (orderId: string) => {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/task/get/apply/ticket`, {
        order_id: +orderId,
      });
    };
    // 获取单据审核记录
    const getOrderAuditRecords = (orderId: string) => {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/task/get/apply/ticket/audit`, {
        order_id: +orderId,
      });
    };
    onMounted(async () => {
      const { data } = await getOrderDetail(route.params.id as string);
      detail.value = data;
      const { data: auditData } = await getOrderAuditRecords(route.params.id as string);
      applyRecord.value = auditData;
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
            {detail.value.suborders?.some(({ resource_type }) => resource_type === 'QCLOUDCVM') && (
              <>
                <p class={'mt16 mb8'}>云主机</p>
                <Table data={detail.value.suborders} columns={CloudHostcolumns} />
              </>
            )}
            {detail.value.suborders?.some(({ resource_type }) => resource_type === 'IDCDVM') && (
              <>
                <p class={'mt16 mb8'}>物理机</p>
                <Table data={detail.value.suborders} columns={PhysicalMachinecolumns} />
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
