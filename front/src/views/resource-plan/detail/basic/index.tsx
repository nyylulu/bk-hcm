import { defineComponent, PropType, ref, watch } from 'vue';
import Panel from '@/components/panel';
import cssModule from '../index.module.scss';
import { TicketBaseInfo } from '@/typings/resourcePlan';
export default defineComponent({
  props: {
    baseData: {
      type: Object as PropType<TicketBaseInfo>,
    },
  },
  setup(props) {
    const baseList = ref([
      {
        label: '业务名称：',
        id: 'bk_biz_name',
        value: '1111',
      },
      {
        label: '部门：',
        id: 'plan_product_name',
        value: '1111',
      },
      {
        label: '运营产品：',
        id: 'bk_product_name',
        value: '1111',
      },
      {
        label: '预测类型：',
        id: 'demand_class',
        value: '1111',
      },
      {
        label: '规划产品：',
        id: 'plan_product_name',
        value: '1111',
      },
      {
        label: '创建时间：',
        id: '',
        value: '1111',
      },
    ]);

    const setBaseListVal = () => {
      baseList.value = baseList.value.map((item) => {
        if (item.id in props.baseData) {
          item.value = props.baseData[item.id];
        }
        return item;
      });
    };

    watch(
      () => props.baseData,
      () => {
        setBaseListVal();
      },
    );
    return () => (
      <Panel title='基本信息' class={cssModule['mb-16']}>
        <div class={cssModule['base-grid']}>
          {baseList.value.map((item) => (
            <div>
              {item.label} <span class={cssModule['base-text']}>{item.value}</span>
            </div>
          ))}
        </div>
      </Panel>
    );
  },
});
