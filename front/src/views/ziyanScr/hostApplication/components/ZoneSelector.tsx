import { defineComponent, ref, watch, onMounted, PropType } from 'vue';
import apiService from '../../../../api/scrApi';
export default defineComponent({
  name: 'ZoneSelector',
  props: {
    value: {
      type: String as PropType<string>,
      default: '',
    },
    area: {
      type: String as PropType<string>,
      required: true,
      default: '',
    },
  },
  emits: ['change'],
  setup(props, { emit, attrs }) {
    const options = ref<{ label: string; value: string }[]>([]);
    const loading = ref(false);

    const loadOptions = async () => {
      if (props.area) {
        loading.value = true;
        try {
          const res = await apiService.getZones(props.area);
          options.value = res.data.zoneList;
        } finally {
          loading.value = false;
        }
      }
    };

    const handleSelectorChange = (value: string) => {
      emit('change', value);
    };

    // const selectedLabel = computed(() => {
    //   const { value } = props;
    //   return value ? options.value.find((i) => i.value === value)?.label || '' : '';
    // });

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
      if (props.area) {
        loadOptions();
      }
    });

    return () => (
      <bk-select
        filterable
        default-first-option
        v-model={props.value}
        loading={loading.value}
        {...attrs}
        onChange={handleSelectorChange}>
        {options.value.map((item, index) => (
          <bk-option key={index} value={item.value} label={item.label}></bk-option>
        ))}
      </bk-select>
    );
  },
});
