import { defineComponent } from 'vue';
import { useI18n } from 'vue-i18n';
import cssModule from './index.module.scss';
import { useRouter, useRoute } from 'vue-router';
import { MENU_SERVICE_TICKET_MANAGEMENT, MENU_BUSINESS_TICKET_MANAGEMENT } from '@/constants/menu-symbol';

export default defineComponent({
  props: {
    id: {
      type: String,
    },
    isBiz: {
      type: Boolean,
      required: true,
    },
  },

  setup(props) {
    const { t } = useI18n();
    const router = useRouter();
    const route = useRoute();

    const handleClick = () => {
      const name = props.isBiz ? MENU_BUSINESS_TICKET_MANAGEMENT : MENU_SERVICE_TICKET_MANAGEMENT;
      router.push({
        name,
        query: {
          ...route.query,
          type: 'resource_plan',
        },
      });
      // router.go(-1);
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
