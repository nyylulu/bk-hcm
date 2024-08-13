import CommonCard from '@/components/CommonCard';
import { Form } from 'bkui-vue';
import { defineComponent, PropType, ref, watch, nextTick, computed } from 'vue';
import AccountSelector from '@/components/account-selector/index.vue';
import { VendorEnum } from '@/common/constant';
import { Senarios, useWhereAmI } from './useWhereAmI';
import { FilterType, QueryRuleOPEnum } from '@/typings';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';

const { FormItem } = Form;

export const useAccountSelectorCard = () => {
  const isAccountShow = ref(false);

  const AccountSelectorCard = defineComponent({
    props: {
      modelVale: String,
      bkBizId: String,
      onAccountChange: Function as PropType<(acount: any) => void>,
      filter: {
        type: Object as PropType<FilterType>,
        default() {
          return { op: QueryRuleOPEnum.AND, rules: [] };
        },
      },
    },
    emits: ['update:modelValue', 'vendorChange'],
    setup(props, { emit }) {
      const resourceAccountStore = useResourceAccountStore();

      const selectedVal = ref(props.modelVale);
      const isDisabled = computed(() => {
        return (
          resourceAccountStore.resourceAccount?.id && resourceAccountStore.resourceAccount?.vendor !== VendorEnum.ZIYAN
        );
      });
      const { whereAmI, isBusinessPage } = useWhereAmI();
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
          if (whereAmI.value !== Senarios.business) return;
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
      const handleChange = async (account: any) => {
        isAccountShow.value = account?.vendor === VendorEnum.ZIYAN;
        await nextTick();
        props.onAccountChange?.(account);
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
                  ref={AccountSelectorRef}
                  v-model={selectedVal.value}
                  must-biz={isBusinessPage}
                  biz-id={props.bkBizId}
                  onChange={handleChange}
                  type={'resource'}
                  filter={props.filter}
                  disabled={isDisabled.value}
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
