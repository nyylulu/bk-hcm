import { defineComponent, ref, watch, onMounted, PropType } from 'vue';
import apiService from '../../../../api/scrApi';
export default defineComponent({
  props: {
    value: {
      type: String as PropType<string>,
      default: '',
    },
    area: {
      type: String as PropType<string>,
      default: '',
    },
    zone: {
      type: String as PropType<string>,
      default: '',
    },
  },
  emits: ['change'],
  setup(props, { emit }) {
    const options = ref<{ label: string; value: string }[]>([]);
    const loading = ref(false);

    const loadOptions = async () => {
      if (!props.zone && !props.area) return;
      loading.value = true;
      try {
        const res = await apiService.getCvmTypes(props.zone, props.area);
        options.value = res.data.cvmType;
      } finally {
        loading.value = false;
      }
    };

    const handleSelectorChange = (value: string) => {
      emit('change', value);
    };

    watch(
      () => props.zone,
      (newVal) => {
        if (newVal) {
          loadOptions();
          emit('change', '');
        }
      },
    );

    watch(
      () => props.area,
      (newVal) => {
        if (newVal) {
          loadOptions();
          emit('change', '');
        }
      },
    );

    onMounted(() => {
      if (props.zone) {
        loadOptions();
      }
    });

    return () => (
      <bk-select
        filterable
        default-first-option
        modelValue={props.value}
        loading={loading.value}
        onChange={handleSelectorChange}>
        {options.value.map((item, index) => (
          <bk-option key={index} value={item.value} label={item.label}></bk-option>
        ))}
      </bk-select>
    );
  },
});
