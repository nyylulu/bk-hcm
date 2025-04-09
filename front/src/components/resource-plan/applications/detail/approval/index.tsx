import { defineComponent, VNode, type PropType } from 'vue';
import StatusUnknown from '@/assets/image/Status-unknown.png';

import { useI18n } from 'vue-i18n';

import Panel from '@/components/panel';
import ExpeditingBtn from '@/views/ziyanScr/components/ticket-audit/children/expediting-btn.vue';

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
          return <bk-loading style='transform: scale(0.5)' mode='spin' theme='primary' loading></bk-loading>;
        case 'rejected':
          return <i class='hcm-icon bkhcm-icon-38moxingshibai-01'></i>;
        case 'done':
          return <i class='hcm-icon bkhcm-icon-7chenggong-01'></i>;
        case 'failed':
          return <i class='hcm-icon bkhcm-icon-close-circle-fill'></i>;
        case 'revoked':
          return <img src={StatusUnknown} style={{ width: '22px', marginRight: '10px' }} />;
        default:
          return <i class='hcm-icon bkhcm-icon-jiazai'></i>;
      }
    };

    const renderAuditStatus = (): VNode => {
      const { itsm_audit, crp_audit } = props.ticketAuditDetail || {};
      if (!(itsm_audit?.status === 'auditing' || crp_audit?.status === 'auditing')) return null;

      if (itsm_audit?.status === 'auditing' || crp_audit?.status === 'auditing') {
        const { processors = [], processors_auth = {} } =
          itsm_audit?.current_steps?.[0] || crp_audit?.current_steps?.[0] || {};

        // 过滤无效审批人
        const displayProcessors = processors.filter((processor) => processor);

        const processorsWithBizAccess = displayProcessors.filter((processor) => {
          if (!props.isBiz) return processor; // 资源下不判断权限
          return processors_auth[processor];
        }); // 有权限的审批人
        const processorsWithoutBizAccess = displayProcessors.filter((processor) => !processors_auth[processor]); // 无权限的审批人

        const platform = itsm_audit?.status === 'auditing' ? 'ITSM' : 'CRP';
        const tagText = `${platform}${t('平台')}`;
        const copyText = `${t('复制')} ${platform} ${t('审批单')}`;
        const ticketLink = itsm_audit?.status === 'auditing' ? itsm_audit?.itsm_url : crp_audit?.crp_url;

        return (
          <div class='flex-row align-items-center'>
            <span class={cssModule['audit-status']}>
              {t('当前处于')}
              <bk-tag theme='success' class='ml4 mr4'>
                {tagText}
              </bk-tag>
              {t('审批环节')}
            </span>
            <ExpeditingBtn
              checkPermission={platform !== 'CRP'}
              processors={displayProcessors}
              processorsWithBizAccess={processorsWithBizAccess}
              processorsWithoutBizAccess={processorsWithoutBizAccess}
              copyText={copyText}
              ticketLink={ticketLink}
              defaultShow={crp_audit?.status === 'auditing'} // 处于CRP审批环境时，默认显示催单弹框
            />
          </div>
        );
      }
      return null;
    };

    return () => (
      <Panel class={cssModule.home}>
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
      </Panel>
    );
  },
});
