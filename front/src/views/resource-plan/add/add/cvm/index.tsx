import { defineComponent, ref, onBeforeMount, type PropType, watch } from 'vue';
import Panel from '@/components/panel';
import { useI18n } from 'vue-i18n';
import { useResourcePlanStore } from '@/store';
import cssModule from './index.module.scss';

import type { IPlanTicketDemand, IDeviceType } from '@/typings/resourcePlan';

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
    const deviceClasses = ref<string[]>([]);
    const deviceTypes = ref<IDeviceType[]>([]);
    const isLoadingDeviceClasses = ref(false);
    const isLoadingDeviceTypes = ref(false);

    const handleUpdatePlanTicketDemand = (key: string, value: unknown) => {
      emit('update:planTicketDemand', {
        ...props.planTicketDemand,
        cvm: {
          ...props.planTicketDemand.cvm,
          [key]: value,
        },
      });
    };

    const getDeviceClasses = () => {
      isLoadingDeviceClasses.value = true;
      resourcePlanStore
        .getDeviceClasses()
        .then((data: { details: string[] }) => {
          deviceClasses.value = data.details || [];
        })
        .finally(() => {
          isLoadingDeviceClasses.value = false;
        });
    };

    const getDeviceTypes = () => {
      // 重置机型规格
      handleUpdatePlanTicketDemand('device_type', '');
      // 重置机型规格列表
      if (props.planTicketDemand.cvm.device_class) {
        isLoadingDeviceTypes.value = true;
        resourcePlanStore
          .getDeviceTypes([props.planTicketDemand.cvm.device_class])
          .then((data: { details: IDeviceType[] }) => {
            deviceTypes.value = data.details || [];
          })
          .finally(() => {
            isLoadingDeviceTypes.value = false;
          });
      } else {
        deviceTypes.value = [];
      }
    };

    const calcCpuAndMemory = () => {
      const deviceType = deviceTypes.value.find(
        (deviceType) => deviceType.device_type === props.planTicketDemand.cvm.device_type,
      );
      handleUpdatePlanTicketDemand('cpu_core', (deviceType?.cpu_core || 0) * props.planTicketDemand.cvm.os);
      handleUpdatePlanTicketDemand('memory', (deviceType?.memory || 0) * props.planTicketDemand.cvm.os);
    };

    const validate = () => {
      return formRef.value?.validate();
    };

    watch(() => props.planTicketDemand.cvm.device_class, getDeviceTypes, {
      immediate: true,
    });

    watch([() => props.planTicketDemand.cvm.device_type, () => props.planTicketDemand.cvm.os], calcCpuAndMemory);

    onBeforeMount(() => {
      getDeviceClasses();
    });

    expose({
      validate,
    });

    return () =>
      props.resourceType === 'cvm' ? (
        <Panel title={t('CVM云主机信息')}>
          <bk-form form-type='vertical' model={props.planTicketDemand.cvm} ref={formRef} class={cssModule.home}>
            <bk-form-item label={t('资源模式')} class={cssModule['span-6']}>
              <bk-radio-group modelValue={props.planTicketDemand.cvm.res_mode}>
                <bk-radio-button label='按机型' />
                <bk-radio-button label='按机型族' disabled />
              </bk-radio-group>
            </bk-form-item>
            <bk-form-item label={t('机型类型')} property='device_class' required class={cssModule['span-3']}>
              <bk-select
                clearable
                loading={isLoadingDeviceClasses.value}
                modelValue={props.planTicketDemand.cvm.device_class}
                onChange={(val: string) => handleUpdatePlanTicketDemand('device_class', val)}>
                {deviceClasses.value.map((deviceClass) => (
                  <bk-option id={deviceClass} name={deviceClass}></bk-option>
                ))}
              </bk-select>
            </bk-form-item>
            <bk-form-item label={t('机型规格')} property='device_type' required class={cssModule['span-3']}>
              <bk-select
                clearable
                loading={isLoadingDeviceTypes.value}
                modelValue={props.planTicketDemand.cvm.device_type}
                onChange={(val: string) => handleUpdatePlanTicketDemand('device_type', val)}>
                {deviceTypes.value.map((deviceType) => (
                  <bk-option id={deviceType.device_type} name={deviceType.core_type}></bk-option>
                ))}
              </bk-select>
            </bk-form-item>
            <bk-form-item label={t('实例数量')} property='os' class={cssModule['span-2']}>
              <bk-input
                type='number'
                suffix={t('台')}
                modelValue={props.planTicketDemand.cvm.os}
                onChange={(val: number) => handleUpdatePlanTicketDemand('os', val || 0)}
                clearable
              />
            </bk-form-item>
            <bk-form-item label={t('CPU总核数')} property='name'>
              <span class={cssModule.number}>{props.planTicketDemand.cvm.cpu_core} 核</span>
            </bk-form-item>
            <bk-form-item label={t('内存总量')} property='name'>
              <span class={cssModule.number}>{props.planTicketDemand.cvm.memory} GB</span>
            </bk-form-item>
          </bk-form>
        </Panel>
      ) : (
        ''
      );
  },
});
