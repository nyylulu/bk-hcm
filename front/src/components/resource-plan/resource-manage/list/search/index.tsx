import { defineComponent } from 'vue';
import cssModule from './index.module.scss';
import Panel from '@/components/panel';

export default defineComponent({
  props: {
    isBiz: {
      type: Boolean,
      required: true,
    },
  },
  setup() {
    return () => (
      <Panel class={cssModule['mb-16']}>
        <section class={cssModule.home}>资源预测 -- 搜索</section>
      </Panel>
    );
  },
});
