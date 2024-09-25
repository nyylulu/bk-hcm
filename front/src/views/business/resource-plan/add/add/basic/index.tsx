import dayjs from 'dayjs';
import { defineComponent, type PropType, onBeforeMount, ref, watch, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import Panel from '@/components/panel';
import { useResourcePlanStore } from '@/store';
import cssModule from './index.module.scss';

import type { IPlanTicketDemand, IRegion, IZone } from '@/typings/resourcePlan';

export default defineComponent({
  props: {
    planTicketDemand: Object as PropType<IPlanTicketDemand>,
    resourceType: String,
  },

  emits: ['update:planTicketDemand', 'update:resourceType'],

  setup(props, { emit, expose }) {
    const { t } = useI18n();
    const resourcePlanStore = useResourcePlanStore();

    const projectTypes = ref<string[]>([]);
    const regions = ref<IRegion[]>([]);
    const zones = ref<IZone[]>([]);
    const sources = ref<string[]>([]);
    const formRef = ref();
    const isLoadingProjectType = ref(false);
    const isLoadingRegion = ref(false);
    const isLoadingZone = ref(false);
    const isLoadingSource = ref(false);

    const handleUpdatePlanTicketDemand = (key: string, value: unknown) => {
      emit('update:planTicketDemand', {
        ...props.planTicketDemand,
        [key]: value,
      });
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

    const getDisabledDate = (date: string) => {
      return dayjs(date).isBefore('2024-10-01');
    };

    const validate = () => {
      return formRef.value.validate();
    };

    const clearValidate = () => {
      return formRef.value?.clearValidate();
    };

    watch(
      () => props.planTicketDemand.region_id,
      () => {
        handleUpdatePlanTicketDemand('zone_id', '');
        getZones();
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
      <Panel title={t('基础信息')}>
        <bk-form form-type='vertical' ref={formRef} model={props.planTicketDemand} class={cssModule.home}>
          <bk-form-item label={t('资源类型')}>
            <bk-radio-group modelValue={props.resourceType} onChange={handleUpdateResourceType}>
              <bk-radio-button label='cvm'>CVM</bk-radio-button>
              <bk-radio-button label='cbs'>CBS</bk-radio-button>
            </bk-radio-group>
          </bk-form-item>
          <bk-form-item label={t('项目类型')} property='obs_project' required>
            <bk-select
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
              clearable
              modelValue={props.planTicketDemand.expect_time}
              disabledDate={getDisabledDate}
              onChange={(val: string) => handleUpdatePlanTicketDemand('expect_time', val)}
            />
          </bk-form-item>
          <bk-form-item label={t('变更原因')} property='demand_source'>
            <bk-select
              clearable
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
        </bk-form>
      </Panel>
    );
  },
});
