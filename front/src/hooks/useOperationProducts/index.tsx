import useBillStore from '@/store/useBillStore';
import { Select } from 'bkui-vue';
import { defineComponent, onMounted, reactive, ref, watch } from 'vue';
import './index.scss';
const { Option } = Select;

export const useOperationProducts = () => {
  const billStore = useBillStore();
  const list = ref([]);
  const pagination = reactive({
    limit: 200,
    start: 0,
  });
  const allCounts = ref(0);

  const getList = async (op_product_name?: string, op_product_ids?: number[]) => {
    const [detailRes, countRes] = await Promise.all(
      [false, true].map((isCount) =>
        billStore.list_operation_products({
          op_product_name,
          op_product_ids,
          page: {
            start: op_product_name === undefined && op_product_ids === undefined ? pagination.start : 0,
            limit: isCount ? 0 : pagination.limit,
            count: isCount,
          },
        }),
      ),
    );
    allCounts.value = countRes.data.count;
    list.value = detailRes.data.details;
    return list.value;
  };

  const getTranslatorMap = async (ids: number[]) => {
    const map = new Map();
    const dataList = await getList(undefined, ids);
    for (const { op_product_id, op_product_name } of dataList) {
      map.set(op_product_id, op_product_name);
    }
    return map;
  };

  onMounted(() => {
    getList();
  });

  const OperationProductsSelector = defineComponent({
    props: {
      modelValue: String,
      isShowManagers: Boolean,
      multiple: Boolean,
    },
    setup(props, { emit }) {
      const selectedVal = ref(props.modelValue);
      const isScrollLoading = ref(false);

      watch(
        () => selectedVal.value,
        (val) => {
          emit('update:modelValue', val);
        },
      );

      return () => (
        <div class={'selector-wrapper'}>
          <Select
            v-model={selectedVal.value}
            scrollLoading={isScrollLoading.value}
            filterable
            multiple={props.multiple}
            multipleMode={props.multiple ? 'tag' : undefined}
            remoteMethod={(val) => getList(val)}
            onScroll-end={async () => {
              if (list.value.length >= allCounts.value || isScrollLoading.value) return;
              isScrollLoading.value = true;
              pagination.start += pagination.limit;
              const { data } = await billStore.list_operation_products({
                page: {
                  start: pagination.start,
                  count: false,
                  limit: pagination.limit,
                },
              });
              list.value.push(...data.details);
              isScrollLoading.value = false;
            }}>
            {list.value.map(({ op_product_name, op_product_id, op_product_managers }) => (
              <Option name={op_product_name} id={op_product_id} key={op_product_id}>
                <span>
                  {op_product_name}
                  {props.isShowManagers && (
                    <span class={'op-production-memo-info'}>&nbsp;({`负责人: ${op_product_managers}`})</span>
                  )}
                </span>
              </Option>
            ))}
          </Select>
        </div>
      );
    },
  });

  return {
    OperationProductsSelector,
    getTranslatorMap,
  };
};
