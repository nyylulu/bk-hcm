import { defineComponent, ref } from 'vue';

import { Button, Table } from 'bkui-vue';
import CommonSideslider from '@/components/common-sideslider';
import SuborderDetailDialog, { SubOrderInfo } from '../suborder-detail';

import { timeFormatter } from '@/common/util';

export default defineComponent({
  props: { details: Array },
  setup(props, { expose }) {
    const isSidesliderShow = ref(false);
    const suborderDetailDialogRef = ref();

    const curSuborder = ref<SubOrderInfo>({
      step_name: '',
      step_id: 1,
      suborder_id: 0,
    });

    const column = [
      { field: 'step_id', label: 'ID', width: 80 },
      { field: 'step_name', label: '步骤名称', width: 100 },
      {
        field: 'status',
        label: '状态',
        width: 80,
        render({ data }: any) {
          if (data.status === -1) return <span>未执行</span>;
          if (data.status === 0) return <span>成功</span>;
          if (data.status === 1) return <span>执行中</span>;
          return <span>失败</span>;
        },
      },
      { field: 'message', label: '状态说明', width: 100 },
      {
        label: '概要',
        width: 220,
        render({ data }: any) {
          return (
            <div>
              <span>
                <span class='c-text-2 fz-12'>总数：</span>
                <span>{data.total_num || '-'}</span>
              </span>
              <span class='ml-10'>
                <span class='c-text-2 fz-12'>成功：</span>
                <span class='c-success'>{data.success_num || '-'}</span>
              </span>
              <span class='ml-10'>
                <span class='c-text-2 fz-12'>进行中：</span>
                <span>{data.running_num || '-'}</span>
              </span>
              <span class='ml-10'>
                <span class='c-text-2 fz-12'>失败：</span>
                <span class='c-danger'>{data.fail_num || '-'}</span>
              </span>
            </div>
          );
        },
      },
      {
        field: 'start_at',
        label: '开始时间',
        render: ({ data }: any) => (data.status === -1 ? '-' : timeFormatter(data.start_at)),
      },
      {
        field: 'end_at',
        label: '结束时间',
        render: ({ data }: any) => (![0, 2].includes(data.status) ? '-' : timeFormatter(data.end_at)),
      },
      {
        field: 'operation',
        label: '操作',
        width: 100,
        render: ({ data }: any) => (
          <div>
            {data.step_id > 1 ? (
              <Button
                text
                theme='primary'
                onClick={() => {
                  curSuborder.value = data;
                  suborderDetailDialogRef.value.triggerShow(true);
                }}>
                查看详情
              </Button>
            ) : (
              '--'
            )}
          </div>
        ),
      },
    ];

    const triggerShow = (v: boolean) => {
      isSidesliderShow.value = v;
    };

    expose({ triggerShow });

    return () => (
      <>
        <CommonSideslider v-model:isShow={isSidesliderShow.value} title='资源匹配详情' width={1100} noFooter>
          <Table showOverflowTooltip border={['outer', 'col', 'row']} data={props.details} columns={column} />
        </CommonSideslider>
        <SuborderDetailDialog ref={suborderDetailDialogRef} subOrderInfo={curSuborder.value} />
      </>
    );
  },
});
