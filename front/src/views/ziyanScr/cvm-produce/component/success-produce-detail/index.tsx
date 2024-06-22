import { defineComponent, ref, watch } from 'vue';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import { Dialog, Table } from 'bkui-vue';
export default defineComponent({
  props: {
    modelValue: {
      type: Boolean,
      default: false,
    },
    title: {
      type: String,
      default: '成功生产资源详情',
    },
    tableData: {
      type: Array,
      default: () => {
        return [];
      },
    },
  },
  emits: ['update:modelValue'],
  setup(props, { attrs, emit }) {
    const { columns } = useColumns('cvmProduceDetailQuery');
    const isDisplay = ref(false);
    watch(
      () => props.modelValue,
      (val) => {
        isDisplay.value = val;
      },
      {
        immediate: true,
      },
    );
    const updateShowValue = () => {
      emit('update:modelValue', false);
    };
    return () => (
      <Dialog
        v-bind={attrs}
        width='700'
        v-model:isShow={isDisplay.value}
        title={props.title}
        onClosed={updateShowValue}>
        {{
          default: () => (
            <div>
              <Table columns={columns} data={props.tableData} />
            </div>
          ),
          footer: () => null,
        }}
      </Dialog>
    );
  },
});
