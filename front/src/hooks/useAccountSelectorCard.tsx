import CommonCard from '@/components/CommonCard';
import { Form } from 'bkui-vue';
import { defineComponent, PropType, ref, watch, nextTick } from 'vue';
import AccountSelector from '@/components/account-selector/index-new.vue';
import { VendorEnum } from '@/common/constant';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import { useAccountSelectorStore } from '@/store/account-selector';
import { useWhereAmI } from './useWhereAmI';

const { FormItem } = Form;

export const useAccountSelectorCard = () => {
  const isAccountShow = ref(false);

  const AccountSelectorCard = defineComponent({
    props: {
      modelVale: String,
      bkBizId: Number,
      onAccountChange: Function as PropType<(acount: any) => void>,
      disabled: Boolean,
      placeholder: String,
    },
    emits: ['update:modelValue', 'vendorChange'],
    setup(props, { emit }) {
      const resourceAccountStore = useResourceAccountStore();
      const accountSelectorStore = useAccountSelectorStore();

      const { isBusinessPage } = useWhereAmI();

      const selectedVal = ref(props.modelVale);
      watch(
        () => selectedVal.value,
        (val) => {
          emit('update:modelValue', val);
        },
      );
      watch(
        () => accountSelectorStore.businessAccountList,
        async (accountList) => {
          const selectedAccount = accountList.find((account) => account.vendor === VendorEnum.ZIYAN);
          selectedVal.value = selectedAccount?.id;
          isAccountShow.value = true;
        },
        { deep: true },
      );
      const handleChange = async (account: any) => {
        isAccountShow.value = account?.vendor === VendorEnum.ZIYAN;
        nextTick(() => {
          props.onAccountChange?.(account);
        });
      };

      watch(
        () => resourceAccountStore.resourceAccount?.id,
        (id) => {
          // 限定切换账号时，才可以切换当前云账号的表单值。避免其他表单项因为依赖 account_id 而导致请求失败（account_id为空）
          id && (selectedVal.value = id);
          // 切换自研云账号时，清空相关表单值
          if (id && resourceAccountStore.resourceAccount.vendor === VendorEnum.ZIYAN) {
            selectedVal.value = '';
            resourceAccountStore.setResourceAccount(null);
            emit('vendorChange', '');
          }
        },
        { immediate: true },
      );

      return () => (
        <div>
          <CommonCard class='mb16' title={() => '所属账号'} layout={'grid'}>
            <Form formType='vertical'>
              <FormItem label={'云账号'} required>
                <AccountSelector
                  v-model={selectedVal.value}
                  bizId={props.bkBizId}
                  disabled={props.disabled}
                  placeholder={props.placeholder}
                  onChange={handleChange}
                  style={isBusinessPage ? { width: '620px' } : {}}
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
