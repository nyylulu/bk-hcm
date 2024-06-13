import { defineComponent, ref, onMounted } from 'vue';
import { Dialog } from 'bkui-vue';
import './index.scss';
import AreaSelector from '../../hostApplication/components/AreaSelector';
import ZoneSelector from '../../hostApplication/components/ZoneSelector';
import apiService from '@/api/scrApi';
export default defineComponent({
  name: 'AllhostInventoryManager',
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
  },
  emits: ['update:visible', 'query-list'],
  setup(props, { emit }) {
    const handleConfirm = async () => {
      await apiService.createCvmDevice(EditForm.value);
      emit('update:visible', false);
      emit('query-list');
    };
    const triggerShow = (val) => {
      emit('update:visible', val);
      emit('query-list');
    };
    const EditForm = ref({
      remark: '',
      requireType: '',
      region: [],
      zone: [],
      deviceGroup: '',
      deviceType: '',
      cpu: 0,
      mem: 1,
    });
    const order = ref({
      requireTypes: [],
    });
    const onEditFormRegionChange = () => {
      EditForm.value.zone = [];
    };
    onMounted(() => {
      getrequireTypes();
    });
    const getrequireTypes = async () => {
      const { info } = await apiService.getRequireTypes();
      order.value.requireTypes = info;
    };
    return () => (
      <div>
        <Dialog
          class='common-dialog'
          isShow={props.visible}
          title='批量更新'
          width={600}
          onConfirm={handleConfirm}
          onClosed={() => triggerShow(false)}>
          <bk-form v-loading='$isLoading(deviceTypeConfigs.updateRequestId)'>
            <bk-form-item label='需求类型' required property='requireType'>
              <bk-select v-model={EditForm.value.requireType} style='width: 192px'>
                {order.value.requireTypes.map((item: { require_type: any; require_name: any }) => (
                  <bk-option key={item.require_type} value={item.require_type} label={item.require_name}></bk-option>
                ))}
              </bk-select>
            </bk-form-item>
            <bk-form-item class='mr16' label='地域' required>
              <AreaSelector
                ref='areaSelector'
                multiple
                style='width: 192px'
                v-model={EditForm.value.region}
                params={{ resourceType: 'QCLOUDCVM' }}
                onChange={onEditFormRegionChange}></AreaSelector>
            </bk-form-item>
            <bk-form-item label='园区' property='zone'>
              <ZoneSelector
                ref='zoneSelector'
                multiple
                style='width: 192px'
                v-model={EditForm.value.zone}
                params={{
                  resourceType: 'QCLOUDCVM',
                  region: EditForm.value.region,
                }}
              />
            </bk-form-item>
            <bk-form-item label='实例族' prop='deviceGroup'>
              <bk-input v-model={EditForm.value.deviceGroup} style='width: 192px' />
            </bk-form-item>
            <bk-form-item label='机型' prop='deviceType'>
              <bk-input v-model={EditForm.value.deviceType} style='width: 192px' />
            </bk-form-item>
            <bk-form-item label='CPU(核)' prop='cpu'>
              <bk-input type='number' v-model={EditForm.value.cpu} min={0} style='width: 192px' />
            </bk-form-item>
            <bk-form-item label='内存(G)' prop='mem'>
              <bk-input type='number' v-model={EditForm.value.mem} min={0} style='width: 192px' />
            </bk-form-item>
            <bk-form-item labbk='其他信息'>
              <remark-textarea v-model={EditForm.value.remark} type='textarea' maxlength='128' show-limit />
            </bk-form-item>
          </bk-form>
        </Dialog>
      </div>
    );
  },
});
