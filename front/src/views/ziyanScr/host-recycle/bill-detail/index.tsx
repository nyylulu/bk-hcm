import { defineComponent, reactive, ref, computed, onMounted } from 'vue';
import { getResourceTypeName, getReturnPlanName } from '@/utils';
import { getRecycleTaskStatusLabel } from '@/views/ziyanScr/host-recycle/field-dictionary/recycleStatus';
import { dateTimeTransform } from '@/views/ziyanScr/host-recycle/field-dictionary/dateTime';
// import { exportTableToExcel } from '@/utils';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
// import { getRecycleHosts, getRecycleOrders, retryOrder, submitOrder, stopOrder, auditOrder } from '@/api/host/recycle';
import { useTable } from '@/hooks/useTable/useTable';
import { Loading } from 'bkui-vue/lib/icon';
import './index.scss';
export default defineComponent({
  props: {
    suborderIdStr: {
      type: String,
      default: '',
    },
  },
  setup(props) {
    const billBaseInfo = reactive({});
    const billStatus = computed(() => {
      const { status = '' } = billBaseInfo;
      if (status === 'DONE') {
        return 'success';
      }
      const isError = status.includes('FAILED') || status.includes('REJECTED') || status.includes('TERMINATE');
      return isError ? 'error' : 'process';
    });
    const currentParams = ref({});
    const suborderId = computed(() => {
      // 路由写法
      // return this.$route.query.suborderId?.split(',') || [];
      // 转参写法
      return props.suborderIdStr.split(',') || [];
    });
    const loadOrders = () => {
      const page = {
        start: 0,
        limit: 10,
        enable_count: false,
      };
      const params = {
        suborderId,
        page,
      };
      currentParams.value = { ...params };
      // return getRecycleOrders(params)
      //   .then((res) => {
      //     const orders = res.data.info || [{}];
      //     const [order = {}] = orders;
      //     billBaseInfo = order;
      //   })
      //   .catch(() => {
      //     billBaseInfo = {};
      //   });
    };
    const fetchRetryOrder = () => {
      // retryOrder({
      //   suborderId: suborderId,
      // }).then((res) => {
      //   if (res.code === 0) {
      //     // this.$message.success('重试成功');
      //     loadOrders();
      //   } else {
      //     // this.$message.error(res.message);
      //   }
      // });
    };
    const fetchStopOrder = () => {
      // stopOrder({
      //   suborderId: suborderId,
      // }).then((res) => {
      //   if (res.code === 0) {
      //     // this.$message.success('终止成功');
      //     loadOrders();
      //   } else {
      //     // this.$message.error(res.message);
      //   }
      // });
    };
    const fetchSubmitOrder = () => {
      // submitOrder({
      //   suborderId: suborderId,
      // }).then((res) => {
      //   if (res.code === 0) {
      //     // this.$message.success('提交成功');
      //     loadOrders();
      //   } else {
      //     // this.$message.error(res.message);
      //   }
      // });
    };

    const activeStep = computed(() => {
      const list = ['DETECT', 'AUDIT', 'TRANSIT', 'RETURN', 'DONE'];
      return list.indexOf(billBaseInfo.stage);
    });
    const stepArr = computed(() => {
      const list = [
        { title: '预检', description: activeStep.value === 0 ? billBaseInfo.message : '' },
        {
          title: 'BG管理员审核',
          description:
            activeStep.value === 0 && billBaseInfo.status === 'REJECTED' ? `驳回原因：${billBaseInfo.message}` : '',
        },
        { title: 'CR系统处理中', description: activeStep.value === 2 ? billBaseInfo.message : '' },
        { title: '公司资源系统处理中', description: activeStep.value === 3 ? billBaseInfo.message : '' },
        { title: '完成' },
      ];
      return list;
    });
    const remark = ref('');
    // const admins = ref(['dommyzhang', 'forestchen']);
    // approval 参数
    const fetchAuditOrder = (approval) => {
      return approval;
      // auditOrder({
      //   suborderId,
      //   approval,
      //   remark.value,
      // }).then(res => {
      //   if (res.code === 0) {
      //     // this.$message.success('请求成功');
      //     loadOrders();
      //   } else {
      //     // this.$message.error(res.message);
      //   }
      // })
    };

    const exportLoading = ref(false);
    const selectedRows = ref([]);
    // const deviceDestroyList = ref([]);
    const { columns } = useColumns('deviceDestroy');
    const operateColList = [
      {
        label: '查看',
        width: 80,
        render: () => {
          return (
            <div>
              {/* <router-link to={{ name: 'precheck-detail', query: { suborderId: row.suborderId, ip: row.ip } }}> */}
              <bk-button size='small' theme='primary' text>
                详情
              </bk-button>
              {/* </router-link> */}
            </div>
          );
        },
      },
    ];
    const tableColumns = [...columns, ...operateColList];
    // const page = ref({
    //   start: 0,
    //   limit: 10,
    //   total: 0,
    // });
    const { CommonTable } = useTable({
      tableOptions: {
        columns: tableColumns,
      },
      requestOption: {
        dataPath: 'data.info',
      },
      scrConfig: () => {
        return {
          url: '/api/v1/woa/config/findmany/config/cvm/device/detail',
          payload: {
            filter: {
              condition: 'AND',
              rules: [
                {
                  field: 'require_type',
                  operator: 'equal',
                  value: 1,
                },
                {
                  field: 'label.device_group',
                  operator: 'in',
                  value: ['标准型'],
                },
              ],
            },
            page: [],
          },
          filter: { simpleConditions: true, requestId: 'devices' },
        };
      },
    });
    const getDeviceDestroyList = (enableCount = false) => {
      return enableCount;
      // getRecycleHosts({
      //   suborderId,
      //   page.value,
      // }, {
      //   // requestId,
      //   enableCount,
      // }).then(res => {
      //   if (enableCount) page.value.total = res.data?.count
      //   deviceDestroyList.value = res.data?.info || []
      // })
    };
    const exportToExcel = () => {
      exportLoading.value = true;

      // getRecycleHosts({
      //   suborderId,
      //   page: {
      //     start: 0,
      //     limit: 500,
      //   },
      // }, {
      //   enableCount: false,
      // }).then(res => {
      //   const totalList = res.data?.info || []
      //   exportTableToExcel(totalList, tableColumns, '设备销毁详情');
      // }).finally(() => {
      //   exportLoading.value = false
      // })
    };
    const pollObj = ref(null);
    const pollOrders = () => {
      // return getRecycleOrders(currentParams.value, {
      //   alertError: false,
      // })
      //   .then((res) => {
      //     const orders = res.data.info || [{}];
      //     const [order = {}] = orders;
      //     billBaseInfo = order;
      //   })
      //   .catch(() => {
      //     // console.error('poll orders failed');
      //   });
    };
    onMounted(() => {
      loadOrders();
      getDeviceDestroyList();
      if (pollObj.value) clearInterval(pollObj.value);
      pollObj.value = setInterval(pollOrders, 5000);
    });
    return () => (
      <div class='bill-detail'>
        <div class='base-info'>
          <h2>基本信息</h2>
          <div class='base-info-top'>
            <div>
              <label>主单号:</label>
              <span>{billBaseInfo.orderId}</span>
            </div>
            <div>
              <label>子单号:</label>
              <span>{billBaseInfo.suborderId}</span>
            </div>
            <div>
              <label>业务:</label>
              <span>{billBaseInfo.bkBizName}</span>
            </div>
            <div>
              <label>资源类型:</label>
              <span>{getResourceTypeName(billBaseInfo.resourceType)}</span>
            </div>
            <div>
              <label>退回策略:</label>
              <span>{getReturnPlanName(billBaseInfo.returnPlan, billBaseInfo.resourceType)}</span>
            </div>
            <div>
              <label>单据状态:</label>
              <span class={{ 'c-danger': billStatus.value === 'error', 'c-success': billStatus.value === 'success' }}>
                {billStatus.value === 'process' ? <Loading /> : ''}
                {getRecycleTaskStatusLabel(billBaseInfo.status)}
              </span>
            </div>
            <div>
              <label>设备总数/销毁完成数量:</label>
              <span>
                {billBaseInfo.totalNum}/{billBaseInfo.successNum}
              </span>
            </div>
            <div>
              <label>处理人:</label>
              <span>{billBaseInfo.handler}</span>
            </div>
            <div>
              <label>回收人:</label>
              <span>{billBaseInfo.bkUsername}</span>
            </div>
            <div>
              <label>提单时间:</label>
              <span>{dateTimeTransform(billBaseInfo.createAt)}</span>
            </div>
          </div>
          <div class='base-info-bottom'>
            <bk-button
              //   :loading="$isLoading(requestId)"
              onClick={fetchRetryOrder}>
              重试
            </bk-button>
            <bk-button
              //   :loading="$isLoading(requestId)"
              onClick={fetchStopOrder}>
              终止
            </bk-button>
            <bk-button
              //   :loading="$isLoading(requestId)"
              onClick={fetchSubmitOrder}>
              去除预检失败IP提交
            </bk-button>
          </div>
        </div>
        <div class='handle-process'>
          <h2>处理流程</h2>
          <bk-steps line-type='solid' steps={stepArr} />
          {/* v-if="admins.includes($store.getters.name) && order.status === 'FOR_AUDIT'" TODO 下面 展示 条件 */}
          <div class='check-export'>
            <div class='check-export-man'>
              <span>当前审批步骤：BG管理员审核</span>
              <span>审核人：</span>
              {/* <w-name
          v-for="item in admins"
          :key="item"
          :username="item"
        /> */}
            </div>
            <div>
              <bk-input v-model={remark} type='textarea' rows={2} placeholder='请输入审核意见' />

              <div class='check-export-trigger'>
                <bk-button
                  // :loading="$isLoading(requestId)"
                  theme='primary'
                  onClick={fetchAuditOrder(true)}>
                  审核通过
                </bk-button>
                <bk-button
                  // :loading="$isLoading(requestId)"
                  theme='danger'
                  onClick={fetchAuditOrder(false)}>
                  驳回
                </bk-button>
              </div>
            </div>
          </div>
        </div>
        <div class='device-destroy'>
          <h2>设备销毁详情</h2>
          <div class='device-destroy-top'>
            <bk-button
              // v-clipboard="clipHostIp"
              // v-clipboard:success="()=> $message.info('已复制')"
              // icon="el-icon-document-copy"
              disabled={!selectedRows.value.length}>
              复制选中IP
            </bk-button>
            <bk-button
              // icon="el-icon-download"
              // loading={exportLoading}
              onClick={exportToExcel}>
              导出全部
            </bk-button>
          </div>
          <CommonTable></CommonTable>
        </div>
      </div>
    );
  },
});
