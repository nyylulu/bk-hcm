import { PropType, defineComponent, onMounted, ref, watch } from 'vue';
import { Select } from 'bkui-vue';
import './index.scss';

export default defineComponent({
  name: 'ScrCreateFilterSelector',
  props: { modelValue: Array as PropType<string[]>, api: Function as PropType<() => Promise<any>> },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    const list = ref([]);
    const selected = ref(props.modelValue);

    onMounted(() => {
      const getOptionList = async () => {
        const res = await props.api();
        list.value = res.data.info || [];
      };

      getOptionList();
    });

    watch(
      selected,
      (val) => {
        emit('update:modelValue', val);
      },
      { deep: true },
    );

    watch(
      () => props.modelValue,
      (val) => {
        selected.value = val;
      },
      {
        deep: true,
      },
    );

    return () => (
      <Select v-model={selected.value} multiple multipleMode='tag' collapseTags>
        {list.value.map((item) => (
          <Select.Option key={item} id={item} name={item} />
        ))}
      </Select>
    );
  },
});
