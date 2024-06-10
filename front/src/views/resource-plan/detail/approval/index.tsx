import { defineComponent, type PropType } from 'vue';
import { Share } from 'bkui-vue/lib/icon';

import { useI18n } from 'vue-i18n';

import Panel from '@/components/panel';

import cssModule from './index.module.scss';

import type { TicketByIdResult } from '@/typings/resourcePlan';

export default defineComponent({
  props: {
    statusInfo: Object as PropType<TicketByIdResult['status_info']>,
  },

  setup(props) {
    const { t } = useI18n();

    const renderIcon = () => {
      switch (props.statusInfo?.status) {
        case 'todo':
        case 'auditing':
          return <i class='hcm-icon bkhcm-icon-jiazai'></i>;
        case 'rejected':
          return <i class='hcm-icon bkhcm-icon-38moxingshibai-01'></i>;
        case 'done':
          return <i class='hcm-icon bkhcm-icon-7chenggong-01'></i>;
        default:
          return <i class='hcm-icon bkhcm-icon-jiazai'></i>;
      }
    };

    return () => (
      <Panel>
        <section class={cssModule.home}>
          <span class={cssModule.status}>
            {renderIcon()}
            <span>{props.statusInfo?.status_name}</span>
          </span>
          <span class={cssModule.links}>
            <bk-link theme='primary' target='_blank' class={cssModule.link} href={props.statusInfo?.itsm_url}>
              {t('ITSM单据')}
              <Share />
            </bk-link>
            <bk-link theme='primary' target='_blank' class={cssModule.link} href={props.statusInfo?.crp_url}>
              {t('CRP单据')}
              <Share />
            </bk-link>
          </span>
        </section>
      </Panel>
    );
  },
});
