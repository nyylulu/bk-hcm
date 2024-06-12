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

    const rules = {
      disk_io: [
        {
          validator: (value: number) => {
            const cloundType = props.planTicketDemand.cbs.disk_type;
            if (cloundType === 'CLOUD_PREMIUM') {
              return value <= 150;
            }
            if (cloundType === 'CLOUD_SSD') {
              return value <= 260;
            }
          },
          message: t('高性能云盘上限150，SSD云硬盘上限260'),
          trigger: 'change',
        },
        {
          validator: (value: number) => {
            return value > 0;
          },
          message: t('云硬盘 数量应大于0'),
          trigger: 'change',
        },
      ],
      disk_per_size: [
        {
          validator: (value: number) => {
            return value > 0;
          },
          message: t('云磁盘容量/实例 数量应大于0'),
          trigger: 'change',
        },
      ],
    };

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

    const clearValidate = () => {
      return formRef.value?.clearValidate();
    };

    const getDiskTypes = () => {
      isLoadingDiskTypes.value = true;
      resourcePlanStore
        .getDiskTypes()
        .then((data: { data: { details: IDiskType[] } }) => {
          diskTypes.value = data?.data?.details || [];
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
      clearValidate,
    });

    return () => (
      <Panel title={t('CBS云磁盘信息')}>
        <bk-form
          form-type='vertical'
          ref={formRef}
          rules={rules}
          model={props.planTicketDemand.cbs}
          class={cssModule.home}>
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
            label={t('云磁盘容量/实例')}
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
          <bk-form-item
            label={t('云盘总量')}
            description={props.resourceType === 'cbs' ? t('需要的云磁盘总量') : t('所有实例的系统盘，数据盘总容量')}
            property='name'>
            <span class={cssModule.number}>{props.planTicketDemand.cbs.disk_size} GB</span>
          </bk-form-item>
          {props.resourceType === 'cbs' ? (
            <bk-form-item
              label={t('所需数量')}
              description={t('需要的云磁盘块数')}
              property='disk_num'
              required
              class={cssModule['span-line']}>
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
          <bk-form-item
            label={t('单实例磁盘IO')}
            description={t('磁盘IO吞吐需求，无特殊要求填写15；高性能云盘上限150，SSD云硬盘上限260')}
            property='disk_io'
            class={cssModule['span-line']}>
            <bk-input
              type='number'
              disabled={!props.planTicketDemand.cbs.disk_type}
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
