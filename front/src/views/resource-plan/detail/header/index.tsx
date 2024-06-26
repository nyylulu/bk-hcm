import { defineComponent } from 'vue';
import { useI18n } from 'vue-i18n';
import cssModule from './index.module.scss';
import { useRouter } from 'vue-router';

export default defineComponent({
  props: {
    id: {
      type: String,
    },
  },

  setup(props) {
    const { t } = useI18n();
    const router = useRouter();

    const handleClick = () => {
      router.push({
        path: '/resource-plan/list',
      });
    };

    return () => (
      <span class={cssModule.home}>
        <i class={`${cssModule.arrow} hcm-icon bkhcm-icon-arrows--left-line`} onClick={handleClick}></i>
        {t('申请单详情')}
        <span class={cssModule.id}> - {props.id}</span>
      </span>
    );
  },
});
