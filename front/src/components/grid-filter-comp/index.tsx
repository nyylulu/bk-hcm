import { defineComponent, onMounted, PropType, VNode } from 'vue';
import cssModule from './index.module.scss';
import { Button } from 'bkui-vue';

interface RuleCompInfo {
  title: string;
  content: VNode;
}

export default defineComponent({
  props: {
    rules: {
      type: Array as PropType<RuleCompInfo[]>,
      required: true,
    },
    loading: Boolean,
    col: { type: Number, default: 3 },
    immediate: Boolean,
  },
  emits: ['search', 'reset'],
  setup(props, { emit }) {
    onMounted(() => {
      props.immediate && emit('search');
    });

    return () => (
      <section class={cssModule.filter}>
        <div class={cssModule.rules} style={{ gridTemplateColumns: `repeat(${props.col}, 1fr)` }}>
          {props.rules.map(({ title, content }) => (
            <div class={cssModule.item}>
              <div class={cssModule.title}>{title}</div>
              <div class={cssModule.content}>{content}</div>
            </div>
          ))}
        </div>
        <div class={cssModule.buttons}>
          <Button class={cssModule.button} theme='primary' onClick={() => emit('search')} loading={props.loading}>
            查询
          </Button>
          <Button class={cssModule.button} onClick={() => emit('reset')}>
            重置
          </Button>
        </div>
      </section>
    );
  },
});
