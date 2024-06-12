import { defineComponent, ref, watch, PropType, computed } from 'vue';
import apiService from '../../../../api/scrApi';
import isEqual from 'lodash/isEqual';
import isEmpty from 'lodash/isEmpty';
export default defineComponent({
  name: 'ZoneSelector',
  props: {
    value: {
      type: String as PropType<string>,
      default: '',
    },
    params: {
      type: Object,
      default: () => ({}),
    },
    valueKey: {
      type: String,
      default: '',
      validator: (val) => ['cmdbZoneName', ''].includes(val),
    },
    separateCampus: {
      type: Boolean,
      default: true,
    },
  },
  emits: ['change'],
  setup(props, { attrs, emit }) {
    const options = ref<{ label: string; value: string }[]>([]);
    const optionsRequestId = ref();
    const selectedValue = ref();
    const regionsIsEmpty = computed(() => {
      return isEmpty(props.params.region);
    });
    const regions = computed(() => {
      if (typeof props.params.region === 'string') {
        if (!props.params.region) return [];
        return [props.params.region];
      }

      return props.params.region || [];
    });
    const handleSelectorChange = (value: string) => {
      emit('change', value);
    };
    const initValue = () => {
      selectedValue.value = props.value;
    };
    const fetchOptions = () => {
      const { resourceType } = props.params;

      if (['QCLOUDCVM', 'QCLOUDDVM'].includes(resourceType)) {
        loadQcloudZones();
      }

      if (['IDCDVM', 'IDCPM'].includes(resourceType)) {
        loadIdcZones();
      }

      initValue();
    };
    const loadQcloudZones = async () => {
      const { info } = await apiService.getZones(
        { vendor: 'qcloud', region: regions.value, isCmdbRegion: props.valueKey === 'cmdbZoneName' },
        {
          requestId: optionsRequestId.value,
        },
      );

      options.value = info.map((item) => {
        return {
          label: `${item.zone_cn}(${item.cmdb_zone_name})`,
          value: props.valueKey === 'cmdbZoneName' ? item.cmdb_zone_name : item.zone,
        };
      });
    };
    const loadIdcZones = async () => {
      const { info } = await apiService.getZones(
        { vendor: 'idc', region: regions.value },
        {
          requestId: optionsRequestId.value,
        },
      );

      options.value = info.map((item) => {
        return {
          value: item.cmdb_zone_name,
          label: item.cmdb_zone_name,
        };
      });
    };
    watch(
      () => props.value,
      (val: any) => {
        if (!isEqual(val, selectedValue.value) && options.value.length !== 0) {
          initValue();
        }
      },
      { immediate: true },
    );
    watch(
      () => props.params,
      (newVal: any, oldVal: any) => {
        if (!isEqual(newVal, oldVal)) {
          fetchOptions();
        }
      },
      { immediate: true },
    );
    return () => (
      <bk-select v-bind={attrs} filterable default-first-option v-model={props.value} onChange={handleSelectorChange}>
        {options.value.map((item, index) => (
          <bk-option key={index} value={item.value} label={item.label}></bk-option>
        ))}
        {props.separateCampus && !regionsIsEmpty.value && <bk-option label='分Campus' value='cvm_separate_campus' />}
      </bk-select>
    );
  },
});
