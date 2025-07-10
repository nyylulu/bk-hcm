import { defineComponent, ref, reactive, computed, watch } from 'vue';
import './index.scss';
import AreaSelector from '../../hostApplication/components/AreaSelector';
import ZoneSelector from '../../hostApplication/components/ZoneSelector';
import RequirementTypeSelector from '@/components/scr/requirement-type-selector';
import apiService from '@/api/scrApi';
import DialogFooter from '@/components/common-dialog/dialog-footer.vue';

export interface ICvmDeviceCreateModel {
  require_type: number[];
  zone: string[];
  device_group: string;
  device_size: '小核心' | '中核心' | '大核心';
  device_type: string;
  cpu: number;
  mem: number;
  remark?: string;
  region?: string[];
}

export default defineComponent({
  name: 'AllhostInventoryManager',
  props: { isShow: Boolean },
  emits: ['update:isShow', 'submit-success', 'hidden'],
  setup(props, { emit }) {
    const deviceSizeNames = { 小核心: '小核心', 中核心: '中核心', 大核心: '大核心' };

    const isShow = computed({
      get() {
        return props.isShow;
      },
      set(val) {
        emit('update:isShow', val);
      },
    });

    const formRef = ref();
    const formModel = reactive<ICvmDeviceCreateModel>({
      require_type: [],
      zone: [],
      region: [],
      device_group: '',
      device_size: '小核心',
      device_type: '',
      cpu: 0,
      mem: 1,
      remark: '',
    });
    const isForceCreate = ref(false);

    const handleRegionChange = () => {
      formModel.zone = [];
    };

    const validateDeviceState = reactive({ code: 0, message: '' });
    const isSubmitLoading = ref(false);
    const disabledSubmit = computed(() => {
      const rules = [
        validateDeviceState.code === 2000022,
        validateDeviceState.code === 2000023 && !isForceCreate.value,
      ];
      return rules.some((rule) => rule);
    });

    const handleConfirm = async () => {
      await formRef.value.validate();
      isSubmitLoading.value = true;
      try {
        const res = await apiService.createCvmDevice(
          { ...formModel, force_create: isForceCreate.value },
          { globalError: false },
        );

        if (res.code === 0) {
          emit('submit-success');
          isShow.value = false;
        } else {
          Object.assign(validateDeviceState, res);
        }
      } finally {
        isSubmitLoading.value = false;
      }
    };

    const handleClosed = () => {
      isShow.value = false;
    };

    watch(
      formModel,
      () => {
        // 当表单值变化时，重置校验状态
        if (validateDeviceState.code !== 0) {
          validateDeviceState.code = 0;
          isForceCreate.value = false;
        }
      },
      { deep: true },
    );

    return () => (
      <bk-dialog
        v-model:isShow={isShow.value}
        class='common-dialog'
        close-icon={false}
        title='创建新机型'
        width={600}
        onHidden={() => emit('hidden')}>
        {{
          default: () => (
            <bk-form model={formModel} ref={formRef} class='cvm-device-create-form'>
              <bk-form-item label='需求类型' property='require_type' required>
                <RequirementTypeSelector v-model={formModel.require_type} multiple class='i-form-control' />
              </bk-form-item>
              <bk-form-item label='地域' property='region' required>
                <AreaSelector
                  ref='areaSelector'
                  v-model={formModel.region}
                  params={{ resourceType: 'QCLOUDCVM' }}
                  multiple
                  onChange={handleRegionChange}
                  class='i-form-control'
                />
              </bk-form-item>
              <bk-form-item label='园区' property='zone' required>
                <ZoneSelector
                  ref='zoneSelector'
                  v-model={formModel.zone}
                  params={{ resourceType: 'QCLOUDCVM', region: formModel.region }}
                  separateCampus={false}
                  multiple
                  class='i-form-control'
                />
              </bk-form-item>
              <bk-form-item label='实例族' property='device_group' required>
                <bk-input v-model={formModel.device_group} class='i-form-control' />
              </bk-form-item>
              <bk-form-item label='机型' property='device_type' required>
                <bk-input v-model={formModel.device_type} class='i-form-control' />
              </bk-form-item>
              <bk-form-item label='核心类型' property='cpu' required>
                <hcm-form-enum v-model={formModel.device_size} option={deviceSizeNames} class='i-form-control' />
              </bk-form-item>
              <bk-form-item label='CPU(核)' property='cpu' required>
                <bk-input type='number' v-model={formModel.cpu} min={0} class='i-form-control' />
              </bk-form-item>
              <bk-form-item label='内存(G)' property='mem' required>
                <bk-input type='number' v-model={formModel.mem} min={0} class='i-form-control' />
              </bk-form-item>
              <bk-form-item label='其他信息'>
                <bk-input v-model={formModel.remark} autosize type='textarea' maxlength={128} class='i-form-control' />
              </bk-form-item>
            </bk-form>
          ),
          footer: () => (
            <div class='create-device-dialog-footer'>
              {validateDeviceState.code === 2000022 && (
                <span class='mr-auto text-danger'>机型已存在，无需重复添加</span>
              )}
              {validateDeviceState.code === 2000023 && (
                <bk-checkbox v-model={isForceCreate.value} class='mr-auto'>
                  该机型在CRP不存在，请确认是否添加
                </bk-checkbox>
              )}
              <DialogFooter
                disabled={disabledSubmit.value}
                loading={isSubmitLoading.value}
                onConfirm={handleConfirm}
                onClosed={handleClosed}
              />
            </div>
          ),
        }}
      </bk-dialog>
    );
  },
});
