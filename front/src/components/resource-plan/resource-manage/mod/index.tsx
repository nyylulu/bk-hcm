import { computed, defineComponent, onBeforeUnmount, onMounted, reactive, Ref, ref, watch } from 'vue';
import './index.scss';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import { Button, DatePicker, Dialog, Form, InfoBox, Message, PopConfirm, Table } from 'bkui-vue';
import { cloneDeep } from 'lodash';
import useScrColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import useFormModel from '@/hooks/useFormModel';
import usePlanStore from '@/store/usePlanStore';
import { IPlanTicketDemand } from '@/typings/resourcePlan';
import EditPlan from '../../add';
import { AdjustType, IExceptTimeRange } from '@/typings/plan';
import { useModColumn } from './useModColumn';
import { timeFormatter } from '@/common/util';
import dayjs from 'dayjs';
import Panel from '@/components/panel';
import { isDateInRange } from '@/utils/plan';
import CommonDialog from '@/components/common-dialog';

const { FormItem } = Form;

export default defineComponent({
  props: {
    currentGlobalBusinessId: Number,
  },
  setup(props) {
    const planStore = usePlanStore();
    const tableData = ref([]);
    const originData = ref([]);
    const { generateColumnsSettings } = useScrColumns('planDemandModColumns');
    const columns = useModColumn(originData);
    const router = useRouter();
    const route = useRoute();
    const { t } = useI18n();
    const curEditData: Ref<IPlanTicketDemand> = ref<IPlanTicketDemand>({});
    let curEditOriginData: IPlanTicketDemand;
    const { handleSelectionChange, selections, resetSelections } = useSelection();
    const isRemoveDialogShow = ref(false);
    const isDelayDialogShow = ref(false);
    const planUpdateSidesliderState = reactive({ isShow: false, isHidden: true });
    const tableRef = ref();
    const delayFormRef = ref();
    const { formModel } = useFormModel({
      time: dayjs().add(13, 'week').format('YYYY-MM-DD'),
    });
    const { formModel: timeRange, setFormValues: setTimeRange } = useFormModel<IExceptTimeRange>({});
    const timeStrictRange = computed(() => ({
      start: timeRange.date_range_in_week?.start || '',
      end: timeRange.date_range_in_week?.end || '',
    }));
    const tableColumns = [
      ...columns,
      {
        label: '操作',
        field: 'actions',
        isDefaultShow: true,
        render: ({ data }: any) => (
          <div>
            <Button
              text
              theme='primary'
              class={'mr8'}
              onClick={() => {
                const idx = tableData.value.findIndex(({ demand_id }) => demand_id === data.demand_id);
                curEditData.value = planStore.convertToPlanTicketDemand(tableData.value[idx]);
                curEditOriginData = planStore.convertToPlanTicketDemand(
                  originData.value.find(({ demand_id }) => demand_id === data.demand_id),
                );
                resetSelections();
                planUpdateSidesliderState.isHidden = false;
                planUpdateSidesliderState.isShow = true;
              }}>
              编辑
            </Button>
            <PopConfirm
              content={t('移除操作无法撤回，请谨慎操作')}
              title={t('确认移除该条数据？')}
              width={288}
              trigger='click'
              onConfirm={() => {
                const idx = tableData.value.findIndex(({ demand_id }) => demand_id === data.demand_id);
                tableData.value.splice(idx, 1);
                originData.value.splice(idx, 1);
              }}>
              <Button text theme='primary'>
                移除
              </Button>
            </PopConfirm>
          </div>
        ),
      },
    ];
    const tableSettings = generateColumnsSettings(tableColumns);

    const toRemoveSum = computed(() =>
      tableData.value.reduce((acc, cur) => {
        if (cur.adjustType === AdjustType.none) acc += 1;
        return acc;
      }, 0),
    );

    const computeDifference = (dataKey: string) =>
      computed(() => {
        const origin = originData.value.reduce((acc, cur) => {
          acc += cur[dataKey];
          return acc;
        }, 0);
        const cur = tableData.value.reduce((acc, cur) => {
          acc += cur[dataKey];
          return acc;
        }, 0);
        if (origin === cur) return <span>{t('无变动')}</span>;
        return (
          <span class={'adjust-txt'}>
            <span>调整前 {origin}</span>
            <span class={'ml16'}>调整后 {cur}</span>
          </span>
        );
      });

    const computedTotalCpus = computeDifference('remained_cpu_core');
    const computedTotalMemory = computeDifference('remained_memory');
    const computedTotalDiskGB = computeDifference('remained_disk_size');

    const clearSelection = () => {
      tableRef?.value?.clearSelection();
      resetSelections();
    };

    onBeforeRouteLeave((_to, _from, next) => {
      if (['BizInvoiceResourceDetail'].includes(_to.name as string)) next();
      else
        InfoBox({
          title: '确定离开当前页面?',
          subTitle: '离开会导致编辑的内容丢失',
          onConfirm: () => next(),
        });
    });

    function confirmLeave(event: any) {
      (event || window.event).returnValue = '关闭提示';
      return '关闭提示';
    }

    onBeforeUnmount(() => {
      window.removeEventListener('beforeunload', confirmLeave);
    });

    onMounted(async () => {
      window.addEventListener('beforeunload', confirmLeave);
    });

    watch(
      () => route.query,
      async () => {
        const planIds = (route.query.planIds as string)?.split(',').map((v) => v) || [];
        const timeRange = {
          start: route.query.start as string,
          end: route.query.end as string,
        };
        const res = await planStore.list_biz_resource_plan_demand(planIds, timeRange);
        const data = res.data.details.map((v) => {
          return {
            ...v,
            adjustType: AdjustType.none,
            res_mode: '按机型',
          };
        });
        originData.value = data;
        tableData.value = cloneDeep(data);
      },
      {
        immediate: true,
      },
    );

    // 提交
    const handleSubmit = async () => {
      const N = tableData.value.length;
      const adjusts = [];
      for (let i = 0; i < N; i++) {
        const originDetail = originData.value[i];
        const updatedDetail = tableData.value[i];
        const info = planStore.convertToAdjust(originDetail, updatedDetail);
        adjusts.push(info);
      }
      const { data } = await planStore.adjust_biz_resource_plan_demand({ adjusts });
      if (!data.id) return;
      Message({
        message: t('提交成功'),
        theme: 'success',
      });
      router.push({
        path: '/business/applications/resource-plan/detail',
        query: {
          id: data.id,
        },
      });
    };

    // 一键移除未修改
    const handleRemoveAll = () => {
      isRemoveDialogShow.value = true;
    };
    // 确定一键移除未修改
    const handleSubmitRemoveAll = () => {
      const idxs = tableData.value.reduce((acc, cur, idx) => {
        if (cur.adjustType === AdjustType.none) acc.push(idx);
        return acc;
      }, []);
      tableData.value = tableData.value.filter((v) => v.adjustType !== AdjustType.none);
      originData.value = originData.value.filter((_, idx) => !idxs.includes(idx));
      isRemoveDialogShow.value = false;
    };

    const isCurRowSelectable = ({ row }: any) => {
      return ![AdjustType.config].includes(row?.adjustType);
    };

    watch(
      () => formModel.time,
      async (time) => {
        if (!time) return;
        const expect_time = timeFormatter(time, 'YYYY-MM-DD');
        const { data } = await planStore.get_demand_available_time(expect_time);
        setTimeRange(data);
      },
      {
        deep: true,
        immediate: true,
      },
    );

    return () => (
      <div class={'plan-mod-container'}>
        <DetailHeader>调整预测需求</DetailHeader>

        <section class={'mod-panel-wrapper'}>
          <Panel class={'plan-mod-table-wrapper'} title={`${t('调整预测需求')}`}>
            <div>
              <section class={'mb16'}>
                <Button
                  theme='primary'
                  onClick={handleRemoveAll}
                  disabled={toRemoveSum.value === 0}
                  v-bk-tooltips={{
                    content: `${t('不存在未修改数据')}`,
                    disabled: toRemoveSum.value > 0,
                  }}>
                  {t('一键移除未修改')}
                </Button>
                <Button
                  class={'ml16'}
                  disabled={!selections.value.filter((v) => ![AdjustType.config].includes(v.adjustType))?.length}
                  onClick={() => {
                    isDelayDialogShow.value = true;
                  }}>
                  {t('批量延期')}
                </Button>
              </section>
              <Table
                ref={tableRef}
                key={'demand_id'}
                rowKey={'demand_id'}
                columns={tableColumns}
                data={tableData.value}
                settings={tableSettings.value}
                isRowSelectEnable={isCurRowSelectable}
                onSelectionChange={(selections: any) => handleSelectionChange(selections, isCurRowSelectable)}
                onSelectAll={(selections: any) => handleSelectionChange(selections, isCurRowSelectable, true)}
              />
            </div>
          </Panel>

          <Panel class={'plan-mod-statistics-panel'} title={`${t('修改预览')}`}>
            <div class={'statistics'}>
              <p class={'item'}>
                <span class={'label'}>{t('CPU调整数(核)：')}</span>
                <span class={'value'}>{computedTotalCpus.value}</span>
              </p>
              <p class={'item'}>
                <span class={'label'}>{t('内存调整量(GB)：')}</span>
                <span class={'value'}>{computedTotalMemory.value}</span>
              </p>
              <p class={'item'}>
                <span class={'label'}>{t('云硬盘调整量(GB)：')}</span>
                <span class={'value'}>{computedTotalDiskGB.value}</span>
              </p>
            </div>
          </Panel>

          <div class={'plan-mod-operation-bar'}>
            <Button
              theme='primary'
              class={'mr8'}
              disabled={toRemoveSum.value > 0}
              onClick={handleSubmit}
              v-bk-tooltips={{
                content: '预测需求未全部修改完，未修改的请手动移除后再提交',
                disabled: toRemoveSum.value === 0,
              }}>
              {t('提交')}
            </Button>
            <Button class={'mr8'} onClick={() => router.back()}>
              {t('取消')}
            </Button>
          </div>
        </section>

        <Dialog
          title='一键移除未调整数据'
          width={680}
          isShow={isRemoveDialogShow.value}
          onConfirm={handleSubmitRemoveAll}
          onClosed={() => (isRemoveDialogShow.value = false)}>
          <span>
            未调整数据有<span class={'plan-mod-remove-num-txt'}> {toRemoveSum.value} </span>
            个，此操作将批量移除表格中未调整数据
          </span>
        </Dialog>

        <CommonDialog title='批量延期' width={680} v-model:isShow={isDelayDialogShow.value}>
          {{
            default: () => (
              <Form model={formModel} ref={delayFormRef}>
                <FormItem label={t('已选预测需求：')}>
                  {selections.value.filter((v) => ![AdjustType.config].includes(v.adjustType)).length}
                </FormItem>
                <FormItem label={t('期望到货日期：')} required property='time'>
                  <div>
                    <DatePicker
                      v-model={formModel.time}
                      appendToBody
                      clearable
                      disabledDate={(date) => dayjs(date).isBefore(dayjs())}
                    />
                    {!!Object.keys(timeRange).length && (
                      <p class={'plan-mod-timepicker-tip'}>
                        注意：日期落在
                        <span class={'time-txt'}>
                          {timeRange.year_month_week.year}年{timeRange.year_month_week.month}月W
                          {timeRange.year_month_week.week}
                        </span>
                        ，需要在
                        <span class={'time-txt'}>
                          {timeStrictRange.value.start}~{timeStrictRange.value.end}
                        </span>
                        之间申领，超过
                        <span class={'time-txt'}>{timeRange.date_range_in_month.end}</span>将无法申领
                      </p>
                    )}
                  </div>
                </FormItem>
              </Form>
            ),
            footer: () => (
              <>
                <Button
                  theme='primary'
                  class={'mr8'}
                  onClick={async () => {
                    await delayFormRef.value.validate();
                    const ids = selections.value
                      .filter((v) => ![AdjustType.config].includes(v.adjustType))
                      .map((v) => v.demand_id);
                    tableData.value = tableData.value.map((v) => {
                      if (ids.includes(v.demand_id))
                        return {
                          ...v,
                          expect_time: timeFormatter(formModel.time, 'YYYY-MM-DD'),
                          adjustType: AdjustType.time,
                        };
                      return v;
                    });
                    isDelayDialogShow.value = false;
                    clearSelection();
                  }}
                  v-bk-tooltips={{
                    content: `日期落在${timeRange.year_month_week?.year}年${timeRange.year_month_week?.month}月W${timeRange.year_month_week?.week},需要在${timeStrictRange.value.start}~${timeStrictRange.value.end}间申领`,
                    disabled: isDateInRange(timeFormatter(formModel.time, 'YYYY-MM-DD'), timeStrictRange.value),
                  }}
                  disabled={!isDateInRange(timeFormatter(formModel.time, 'YYYY-MM-DD'), timeStrictRange.value)}>
                  提交
                </Button>
                <Button
                  onClick={() => {
                    isDelayDialogShow.value = false;
                    clearSelection();
                  }}>
                  取消
                </Button>
              </>
            ),
          }}
        </CommonDialog>

        {!planUpdateSidesliderState.isHidden && (
          <EditPlan
            v-model={curEditData.value}
            v-model:isShow={planUpdateSidesliderState.isShow}
            isEdit
            initDemand={curEditData.value}
            originDemand={curEditOriginData}
            currentGlobalBusinessId={props.currentGlobalBusinessId}
            onUpdateDemand={(val) => {
              const idx = tableData.value.findIndex(({ demand_id }) => demand_id === val.demand_id);
              const originItem = tableData.value[idx];
              const tmp = planStore.convertToDemandListDetail(val, originItem);
              tableData.value.splice(idx, 1, tmp);
            }}
            onHidden={() => (planUpdateSidesliderState.isHidden = true)}
          />
        )}
      </div>
    );
  },
});
