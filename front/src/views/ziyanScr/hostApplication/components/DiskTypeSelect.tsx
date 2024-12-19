import { defineComponent, ref, onMounted, PropType, watch } from 'vue';
import apiService from '../../../../api/scrApi';
import isEqual from 'lodash/isEqual';

interface IDiskType {
  disk_name: string;
  disk_type: string;
}

export default defineComponent({
  name: 'AreaSelector',
  props: {
    value: { type: String as PropType<string | string[]>, default: '' },
    multiple: { type: Boolean, default: false },
  },
  emits: ['change'],
  setup(props, { emit }) {
    const selectedValue = ref<string | string[]>(props.multiple ? [] : '');
    const isLoading = ref(false);
    const options = ref([]);

    const handleSelectorChange = (value: string | string[]) => {
      emit('change', value);
    };

    const fetchOptions = async () => {
      isLoading.value = true;
      try {
        const { info } = await apiService.getDiskTypes();
        options.value = info.map(({ disk_type, disk_name }: IDiskType) => ({ value: disk_type, label: disk_name }));
      } catch (error) {
        console.error(error);
        options.value = [];
      } finally {
        isLoading.value = false;
      }
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
      <bk-select
        v-model={selectedValue.value}
        loading={isLoading.value}
        multiple={props.multiple}
        filterable
        onChange={handleSelectorChange}>
        {options.value.map((item) => (
          <bk-option key={item.value} value={item.value} label={item.label}></bk-option>
        ))}
      </bk-select>
    );
  },
});
