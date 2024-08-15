import ScrCreateFilterSelector from '@/views/ziyanScr/resource-manage/create/ScrCreateFilterSelector';
import { defineComponent, ref, watch } from 'vue';
import http from '@/http';
import './index.scss';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default defineComponent({
  props: {
    modelValue: {
      type: String,
    },
    multiple: {
      type: Boolean,
      default: false,
    },
  },
  setup(props, { emit }) {
    const selected = ref(props.modelValue);
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
      <ScrCreateFilterSelector
        v-model={selected.value}
        api={() => http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/find/config/apply/stage`)}
        multiple={props.multiple}
        optionIdPath='stage'
        optionNamePath='description'
      />
    );
  },
});
