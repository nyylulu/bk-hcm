import { defineComponent, ref, onMounted, PropType, watch } from 'vue';
import apiService from '../../../../api/scrApi';
import isEqual from 'lodash/isEqual';
export default defineComponent({
  name: 'AreaSelector',
  props: {
    value: {
      type: String as PropType<string>,
      default: '',
    },
  },
  emits: ['change'],
  setup(props, { emit }) {
    const selectedValue = ref('');
    const options = ref([]);
    const handleSelectorChange = (value: string) => {
      emit('change', value);
    };
    const fetchOptions = async () => {
      const { info } = await apiService.getDiskTypes();
      options.value = info.map((item) => {
        return {
          value: item.disk_type,
          label: item.disk_name,
        };
      });
    };
    watch(
      () => props.value,
      (val) => {
        if (val && !isEqual(val, selectedValue.value)) {
          selectedValue.value = val;
        }
      },
      { immediate: true },
    );
    onMounted(() => {
      fetchOptions();
    });

    return () => (
      <bk-select filterable default-first-option v-model={selectedValue.value} onChange={handleSelectorChange}>
        {options.value.map((item) => (
          <bk-option key={item.value} value={item.value} label={item.label}></bk-option>
        ))}
      </bk-select>
    );
  },
});
