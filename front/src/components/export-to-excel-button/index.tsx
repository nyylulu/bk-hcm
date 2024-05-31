import { defineComponent } from 'vue';
import { exportTableToExcel } from '@/utils';
import { Plus } from 'bkui-vue/lib/icon';
export default defineComponent({
  props: {
    data: {
      type: Array,
      required: true,
    },
    columns: {
      type: Array,
      required: true,
    },
    filename: {
      type: String,
      default: `导出文件`,
    },
  },
  emits: ['update:modelValue'],
  setup(props, { attrs }) {
    const exportToExcel = () => {
      exportTableToExcel(props.data, props.columns, props.filename);
    };
    return () => (
      <bk-button v-bind={attrs} onClick={exportToExcel}>
        <Plus />
        导出
      </bk-button>
    );
  },
});
