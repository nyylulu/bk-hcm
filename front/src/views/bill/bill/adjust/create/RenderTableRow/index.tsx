import { PropType, defineComponent, ref, watch } from 'vue';
import { InputColumn, OperationColumn, TextPlainColumn } from '@blueking/ediatable';
import AdjustTypeSelector, { AdjustTypeEnum } from './components/AdjustTypeSelector';
import SubAccountSelector from '../../../components/search/sub-account-selector';
import { VendorEnum } from '@/common/constant';
import { useOperationProducts } from '@/hooks/useOperationProducts';
import useFormModel from '@/hooks/useFormModel';

export default defineComponent({
  props: {
    removeable: {
      required: true,
      type: Boolean,
      default: false,
    },
    vendor: {
      required: true,
      type: String as PropType<VendorEnum>,
    },
    rootAccountId: {
      required: true,
      type: String,
    },
    editData: {
      required: true,
      type: Object,
      default: {},
    },
    edit: {
      required: true,
      type: Boolean,
    },
  },
  emits: ['add', 'remove', 'copy', 'change'],
  setup(props, { emit, expose }) {
    const { formModel, resetForm, setFormValues } = useFormModel({
      type: AdjustTypeEnum.Increase,
      product_id: '',
      main_account_id: '',
      cost: '',
      memo: '',
    });

    const costRef = ref();
    const memoRef = ref();
    const productRef = ref();
    const mainAccountRef = ref();

    const { OperationProductsSelector, getAppendixList } = useOperationProducts(!props.edit);

    const handleAdd = () => {
      emit('add');
    };

    const handleRemove = () => {
      emit('remove');
    };

    const handleCopy = () => {
      emit('copy', formModel);
    };

    watch(
      () => props.editData,
      (data) => {
        if (data.product_id) getAppendixList(data.product_id);
        setFormValues(data);
      },
      {
        deep: true,
        immediate: true,
      },
    );

    watch(
      () => formModel,
      (val) => {
        emit('change', val);
      },
      {
        deep: true,
      },
    );

    watch(
      () => props.rootAccountId,
      () => {
        formModel.main_account_id = '';
      },
    );

    expose({
      getValue: async () => {
        return await Promise.all([
          costRef.value!.getValue(),
          memoRef.value!.getValue(),
          productRef.value!.getValue(),
          mainAccountRef.value!.getValue(),
        ]).then(() => {
          return formModel;
        });
      },
      reset: resetForm,
      getRowValue: () => {
        return formModel;
      },
    });

    return () => (
      <>
        <tr>
          <td>
            <AdjustTypeSelector v-model={formModel.type} />
          </td>
          <td>
            <OperationProductsSelector v-model={formModel.product_id} ref={productRef} isEdiatable />
          </td>
          <td>
            <SubAccountSelector
              isEditable={true}
              v-model={formModel.main_account_id}
              ref={mainAccountRef}
              vendor={[props.vendor]}
              rootAccountId={[props.rootAccountId]}
            />
          </td>
          <td>
            <TextPlainColumn>人工调账</TextPlainColumn>
          </td>
          <td>
            <InputColumn
              type='number'
              precision={3}
              ref={costRef}
              v-model={formModel.cost}
              rules={[
                {
                  validator: (value: string) => Boolean(value),
                  message: '金额不能为空',
                },
              ]}
            />
          </td>
          <td>
            <InputColumn ref={memoRef} v-model={formModel.memo} />
          </td>
          {!props.edit && (
            <OperationColumn
              removeable={props.removeable}
              onAdd={handleAdd}
              onRemove={handleRemove}
              showCopy
              onCopy={handleCopy}
            />
          )}
        </tr>
      </>
    );
  },
});
