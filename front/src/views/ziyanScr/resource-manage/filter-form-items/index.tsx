import { PropType, VNode, defineComponent } from 'vue';
import { Button } from 'bkui-vue';
import { Search } from 'bkui-vue/lib/icon';
import './index.scss';

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
  setup(props, { slots }) {
    const renderFilterItems: FilterItemConfig[] = [
      ...props.config,
      {
        render: () => (
          <Button theme='primary' onClick={props.handleSearch}>
            <Search />
            查询
          </Button>
        ),
      },
      {
        render: () => <Button onClick={props.handleClear}>清空</Button>,
      },
    ];

    return () => (
      <div class='filter-container'>
        {renderFilterItems.map(
          ({ label, render, hidden }) =>
            !hidden && (
              <div class='filter-item mr8'>
                {label && <span class='mr8'>{label}</span>}
                {render()}
              </div>
            ),
        )}
        {slots.end?.()}
      </div>
    );
  },
});
