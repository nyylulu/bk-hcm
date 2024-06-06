import { defineComponent, ref } from 'vue';
import CommonSideslider from '@/components/common-sideslider';
import Basic from './basic';
import CVM from './cvm';
import CBS from './cbs';
import cssModule from './index.module.scss';
import { useI18n } from 'vue-i18n';

export default defineComponent({
  setup() {
    const { t } = useI18n();

    const isShow = ref(false);

    const handleClose = () => {
      isShow.value = false;
    };

    return () => (
      <CommonSideslider
        class={cssModule.home}
        isShow={isShow.value}
        width='960'
        handleClose={handleClose}
        title={t('增加预测类型')}>
        <Basic></Basic>
        <CVM class={cssModule.mt16}></CVM>
        <CBS class={cssModule.mt16}></CBS>
      </CommonSideslider>
    );
  },
});
