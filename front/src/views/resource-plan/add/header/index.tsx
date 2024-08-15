import { defineComponent } from 'vue';
import { useI18n } from 'vue-i18n';
import cssModule from './index.module.scss';
import { useRouter } from 'vue-router';

export default defineComponent({
  setup() {
    const { t } = useI18n();
    const router = useRouter();

    const handleClick = () => {
      router.push({
        path: '/service/resource-plan/list',
      });
    };

    return () => (
      <span class={cssModule.home}>
        <i class={`${cssModule.arrow} hcm-icon bkhcm-icon-arrows--left-line`} onClick={handleClick}></i>
        {t('新增资源预测')}
      </span>
    );
  },
});
