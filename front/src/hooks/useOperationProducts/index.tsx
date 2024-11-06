import useBillStore from '@/store/useBillStore';
import { Select } from 'bkui-vue';
import { defineComponent, onMounted, reactive, Ref, ref, watch } from 'vue';
import './index.scss';
import { SelectColumn } from '@blueking/ediatable';
const { Option } = Select;

export const useOperationProducts = (
  immediate = true,
  filterParams?: Ref<{ op_product_ids?: number[]; op_product_name?: string; dept_ids?: number[]; bg_ids?: number[] }>,
) => {
  const billStore = useBillStore();
  const list = ref([]);
  const pagination = reactive({ limit: 200, start: 0 });
  const allCounts = ref(0);

  const getList = async (op_product_name?: string, op_product_ids?: number[]) => {
    const [detailRes, countRes] = await Promise.all(
      [false, true].map((isCount) =>
        billStore.list_operation_products({
          op_product_name,
          op_product_ids,
          page: {
            start: op_product_name === undefined && op_product_ids === undefined && !isCount ? pagination.start : 0,
            limit: isCount ? 0 : pagination.limit,
            count: isCount,
          },
          ...(filterParams?.value || {}),
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

  const getAppendixList = async (id?: number) => {
    if (list.value.findIndex((v) => v.op_product_id === id) !== -1) return;
    await getList();
    const detailRes = await billStore.list_operation_products({
      op_product_ids: [id],
      page: { start: 0, limit: pagination.limit, count: false },
      ...(filterParams?.value || {}),
    });
    list.value = list.value.concat(detailRes.data.details);
  };

  const OperationProductsSelector = defineComponent({
    props: {
      modelValue: String,
      isShowManagers: Boolean,
      multiple: Boolean,
      isEdiatable: Boolean,
    },
    emits: ['update:modelValue'],
    setup(props, { emit, expose }) {
      const selectedVal = ref(props.modelValue);
      const isScrollLoading = ref(false);
      const selectRef = ref();

      const getValue = () => {
        return selectRef.value.getValue().then(() => selectedVal.value);
      };

      const handleScrollEnd = async () => {
        if (list.value.length >= allCounts.value || isScrollLoading.value) return;
        isScrollLoading.value = true;
        pagination.start += pagination.limit;
        const { data } = await billStore.list_operation_products({
          page: { start: pagination.start, count: false, limit: pagination.limit },
          ...(filterParams?.value || {}),
        });
        list.value.push(...data.details);
        isScrollLoading.value = false;
      };

      const handleClear = () => {
        selectedVal.value = '';
      };

      watch(selectedVal, (val) => emit('update:modelValue', val), { deep: true });

      watch(
        () => props.modelValue,
        (val) => (selectedVal.value = val),
        { deep: true },
      );

      onMounted(() => {
        immediate && getList();
      });

      expose({
        getValue,
      });

      if (props.isEdiatable)
        return () => (
          <SelectColumn
            v-model={selectedVal.value}
            ref={selectRef}
            scrollLoading={isScrollLoading.value}
            filterable
            list={list.value.map((v) => ({
              label: v.op_product_name,
              key: v.op_product_id,
              value: v.op_product_id,
            }))}
            multiple={props.multiple}
            multipleMode={props.multiple ? 'tag' : undefined}
            remoteMethod={(val) => getList(val)}
            onScroll-end={handleScrollEnd}
            onClear={handleClear}
            rules={[
              {
                validator: (value: string) => Boolean(value),
                message: '运营产品不能为空',
              },
            ]}
          />
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
            onScroll-end={handleScrollEnd}
            onClear={handleClear}>
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
    getAppendixList,
  };
};
