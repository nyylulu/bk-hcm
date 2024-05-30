import { defineComponent, ref, watch, onMounted } from 'vue';
import apiService from '../../../../api/scrApi';
import type { PropType } from 'vue';

export default defineComponent({
  name: 'ImageSelector',
  props: {
    value: {
      type: String as PropType<string>,
      default: '',
    },
    area: {
      type: String as PropType<string>,
      required: true,
    },
  },
  emits: ['change'],
  setup(props, { emit, attrs }) {
    const options = ref<{ label: string; value: string }[]>([]);
    const loading = ref(false);
    // const selectedLabel = computed(() => {
    //   return props.value ? options.value.find((i) => i.value === props.value)?.label ?? '' : '';
    // });

    const loadOptions = async () => {
      loading.value = true;
      const res = await apiService.getImages(props.area);
      options.value = res.data.imageName;
      loading.value = false;
    };

    const handleSelectorChange = (value: string) => {
      emit('change', value);
    };

    watch(
      () => props.area,
      (newVal) => {
        if (newVal) {
          loadOptions();
          emit('change', ''); // Reset value when area changes
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
        defaultFirstOption
        loading={loading.value}
        value={props.value}
        onChange={handleSelectorChange}
        {...attrs}>
        {options.value.map((item, i) => (
          <bk-option key={i} label={item.label} value={item.value} />
        ))}
      </bk-select>
    );
  },
});
