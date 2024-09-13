import { defineComponent, type PropType, ref, watch } from 'vue';
import CommonSideslider from '@/components/common-sideslider';
import Basic from './basic';
import CVM from './cvm';
import CBS from './cbs';
import cssModule from './index.module.scss';
import { useI18n } from 'vue-i18n';

import type { IPlanTicket, IPlanTicketDemand } from '@/typings/resourcePlan';

export default defineComponent({
  props: {
    isShow: {
      type: Boolean,
    },
    modelValue: {
      type: Object as PropType<IPlanTicket>,
    },
    initDemand: {
      type: Object as PropType<IPlanTicketDemand>,
    },
  },

  emits: ['update:isShow', 'update:modelValue'],

  setup(props, { emit }) {
    const { t } = useI18n();

    const basicRef = ref(null);
    const cvmRef = ref(null);
    const cbsRef = ref(null);
    const resourceType = ref('cvm');
    const planTicketDemand = ref<IPlanTicketDemand>();

    const initData = () => {
      resourceType.value = props.initDemand?.demand_res_types.length < 2 ? 'cbs' : 'cvm';
      planTicketDemand.value = {
        obs_project: '',
        expect_time: '2024-10-01',
        region_id: '',
        zone_id: '',
        demand_source: '指标变化',
        remark: '',
        demand_res_types: ['CVM', 'CBS'],
        cvm: {
          res_mode: '按机型',
          device_class: '',
          device_type: '',
          os: 0,
          cpu_core: 0,
          memory: 0,
        },
        cbs: {
          disk_type: '',
          disk_type_name: '',
          disk_io: 15,
          disk_size: 0,
          disk_num: 0,
          disk_per_size: 0,
        },
        ...props.initDemand,
      };
    };

    const handleClose = () => {
      emit('update:isShow', false);
    };

    const handleSubmit = async () => {
      await validate();
      if (props.initDemand) {
        const demandIndex = props.modelValue.demands.findIndex((demand) => demand === props.initDemand);
        emit('update:modelValue', {
          ...props.modelValue,
          demands: [
            ...props.modelValue.demands.slice(0, demandIndex),
            planTicketDemand.value,
            ...props.modelValue.demands.slice(demandIndex + 1),
          ],
        });
      } else {
        emit('update:modelValue', {
          ...props.modelValue,
          demands: [...props.modelValue.demands, planTicketDemand.value],
        });
      }
      handleClose();
    };

    const validate = () => {
      return Promise.all([basicRef.value.validate(), cvmRef.value.validate(), cbsRef.value.validate()]);
    };

    const clearValidate = () => {
      Promise.all([basicRef.value.clearValidate(), cvmRef.value.clearValidate(), cbsRef.value.clearValidate()]);
    };

    const handleShown = () => {
      clearValidate();
    };

    watch(
      () => props.isShow,
      () => {
        if (props.isShow) {
          initData();
        }
      },
    );

    watch(() => resourceType.value, clearValidate);

    return () => (
      <CommonSideslider
        width='960'
        class={cssModule.home}
        isShow={props.isShow}
        title={props.initDemand ? t('修改预测需求') : t('增加预测需求')}
        handleClose={handleClose}
        onUpdate:isShow={handleClose}
        onHandleSubmit={handleSubmit}
        onHandleShown={handleShown}>
        <Basic
          ref={basicRef}
          v-model:planTicketDemand={planTicketDemand.value}
          v-model:resourceType={resourceType.value}
        />
        <CVM
          ref={cvmRef}
          v-model:planTicketDemand={planTicketDemand.value}
          resourceType={resourceType.value}
          class={cssModule.mt16}
        />
        <CBS
          ref={cbsRef}
          v-model:planTicketDemand={planTicketDemand.value}
          resourceType={resourceType.value}
          class={cssModule.mt16}
        />
      </CommonSideslider>
    );
  },
});
