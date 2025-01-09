import { defineComponent, onMounted, ref, watch, nextTick } from 'vue';
import { useUserStore } from '@/store';
import { getRequireTypes } from '@/api/host/task';
import { createCvmProduceOrder } from '@/api/host/cvm';
import useCvmChargeType from '@/views/ziyanScr/hooks/use-cvm-charge-type';
import CvmForm from './cvm-form';
import { Dialog, Form, Message, Select } from 'bkui-vue';
import './index.scss';
const { FormItem } = Form;
export default defineComponent({
  components: {
    CvmForm,
  },
  props: {
    modelValue: {
      type: Boolean,
      default: false,
    },
    title: {
      type: String,
      default: '创建单据',
    },
    dataInfo: {
      type: Object,
      default: () => {
        return {};
      },
    },
  },
  emits: ['update:modelValue', 'clearDataInfo', 'updateProduceData'],
  setup(props, { attrs, emit }) {
    const isDisplay = ref(false);
    watch(
      () => props.modelValue,
      (val) => {
        isDisplay.value = val;
      },
      {
        immediate: true,
      },
    );
    const updateShowValue = () => {
      emit('update:modelValue', false);
      // 清空props.dataInfo
      emit('clearDataInfo');
    };
    const defaultTopModel = () => ({
      bk_biz_id: 931,
      bk_module_id: 29309,
      require_type: 1,
      spec: {},
    });
    const topModelForm = ref(defaultTopModel());
    const topRulesForm = ref({
      require_type: [{ required: true, message: '请选择需求类型' }],
    });
    const userStore = useUserStore();
    const { cvmChargeTypes } = useCvmChargeType();
    const defaultBottomModel = () => {
      return {
        replicas: 1,
        antiAffinityLevel: 'ANTI_NONE',
        remark: '',
        enableDiskCheck: false,
        spec: {
          region: '',
          zone: '',
          device_type: '',
          image_id: '',
          disk_size: 0,
          disk_type: 'CLOUD_PREMIUM',
          networkType: 'TENTHOUSAND', // 写成一个常量
          vpc: '',
          subnet: '',
          charge_type: cvmChargeTypes.PREPAID,
          charge_months: 36,
        },
      };
    };
    const bottomModelForm = ref(defaultBottomModel());
    watch(
      () => props.dataInfo,
      (val) => {
        if (!Object.keys(props.dataInfo).length) return;
        topModelForm.value.require_type = val.require_type;
        bottomModelForm.value.spec.region = val.region;
        bottomModelForm.value.spec.zone = val.zone;
        bottomModelForm.value.spec.device_type = val.device_type;
      },
      {
        deep: true,
      },
    );
    // 需求类型
    const requireTypeList = ref([]);
    const fetchRequireType = async () => {
      const res = await getRequireTypes();
      requireTypeList.value = res.data.info.map((item) => ({
        label: item.require_name,
        value: item.require_type,
      }));
    };
    const topModelFormRef = ref(null);
    const bottomModelFormRef = ref(null);
    const createProduceOrder = () => {
      createCvmProduceOrder({
        ...topModelForm.value,
        ...bottomModelForm.value,
        bk_username: userStore.username,
      })
        .then(() => {
          Message({ theme: 'success', message: '提交成功' });
          emit('updateProduceData');
        })
        .catch(() => {});
    };
    const handleOrderFormSubmit = async () => {
      await topModelFormRef.value.validate();
      await bottomModelFormRef.value.validate();
      createProduceOrder();
      handleOrderFormCancel();
    };
    const handleOrderFormCancel = () => {
      topModelForm.value = defaultTopModel();
      bottomModelForm.value = defaultBottomModel();
      nextTick(() => {
        bottomModelFormRef.value.clearValidate();
      });
      updateShowValue();
    };
    onMounted(() => {
      fetchRequireType();
    });
    return () => (
      <Dialog
        class='cvm-produce-create-order-dialog'
        v-bind={attrs}
        width='1300'
        v-model:isShow={isDisplay.value}
        title={props.title}
        onClosed={handleOrderFormCancel}
        renderDirective='if'>
        {{
          default: () => (
            <div>
              <Form ref={topModelFormRef} model={topModelForm.value} rules={topRulesForm.value}>
                <div class='form-item-container'>
                  <FormItem label='业务' required>
                    {`资源运营服务`}
                  </FormItem>
                  <FormItem label='模块' required>
                    SA云化池
                  </FormItem>
                </div>
                <div class='form-item-container'>
                  <FormItem label='需求类型' required property='require_type'>
                    <Select v-model={topModelForm.value.require_type} clearable class='i-form-control'>
                      {requireTypeList.value.map(({ label, value }) => {
                        return <Select.Option key={value} name={label} id={value} />;
                      })}
                    </Select>
                  </FormItem>
                </div>
              </Form>
              <cvm-form
                ref={bottomModelFormRef}
                v-model={bottomModelForm.value}
                requireType={topModelForm.value.require_type}
              />
            </div>
          ),
          footer: () => (
            <div class='dialog-footer-btn'>
              <bk-button
                theme='primary'
                onClick={handleOrderFormSubmit}
                disabled={bottomModelFormRef.value?.isSubmitDisabled}>
                提交
              </bk-button>
              <bk-button onClick={handleOrderFormCancel}>取消</bk-button>
            </div>
          ),
        }}
      </Dialog>
    );
  },
});
