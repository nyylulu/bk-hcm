import { defineComponent, PropType, inject } from 'vue';
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

    const validate = inject<() => Promise<void>>('validate');

    const handleClick = async () => {
      await validate();
      await resourcePlanStore.createPlan(props.modelValue);
      router.back();
    };

    const handleCancel = () => {
      router.back();
    };

    return () => (
      <section>
        <bk-button onClick={handleClick} theme='primary' class={cssModule.button}>
          {t('提交')}
        </bk-button>
        <bk-button onClick={handleCancel} class={cssModule.button}>
          {t('取消')}
        </bk-button>
      </section>
    );
  },
});
