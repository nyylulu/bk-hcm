import { defineComponent, ref, onMounted } from 'vue';
import apiService from '../../../../api/scrApi';

export default defineComponent({
  name: 'AntiAffinityLevelSelect',
  props: {
    value: {
      type: [String, Array],
      default: '',
    },
    params: {
      type: Object,
      default: () => ({}),
    },
  },
  emits: ['value-change'],
  setup(props) {
    const options = ref<{ value: string; label: string }[]>([]);
    const optionsRequestId = Symbol();

    const fetchOptions = async () => {
      const { resourceType, hasZone } = props.params;

      try {
        const { info } = await apiService.getAntiAffinityLevels(resourceType, hasZone, {
          requestId: optionsRequestId,
        });
        options.value =
          info.map((item: any) => ({
            value: item.level,
            label: item.description,
          })) || [];
      } catch {
        options.value = [];
      }
    };
    onMounted(() => {
      fetchOptions();
    });

    return () => (
      <div>
        <bk-select v-model={props.value}>
          {options.value.map((opt) => (
            <bk-option key={opt.value} label={opt.label} value={opt.value} />
          ))}
        </bk-select>
      </div>
    );
  },
});
