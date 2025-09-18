<script setup lang="ts">
import Notice from '@/assets/image/notice.svg';
import Copy from '@/assets/image/copy.svg';
import CopyToClipboard from '@/components/copy-to-clipboard/index.vue';
import { computed, onBeforeMount, ref } from 'vue';
import ExpeditingBtn from '@/views/ziyanScr/components/ticket-audit/children/expediting-btn.vue';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { useResSubTicketStore, SubTicketItem, SubTicketAudit } from '@/store/ticket/res-sub-ticket';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import { useRoute } from 'vue-router';

interface Props {
  text?: string;
  showActions: boolean;
  ticketData: SubTicketItem;
  showExpeditingBtn: boolean;
}
const props = withDefaults(defineProps<Props>(), {
  text: '',
  showActions: false,
});
const store = useResSubTicketStore();
const { isBusinessPage } = useWhereAmI();
const route = useRoute();
const bizId = computed(() => Number(route.query[GLOBAL_BIZS_KEY]));

// 请求查询审批流接口
const auidtData = ref<SubTicketAudit>();
const getAduitData = async () => {
  const { data } = await store.getAudit(props.ticketData.id, bizId.value);
  auidtData.value = data;
};

const expeditingBtnProps = computed(() => {
  if (!auidtData.value) return {};
  const { crp_audit } = auidtData.value;
  const { processors = [], processors_auth = {} } = crp_audit?.current_steps?.[0] || {};

  // 过滤无效审批人
  const displayProcessors = processors.filter((processor) => processor);

  const processorsWithBizAccess = displayProcessors.filter((processor) => {
    if (!isBusinessPage) return processor; // 资源下不判断权限
    return processors_auth[processor];
  }); // 有权限的审批人
  const processorsWithoutBizAccess = displayProcessors.filter((processor) => !processors_auth[processor]); // 无权限的审批人

  const copyText = `复制CRP审批单`;
  const ticketLink = crp_audit?.crp_url;

  return {
    checkPermission: false,
    processors: displayProcessors,
    processorsWithBizAccess,
    processorsWithoutBizAccess,
    copyText,
    ticketLink,
    defaultShow: false,
  };
});

const alwaysShowActionBtn = ref(false);
const handleShow = () => {
  alwaysShowActionBtn.value = true;
};
const handleHidden = () => {
  alwaysShowActionBtn.value = false;
};

onBeforeMount(() => {
  getAduitData();
});
</script>

<template>
  <div class="stage">
    <span>{{ text }}</span>
    <div class="stage" v-show="showActions || alwaysShowActionBtn">
      <CopyToClipboard :content="ticketData.crp_url" success-msg="复制成功">
        <img v-bk-tooltips="{ content: '复制CRP审批链接' }" :src="Copy" width="13" height="13" />
      </CopyToClipboard>
      <ExpeditingBtn v-if="showExpeditingBtn" v-bind="expeditingBtnProps" @show="handleShow" @hidden="handleHidden">
        <img v-bk-tooltips="{ content: '催单' }" :src="Notice" width="13" height="13" />
      </ExpeditingBtn>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.stage {
  display: flex;
  align-items: center;

  img {
    margin-left: 13px;
    cursor: pointer;
  }
}
</style>
