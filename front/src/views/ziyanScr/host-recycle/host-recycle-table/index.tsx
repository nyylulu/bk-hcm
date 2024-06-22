import { defineComponent, ref, computed, watch, onMounted } from 'vue';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useTable } from '@/hooks/useTable/useTable';
import { getRecycleStageOpts, retryOrder, submitOrder, stopOrder } from '@/api/host/recycle';
import { removeEmptyFields } from '@/utils/scr/remove-query-fields';
import BusinessSelector from '@/components/business-selector/index.vue';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import RequireNameSelect from './require-name-select';
import { Button, Form, Popover } from 'bkui-vue';
import MemberSelect from '@/components/MemberSelect';
import ExportToExcelButton from '@/components/export-to-excel-button';
import { GragFill, Search } from 'bkui-vue/lib/icon';
import BillDetail from '../bill-detail';
import FloatInput from '@/components/float-input';
import './index.scss';
import { useRouter } from 'vue-router';
import dayjs from 'dayjs';
const { FormItem } = Form;
export default defineComponent({
  components: {
    RequireNameSelect,
    MemberSelect,
    ExportToExcelButton,
    BillDetail,
    FloatInput,
  },
  props: {
    subBizBillNum: {
      type: Object,
      default: () => {},
    },
  },
  setup(props) {
    const router = useRouter();
    const defaultRecycleForm = () => {
      return {
        bk_biz_id: '',
        order_id: [],
        suborder_id: [],
        resource_type: [],
        recycle_type: [],
        return_plan: [],
        stage: [],
        bk_username: [],
      };
    };
    const defaultTime = () => [new Date(dayjs().subtract(30, 'day').format('YYYY-MM-DD')), new Date()];
    const recycleForm = ref(defaultRecycleForm());
    watch(
      () => recycleForm.value.bk_biz_id,
      (newVal, oldVal) => {
        if (!oldVal.length) {
          getListData();
        }
      },
    );
    const timeForm = ref(defaultTime());
    const handleTime = (time) => (!time ? '' : dayjs(time).format('YYYY-MM-DD'));
    const timeObj = computed(() => {
      return {
        start: handleTime(timeForm.value[0]),
        end: handleTime(timeForm.value[1]),
      };
    });
    const pageInfo = ref({
      start: 0,
      limit: 10,
      enable_count: false,
    });
    const requestListParams = computed(() => {
      const params = {
        ...recycleForm.value,
        ...timeObj.value,
        page: pageInfo.value,
      };
      params.bk_biz_id = params.bk_biz_id === 'all' ? 0 : params.bk_biz_id;
      params.order_id = params.order_id.length ? params.order_id.map((v) => +v) : [];
      removeEmptyFields(params);
      return params;
    });
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
    const stageList = ref([]);
    const fetchStageList = async () => {
      const data = await getRecycleStageOpts();
      stageList.value = data?.info || [];
    };
    const loadOrders = () => {
      getListData();
    };
    const filterOrders = () => {
      pageInfo.value.start = 0;
      loadOrders();
    };
    const clearFilter = () => {
      const initForm = defaultRecycleForm();
      initForm.bk_biz_id = businessRef.value.defaultBusiness;
      recycleForm.value = initForm;
      timeForm.value = defaultTime();
      filterOrders();
    };
    const goToPrecheck = () => {
      router.push({
        path: '/ziyanScr/hostRecycling/preDetail',
        query: {
          suborder_id: getBatchSuborderId().join('\n'),
        },
      });
    };
    const getBatchSuborderId = () => {
      return selections.value.map((item) => {
        return item.suborder_id;
      });
    };
    const retryOrderFunc = (id: string) => {
      const suborderId = id === 'isBatch' ? getBatchSuborderId() : [id];
      retryOrder(
        {
          suborderId,
        },
        {},
      ).then((res) => {
        if (res.code === 0) {
          loadOrders();
        }
      });
    };
    const stopOrderFunc = (id: string) => {
      stopOrder(
        {
          suborderId: [id],
        },
        {},
      ).then((res) => {
        if (res.code === 0) {
          loadOrders();
        }
      });
    };
    const submitOrderFunc = (id: string) => {
      const suborderId = id === 'isBatch' ? getBatchSuborderId() : [id];
      submitOrder(
        {
          suborderId,
        },
        {},
      ).then((res) => {
        if (res.code === 0) {
          loadOrders();
        }
      });
    };
    const { selections, handleSelectionChange } = useSelection();
    const { columns } = useColumns('hostRecycle');
    const enterDetail = (row) => {
      router.push({
        path: '/ziyanScr/hostRecycling/docDetail',
        query: {
          suborderId: row.suborder_id,
          bkBizId: row.bk_biz_id,
        },
      });
    };
    watch(
      () => props.subBizBillNum,
      (newVal) => {
        enterDetail(newVal);
      },
    );
    const operateColList = [
      {
        label: '操作',
        width: '150px',
        render: ({ row }) => {
          return (
            <div class='recycle-operation'>
              <bk-button size='small' theme='primary' text onClick={() => returnPreDetails(row)}>
                预检详情
              </bk-button>
              <bk-button size='small' theme='primary' text onClick={() => enterDetail(row)}>
                单据详情
              </bk-button>
              {!['DONE', 'TERMINATE'].includes(row.status) ? (
                <>
                  <Popover theme='light'>
                    {{
                      default: () => (
                        <Button size='small' theme='primary' text>
                          <GragFill />
                        </Button>
                      ),
                      content: () => (
                        <>
                          <div>
                            <Button onClick={() => retryOrderFunc(row.suborder_id)} size='small' theme='primary' text>
                              重试
                            </Button>
                          </div>
                          <div>
                            <Button onClick={() => stopOrderFunc(row.suborder_id)} size='small' theme='primary' text>
                              终止
                            </Button>
                          </div>
                          <div>
                            <Button onClick={() => submitOrderFunc(row.suborder_id)} size='small' theme='primary' text>
                              去除预检失败IP提交
                            </Button>
                          </div>
                        </>
                      ),
                    }}
                  </Popover>
                </>
              ) : null}
            </div>
          );
        },
      },
    ];
    // 在第三个加子单号，需要跳转到单据详情，未用到路由
    columns.splice(2, 0, {
      label: '子单号',
      field: 'suborder_id',
      width: 80,
      render: ({ row }) => {
        return (
          // 单据详情
          <span class='sub-order-num' onClick={() => enterDetail(row)}>
            {row.suborder_id}
          </span>
        );
      },
    });
    const tableColumns = [...columns, ...operateColList];
    const { CommonTable, getListData, dataList } = useTable({
      tableOptions: {
        columns: tableColumns,
        extra: {
          onSelect: (selections: any) => {
            handleSelectionChange(selections, () => true, false);
          },
          onSelectAll: (selections: any) => {
            handleSelectionChange(selections, () => true, true);
          },
        },
      },
      requestOption: {
        dataPath: 'data.info',
        sortOption: {
          sort: 'create_at',
          order: 'DESC',
        },
        immediate: false,
      },
      scrConfig: () => {
        return {
          url: '/api/v1/woa/task/findmany/recycle/order',
          payload: {
            ...requestListParams.value,
          },
        };
      },
    });
    const returnPreDetails = (row: { suborder_id: any; bk_biz_id: any }) => {
      router.push({
        path: '/ziyanScr/hostRecycling/preDetail',
        query: {
          suborder_id: row.suborder_id,
        },
      });
    };
    const returnRecyclingResources = () => {
      router.push({
        path: '/ziyanScr/hostRecycling/resources',
      });
    };
    const businessRef = ref(null);
    const renderNodes = () => {
      return (
        <div class={'apply-list-container'}>
          <div class={'filter-container'}>
            <Form model={recycleForm} class={'scr-form-wrapper'}>
              <FormItem label='业务'>
                <BusinessSelector
                  ref={businessRef}
                  v-model={recycleForm.value.bk_biz_id}
                  autoSelect
                  authed
                  clearable={false}
                />
              </FormItem>
              <FormItem label='OBS项目类型'>
                <require-name-select v-model={recycleForm.value.recycle_type} multiple clearable collapse-tags />
              </FormItem>
              <FormItem label='单号'>
                <FloatInput v-model={recycleForm.value.order_id} placeholder='请输入单号，多个换行分割' />
              </FormItem>
              <FormItem label='子单号'>
                <FloatInput v-model={recycleForm.value.suborder_id} placeholder='请输入子单号，多个换行分割' />
              </FormItem>
              <FormItem label='资源类型'>
                <bk-select v-model={recycleForm.value.resource_type} multiple clearable placeholder='请选择资源类型'>
                  {resourceTypeList.map(({ key, value }) => {
                    return <bk-option key={key} label={value} value={key}></bk-option>;
                  })}
                </bk-select>
              </FormItem>
              <FormItem label='回收类型'>
                <bk-select v-model={recycleForm.value.return_plan} multiple clearable placeholder='请选择回收类型'>
                  {returnPlanList.map(({ key, value }) => {
                    return <bk-option key={key} label={value} value={key}></bk-option>;
                  })}
                </bk-select>
              </FormItem>
              <FormItem label='状态'>
                <bk-select v-model={recycleForm.value.stage} multiple clearable placeholder='请选择状态'>
                  {stageList.value.map(({ stage, description }) => {
                    return <bk-option key={stage} label={description} value={stage}></bk-option>;
                  })}
                </bk-select>
              </FormItem>
              <FormItem label='回收人'>
                <member-select
                  v-model={recycleForm.value.bk_username}
                  multiple
                  clearable
                  placeholder='请输入企业微信名'
                />
              </FormItem>
              <FormItem label='回收时间'>
                <bk-date-picker v-model={timeForm.value} type='daterange' />
              </FormItem>
            </Form>
          </div>
          <div>
            <CommonTable>
              {{
                tabselect: () => (
                  <bk-form label-width='110' class='bill-filter-form' model={recycleForm}>
                    <bk-form-item label-width='0' class='bill-form-btn'>
                      <bk-button theme='primary' onClick={returnRecyclingResources} icon='el-icon-plus'>
                        回收资源
                      </bk-button>
                      <bk-button theme='primary' onClick={filterOrders}>
                        <Search />
                        查询
                      </bk-button>
                      <bk-button onClick={() => clearFilter()}>清空</bk-button>
                      <export-to-excel-button data={dataList.value} columns={tableColumns} filename='回收单据列表' />
                      <bk-button disabled={!selections.value.length} onClick={goToPrecheck}>
                        批量查看预检详情
                      </bk-button>
                      <bk-button disabled={!selections.value.length} onClick={() => retryOrderFunc('isBatch')}>
                        批量重试
                      </bk-button>
                      <bk-button disabled={!selections.value.length} onClick={() => submitOrderFunc('isBatch')}>
                        批量去除预检失败IP提交
                      </bk-button>
                    </bk-form-item>
                  </bk-form>
                ),
              }}
            </CommonTable>
          </div>
        </div>
      );
    };
    onMounted(() => {
      fetchStageList();
    });
    return renderNodes;
  },
});
