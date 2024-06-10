import { defineComponent, PropType, inject, ref } from 'vue';
import { Message } from 'bkui-vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { useResourcePlanStore } from '@/store';
import cssModule from './index.module.scss';

import type { IPlanTicket } from '@/typings/resourcePlan';

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
        await resourcePlanStore.createPlan(props.modelValue);
        router.push({
          path: '/resource-plan/list',
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
        path: '/resource-plan/list',
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
