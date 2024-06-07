import { defineComponent, ref } from 'vue';
import CommonSideslider from '@/components/common-sideslider';
import Basic from './basic';
import CVM from './cvm';
import CBS from './cbs';
import cssModule from './index.module.scss';
import { useI18n } from 'vue-i18n';
import { useResourcePlanStore } from '@/store';

export default defineComponent({
  setup() {
    const { t } = useI18n();
    const resourcePlanStore = useResourcePlanStore();

    const isShow = ref(true);

    resourcePlanStore.getDiskTypes();

    const handleClose = () => {
      isShow.value = false;
    };

    const handleSubmit = () => {
      handleClose();
    };

    return () => (
      <CommonSideslider
        class={cssModule.home}
        isShow={isShow.value}
        width='960'
        handleClose={handleClose}
        title={t('增加预测类型')}
        onUpdate:isShow={handleClose}
        onHandleSubmit={handleSubmit}>
        <Basic></Basic>
        <CVM class={cssModule.mt16}></CVM>
        <CBS class={cssModule.mt16}></CBS>
      </CommonSideslider>
    );
  },
});
