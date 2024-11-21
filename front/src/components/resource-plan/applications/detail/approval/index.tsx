import { defineComponent, VNode, type PropType } from 'vue';

import { useI18n } from 'vue-i18n';

import Panel from '@/components/panel';

import cssModule from './index.module.scss';

import type { IPlanTicketAudit, TicketByIdResult } from '@/typings/resourcePlan';

export default defineComponent({
  props: {
    statusInfo: Object as PropType<TicketByIdResult['status_info']>,
    isBiz: {
      type: Boolean,
      required: true,
    },
    errorMessage: String,
    ticketAuditDetail: Object as PropType<IPlanTicketAudit>,
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

    const renderAuditStatus = (): VNode => {
      const { itsm_audit, crp_audit } = props.ticketAuditDetail || {};
      if (itsm_audit?.status === 'auditing' || crp_audit?.status === 'auditing') {
        return (
          <span class={cssModule['audit-status']}>
            {t('当前处于')}
            <bk-tag theme='info' class='ml4 mr4'>
              {itsm_audit?.status === 'auditing' ? t('ITSM平台') : t('CRP平台')}
            </bk-tag>
            {t('审批')}
          </span>
        );
      }
      return null;
    };

    return () => (
      <Panel class={cssModule.title}>
        <section class={cssModule.home}>
          <span class={cssModule.status}>
            {renderIcon()}
            <span>{props.statusInfo?.status_name}</span>
            {renderAuditStatus()}
            {props.errorMessage && (
              <div class={cssModule['error-message']}>
                <i class={`hcm-icon bkhcm-icon-alert ${cssModule['error-message-color']}`} />
                <span>{props.errorMessage}</span>
              </div>
            )}
          </span>
        </section>
      </Panel>
    );
  },
});