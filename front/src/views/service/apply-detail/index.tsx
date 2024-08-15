import { defineComponent, onMounted, onUnmounted, ref, watch } from 'vue';
import './index.scss';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import ApplyDetail from '@/views/service/my-apply/components/apply-detail/index.vue';
import { useAccountStore } from '@/store';
import { useRoute } from 'vue-router';
import { ACCOUNT_TYPES, APPLICATION_TYPE_MAP, COMMON_TYPES } from '../apply-list/constants';
import AccountApplyDetail, { IDetail } from './account-apply-detail';
import BpassApplyDetail, { BpaasEndStatus } from '../my-apply/components/bpass-apply-detail';
import useFormModel from '@/hooks/useFormModel';

export const ApplicationEndStatus = ['rejected', 'pass', 'canceled', 'completed', 'deliver_error'];

export default defineComponent({
  setup() {
    const accountStore = useAccountStore();
    const isLoading = ref(false);
    const currentApplyData = ref<IDetail & { BpaasName?: string }>({});
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
        if (ApplicationEndStatus.includes(currentApplyData.value.status)) clearInterval(interval);
        if (isBpaas.value) {
          setBpassPayload({
            applicant: res.data.applicant,
            account_id: JSON.parse(res.data.content).account_id,
            id: res.data.id,
            bpaas_sn: +res.data.sn,
          });
          const bpaasRes = await accountStore.getBpassDetail(bpassPayload);
          currentApplyData.value = bpaasRes.data;
          if (BpaasEndStatus.includes(bpaasRes.data.Status)) clearInterval(interval);
        }
        curApplyKey.value = res.data.id;
      } finally {
        isLoading.value = false;
      }
    };

    onMounted(() => {
      interval = setInterval(() => getMyApplyDetail(route.query.id as string), 5000);
    });

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

    return () => (
      <div class={'apply-detail-container'}>
        <DetailHeader>
          <span class={'title'}>申请单详情</span>
          <span class={'sub-title'}>
            &nbsp;-&nbsp;
            {APPLICATION_TYPE_MAP[currentApplyData.value.type as keyof typeof APPLICATION_TYPE_MAP] ||
              currentApplyData.value.BpaasName}
          </span>
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
            />
          </div>
        )}
      </div>
    );
  },
});
