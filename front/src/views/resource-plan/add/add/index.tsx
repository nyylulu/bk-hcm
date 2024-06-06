import { defineComponent, ref } from 'vue';
import CommonSideslider from '@/components/common-sideslider';
import Basic from './basic';
import CVM from './cvm';
import CBS from './cbs';
import cssModule from './index.module.scss';

export default defineComponent({
  setup() {
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
        title='增加预测类型'>
        <Basic></Basic>
        <CVM></CVM>
        <CBS></CBS>
      </CommonSideslider>
    );
  },
});
