import { defineComponent, type PropType, ref, watch } from 'vue';
import CommonSideslider from '@/components/common-sideslider';
import Basic from './basic';
import CVM from './cvm';
import CBS from './cbs';
import cssModule from './index.module.scss';
import { useI18n } from 'vue-i18n';
import type { IPlanTicket, IPlanTicketDemand } from '@/typings/resourcePlan';
import { Button, Form } from 'bkui-vue';
import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import { AdjustType } from '@/typings/plan';
import Panel from '@/components/panel';
const { FormItem } = Form;

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

  emits: ['update:isShow', 'update:modelValue', 'update:demand'],

  setup(props, { emit }) {
    const { t } = useI18n();

    const basicRef = ref(null);
    const cvmRef = ref(null);
    const cbsRef = ref(null);
    const resourceType = ref('cvm');
    const planTicketDemand = ref<IPlanTicketDemand>();
    const adjust_type = ref(props.initDemand.adjustType === AdjustType.time ? AdjustType.time : AdjustType.config);

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
      emit('update:demand', { ...planTicketDemand.value, adjustType: adjust_type.value });
      handleClose();
    };

    const validate = () => {
      return Promise.all([basicRef.value.validate(), cvmRef.value.validate(), cbsRef.value.validate()]);
    };

    const clearValidate = () => {
      Promise.all([basicRef.value?.clearValidate(), cvmRef.value?.clearValidate(), cbsRef.value?.clearValidate()]);
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

    watch(
      () => props.initDemand.adjustType,
      (type) => {
        if (type === AdjustType.time) adjust_type.value = AdjustType.time;
        else adjust_type.value = AdjustType.config;
      },
    );

    watch(() => resourceType.value, clearValidate);

    return () => (
      <CommonSideslider
        noFooter
        width='960'
        class={cssModule.home}
        isShow={props.isShow}
        title={props.initDemand ? t('修改预测需求') : t('增加预测需求')}
        handleClose={handleClose}
        onUpdate:isShow={handleClose}
        onHandleSubmit={handleSubmit}
        renderType='if'
        onHandleShown={handleShown}>
        <Panel class={'mb16'} title={`${'调整类型'}`}>
          <Form formType='vertical'>
            <FormItem label={t('调整方式')}>
              <BkRadioGroup v-model={adjust_type.value}>
                <BkRadioButton label={AdjustType.config} disabled={props.initDemand.adjustType === AdjustType.time}>
                  {t('调整配置')}
                </BkRadioButton>
                <BkRadioButton label={AdjustType.time} disabled={props.initDemand.adjustType === AdjustType.config}>
                  {t('调整时间')}
                </BkRadioButton>
              </BkRadioGroup>
            </FormItem>
          </Form>
        </Panel>
        <Basic
          type={adjust_type.value}
          ref={basicRef}
          v-model:planTicketDemand={planTicketDemand.value}
          v-model:resourceType={resourceType.value}
        />
        <CVM
          type={adjust_type.value}
          ref={cvmRef}
          v-model:planTicketDemand={planTicketDemand.value}
          resourceType={resourceType.value}
          class={cssModule.mt16}
        />
        <CBS
          type={adjust_type.value}
          ref={cbsRef}
          v-model:planTicketDemand={planTicketDemand.value}
          resourceType={resourceType.value}
          class={cssModule.mt16}
        />
        <section class={'mt16'}>
          <Button theme='primary' class={'mr16'} onClick={handleSubmit}>
            提交
          </Button>
          <Button onClick={handleClose}>取消</Button>
        </section>
      </CommonSideslider>
    );
  },
});
