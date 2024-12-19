import { defineComponent, ref, nextTick, reactive } from 'vue';
import './index.scss';
import AreaSelector from '../../hostApplication/components/AreaSelector';
import ZoneSelector from '../../hostApplication/components/ZoneSelector';
import RequirementTypeSelector from '@/components/scr/requirement-type-selector';
import apiService from '@/api/scrApi';

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
  emits: ['queryList'],
  setup(_, { emit, expose }) {
    const deviceSizeNames = { 小核心: '小核心', 中核心: '中核心', 大核心: '大核心' };

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
    });

    const handleRegionChange = () => {
      formModel.zone = [];
    };

    const handleConfirm = async () => {
      await formRef.value.validate();
      const { require_type, zone, device_group, device_size, device_type, cpu, mem, remark } = formModel;
      const params = { require_type, zone, device_group, device_size, device_type, cpu, mem, remark };
      await apiService.createCvmDevice(params);
      emit('queryList');
      clearValidate();
    };

    const clearValidate = () => {
      nextTick(() => {
        formRef.value.clearValidate();
      });
    };

    expose({ handleConfirm, clearValidate });

    return () => (
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
    );
  },
});
