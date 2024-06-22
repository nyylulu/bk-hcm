import { defineComponent, ref, nextTick } from 'vue';
import './index.scss';
import AreaSelector from '../../hostApplication/components/AreaSelector';
import ZoneSelector from '../../hostApplication/components/ZoneSelector';
import RequirementTypeSelector from '@/components/scr/requirement-type-selector';
import apiService from '@/api/scrApi';
export default defineComponent({
  name: 'AllhostInventoryManager',
  emits: ['queryList'],
  setup(props, { emit, expose }) {
    const handleConfirm = async () => {
      await formInstance.value.validate();
      await apiService.createCvmDevice(EditForm.value);
      emit('queryList');
      clearValidate();
    };
    const clearValidate = () => {
      nextTick(() => {
        formInstance.value.clearValidate();
      });
    };
    const EditForm = ref({
      remark: '',
      requireType: [],
      region: [],
      zone: [],
      deviceGroup: '',
      deviceType: '',
      cpu: 0,
      mem: 1,
    });
    const onEditFormRegionChange = () => {
      EditForm.value.zone = [];
    };
    const formInstance = ref();
    const formRules = ref({
      requireType: [{ required: true, message: '请选择需求类型', trigger: 'change' }],
      region: [{ required: true, message: '请选择地域', trigger: 'change' }],
      zone: [{ required: true, message: '请选择园区', trigger: 'change' }],
      deviceGroup: [{ required: true, message: '请输入实例族', trigger: 'change' }],
      deviceType: [{ required: true, message: '请输入机型', trigger: 'change' }],
      cpu: [{ required: true, message: '请输入CPU', trigger: 'change' }],
      mem: [{ required: true, message: '请输入内存', trigger: 'change' }],
    });
    expose({ handleConfirm, clearValidate });
    return () => (
      <div>
        <bk-form model={EditForm.value} ref={formInstance} rules={formRules.value}>
          <bk-form-item label='需求类型' required property='requireType'>
            <RequirementTypeSelector style='width: 250px' v-model={EditForm.value.requireType} multiple />
          </bk-form-item>
          <bk-form-item class='mr16' label='地域' required property='region'>
            <AreaSelector
              ref='areaSelector'
              multiple
              style='width: 250px'
              v-model={EditForm.value.region}
              params={{ resourceType: 'QCLOUDCVM' }}
              onChange={onEditFormRegionChange}></AreaSelector>
          </bk-form-item>
          <bk-form-item label='园区' required property='zone'>
            <ZoneSelector
              ref='zoneSelector'
              multiple
              style='width: 250px'
              separateCampus={false}
              v-model={EditForm.value.zone}
              params={{
                resourceType: 'QCLOUDCVM',
                region: EditForm.value.region,
              }}
            />
          </bk-form-item>
          <bk-form-item label='实例族' required prop='deviceGroup'>
            <bk-input v-model={EditForm.value.deviceGroup} style='width: 250px' />
          </bk-form-item>
          <bk-form-item label='机型' required prop='deviceType'>
            <bk-input v-model={EditForm.value.deviceType} style='width: 250px' />
          </bk-form-item>
          <bk-form-item label='CPU(核)' required prop='cpu'>
            <bk-input type='number' v-model={EditForm.value.cpu} min={0} style='width: 250px' />
          </bk-form-item>
          <bk-form-item label='内存(G)' required prop='mem'>
            <bk-input type='number' v-model={EditForm.value.mem} min={0} style='width: 250px' />
          </bk-form-item>
          <bk-form-item label='其他信息'>
            <bk-input
              v-model={EditForm.value.remark}
              class='mb8'
              autosize
              style='width: 250px'
              type='textarea'
              maxlength={128}
            />
          </bk-form-item>
        </bk-form>
      </div>
    );
  },
});
