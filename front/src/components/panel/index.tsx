import { PropType, VNode, defineComponent, computed } from 'vue';
import cssModule from './index.module.scss';

export default defineComponent({
  props: {
    title: {
      type: [Function, String] as PropType<(() => string | HTMLElement | VNode) | String>,
      default: () => '',
    },
  },

  setup(props, { slots }) {
    const renderTitle = computed(() => (typeof props?.title === 'function' ? props.title() : props.title));
    return () => (
      <section class={cssModule.home}>
        {slots.title ? slots.title() : <span class={cssModule.title}>{renderTitle.value}</span>}
        {slots.default()}
      </section>
    );
  },
});
