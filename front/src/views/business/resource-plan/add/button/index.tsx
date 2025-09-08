import { defineComponent, PropType, inject, ref } from 'vue';
import { Message } from 'bkui-vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { useResourcePlanStore } from '@/store';
import cssModule from './index.module.scss';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import type { IPlanTicket } from '@/typings/resourcePlan';
import { MENU_BUSINESS_TICKET_RESOURCE_PLAN_DETAILS } from '@/constants/menu-symbol';

export default defineComponent({
  props: {
    modelValue: Object as PropType<IPlanTicket>,
  },

  setup(props) {
    const { t } = useI18n();
    const router = useRouter();
    const resourcePlanStore = useResourcePlanStore();

    const isLoading = ref(false);

    const validate = inject<() => Promise<void>>('validate');

    const handleClick = async () => {
      try {
        isLoading.value = true;
        await validate();
        const data = await resourcePlanStore.createBizPlan(props.modelValue, props.modelValue.bk_biz_id);
        router.push({
          name: MENU_BUSINESS_TICKET_RESOURCE_PLAN_DETAILS,
          query: {
            id: data.data.id,
            [GLOBAL_BIZS_KEY]: props.modelValue.bk_biz_id,
          },
        });
      } catch (error: any) {
        Message({
          message: error.message || error,
          theme: 'error',
        });
      } finally {
        isLoading.value = false;
      }
    };

    const handleCancel = () => {
      router.push({
        path: '/business/resource-plan',
      });
    };

    return () => (
      <section>
        <bk-button onClick={handleClick} loading={isLoading.value} theme='primary' class={cssModule.button}>
          {t('提交')}
        </bk-button>
        <bk-button onClick={handleCancel} disabled={isLoading.value} class={cssModule.button}>
          {t('取消')}
        </bk-button>
      </section>
    );
  },
});
