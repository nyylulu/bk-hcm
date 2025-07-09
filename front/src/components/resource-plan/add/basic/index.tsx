import { defineComponent, type PropType, onBeforeMount, ref, watch, nextTick, computed } from 'vue';
import dayjs from 'dayjs';
import isBetween from 'dayjs/plugin/isBetween';
import isoWeek from 'dayjs/plugin/isoWeek';
import { useI18n } from 'vue-i18n';
import Panel from '@/components/panel';
import { useResourcePlanStore } from '@/store';
import cssModule from './index.module.scss';
import usePlanStore from '@/store/usePlanStore';
import type { IPlanTicketDemand, IRegion, IZone } from '@/typings/resourcePlan';
import { timeFormatter } from '@/common/util';
import { AdjustType, IExceptTimeRange } from '@/typings/plan';
import { isDateInRange } from '@/utils/plan';
import useFormModel from '@/hooks/useFormModel';
import { isEqual } from 'lodash';

dayjs.extend(isBetween);
dayjs.extend(isoWeek);

export default defineComponent({
  props: {
    planTicketDemand: Object as PropType<IPlanTicketDemand>,
    resourceType: String,
    type: String as PropType<AdjustType>,
    submitTooltips: Object as PropType<string | { content: string; disabled: boolean }>,
    originPlanTicketDemand: Object as PropType<IPlanTicketDemand>,
  },

  emits: [
    'update:planTicketDemand',
    'update:resourceType',
    'update:submitTooltips',
    'update:isSubmitDisabled',
    'disableTypeChange',
  ],

  setup(props, { emit, expose }) {
    const planStore = usePlanStore();
    const { t } = useI18n();
    const resourcePlanStore = useResourcePlanStore();
    const { formModel: timeRange, setFormValues: setTimeRange } = useFormModel<IExceptTimeRange>({
      year_month_week: null,
      date_range_in_week: null,
      date_range_in_month: null,
    });
    const timeStrictRange = computed(() => ({
      start: timeRange.date_range_in_week?.start || '',
      end: timeRange.date_range_in_week?.end || '',
    }));

    const projectTypes = ref<string[]>([]);
    const regions = ref<IRegion[]>([]);
    const zones = ref<IZone[]>([]);
    const sources = ref<string[]>([]);
    const formRef = ref();
    const isLoadingProjectType = ref(false);
    const isLoadingRegion = ref(false);
    const isLoadingZone = ref(false);
    const isLoadingSource = ref(false);

    const isThirteenDate = computed(() => {
      const expectedDate = dayjs(props.planTicketDemand.expect_time);
      const futureDate = dayjs().add(13, 'week');
      return expectedDate.isSame(futureDate, 'day');
    });

    const DISABLE_TIME_KEYS = ['obs_project', 'region_id', 'zone_id'];
    const DISABLE_CONFIG_KEY = 'expect_time';

    const handleUpdatePlanTicketDemand = (key: string, value: unknown) => {
      const newDemand = { ...props.planTicketDemand, [key]: value };

      if (props.type === AdjustType.none) {
        emit('update:planTicketDemand', newDemand);
        return;
      }

      const isValueChanged = !isEqual(props.originPlanTicketDemand[key], value);

      if (isValueChanged) {
        if (DISABLE_TIME_KEYS.includes(key)) {
          emit('disableTypeChange', AdjustType.time);
        } else if (key === DISABLE_CONFIG_KEY) {
          emit('disableTypeChange', AdjustType.config);
        }
      } else if (isEqual(newDemand, props.originPlanTicketDemand)) {
        emit('disableTypeChange', AdjustType.none);
      }

      emit('update:planTicketDemand', newDemand);
    };

    const handleUpdateResourceType = (value: string) => {
      handleUpdatePlanTicketDemand('demand_res_types', value === 'cvm' ? ['CVM', 'CBS'] : ['CBS']);
      emit('update:resourceType', value);
    };

    const handleChooseZone = (id: string) => {
      const zone = zones.value.find((zone) => zone.zone_id === id);
      nextTick(() => {
        handleUpdatePlanTicketDemand('zone_id', zone?.zone_id || '');

        nextTick(() => {
          handleUpdatePlanTicketDemand('zone_name', zone?.zone_name || '');
        });
      });
    };

    const handleChooseRegion = (id: string) => {
      const region = regions.value.find((region) => region.region_id === id);
      nextTick(() => {
        handleUpdatePlanTicketDemand('region_id', region?.region_id || '');

        nextTick(() => {
          handleUpdatePlanTicketDemand('region_name', region?.region_name || '');
        });
      });
    };

    const getProjectTypes = () => {
      isLoadingProjectType.value = true;
      resourcePlanStore
        .getObsProjects()
        .then((data: { data: { details: string[] } }) => {
          projectTypes.value = data?.data?.details || [];
        })
        .finally(() => {
          isLoadingProjectType.value = false;
        });
    };

    const getRegions = () => {
      isLoadingRegion.value = true;
      resourcePlanStore
        .getRegions()
        .then((data: { data: { details: IRegion[] } }) => {
          regions.value = data?.data?.details || [];
        })
        .finally(() => {
          isLoadingRegion.value = false;
        });
    };

    const getZones = () => {
      if (props.planTicketDemand.region_id) {
        isLoadingZone.value = true;
        resourcePlanStore
          .getZones([props.planTicketDemand.region_id])
          .then((data: { data: { details: IZone[] } }) => {
            zones.value = data?.data?.details || [];
          })
          .finally(() => {
            isLoadingZone.value = false;
          });
      } else {
        zones.value = [];
      }
    };

    const getSources = () => {
      isLoadingSource.value = true;
      resourcePlanStore
        .getSources()
        .then((data: { data: { details: string[] } }) => {
          sources.value = data?.data?.details || [];
        })
        .finally(() => {
          isLoadingSource.value = false;
        });
    };

    const getDisabledDate = (date: Date) => {
      const currentDate = dayjs(date);

      const startOfWeek = dayjs().startOf('isoWeek');
      const endOfWeek = dayjs().endOf('isoWeek');

      // 检查给定日期是否在本周内
      if (currentDate.isBetween(startOfWeek, endOfWeek, 'day', '[]')) {
        return false;
      }
      return dayjs(currentDate).isBefore(dayjs());
    };

    const validate = () => {
      return formRef.value.validate();
    };

    const clearValidate = () => {
      return formRef.value?.clearValidate();
    };

    const handleDateWithThirteen = () => {
      if (isThirteenDate.value) {
        return;
      }
      handleUpdatePlanTicketDemand('expect_time', dayjs().add(13, 'week').format('YYYY-MM-DD'));
    };

    watch(
      () => props.planTicketDemand.region_id,
      () => {
        handleUpdatePlanTicketDemand('zone_id', '');
        getZones();
      },
    );

    // 当存在初始的region_id和zone_id时，回填对应的name
    watch(regions, () => {
      if (props.planTicketDemand.region_id) {
        handleChooseRegion(props.planTicketDemand.region_id);
      }
    });
    watch(zones, () => {
      if (props.planTicketDemand.zone_id) {
        handleChooseZone(props.planTicketDemand.zone_id);
      }
    });

    const isExpectTimeTipsReady = ref(false);
    watch(
      () => props.planTicketDemand.expect_time,
      async (time) => {
        if (!time) {
          isExpectTimeTipsReady.value = false;
          return;
        }
        // 当前日期的13周后日期
        const expect_time = timeFormatter(time, 'YYYY-MM-DD');

        const { data } = await planStore.get_demand_available_time(expect_time);
        isExpectTimeTipsReady.value = true;

        // data 复制到 timeRange
        setTimeRange(data);
        // 判断当前日期是否在周期内
        const disabledWithDateRange = isDateInRange(timeFormatter(time, 'YYYY-MM-DD'), timeStrictRange.value);

        emit('update:submitTooltips', {
          content: t(
            `日期落在${timeRange.year_month_week?.year}年${timeRange.year_month_week?.month}月W${timeRange.year_month_week?.week}, 需要选择${timeStrictRange.value.start}~${timeStrictRange.value.end}的日期`,
          ),
          disabled: disabledWithDateRange,
        });
        emit('update:isSubmitDisabled', !disabledWithDateRange);
      },
      {
        immediate: true,
      },
    );

    onBeforeMount(() => {
      getProjectTypes();
      getZones();
      getRegions();
      getSources();
    });

    expose({
      validate,
      clearValidate,
    });

    return () => (
      <Panel title={t('基础信息')} noShadow>
        <bk-form form-type='vertical' ref={formRef} model={props.planTicketDemand} class={cssModule.home}>
          <bk-form-item label={t('资源类型')}>
            <bk-radio-group
              modelValue={props.resourceType}
              onChange={handleUpdateResourceType}
              disabled={props.type !== AdjustType.none}>
              <bk-radio-button label='cvm'>CVM</bk-radio-button>
              <bk-radio-button label='cbs'>CBS</bk-radio-button>
            </bk-radio-group>
          </bk-form-item>
          <bk-form-item label={t('项目类型')} property='obs_project' required>
            <bk-select
              disabled={props.type === AdjustType.time}
              clearable
              loading={isLoadingProjectType.value}
              modelValue={props.planTicketDemand.obs_project}
              onChange={(val: string) => handleUpdatePlanTicketDemand('obs_project', val)}>
              {projectTypes.value.map((projectType) => (
                <bk-option id={projectType} name={projectType}></bk-option>
              ))}
            </bk-select>
          </bk-form-item>
          <bk-form-item label={t('城市')} property='region_id' required>
            <bk-select
              disabled={props.type === AdjustType.time}
              clearable
              loading={isLoadingRegion.value}
              modelValue={props.planTicketDemand.region_id}
              onChange={(val: string) => handleChooseRegion(val)}>
              {regions.value.map((region) => (
                <bk-option id={region.region_id} name={region.region_name}></bk-option>
              ))}
            </bk-select>
          </bk-form-item>
          <bk-form-item label={t('可用区')} property='zone_id'>
            <bk-select
              disabled={props.type === AdjustType.time}
              clearable
              loading={isLoadingZone.value}
              modelValue={props.planTicketDemand.zone_id}
              onChange={(val: string) => handleChooseZone(val)}>
              {zones.value.map((zone) => (
                <bk-option id={zone.zone_id} name={zone.zone_name}></bk-option>
              ))}
            </bk-select>
          </bk-form-item>
          <bk-form-item label={t('期望到货日期')} property='expect_time' required>
            <bk-date-picker
              disabled={props.type === AdjustType.config}
              clearable
              modelValue={props.planTicketDemand.expect_time}
              disabledDate={getDisabledDate}
              onChange={(val: string) => handleUpdatePlanTicketDemand('expect_time', val)}>
              {{
                footer: () => (
                  <div
                    class={[`${isThirteenDate.value ? 'is-thirteen' : ''}`, cssModule['in-thirteen-weeks']]}
                    onClick={handleDateWithThirteen}>
                    {t('13周后')}
                  </div>
                ),
              }}
            </bk-date-picker>
            <p v-show={isExpectTimeTipsReady.value} class={cssModule['plan-mod-timepicker-tip']}>
              {t('注意：日期落在')}
              <span class={cssModule['time-txt']}>
                {t(
                  `${timeRange?.year_month_week?.year}年${timeRange?.year_month_week?.month}月W${timeRange?.year_month_week?.week}`,
                )}
              </span>
              {t(',需要在')}
              <span class={cssModule['time-txt']}>
                {t(`${timeStrictRange.value?.start}~${timeStrictRange.value?.end}`)}
              </span>
              {t('之间申领，超过')}
              <span class={cssModule['time-txt']}>{t(`${timeRange.date_range_in_month?.end}`)}</span>
              {t('将无法申领')}
            </p>
          </bk-form-item>
          {/* 变更原因、需求备注仅和单据绑定，编辑时无需更改 */}
          {props.type === AdjustType.none && (
            <>
              <bk-form-item label={t('变更原因')} property='demand_source'>
                <bk-select
                  clearable={false}
                  loading={isLoadingSource.value}
                  modelValue={props.planTicketDemand.demand_source}
                  onChange={(val: string) => handleUpdatePlanTicketDemand('demand_source', val)}>
                  {sources.value.map((source) => (
                    <bk-option id={source} name={source}></bk-option>
                  ))}
                </bk-select>
              </bk-form-item>
              <bk-form-item label={t('需求备注')} property='remark' class={cssModule['span-2']}>
                <bk-input
                  clearable
                  type='textarea'
                  maxlength={100}
                  showWordLimit
                  modelValue={props.planTicketDemand.remark}
                  onChange={(val: string) => handleUpdatePlanTicketDemand('remark', val)}
                />
              </bk-form-item>
            </>
          )}
        </bk-form>
      </Panel>
    );
  },
});
