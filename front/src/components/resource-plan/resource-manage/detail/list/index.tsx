import { defineComponent } from 'vue';
import cssModule from './index.module.scss';

export default defineComponent({
  props: {
    isBiz: {
      type: Boolean,
      required: true,
    },
  },
  setup(props) {
    return () => (
      <section class={cssModule.home}>
        {props.isBiz ? '资源管理 资源预测详情  调整记录 * 列表' : '服务管理 资源预测详情  调整记录 * 列表'}
      </section>
    );
  },
});
