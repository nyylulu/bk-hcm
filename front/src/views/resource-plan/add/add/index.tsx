import { defineComponent, type PropType, ref, watch } from 'vue';
import CommonSideslider from '@/components/common-sideslider';
import Basic from './basic';
import CVM from './cvm';
import CBS from './cbs';
import cssModule from './index.module.scss';
import { useI18n } from 'vue-i18n';
import dayjs from 'dayjs';

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

    const initPlanTicketDemand = () => {
      planTicketDemand.value = {
        obs_project: '',
        expect_time: dayjs().format('YYYY-MM-DD'),
        region: '',
        zone: '',
        demand_source: '',
        remark: '',
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
          disk_io: 0,
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

    const handleSubmit = () => {
      Promise.all([basicRef.value.validate(), cvmRef.value.validate(), cbsRef.value.validate()]).then(() => {
        emit('update:modelValue', {
          ...props.modelValue,
          demands: [...props.modelValue.demands, planTicketDemand.value],
        });
        handleClose();
      });
    };

    watch(
      () => props.isShow,
      () => {
        if (props.isShow) {
          initPlanTicketDemand();
        }
      },
    );

    return () => (
      <CommonSideslider
        width='960'
        class={cssModule.home}
        isShow={props.isShow}
        title={t('增加预测类型')}
        handleClose={handleClose}
        onUpdate:isShow={handleClose}
        onHandleSubmit={handleSubmit}>
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
