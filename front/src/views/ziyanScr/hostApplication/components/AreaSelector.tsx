import { defineComponent, ref, onMounted, PropType } from 'vue';
import apiService from '../../../../api/scrApi';
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
    const options = ref<{ label: string; value: string }[]>([]);
    const loading = ref(false);

    const loadOptions = async () => {
      if (!options.value.length) {
        loading.value = true;
        try {
          const res = await apiService.getAreas();
          options.value = res.data.areaList;
        } finally {
          loading.value = false;
        }
      }
    };

    const handleSelectorChange = (value: string) => {
      emit('change', value);
    };
    onMounted(() => {
      loadOptions();
    });

    return () => (
      <bk-select
        filterable
        default-first-option
        v-model={props.value}
        loading={loading.value}
        onChange={handleSelectorChange}>
        {options.value.map((item, index) => (
          <bk-option key={index} value={item.value} label={item.label}></bk-option>
        ))}
      </bk-select>
    );
  },
});
