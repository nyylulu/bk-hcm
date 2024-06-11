import { defineComponent, ref, watch } from 'vue';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
export default defineComponent({
  props: {
    modelValue: {
      type: Boolean,
      default: false,
    },
    title: {
      type: String,
      default: '回收预检详情',
    },
    dataInfo: {
      type: Object,
      default: () => {
        return {};
      },
    },
  },
  emits: ['update:modelValue'],
  setup(props, { attrs, emit }) {
    const { columns } = useColumns('ExecutionRecords');
    const requestParams = ref({});
    const { CommonTable, getListData } = useTable({
      tableOptions: {
        columns,
      },
      requestOption: {
        dataPath: 'data.info',
      },
      scrConfig: () => {
        return {
          payload: {
            ...requestParams.value,
          },
          url: '/api/v1/woa/task/findmany/recycle/detect/step',
        };
      },
    });
    const isDisplay = ref(false);
    watch(
      () => props.modelValue,
      (val) => {
        isDisplay.value = val;
      },
      {
        immediate: true,
      },
    );
    watch(
      () => props.dataInfo,
      () => {
        requestParams.value = {
          suborder_id: [props.dataInfo.suborderId],
          ip: [props.dataInfo.ip],
          page: props.dataInfo.page,
        };
        getListData();
      },
      { deep: true },
    );
    const updateShowValue = () => {
      emit('update:modelValue', false);
    };
    return () => (
      <bk-sideslider
        class='common-sideslider'
        v-bind={attrs}
        width='700'
        v-model:isShow={isDisplay.value}
        title={props.title}
        before-close={updateShowValue}>
        {{
          default: () => (
            <div class='common-sideslider-content'>
              <div class='execute-record-top'>IP : {props.dataInfo.ip}</div>
              <CommonTable />
            </div>
          ),
        }}
      </bk-sideslider>
    );
  },
});
