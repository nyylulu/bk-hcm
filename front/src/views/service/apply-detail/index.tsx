import { computed, defineComponent, onUnmounted, ref, watch } from 'vue';
import './index.scss';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import ApplyDetail from '@/views/service/my-apply/components/apply-detail/index.vue';
import { useAccountStore } from '@/store';
import { useRoute } from 'vue-router';
import { ACCOUNT_TYPES, APPLICATION_TYPE_MAP, COMMON_TYPES } from '../apply-list/constants';
import AccountApplyDetail from './account-apply-detail';
import BpassApplyDetail, { BpaasEndStatus } from '../my-apply/components/bpass-apply-detail';
import Clb from './clb.vue';
import useFormModel from '@/hooks/useFormModel';
import routerAction from '@/router/utils/action';
import { GLOBAL_BIZS_KEY, VendorEnum } from '@/common/constant';

export enum ApplicationStatus {
  pending = 'pending',
  pass = 'pass',
  rejected = 'rejected',
  cancelled = 'cancelled',
  delivering = 'delivering',
  completed = 'completed',
  deliver_partial = 'deliver_partial',
  deliver_error = 'deliver_error',
}

export interface IApplicationDetail {
  id: string;
  source: string;
  sn: string;
  type: string;
  status: ApplicationStatus;
  applicant: string;
  content: string;
  delivery_detail: string;
  memo: string;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
  ticket_url: string;
  [key: string]: any;
}

export default defineComponent({
  setup() {
    const accountStore = useAccountStore();
    const isLoading = ref(false);
    const currentApplyData = ref<IApplicationDetail & { BpaasName?: string }>({});
    const curApplyKey = ref('');
    const isCancelBtnLoading = ref(false);
    const route = useRoute();
    const isBpaas = ref(false);

    const { formModel: bpassPayload, setFormValues: setBpassPayload } = useFormModel({
      bpaas_sn: 0,
      account_id: '',
      id: '',
      applicant: '',
    });

    let interval: NodeJS.Timeout;

    // 获取单据详情
    const getMyApplyDetail = async (id: string) => {
      isLoading.value = true;
      try {
        const res = await accountStore.getApplyAccountDetail(id);
        currentApplyData.value = res.data;
        if (isBpaas.value) {
          setBpassPayload({
            applicant: res.data.applicant,
            account_id: JSON.parse(res.data.content).account_id,
            id: res.data.id,
            bpaas_sn: +res.data.sn,
          });
          const bpaasRes = await accountStore.getBpassDetail(bpassPayload);
          currentApplyData.value = { ...currentApplyData.value, ...bpaasRes.data };
          if (BpaasEndStatus.includes(bpaasRes.data.Status)) clearInterval(interval);
        }
        curApplyKey.value = res.data.id;

        if ([ApplicationStatus.pending, ApplicationStatus.delivering].includes(res.data.status)) {
          clearInterval(interval);
          interval = setInterval(() => getMyApplyDetail(route.query.id as string), 5000);
        } else {
          clearInterval(interval);
        }
      } finally {
        isLoading.value = false;
      }
    };

    onUnmounted(() => {
      clearInterval(interval);
    });

    // 撤销单据
    const handleCancel = async (id: string) => {
      isCancelBtnLoading.value = true;
      try {
        await accountStore.cancelApplyAccount(id);
        getMyApplyDetail(id);
      } finally {
        isCancelBtnLoading.value = false;
      }
    };

    watch(
      [() => route.query.id, () => route.query.source],
      ([id, source]) => {
        isBpaas.value = source === 'bpaas';
        if (id) {
          getMyApplyDetail(id as string);
        }
      },
      {
        immediate: true,
      },
    );

    const isGotoSecurityRuleShow = computed(() =>
      [
        'create_security_group',
        'create_security_group_rule',
        'update_security_group_rule',
        'delete_security_group_rule',
      ].includes(currentApplyData.value?.type),
    );
    const isGotoSecurityRuleDisabled = computed(() => !JSON.parse(currentApplyData.value?.content)?.sg_id);
    const gotoSecurityRule = () => {
      const securityInfo = JSON.parse(currentApplyData.value.content);
      routerAction.open({
        path: '/business/security/detail',
        query: {
          [GLOBAL_BIZS_KEY]: accountStore.bizs,
          id: securityInfo.sg_id,
          vendor: securityInfo.vendor ?? VendorEnum.ZIYAN,
          scene: 'rule',
        },
      });
    };

    const render = () => {
      let vNode = (
        <div class={'apply-detail-container'}>
          <DetailHeader>
            {{
              default: () => (
                <>
                  <span class={'title'}>申请单详情</span>
                  <span class={'sub-title'}>
                    &nbsp;-&nbsp;
                    {APPLICATION_TYPE_MAP[currentApplyData.value.type as keyof typeof APPLICATION_TYPE_MAP] ||
                      currentApplyData.value.BpaasName}
                  </span>
                </>
              ),
              right: () =>
                isGotoSecurityRuleShow.value ? (
                  <bk-button theme='primary' disabled={isGotoSecurityRuleDisabled.value} onClick={gotoSecurityRule}>
                    跳转至安全组规则
                  </bk-button>
                ) : undefined,
            }}
          </DetailHeader>
          {!isBpaas.value ? (
            <div class={'apply-content-wrapper'}>
              {ACCOUNT_TYPES.includes(currentApplyData.value.type) && (
                <AccountApplyDetail detail={currentApplyData.value} />
              )}

              {COMMON_TYPES.includes(currentApplyData.value.type) && (
                <ApplyDetail params={currentApplyData.value} key={curApplyKey.value} onCancel={handleCancel} />
              )}
            </div>
          ) : (
            <div class={'apply-content-wrapper'}>
              <BpassApplyDetail
                loading={isLoading.value}
                params={currentApplyData.value}
                key={curApplyKey.value}
                getBpaasDetail={getMyApplyDetail}
                bpaasPayload={bpassPayload}
                isGotoSecurityRuleShow={isGotoSecurityRuleShow.value}
              />
            </div>
          )}
        </div>
      );
      // 负载均衡详情
      if (route.query.type?.includes('load_balancer')) {
        vNode = <Clb applicationDetail={currentApplyData.value} loading={isLoading.value} />;
      }
      return vNode;
    };

    return render;
  },
});
