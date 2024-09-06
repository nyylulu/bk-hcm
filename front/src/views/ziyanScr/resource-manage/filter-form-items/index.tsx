import { PropType, VNode, defineComponent } from 'vue';
import { Button } from 'bkui-vue';
import cssModule from './index.module.scss';

type FilterItemConfig = {
  label?: string;
  render: () => VNode;
  hidden?: boolean;
};

export default defineComponent({
  name: 'FilterFormItems',
  props: {
    config: Array as PropType<FilterItemConfig[]>,
    handleSearch: Function as PropType<() => void>,
    handleClear: Function as PropType<() => void>,
  },
  setup(props) {
    return () => (
      <>
        <div class={cssModule['filter-container']}>
          <div class={cssModule['filter-items-wrapper']}>
            {props.config.map(
              ({ label, render, hidden }) =>
                !hidden && (
                  <div class={cssModule['filter-item']}>
                    <div class={cssModule.label}>{label}</div>
                    <div class={cssModule.value}>{render()}</div>
                  </div>
                ),
            )}
          </div>
          <div class={cssModule['operation-btn-wrapper']}>
            <Button theme='primary' onClick={props.handleSearch}>
              查询
            </Button>
            <Button onClick={props.handleClear}>重置</Button>
          </div>
        </div>
      </>
    );
  },
});
