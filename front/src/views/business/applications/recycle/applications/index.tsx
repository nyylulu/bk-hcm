import { computed, defineComponent, onMounted, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import cssModule from './index.module.scss';

import { Button, DatePicker, Dropdown, Message, Select } from 'bkui-vue';
import GridFilterComp from '@/components/grid-filter-comp';
import ScrCreateFilterSelector from '@/views/ziyanScr/resource-manage/create/ScrCreateFilterSelector';
import FloatInput from '@/components/float-input';
import MemberSelect from '@/components/MemberSelect';
import ExportToExcelButton from '@/components/export-to-excel-button';

import dayjs from 'dayjs';
import { useI18n } from 'vue-i18n';
import { useUserStore, useZiyanScrStore } from '@/store';
import { useTable } from '@/hooks/useTable/useTable';
import useScrColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import { removeEmptyFields } from '@/utils/scr/remove-query-fields';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import http from '@/http';
import { getEntirePath } from '@/utils';
import { getRecycleStageOpts } from '@/api/host/recycle';

export default defineComponent({
  setup() {
    const router = useRouter();
    const userStore = useUserStore();
    const scrStore = useZiyanScrStore();
    const { getBusinessApiPath, getBizsId } = useWhereAmI();
    const { t } = useI18n();

    const resourceTypeList = [
      { key: 'QCLOUDCVM', value: '腾讯云虚拟机' },
      { key: 'IDCPM', value: 'IDC物理机' },
      { key: 'OTHERS', value: '其他' },
    ];
    const returnPlanList = [
      { key: 'IMMEDIATE', value: '立即销毁' },
      { key: 'DELAY', value: '延迟销毁' },
    ];

    const defaultRecycleForm = () => {
      return {
        bk_biz_id: [] as any[],
        order_id: [] as any[],
        suborder_id: [] as any[],
        resource_type: [] as any[],
        recycle_type: [] as any[],
        return_plan: [] as any[],
        stage: [] as any[],
        bk_username: [userStore.username],
      };
    };
    const defaultTime = () => [new Date(dayjs().subtract(30, 'day').format('YYYY-MM-DD')), new Date()];
    const recycleForm = ref(defaultRecycleForm());
    const timeForm = ref(defaultTime());
    const pageInfo = ref({ start: 0, limit: 10, enable_count: false });
    const currentOperateRowIndex = ref(-1);

    const opBtnDisabled = computed(() => {
      return (status: any) =>
        ['UNCOMMIT', 'COMMITTED', 'DETECTING', 'FOR_AUDIT', 'TRANSITING', 'RETURNING', 'DONE', 'TERMINATE'].includes(
          status,
        );
    });
    const handleTime = (time: any) => (!time ? '' : dayjs(time).format('YYYY-MM-DD'));
    const timeObj = computed(() => {
      return {
        start: handleTime(timeForm.value[0]),
        end: handleTime(timeForm.value[1]),
      };
    });
    const requestListParams = computed(() => {
      const params = {
        ...recycleForm.value,
        ...timeObj.value,
        page: pageInfo.value,
        bk_biz_id: [getBizsId()],
      };
      params.order_id = params.order_id.length ? params.order_id.map((v) => +v) : [];
      removeEmptyFields(params);
      return params;
    });

    const { columns } = useScrColumns('hostRecycle');
    columns.splice(1, 1, {
      label: t('单号/子单号'),
      width: 100,
      render: ({ row }: any) => {
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
      exportFormatter: (data: any) => `${data.order_id}/${data.suborder_id}`,
    });
    const { selections, handleSelectionChange } = useSelection();
    const { CommonTable, getListData, dataList, pagination, isLoading } = useTable({
      tableOptions: {
        columns: [
          ...columns,
          {
            label: t('操作'),
            width: 120,
            fixed: 'right',
            render: ({ row, index }: any) => (
              <div class={cssModule['operation-column']}>
                <Button text theme='primary' onClick={() => returnPreDetails(row)}>
                  {t('预检详情')}
                </Button>
                <Dropdown
                  trigger='click'
                  popoverOptions={{
                    renderType: 'shown',
                    onAfterShow: () => (currentOperateRowIndex.value = index),
                    onAfterHidden: () => (currentOperateRowIndex.value = -1),
                  }}>
                  {{
                    default: () => (
                      <div
                        class={[
                          cssModule['more-action'],
                          { [cssModule['current-operate-row']]: currentOperateRowIndex.value === index },
                        ]}>
                        <i class='hcm-icon bkhcm-icon-more-fill'></i>
                      </div>
                    ),
                    content: () => (
                      <Dropdown.DropdownMenu>
                        <Dropdown.DropdownItem
                          key='retry'
                          class={[
                            cssModule['more-action-item'],
                            { [cssModule.disabled]: opBtnDisabled.value(row.status) },
                          ]}
                          onClick={() => retryOrderFunc(row.suborder_id, opBtnDisabled.value(row.status))}>
                          {t('全部重试')}
                        </Dropdown.DropdownItem>
                        <Dropdown.DropdownItem
                          key='stop'
                          class={[
                            cssModule['more-action-item'],
                            { [cssModule.disabled]: opBtnDisabled.value(row.status) },
                          ]}
                          onClick={() => stopOrderFunc(row.suborder_id, opBtnDisabled.value(row.status))}>
                          {t('全部终止')}
                        </Dropdown.DropdownItem>
                        <Dropdown.DropdownItem
                          key='submit'
                          class={[
                            cssModule['more-action-item'],
                            { [cssModule.disabled]: opBtnDisabled.value(row.status) },
                          ]}
                          onClick={() => submitOrderFunc(row.suborder_id, opBtnDisabled.value(row.status))}>
                          {t('剔除预检失败IP重试')}
                        </Dropdown.DropdownItem>
                      </Dropdown.DropdownMenu>
                    ),
                  }}
                </Dropdown>
              </div>
            ),
          },
        ],
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
        sortOption: { sort: 'create_at', order: 'DESC' },
        immediate: false,
      },
      scrConfig: () => {
        return {
          url: `/api/v1/woa/${getBusinessApiPath()}task/findmany/recycle/order`,
          payload: { ...requestListParams.value },
        };
      },
    });
    const enterDetail = (row: any) => {
      router.push({ name: 'HostRecycleDocDetail', query: { suborderId: row.suborder_id, bkBizId: getBizsId() } });
    };
    const returnPreDetails = (row: any) => {
      router.push({ name: 'HostRecyclePreDetail', query: { suborder_id: row.suborder_id } });
    };
    const getBatchSuborderId = () => selections.value.map((item) => item.suborder_id);
    const goToPrecheck = () => {
      router.push({ name: 'HostRecyclePreDetail', query: { suborder_id: getBatchSuborderId().join('\n') } });
    };
    const textTip = (text: string, theme: 'error' | 'success') => {
      const themeDes = { error: t('失败'), success: t('成功') };
      Message({ message: `${text}${themeDes[theme]}`, theme, duration: 1500 });
    };
    const retryOrderFunc = async (id: string, disabled: boolean) => {
      if (disabled) return;
      const suborderId = id === 'isBatch' ? getBatchSuborderId() : [id];
      const res = await http.post(getEntirePath(`${getBusinessApiPath()}task/start/recycle/order`), {
        suborder_id: suborderId,
      });

      if (res.code === 0) {
        textTip(t('重试'), 'success');
        getListData();
      }
    };
    const stopOrderFunc = async (id: string, disabled: boolean) => {
      if (disabled) return;
      const res = await http.post(getEntirePath(`${getBusinessApiPath()}task/terminate/recycle/order`), {
        suborder_id: [id],
      });

      if (res.code === 0) {
        textTip(t('终止'), 'success');
        getListData();
      }
    };
    const submitOrderFunc = async (id: string, disabled: boolean) => {
      if (disabled) return;
      const suborderId = id === 'isBatch' ? getBatchSuborderId() : [id];
      const res = await http.post(getEntirePath(`${getBusinessApiPath()}task/revise/recycle/order`), {
        suborder_id: suborderId,
      });

      if (res.code === 0) {
        textTip(t('去除预检失败IP提交'), 'success');
        getListData();
      }
    };

    const stageList = ref([]);
    const fetchStageList = async () => {
      const data = await getRecycleStageOpts();
      stageList.value = data?.info || [];
    };

    const filterOrders = () => {
      pagination.start = 0;
      recycleForm.value.bk_biz_id = [getBizsId()];
      getListData();
    };
    const clearFilter = () => {
      const initForm = defaultRecycleForm();
      initForm.bk_biz_id = [getBizsId()];
      recycleForm.value = initForm;
      timeForm.value = defaultTime();
      filterOrders();
    };

    onMounted(() => {
      fetchStageList();
    });

    watch(
      () => userStore.username,
      (username) => {
        recycleForm.value.bk_username = [username];
      },
    );

    return () => (
      <>
        <GridFilterComp
          rules={[
            {
              title: t('需求类型'),
              content: (
                <ScrCreateFilterSelector
                  v-model={recycleForm.value.recycle_type}
                  api={scrStore.getRequirementList}
                  multiple
                  optionIdPath='require_type'
                  optionNamePath='require_name'
                />
              ),
            },
            {
              title: t('单号'),
              content: <FloatInput v-model={recycleForm.value.order_id} placeholder={t('请输入单号，多个换行分割')} />,
            },
            {
              title: t('子单号'),
              content: (
                <FloatInput v-model={recycleForm.value.suborder_id} placeholder={t('请输入子单号，多个换行分割')} />
              ),
            },
            {
              title: t('资源类型'),
              content: (
                <Select v-model={recycleForm.value.resource_type} multiple clearable placeholder={t('请选择资源类型')}>
                  {resourceTypeList.map(({ key, value }) => {
                    return <Select.Option key={key} name={value} id={key} />;
                  })}
                </Select>
              ),
            },
            {
              title: t('回收类型'),
              content: (
                <Select v-model={recycleForm.value.return_plan} multiple clearable placeholder={t('请选择回收类型')}>
                  {returnPlanList.map(({ key, value }) => {
                    return <Select.Option key={key} name={value} id={key}></Select.Option>;
                  })}
                </Select>
              ),
            },
            {
              title: t('状态'),
              content: (
                <Select v-model={recycleForm.value.stage} multiple clearable placeholder={t('请选择状态')}>
                  {stageList.value.map(({ stage, description }) => {
                    return <Select.Option key={stage} name={description} id={stage}></Select.Option>;
                  })}
                </Select>
              ),
            },
            {
              title: t('回收人'),
              content: (
                <MemberSelect
                  v-model={recycleForm.value.bk_username}
                  multiple
                  clearable
                  placeholder={t('请输入企业微信名')}
                  defaultUserlist={[{ username: userStore.username, display_name: userStore.username }]}
                />
              ),
            },
            {
              title: t('回收时间'),
              content: <DatePicker class='full-width' v-model={timeForm.value} type='daterange' />,
            },
          ]}
          onSearch={filterOrders}
          onReset={clearFilter}
          loading={isLoading.value}
          col={4}
          immediate
        />
        <section class={cssModule['table-wrapper']}>
          <div class={[cssModule.buttons, cssModule.mb16]}>
            <ExportToExcelButton
              class={cssModule.button}
              data={dataList.value}
              columns={columns}
              text={t('全部导出')}
              filename={t('回收单据列表')}
            />
            <Button class={cssModule.button} disabled={!selections.value.length} onClick={goToPrecheck}>
              {t('批量查看预检详情')}
            </Button>
            <Button
              class={cssModule.button}
              disabled={!selections.value.length}
              onClick={() => retryOrderFunc('isBatch', false)}>
              {t('批量重试')}
            </Button>
            <Button
              class={cssModule.button}
              disabled={!selections.value.length}
              onClick={() => submitOrderFunc('isBatch', false)}>
              {t('批量去除预检失败IP提交')}
            </Button>
          </div>
          <CommonTable style={{ height: 'calc(100% - 48px)' }} />
        </section>
      </>
    );
  },
});
