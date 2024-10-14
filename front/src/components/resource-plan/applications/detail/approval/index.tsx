import { defineComponent, type PropType } from 'vue';
import { Share } from 'bkui-vue/lib/icon';

import { useI18n } from 'vue-i18n';

import Panel from '@/components/panel';

import cssModule from './index.module.scss';

import type { TicketByIdResult } from '@/typings/resourcePlan';

export default defineComponent({
  props: {
    statusInfo: Object as PropType<TicketByIdResult['status_info']>,
    isBiz: {
      type: Boolean,
      required: true,
    },
    errorMessage: String,
  },
  setup(props) {
    const { t } = useI18n();

    const renderIcon = () => {
      switch (props.statusInfo?.status) {
        case 'auditing':
          return <i class='hcm-icon bkhcm-icon-jiazai'></i>;
        case 'rejected':
          return <i class='hcm-icon bkhcm-icon-38moxingshibai-01'></i>;
        case 'done':
          return <i class='hcm-icon bkhcm-icon-7chenggong-01'></i>;
        case 'failed':
          return <i class='hcm-icon bkhcm-icon-close-circle-fill'></i>;
        default:
          return <i class='hcm-icon bkhcm-icon-jiazai'></i>;
      }
    };

    return () => (
      <Panel class={cssModule.title}>
        <section class={cssModule.home}>
          <span class={cssModule.status}>
            {renderIcon()}
            <span>{props.statusInfo?.status_name}</span>
            {props.errorMessage && (
              <div class={cssModule['error-message']}>
                <i class={`hcm-icon bkhcm-icon-alert ${cssModule['error-message-color']}`} />
                <span>{props.errorMessage}</span>
              </div>
            )}
          </span>
          <span class={cssModule.links}>
            <bk-link
              theme='primary'
              target='_blank'
              class={cssModule.link}
              disabled={!props.statusInfo?.itsm_url}
              href={props.statusInfo?.itsm_url}>
              {t('ITSM单据')}
              <Share />
            </bk-link>
            <bk-link
              theme='primary'
              target='_blank'
              class={cssModule.link}
              disabled={!props.statusInfo?.crp_url}
              href={props.statusInfo?.crp_url}>
              {t('CRP单据')}
              <Share />
            </bk-link>
          </span>
        </section>
      </Panel>
    );
  },
});
