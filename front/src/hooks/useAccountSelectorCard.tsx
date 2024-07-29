import CommonCard from '@/components/CommonCard';
import { Form } from 'bkui-vue';
import { defineComponent, PropType, ref, watch, nextTick } from 'vue';
import AccountSelector from '@/components/account-selector/index.vue';
import { VendorEnum } from '@/common/constant';

const { FormItem } = Form;

export const useAccountSelectorCard = () => {
  const isAccountShow = ref(false);

  const AccountSelectorCard = defineComponent({
    props: {
      modelVale: String,
      bkBizId: String,
      onAccountChange: Function as PropType<(acount: any) => void>,
    },
    emits: ['update:modelValue'],
    setup(props, { emit }) {
      const selectedVal = ref(props.modelVale);
      const AccountSelectorRef = ref();
      watch(
        () => selectedVal.value,
        (val) => {
          emit('update:modelValue', val);
        },
      );
      watch(
        () => AccountSelectorRef.value?.accountList,
        async (newAccountList) => {
          if (newAccountList && Array.isArray(newAccountList)) {
            await nextTick(); // 等待下一次 DOM 更新周期
            const proxyArray = [...AccountSelectorRef.value.accountList];
            const selectedAccount = proxyArray.find((account) => account.vendor === 'tcloud-ziyan');
            selectedVal.value = selectedAccount?.id;
            isAccountShow.value = true;
          }
        },
        { deep: true },
      );
      const handleChange = (account: any) => {
        isAccountShow.value = account?.vendor === VendorEnum.ZIYAN;
        props.onAccountChange?.(account);
      };

      return () => (
        <div>
          <CommonCard class='mb16' title={() => '所属账号'} layout={'grid'}>
            <Form formType='vertical'>
              <FormItem label={'云账号'} required>
                <AccountSelector
                  ref={AccountSelectorRef}
                  v-model={selectedVal.value}
                  must-biz={true}
                  biz-id={props.bkBizId}
                  onChange={handleChange}
                  type={'resource'}
                />
              </FormItem>
            </Form>
          </CommonCard>
        </div>
      );
    },
  });

  return {
    AccountSelectorCard,
    isAccountShow,
  };
};
