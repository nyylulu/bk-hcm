import { computed, defineComponent, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import cssModule from './index.module.scss';

import { Button, Dropdown, Message, Select } from 'bkui-vue';
import GridFilterComp from '@/components/grid-filter-comp';
import RequireNameSelect from '@/views/ziyanScr/host-recycle/host-recycle-table/require-name-select';
import FloatInput from '@/components/float-input';
import MemberSelect from '@/components/MemberSelect';
import ExportToExcelButton from '@/components/export-to-excel-button';
import ScrDatePicker from '@/components/scr/scr-date-picker';

import { useI18n } from 'vue-i18n';
import { useUserStore } from '@/store';
import { QueryRuleOPEnum } from '@/typings';
import { useTable } from '@/hooks/useTable/useTable';
import useScrColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import { removeEmptyFields } from '@/utils/scr/remove-query-fields';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import useSearchQs from '@/hooks/use-search-qs';
import http from '@/http';
import { getEntirePath } from '@/utils';
import { getRecycleStageOpts } from '@/api/host/recycle';
import { useSaveSearchRules } from '../../useSaveSearchRules';
import useFormModel from '@/hooks/useFormModel';
import { applicationTime } from '@/common/util';

export default defineComponent({
  setup() {
    const router = useRouter();
    const userStore = useUserStore();
    const { getBusinessApiPath, getBizsId } = useWhereAmI();
    const { t } = useI18n();
    const route = useRoute();

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
        dateRange: applicationTime(),
      };
    };
    const { formModel, resetForm } = useFormModel(defaultRecycleForm());
    const currentOperateRowIndex = ref(-1);

    const opBtnDisabled = computed(() => {
      return (status: any) =>
        ['UNCOMMIT', 'COMMITTED', 'DETECTING', 'FOR_AUDIT', 'TRANSITING', 'RETURNING', 'DONE', 'TERMINATE'].includes(
          status,
        );
    });
    const requestListParams = computed(() => {
      const params = {
        ...formModel,
        start: formModel.dateRange[0],
        end: formModel.dateRange[1],
        bk_biz_id: [getBizsId()],
      };
      params.order_id = params.order_id.length ? params.order_id.map((v) => +v) : [];
      params.dateRange = undefined;
      removeEmptyFields(params);
      return params;
    });

    const { columns } = useScrColumns('hostRecycleApplication');
    columns.splice(1, 0, {
      label: t('单号/子单号'),
      width: 100,
      render: ({ row }: any) => {
        return (
          <>
            <p>{row.order_id}</p>
            <div>
              <Button theme='primary' text onClick={() => enterDetail(row)}>
                {row.suborder_id}
              </Button>
            </div>
          </>
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
      router.push({
        name: 'HostRecycleDocDetail',
        query: { ...route.query, suborderId: row.suborder_id, bkBizId: getBizsId() },
      });
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

    const searchRulesKey = 'host_recycle_applications_rules';
    const searchQs = useSearchQs({
      key: 'initial_filter',
      properties: [
        { id: 'requireType', type: 'number', name: 'requireType', op: QueryRuleOPEnum.IN },
        { id: 'orderId', type: 'number', name: 'orderId', op: QueryRuleOPEnum.IN },
        { id: 'suborder_id', type: 'number', name: 'suborder_id', op: QueryRuleOPEnum.IN },
      ],
    });
    const filterOrders = (searchRulesStr?: string) => {
      // 合并默认条件值
      Object.assign(formModel, searchQs.get(route.query));
      // 回填
      if (searchRulesStr) {
        // 解决人员选择器搜索问题
        formModel.bk_username.length > 0 &&
          userStore.setMemberDefaultList([...new Set([...userStore.memberDefaultList, ...formModel.bk_username])]);
      }
      formModel.bk_biz_id = [getBizsId()];
      pagination.start = 0;
      getListData();
    };
    const { saveSearchRules, clearSearchRules } = useSaveSearchRules(searchRulesKey, filterOrders, formModel);

    const handleSearch = () => {
      // update query
      saveSearchRules();
    };

    const handleReset = () => {
      resetForm(defaultRecycleForm());
      formModel.bk_biz_id = [getBizsId()];
      // update query
      clearSearchRules();
    };

    watch(
      () => userStore.username,
      (username) => {
        if (route.query[searchRulesKey]) return;
        // 无搜索记录，设置申请人默认值
        formModel.bk_username = [username];
      },
    );

    onMounted(() => {
      fetchStageList();
    });

    return () => (
      <>
        <GridFilterComp
          onSearch={handleSearch}
          onReset={handleReset}
          loading={isLoading.value}
          col={4}
          rules={[
            {
              title: t('需求类型'),
              content: <RequireNameSelect v-model={formModel.recycle_type} multiple clearable collapseTags />,
            },
            {
              title: t('单号'),
              content: <FloatInput v-model={formModel.order_id} placeholder={t('请输入单号，多个换行分割')} />,
            },
            {
              title: t('子单号'),
              content: <FloatInput v-model={formModel.suborder_id} placeholder={t('请输入子单号，多个换行分割')} />,
            },
            {
              title: t('资源类型'),
              content: (
                <Select v-model={formModel.resource_type} multiple clearable placeholder={t('请选择资源类型')}>
                  {resourceTypeList.map(({ key, value }) => {
                    return <Select.Option key={key} name={value} id={key} />;
                  })}
                </Select>
              ),
            },
            {
              title: t('回收类型'),
              content: (
                <Select v-model={formModel.return_plan} multiple clearable placeholder={t('请选择回收类型')}>
                  {returnPlanList.map(({ key, value }) => {
                    return <Select.Option key={key} name={value} id={key}></Select.Option>;
                  })}
                </Select>
              ),
            },
            {
              title: t('状态'),
              content: (
                <Select v-model={formModel.stage} multiple clearable placeholder={t('请选择状态')}>
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
                  v-model={formModel.bk_username}
                  multiple
                  clearable
                  placeholder={t('请输入企业微信名')}
                  defaultUserlist={userStore.memberDefaultList.map((username) => ({
                    username,
                    display_name: username,
                  }))}
                />
              ),
            },
            {
              title: t('回收时间'),
              content: <ScrDatePicker class='full-width' v-model={formModel.dateRange} />,
            },
          ]}
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
