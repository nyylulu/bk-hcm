import { defineComponent, ref, onBeforeMount, type PropType, watch, nextTick } from 'vue';
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

    const rules = {
      os: [
        {
          validator: (value: number) => {
            return value > 0;
          },
          message: t('实例数量应大于0'),
          trigger: 'change',
        },
      ],
    };

    const formRef = ref();
    const deviceClasses = ref<string[]>([]);
    const deviceTypes = ref<IDeviceType[]>([]);
    const isLoadingDeviceClasses = ref(false);
    const isLoadingDeviceTypes = ref(false);
    const deviceTypeInfo = ref('');

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
        .then((data: { data: { details: string[] } }) => {
          deviceClasses.value = data?.data?.details || [];
        })
        .finally(() => {
          isLoadingDeviceClasses.value = false;
        });
    };

    const getDeviceTypes = () => {
      // 重置机型规格列表
      if (props.planTicketDemand.cvm.device_class) {
        isLoadingDeviceTypes.value = true;
        resourcePlanStore
          .getDeviceTypes([props.planTicketDemand.cvm.device_class])
          .then((data: { data: { details: IDeviceType[] } }) => {
            deviceTypes.value = data?.data?.details || [];
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
      deviceTypeInfo.value = deviceType
        ? t('所选机型为{0}，CPU为{1}核，内存为{2}G', [deviceType.core_type, deviceType.cpu_core, deviceType.memory])
        : '';

      const perCpuCore = deviceType?.cpu_core || 0;
      const perMemory = deviceType?.memory || 0;
      const osNum = +props.planTicketDemand.cvm.os;

      nextTick(() => {
        handleUpdatePlanTicketDemand('cpu_core', perCpuCore * osNum);

        nextTick(() => {
          handleUpdatePlanTicketDemand('memory', perMemory * osNum);
        });
      });
    };

    const validate = () => {
      return formRef.value?.validate();
    };

    const clearValidate = () => {
      return formRef.value?.clearValidate();
    };

    watch(
      () => props.planTicketDemand.cvm.device_class,
      () => {
        // 重置机型规格
        handleUpdatePlanTicketDemand('device_type', '');
        // 更新数据
        getDeviceTypes();
      },
    );

    watch([() => props.planTicketDemand.cvm.device_type, () => props.planTicketDemand.cvm.os], calcCpuAndMemory, {
      immediate: true,
    });

    onBeforeMount(() => {
      getDeviceClasses();
      getDeviceTypes();
    });

    expose({
      validate,
      clearValidate,
    });

    return () =>
      props.resourceType === 'cvm' ? (
        <Panel title={t('CVM云主机信息')}>
          <bk-form
            form-type='vertical'
            model={props.planTicketDemand.cvm}
            rules={rules}
            ref={formRef}
            class={cssModule.home}>
            <bk-form-item label={t('资源模式')} class={cssModule['span-6']}>
              <bk-radio-group modelValue={props.planTicketDemand.cvm.res_mode}>
                <bk-radio-button label='按机型' />
                <bk-radio-button label='按机型族' disabled />
              </bk-radio-group>
              <span class={cssModule['tip-text']}>{t('暂不支持按机型族选择')}</span>
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
                  <bk-option id={deviceType.device_type} name={deviceType.device_type}></bk-option>
                ))}
              </bk-select>
              <span class={cssModule.info}>{deviceTypeInfo.value}</span>
            </bk-form-item>
            <bk-form-item label={t('实例数量')} property='os' class={cssModule['span-2']}>
              <bk-input
                type='number'
                suffix={t('台')}
                min={0}
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
