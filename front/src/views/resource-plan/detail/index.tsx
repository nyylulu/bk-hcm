import { defineComponent, ref, reactive } from 'vue';
import Approval from './approval';
import Basic from './basic';
import List from './list';

import { Form } from 'bkui-vue';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useTable } from '@/hooks/useTable/useTable';
import Panel from '@/components/panel';

export default defineComponent({
  setup(props) {
    const { FormItem } = Form;
    const formRef = ref();
    const formModel = reactive({
      forecast_order_number: '' as string, // 预测单号
    });
    const detailTableData = [
      {
        region: '华南地区',
      },
    ];
    const { columns, settings } = useColumns('forecastDemandDetail');

    const { CommonTable } = useTable({
      searchOptions: {
        disabled: true,
      },
      tableOptions: {
        columns,
        reviewData: detailTableData,
        extra: {
          settings: settings.value,
        },
      },
      requestOption: {
        type: '',
      },
    });
    return () => (
      <section>
        <Approval></Approval>
        <Basic></Basic>
        <List></List>
        <Panel title='资源预测详情'>
          <Form ref={formRef} formType='vertical' model={formModel}>
            <FormItem key='base' label='基础信息'>
              {props.baseMap?.map((item) => {
                return (
                  <div>
                    <span>{item.label}：</span>
                    <span>{item.value}</span>
                  </div>
                );
              })}
            </FormItem>
            <FormItem key='forecast' label='资源预测'>
              <CommonTable></CommonTable>
            </FormItem>
          </Form>
        </Panel>
      </section>
    );
  },
});
