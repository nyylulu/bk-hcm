import { CreateRecallTaskModal } from '@/typings/scr';
import { defineComponent, reactive, ref } from 'vue';
import { Dialog, Form, Message } from 'bkui-vue';
import ScrCreateFilterSelector from './ScrCreateFilterSelector';
import QcloudRegionSelector from '@/views/ziyanScr/components/qcloud-resource/region-selector.vue';
import QcloudZoneSelector from '@/views/ziyanScr/components/qcloud-resource/zone-selector.vue';
import InputNumber from '@/components/input-number';
import { useZiyanScrStore } from '@/store';
import { useI18n } from 'vue-i18n';

export default defineComponent({
  name: 'CreateRecallTaskDialog',
  emits: ['reloadTable'],
  setup(_, { emit, expose }) {
    const ziyanScrStore = useZiyanScrStore();

    const { t } = useI18n();
    const isShow = ref(false);
    const isLoading = ref(false);
    const formData = reactive<CreateRecallTaskModal>({
      device_type: '',
      bk_cloud_region: '',
      bk_cloud_zone: '',
      replicas: 0,
      asset_ids: [],
    });

    const triggerShow = (v: boolean, data?: CreateRecallTaskModal) => {
      isShow.value = v;
      data && Object.assign(formData, data);
    };

    const handleRecall = async (data: CreateRecallTaskModal) => {
      isLoading.value = true;
      try {
        await ziyanScrStore.createRecallTask(data);
        Message({ theme: 'success', message: '提交成功' });
        emit('reloadTable');
      } finally {
        isLoading.value = false;
      }
    };

    const handleConfirm = () => {
      handleRecall(formData);
      triggerShow(false);
    };

    expose({ triggerShow });

    return () => (
      <Dialog v-model:isShow={isShow.value} title={t('发起下架')} onConfirm={handleConfirm} isLoading={isLoading.value}>
        <Form model={formData} labelWidth={80}>
          <Form.FormItem property='device_type' label='机型'>
            <ScrCreateFilterSelector
              v-model={formData.device_type}
              api={ziyanScrStore.getDeviceTypeList}
              class='w200'
              multiple={false}
            />
          </Form.FormItem>
          <Form.FormItem property='region' label='地域'>
            <QcloudRegionSelector v-model={formData.bk_cloud_region} multiple={false} />
          </Form.FormItem>
          <Form.FormItem property='zone' label='园区'>
            <QcloudZoneSelector v-model={formData.bk_cloud_zone} multiple={false} region={[formData.bk_cloud_region]} />
          </Form.FormItem>
          <Form.FormItem property='replicas' label='数量'>
            <InputNumber v-model={formData.replicas} min={0} max={500} />
          </Form.FormItem>
        </Form>
      </Dialog>
    );
  },
});
