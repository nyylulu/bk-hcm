import { defineComponent } from 'vue';
import { useI18n } from 'vue-i18n';
import cssModule from './index.module.scss';
import { useRouter, useRoute } from 'vue-router';

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
      const path = props.isBiz ? '/business/applications' : '/service/my-apply';
      router.push({
        path,
        query: {
          ...route.query,
          type: 'resource_plan',
        },
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
