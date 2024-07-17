import { defineComponent, ref, computed, watch, onMounted, h, withDirectives } from 'vue';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import { useTable } from '@/hooks/useTable/useTable';
import { getRecycleStageOpts, retryOrder, submitOrder, stopOrder } from '@/api/host/recycle';
import { removeEmptyFields } from '@/utils/scr/remove-query-fields';
import BusinessSelector from '@/components/business-selector/index.vue';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import RequireNameSelect from './require-name-select';
import { Button, Form, Message, Select, Dropdown } from 'bkui-vue';
import { useUserStore } from '@/store';
import MemberSelect from '@/components/MemberSelect';
import ExportToExcelButton from '@/components/export-to-excel-button';
import { Plus, Search } from 'bkui-vue/lib/icon';
import BillDetail from '../bill-detail';
import FloatInput from '@/components/float-input';
import './index.scss';
import { useRoute, useRouter } from 'vue-router';
import dayjs from 'dayjs';
const { FormItem } = Form;
const { DropdownMenu, DropdownItem } = Dropdown;
export default defineComponent({
  components: {
    RequireNameSelect,
    MemberSelect,
    ExportToExcelButton,
    BillDetail,
    FloatInput,
  },
  setup() {
    const currentOperateRowIndex = ref(-1);
    const userStore = useUserStore();
    const router = useRouter();
    const route = useRoute();
    const defaultRecycleForm = () => {
      return {
        bk_biz_id: [],
        order_id: [],
        suborder_id: [],
        resource_type: [],
        recycle_type: [],
        return_plan: [],
        stage: [],
        bk_username: [userStore.username],
      };
    };
    const defaultTime = () => [new Date(dayjs().subtract(30, 'day').format('YYYY-MM-DD')), new Date()];
    const recycleForm = ref(defaultRecycleForm());
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
        bk_biz_id:
          recycleForm.value.bk_biz_id.length === 0
            ? businessRef.value.businessList.slice(1).map((item: any) => item.id)
            : recycleForm.value.bk_biz_id,
      };
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
    const filterOrders = () => {
      pagination.start = 0;
      recycleForm.value.bk_biz_id =
        recycleForm.value.bk_biz_id.length === 1 && recycleForm.value.bk_biz_id[0] === 'all'
          ? []
          : recycleForm.value.bk_biz_id;
      getListData();
    };
    const clearFilter = () => {
      const initForm = defaultRecycleForm();
      // 因为要保存业务全选的情况, 所以这里 defaultBusiness 可能是 ['all'], 而组件的全选对应着 [], 所以需要额外处理
      // 根源是此处的接口要求全选时携带传递所有业务id, 所以需要与空数组做区分
      initForm.bk_biz_id = businessRef.value.defaultBusiness;
      recycleForm.value = initForm;
      timeForm.value = defaultTime();
      filterOrders();
    };
    const goToPrecheck = () => {
      router.push({
        path: '/service/hostRecycling/preDetail',
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
    const loadHostRecycle = () => {
      getListData();
    };
    const textTip = (text, theme) => {
      const themeDes = {
        error: '失败',
        success: '成功',
      };
      Message({
        message: `${text}${themeDes[theme]}`,
        theme,
        duration: 1500,
      });
    };
    const retryOrderFunc = (id: string, disabled) => {
      if (disabled) return;
      const suborderId = id === 'isBatch' ? getBatchSuborderId() : [id];
      retryOrder(
        {
          suborderId,
        },
        {},
      ).then((res) => {
        if (res.code === 0) {
          textTip('重试', 'success');
          loadHostRecycle();
        }
      });
    };
    const stopOrderFunc = (id: string, disabled) => {
      if (disabled) return;
      stopOrder(
        {
          suborderId: [id],
        },
        {},
      ).then((res) => {
        if (res.code === 0) {
          textTip('终止', 'success');
          loadHostRecycle();
        }
      });
    };
    const submitOrderFunc = (id: string, disabled) => {
      if (disabled) return;
      const suborderId = id === 'isBatch' ? getBatchSuborderId() : [id];
      submitOrder(
        {
          suborderId,
        },
        {},
      ).then((res) => {
        if (res.code === 0) {
          textTip('去除预检失败IP提交', 'success');
          loadHostRecycle();
        }
      });
    };
    const { selections, handleSelectionChange } = useSelection();
    const { columns } = useColumns('hostRecycle');
    const enterDetail = (row) => {
      router.push({
        path: '/service/hostRecycling/docDetail',
        query: {
          suborderId: row.suborder_id,
          bkBizId: row.bk_biz_id,
        },
      });
    };

    watch(
      () => userStore.username,
      (username) => {
        recycleForm.value.bk_username = [username];
      },
    );

    const opBtnDisabled = computed(() => {
      return (status) => {
        if (
          ['UNCOMMIT', 'COMMITTED', 'DETECTING', 'FOR_AUDIT', 'TRANSITING', 'RETURNING', 'DONE', 'TERMINATE'].includes(
            status,
          )
        ) {
          return true;
        }
        return false;
      };
    });
    const operateColList = [
      {
        label: '操作',
        width: 120,
        showOverflowTooltip: false,
        render: ({ row, index }) => {
          return h('div', { class: 'operation-column' }, [
            withDirectives(
              h(
                Button,
                {
                  text: true,
                  theme: 'primary',
                  class: 'mr10',
                  onClick: () => {
                    returnPreDetails(row);
                  },
                },
                '预检详情',
              ),
              [],
            ),
            withDirectives(
              h(
                Dropdown,
                {
                  trigger: 'click',
                  popoverOptions: {
                    renderType: 'shown',
                    onAfterShow: () => (currentOperateRowIndex.value = index),
                    onAfterHidden: () => (currentOperateRowIndex.value = -1),
                  },
                },
                {
                  default: () =>
                    h(
                      'div',
                      {
                        class: [`more-action${currentOperateRowIndex.value === index ? ' current-operate-row' : ''}`],
                      },
                      h('i', { class: 'hcm-icon bkhcm-icon-more-fill' }),
                    ),
                  content: () =>
                    h(DropdownMenu, null, [
                      withDirectives(
                        h(
                          DropdownItem,
                          {
                            key: 'retry',
                            onClick: () => retryOrderFunc(row.suborder_id, opBtnDisabled.value(row.status)),
                            extCls: `more-action-item${opBtnDisabled.value(row.status) ? ' disabled' : ''}`,
                          },
                          '全部重试',
                        ),
                        [],
                      ),
                      withDirectives(
                        h(
                          DropdownItem,
                          {
                            key: 'stop',
                            onClick: () => stopOrderFunc(row.suborder_id, opBtnDisabled.value(row.status)),
                            extCls: `more-action-item${opBtnDisabled.value(row.status) ? ' disabled' : ''}`,
                          },
                          '全部终止',
                        ),
                        [],
                      ),
                      withDirectives(
                        h(
                          DropdownItem,
                          {
                            key: 'submit',
                            onClick: () => submitOrderFunc(row.suborder_id, opBtnDisabled.value(row.status)),
                            extCls: `more-action-item${opBtnDisabled.value(row.status) ? ' disabled' : ''}`,
                          },
                          '剔除预检失败IP重试',
                        ),
                        [],
                      ),
                    ]),
                },
              ),
              [],
            ),
          ]);
        },
      },
    ];
    // 在第三个加子单号，需要跳转到单据详情，未用到路由
    columns.splice(1, 0, {
      label: '单号/子单号',
      width: 100,
      render: ({ row }) => {
        return (
          <div>
            <div>
              <p>{row.order_id}</p>
            </div>
            <div>
              <Button theme='primary' text onClick={() => enterDetail(row)}>
                {row.suborder_id}
              </Button>
            </div>
          </div>
        );
      },
    });
    const tableColumns = [...columns, ...operateColList];
    const { CommonTable, getListData, dataList, pagination } = useTable({
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
        path: '/service/hostRecycling/preDetail',
        query: {
          suborder_id: row.suborder_id,
        },
      });
    };
    const returnRecyclingResources = () => {
      router.push({
        path: '/service/hostRecycling/resources',
        query: { ...route.query },
      });
    };
    const businessRef = ref(null);
    const renderNodes = () => {
      return (
        <div class={'apply-list-container'}>
          <div class={'filter-container'}>
            <Form model={recycleForm} formType='vertical' class={'scr-form-wrapper'}>
              <FormItem label='业务'>
                <BusinessSelector
                  ref={businessRef}
                  v-model={recycleForm.value.bk_biz_id}
                  autoSelect
                  authed
                  clearable={false}
                  isShowAll
                  notAutoSelectAll
                  multiple
                  saveBizs
                  bizsKey='scr_host_bizs'
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
                <Select v-model={recycleForm.value.resource_type} multiple clearable placeholder='请选择资源类型'>
                  {resourceTypeList.map(({ key, value }) => {
                    return <Select.Option key={key} name={value} id={key} />;
                  })}
                </Select>
              </FormItem>
              <FormItem label='回收类型'>
                <Select v-model={recycleForm.value.return_plan} multiple clearable placeholder='请选择回收类型'>
                  {returnPlanList.map(({ key, value }) => {
                    return <Select.Option key={key} name={value} id={key}></Select.Option>;
                  })}
                </Select>
              </FormItem>
              <FormItem label='状态'>
                <Select v-model={recycleForm.value.stage} multiple clearable placeholder='请选择状态'>
                  {stageList.value.map(({ stage, description }) => {
                    return <Select.Option key={stage} name={description} id={stage}></Select.Option>;
                  })}
                </Select>
              </FormItem>
              <FormItem label='回收人'>
                <member-select
                  v-model={recycleForm.value.bk_username}
                  multiple
                  clearable
                  defaultUserlist={[
                    {
                      username: userStore.username,
                      display_name: userStore.username,
                    },
                  ]}
                  placeholder='请输入企业微信名'
                />
              </FormItem>
              <FormItem label='回收时间'>
                <bk-date-picker v-model={timeForm.value} type='daterange' />
              </FormItem>
            </Form>
            <div class='btn-container'>
              <Button theme='primary' onClick={filterOrders}>
                <Search />
                查询
              </Button>
              <Button onClick={() => clearFilter()}>重置</Button>
            </div>
          </div>
          <div class='btn-container oper-btn-pad'>
            <Button theme='primary' onClick={returnRecyclingResources}>
              <Plus />
              回收资源
            </Button>
            <export-to-excel-button data={dataList.value} columns={tableColumns} filename='回收单据列表' />
            <Button disabled={!selections.value.length} onClick={goToPrecheck}>
              批量查看预检详情
            </Button>
            <Button disabled={!selections.value.length} onClick={() => retryOrderFunc('isBatch', false)}>
              批量重试
            </Button>
            <Button disabled={!selections.value.length} onClick={() => submitOrderFunc('isBatch', false)}>
              批量去除预检失败IP提交
            </Button>
          </div>
          <CommonTable class={'filter-common-table'} />
        </div>
      );
    };
    onMounted(() => {
      fetchStageList();
    });

    watch(
      () => businessRef.value?.businessList,
      (val) => {
        if (!val?.length) return;
        getListData();
      },
      { deep: true },
    );

    return renderNodes;
  },
});
