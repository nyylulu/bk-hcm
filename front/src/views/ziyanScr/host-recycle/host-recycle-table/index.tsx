import { defineComponent, ref, onMounted } from 'vue';
// import { useRouter } from 'vue-router';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useTable } from '@/hooks/useTable/useTable';
// import { getRecycleOrders, getRecycleStageOpts, retryOrder, submitOrder, stopOrder } from '@/api/host/recycle';
import AppSelect from '@blueking/app-select';
import RequireNameSelect from './require-name-select';
import MemberSelect from '@/components/MemberSelect';
import ExportToExcelButton from '@/components/export-to-excel-button';
import { Search } from 'bkui-vue/lib/icon';
import BillDetail from '../bill-detail';
import './index.scss';

export default defineComponent({
  components: {
    AppSelect,
    RequireNameSelect,
    MemberSelect,
    ExportToExcelButton,
    BillDetail,
  },
  setup() {
    const defaultRecycleForm = {
      bkBizId: [],
      orderId: [],
      suborderId: [],
      resourceType: [],
      recycle_type: [],
      returnPlan: [],
      stage: [],
      handlerTime: [new Date(), new Date()],
      //   start: getDate('yyyy-MM-dd', -30),
      //   end: getDate('yyyy-MM-dd', 0),
      bkUsername: '',
      //   bkUsername: [this.$store.getters.name],
    };
    const recycleForm = ref(defaultRecycleForm);
    const bussinessList = [];
    const resourceTypeList = [
      {
        key: 'QCLOUDCVM',
        value: '腾讯云虚拟机',
      },
      {
        key: 'IDCPM',
        value: 'IDC物理机',
      },
      {
        key: 'OTHERS',
        value: '其他',
      },
    ];
    const returnPlanList = [
      {
        key: 'IMMEDIATE',
        value: '立即销毁',
      },
      {
        key: 'DELAY',
        value: '延迟销毁',
      },
    ];
    const stageList = [];
    const recycleMen = [];
    const start = ref(0);
    // const selectedSuborderId = ref([]);
    const selectedRows = ref([]);
    const billList = ref([]);
    // const pageInfo = ref({
    //   start: 0,
    //   limit: 10,
    //   total: 0,
    // });
    // const currentParams = ref({});
    const loadOrders = ({ enableCount } = { enableCount: false }) => {
      return enableCount;
      // const params = {
      //   ...recycleForm.value,
      //   page: pageInfo.value,
      // };
      // params.orderId = params.orderId.map((item) => Number(item));

      // currentParams.value = { ...params };
      // return getRecycleOrders(params, {
      //   // requestId: this.orders.requestId,
      //   enableCount,
      // })
      //   .then((res) => {
      //     if (enableCount) pageInfo.value.total = res.data.count;
      //     billList.value = res.data.info;
      //   })
      //   .catch(() => {
      //     billList.value = [];
      //   });
    };
    const filterOrders = () => {
      start.value = 0;
      loadOrders({ enableCount: true });
    };
    const clearFilter = () => {
      // TODO 可能会影响到原始值
      recycleForm.value = defaultRecycleForm;
      filterOrders();
    };
    // const router = new useRouter();
    const goToPrecheck = () => {
      // TODO 不知为啥 pinia
      // this.$store.commit('SET_SUBORDER_IDS', this.selectedSuborderId)
      // router.push({ name: 'precheck-detail' });
    };
    const retryOrderFunc = (id) => {
      return id;
      // const suborderId = id === 'isBatch' ? selectedSuborderId : [id];
      // retryOrder({
      //   suborderId,
      // }).then((res) => {
      //   if (res.code === 0) {
      //       this.$message.success('重试成功');
      //     loadOrders();
      //   } else {
      //       this.$message.error(res.message);
      //   }
      // });
    };
    const stopOrderFunc = (id) => {
      return id;
      // stopOrder({
      //   suborderId: [id],
      // }).then((res) => {
      //   if (res.code === 0) {
      //       this.$message.success('终止成功');
      //     loadOrders();
      //   } else {
      //       this.$message.error(res.message);
      //   }
      // });
    };
    const submitOrderFunc = (id) => {
      return id;
      // const suborderId = id === 'isBatch' ? selectedSuborderId : [id];
      // submitOrder({
      //   suborderId,
      // }).then((res) => {
      //   if (res.code === 0) {
      //       this.$message.success('提交成功');
      //     loadOrders();
      //   } else {
      //       this.$message.error(res.message);
      //   }
      // });
    };
    const { columns } = useColumns('hostRecycle');
    // 0 表示单据列表，1 表示单据详情
    const hostRecyclePage = ref(0);
    const suborderId = ref('');
    const enterDetail = (pageFlag, params) => {
      hostRecyclePage.value = pageFlag;
      suborderId.value = params;
    };
    const operateColList = [
      {
        label: '操作',
        render: ({ row }) => {
          return (
            <div>
              {/* <router-link to={{ name: 'precheck-detail', query: { suborderId: row.suborderId } }}> */}
              <bk-button size='small' type='text'>
                预检详情
              </bk-button>
              {/* </router-link> */}
              {/* <router-link to={{ name: 'recycle-detail', query: { suborderId: row.suborderId } }}> */}
              <bk-button class='ml-10' size='small' type='text' onClick={enterDetail(1, row.suborderId)}>
                单据详情
              </bk-button>
              {/* </router-link> */}
              {!['DONE', 'TERMINATE'].includes(row.status) ? (
                <div>
                  <bk-button
                    onClick={() => retryOrderFunc(row.suborderId)}
                    // loading全局配置
                    // loading={this.$isLoading(`retry-${row.suborderId}`)}
                    class='ml-10'
                    size='small'
                    type='text'>
                    <svg-icon v-bk-tooltips={{ content: '重试' }} icon-class='retry' />
                  </bk-button>
                  <bk-button
                    onClick={() => stopOrderFunc(row.suborderId)}
                    // loading={this.$isLoading(`stop-${row.suborderId}`)}
                    class='ml-10'
                    size='small'
                    type='text'>
                    <svg-icon v-bk-tooltips={{ content: '终止' }} icon-class='stop' />
                  </bk-button>
                  <bk-button
                    onClick={() => submitOrderFunc(row.suborderId)}
                    // loading={this.$isLoading(`submit-${row.suborderId}`)}
                    class='ml-10'
                    size='small'
                    type='text'>
                    <svg-icon v-bk-tooltips={{ content: '去除预检失败IP提交' }} icon-class='refresh' />
                  </bk-button>
                </div>
              ) : null}
            </div>
          );
        },
      },
    ];
    const tableColumns = [...columns, ...operateColList];
    const { CommonTable } = useTable({
      tableOptions: {
        columns: tableColumns,
      },
      requestOption: {
        type: 'load_balancers/with/delete_protection',
        sortOption: { sort: 'created_at', order: 'DESC' },
      },
      slotAllocation: () => {
        return {
          ScrSwitch: true,
          interface: {
            Parameters: {
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
            path: '/api/v1/woa/config/findmany/config/cvm/device/detail',
          },
        };
      },
    });
    const renderNodes = () => {
      if (hostRecyclePage.value === 0) {
        return (
          <CommonTable>
            {{
              tabselect: () => (
                <bk-form label-width='110' class='bill-filter-form' model={recycleForm}>
                  <bk-form-item label-width='0'>
                    {/* to={{ name: 'recycle-create' }} */}
                    {/* <router-link> */}
                    <bk-button theme='primary' icon='el-icon-plus'>
                      回收资源
                    </bk-button>
                    {/* </router-link> */}
                  </bk-form-item>
                  <bk-form-item label='业务'>
                    {/* <AppSelect>
                    {
                      {
                        //   append: () => (
                        //     <div class={'app-action-content'}>
                        //       <i class={'hcm-icon bkhcm-icon-plus-circle app-action-content-icon'} />
                        //       <span class={'app-action-content-text'}>新建业务</span>
                        //     </div>
                        //   ),
                      }
                    }
                  </AppSelect> */}
                    {/* TODO AppSelect使用不对 */}
                    <bk-select v-model={recycleForm.value.resourceType} multiple clearable placeholder='请选择业务'>
                      {bussinessList.map(({ key, value }) => {
                        return <bk-option key={key} label={value} value={key}></bk-option>;
                      })}
                    </bk-select>
                  </bk-form-item>
                  <bk-form-item label='OBS项目类型'>
                    <require-name-select v-model={recycleForm.value.recycle_type} multiple clearable collapse-tags />
                  </bk-form-item>
                  <bk-form-item label='单号'>
                    {/* TODO 是否封装成旧的 */}
                    <bk-input v-model={recycleForm.value.orderId} />
                  </bk-form-item>
                  <bk-form-item label='子单号'>
                    {/* TODO 是否封装成旧的 */}
                    <bk-input v-model={recycleForm.value.suborderId} />
                  </bk-form-item>
                  <bk-form-item label='资源类型'>
                    <bk-select v-model={recycleForm.value.resourceType} multiple clearable placeholder='请选择资源类型'>
                      {resourceTypeList.map(({ key, value }) => {
                        return <bk-option key={key} label={value} value={key}></bk-option>;
                      })}
                    </bk-select>
                  </bk-form-item>
                  <bk-form-item label='回收类型'>
                    <bk-select v-model={recycleForm.value.returnPlan} multiple clearable placeholder='请选择回收类型'>
                      {returnPlanList.map(({ key, value }) => {
                        return <bk-option key={key} label={value} value={key}></bk-option>;
                      })}
                    </bk-select>
                  </bk-form-item>
                  <bk-form-item label='状态'>
                    <bk-select v-model={recycleForm.value.stage} multiple clearable placeholder='请选择回收类型'>
                      {stageList.map(({ key, value }) => {
                        return <bk-option key={key} label={value} value={key}></bk-option>;
                      })}
                    </bk-select>
                  </bk-form-item>
                  <bk-form-item label='回收人'>
                    {/* TODO MemberSelect使用不对 */}
                    {/* <member-select v-model={recycleForm.value.bkUsername} multiple clearable allowCreate /> */}
                    <bk-select v-model={recycleForm.value.resourceType} multiple clearable placeholder='请选择业务'>
                      {recycleMen.map(({ key, value }) => {
                        return <bk-option key={key} label={value} value={key}></bk-option>;
                      })}
                    </bk-select>
                  </bk-form-item>
                  <bk-form-item label='回收时间'>
                    {/* TODO 是否封装成旧的 */}
                    <bk-date-picker v-model={recycleForm.value.handlerTime} type='daterange' />
                  </bk-form-item>
                  <bk-form-item label-width='0' class='bill-form-btn'>
                    <bk-button
                      theme='primary'
                      //   :loading="$isLoading(orders.requestId)"
                      native-type='submit'
                      onClick={filterOrders}>
                      <Search />
                      查询
                    </bk-button>
                    <bk-button
                      //   :loading="$isLoading(orders.requestId)"
                      onClick={clearFilter}>
                      {/* TODO icon='el-icon-refresh' */}
                      <Search />
                      清空
                    </bk-button>
                    <export-to-excel-button data={billList.value} columns={tableColumns} filename='回收单据列表' />
                    <bk-button
                      //   :loading="$isLoading(orders.requestId)"
                      disabled={!selectedRows.value.length}
                      onClick={goToPrecheck}>
                      批量查看预检详情
                    </bk-button>
                    <bk-button
                      //   :loading="$isLoading(orders.requestId)"
                      disabled={!selectedRows.value.length}
                      onClick={retryOrderFunc('isBatch')}>
                      批量重试
                    </bk-button>
                    <bk-button
                      //   :loading="$isLoading(orders.requestId)"
                      disabled={!selectedRows.value.length}
                      onClick={submitOrderFunc('isBatch')}>
                      批量去除预检失败IP提交
                    </bk-button>
                  </bk-form-item>
                </bk-form>
              ),
            }}
          </CommonTable>
        );
      }
      if (hostRecyclePage.value === 1) {
        return <BillDetail />;
      }
      return null;
    };
    onMounted(() => {
      // getListData();
    });
    return () => <div>{renderNodes()}</div>;
  },
});
