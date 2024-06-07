import { defineComponent, ref, watch, onMounted } from 'vue';
import apiService from '../../../../api/scrApi';
import isEqual from 'lodash/isEqual';

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
  setup(props, { emit }) {
    const selectedValue = ref(props.value);
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

    watch(
      () => props.value,
      (newValue) => {
        if (!isEqual(newValue, selectedValue.value)) {
          selectedValue.value = newValue;
        }
      },
      { immediate: true },
    );

    watch(
      () => selectedValue.value,
      (newValue) => {
        emit('value-change', newValue);
      },
      { immediate: true },
    );

    watch(
      () => props.params.resourceType,
      () => {
        fetchOptions();
      },
      { immediate: true },
    );

    onMounted(() => {
      fetchOptions();
    });

    return () => (
      <div>
        <bk-select v-model={selectedValue.value}>
          {options.value.map((opt) => (
            <bk-option key={opt.value} label={opt.label} value={opt.value} />
          ))}
        </bk-select>
        {/* <HelpTooltip class='ml-5'>
          1. 默认情况下，Docker 服务器会尽量打散母机
          <br />
          2. 未指定 Campus 情况下，选择“分Campus”会保障均分到至少 2 个 Campus
        </HelpTooltip> */}
      </div>
    );
  },
});
