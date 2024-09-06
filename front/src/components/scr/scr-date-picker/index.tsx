import { PropType, defineComponent, ref, watch } from 'vue';
import { DatePicker } from 'bkui-vue';
import { timeFormatter } from '@/common/util';

export default defineComponent({
  props: {
    modelValue: {
      type: Array as PropType<string[]>,
    },
  },
  setup(props, { emit }) {
    const range = ref([...props.modelValue]);
    watch(
      () => range.value,
      (newVal, oldVal) => {
        if (String(newVal) === String(oldVal)) return;
        emit(
          'update:modelValue',
          newVal.map((date) => timeFormatter(date, 'YYYY-MM-DD')),
        );
      },
    );

    watch(
      () => props.modelValue,
      (newVal, oldVal) => {
        if (String(newVal) === String(oldVal)) return;
        range.value = [...newVal];
      },
    );

    return () => <DatePicker v-model={range.value} type={'daterange'} />;
  },
});
