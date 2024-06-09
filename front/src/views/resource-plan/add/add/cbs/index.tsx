import { defineComponent, ref, type PropType, onBeforeMount, watch } from 'vue';
import Panel from '@/components/panel';
import { useI18n } from 'vue-i18n';
import { useResourcePlanStore } from '@/store';
import cssModule from './index.module.scss';

import type { IPlanTicketDemand, IDiskType } from '@/typings/resourcePlan';

export default defineComponent({
  props: {
    planTicketDemand: Object as PropType<IPlanTicketDemand>,
    resourceType: String,
  },

  emits: ['update:planTicketDemand'],

  setup(props, { emit, expose }) {
    const { t } = useI18n();
    const resourcePlanStore = useResourcePlanStore();

    const formRef = ref();
    const isLoadingDiskTypes = ref(false);
    const diskTypes = ref<IDiskType[]>([]);

    const handleUpdatePlanTicketDemand = (key: string, value: unknown) => {
      emit('update:planTicketDemand', {
        ...props.planTicketDemand,
        cbs: {
          ...props.planTicketDemand.cbs,
          [key]: value,
        },
      });
    };

    const validate = () => {
      return formRef.value.validate();
    };

    const getDiskTypes = () => {
      isLoadingDiskTypes.value = true;
      resourcePlanStore
        .getDiskTypes()
        .then((data: { details: IDiskType[] }) => {
          diskTypes.value = data.details || [];
        })
        .finally(() => {
          isLoadingDiskTypes.value = false;
        });
    };

    const calcDiskSize = () => {
      const num =
        (props.resourceType === 'cvm' ? props.planTicketDemand.cvm.os : props.planTicketDemand.cbs.disk_num) || 0;
      const perSize = props.planTicketDemand.cbs.disk_per_size || 0;
      handleUpdatePlanTicketDemand('disk_size', num * perSize);
    };

    watch(
      [
        () => props.planTicketDemand.cbs.disk_num,
        () => props.planTicketDemand.cbs.disk_per_size,
        () => props.planTicketDemand.cvm.os,
        () => props.resourceType,
      ],
      calcDiskSize,
    );

    onBeforeMount(getDiskTypes);

    expose({
      validate,
    });

    return () => (
      <Panel title={t('CBS云磁盘信息')}>
        <bk-form form-type='vertical' ref={formRef} model={props.planTicketDemand.cbs} class={cssModule.home}>
          <bk-form-item label={t('云盘类型')} property='disk_type' required class={cssModule['span-line']}>
            <bk-select
              clearable
              loading={isLoadingDiskTypes.value}
              modelValue={props.planTicketDemand.cbs.disk_type}
              onChange={(val: string) => handleUpdatePlanTicketDemand('disk_type', val)}>
              {diskTypes.value.map((diskType) => (
                <bk-option id={diskType.disk_type} name={diskType.disk_type_name}></bk-option>
              ))}
            </bk-select>
          </bk-form-item>
          <bk-form-item
            label={t('云磁盘容量/块')}
            property='disk_per_size'
            required
            class={cssModule['span-half-line']}>
            <bk-input
              type='number'
              suffix={'GB'}
              modelValue={props.planTicketDemand.cbs.disk_per_size}
              onChange={(val: number) => handleUpdatePlanTicketDemand('disk_per_size', val || 0)}
              clearable
            />
          </bk-form-item>
          <bk-form-item label={t('云盘总量')} property='name'>
            <span class={cssModule.number}>{props.planTicketDemand.cbs.disk_size} GB</span>
          </bk-form-item>
          {props.resourceType === 'cbs' ? (
            <bk-form-item label={t('所需数量')} property='disk_num' required class={cssModule['span-line']}>
              <bk-input
                type='number'
                suffix={t('块')}
                modelValue={props.planTicketDemand.cbs.disk_num}
                onChange={(val: number) => handleUpdatePlanTicketDemand('disk_num', val || 0)}
                clearable
              />
            </bk-form-item>
          ) : (
            ''
          )}
          <bk-form-item label={t('单实例磁盘IO')} property='disk_io' class={cssModule['span-line']}>
            <bk-input
              type='number'
              modelValue={props.planTicketDemand.cbs.disk_io}
              onChange={(val: number) => handleUpdatePlanTicketDemand('disk_io', val || 0)}
              clearable
            />
          </bk-form-item>
        </bk-form>
      </Panel>
    );
  },
});
