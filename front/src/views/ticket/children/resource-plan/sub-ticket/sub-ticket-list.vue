<script setup lang="ts">
import Panel from '@/components/panel/panel.vue';
import { computed, h, ref, useTemplateRef, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { Button, Message } from 'bkui-vue';
import { timeFormatter } from '@/common/util';
import { useTable } from '@/hooks/useResourcePlanTable';
import { useI18n } from 'vue-i18n';
import CopyToClipboard from '@/components/copy-to-clipboard/index.vue';
import { IPageQuery } from '@/typings';
import Stage from './components/stage.vue';
import SubTicketDetail from './sub-ticket-detail.vue';
import { useResSubTicketStore, SubTicketItem, STATUS_ENUM, STAGE_ENUM } from '@/store/ticket/res-sub-ticket';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import { debounce } from 'lodash';
import StatusText from './components/status-text.vue';

// 补全类型泛型
const emits = defineEmits<{
  retryTicket: [];
}>();
const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const subTicketStore = useResSubTicketStore();
const detailRef = useTemplateRef('detailRef');
const bizId = computed(() => Number(route.query[GLOBAL_BIZS_KEY]));

// 表格
const hoverIndex = ref(-1);
const columns: any[] = [
  {
    label: '子单号',
    field: 'id',
    render: ({ row }: { row?: SubTicketItem }) => {
      return h(
        Button,
        {
          text: true,
          theme: 'primary',
          onClick: () => {
            router.replace({ query: { ...route.query, subId: row.id } });
          },
        },
        row.id,
      );
    },
  },
  {
    label: '子单状态',
    field: 'status',
    render({ cell, row }: { cell?: string; row?: SubTicketItem }) {
      const ICON_TYPE: Record<string, any> = {
        init: 'loading',
        auditing: 'loading',
        done: 'success',
        rejected: 'failed',
        failed: 'failed',
        invalid: 'default',
      };
      const txt = STATUS_ENUM[cell] || '---';
      return h(StatusText, { text: txt, type: ICON_TYPE[cell], errorMessage: row.message });
    },
  },
  {
    label: '审批步骤',
    field: 'stage',
    render({ cell, row, index }: { cell?: string; index?: number; row?: SubTicketItem }) {
      if (cell === 'crp_audit') {
        return h(Stage, {
          showExpeditingBtn: row.status === 'auditing' || row.status === 'init',
          text: STAGE_ENUM[cell],
          ticketData: row,
          showActions: hoverIndex.value === index,
        });
      }
      return STAGE_ENUM[cell];
    },
  },
  {
    label: '单据类型',
    field: 'ticket_type_name',
  },
  {
    label: 'CPU核数',
    field: 'updated_info.cvm.cpu_core',
    isDefaultShow: true,
    render: ({ row }: { row?: SubTicketItem }) => {
      const type = row.sub_ticket_type;
      const value = row.updated_info.cvm.cpu_core - row.original_info.cvm.cpu_core;
      let color = '';

      switch (type) {
        case 'add':
          color = '#299e56'; // 绿色
          break;
        case 'cancel':
        case 'adjust':
          color = '#ea3636'; // 红色
          break;
        case 'transfer':
          color = value >= 0 ? '#299e56' : '#ea3636'; // 绿色
          break;
        default:
          color = '#ea3636'; // 红色
      }
      if (isNaN(value)) {
        return '--';
      }
      let prefix = value > 0 ? '+' : '';
      if (value === 0) {
        prefix = type === 'cancel' ? '-' : '+';
      }
      return h('span', { style: { color } }, `${prefix}${value}`);
    },
  },
  {
    label: '单据生成时间',
    field: 'created_at',
    render({ cell }: any) {
      return timeFormatter(cell);
    },
  },
  {
    label: '单据完成时间',
    field: 'updated_at',
    render({ cell }: any) {
      return timeFormatter(cell);
    },
  },
];
const getData = (page: IPageQuery) => {
  return subTicketStore.getList(
    {
      page,
      ticket_id: route.query?.id as string,
    },
    bizId.value,
  );
};
const { tableData, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort, triggerApi } =
  useTable(getData);
pagination.value.limit = 500; // 不分页，设置limit为最大值
const retryBtnLoading = ref(false);

// 数据
const ticketLinkArr = computed(() => {
  return tableData.value.reduce((acc, cur) => {
    if (cur.status === 'auditing' && cur.crp_url) acc.push(cur.crp_url);
    return acc;
  }, []);
});
const successMsg = computed(() => {
  return `复制${ticketLinkArr.value.length}条CRP待审批链接`;
});
// 是否有failed的单据
const hasFailedTicket = computed(() => {
  return tableData.value.some((item) => item.status === 'failed');
});
// 方法
const handleFailedTicket = () => {
  retryBtnLoading.value = true;
  subTicketStore
    .retryTickets(route.query?.id as string, bizId.value)
    .then(() => {
      Message({ theme: 'success', message: '重试成功' });
    })
    .catch(() => {
      Message({ theme: 'error', message: `重试失败` });
    })
    .finally(() => {
      retryBtnLoading.value = false;
      emits('retryTicket');
    });
};
// 防抖 handleFailedTicket
const handleFailedTicketDebounce = debounce(handleFailedTicket, 500);

const handleMouseEnter = (e: any, row: any, index: number) => {
  hoverIndex.value = index;
};
const handleMouseLeave = () => {
  hoverIndex.value = -1;
};

const handleSubTicketShowById = async (id: string) => {
  const list = tableData.value?.length ? tableData.value : (await getData({ limit: 500 })).data.details;
  const subTicketItem = list.find((item) => item.id === id);
  if (subTicketItem) {
    detailRef.value.open(subTicketItem);
  }
};
// 监听 route.query.subId
watch(
  () => route.query.subId,
  (val: string | string[]) => {
    if (val) handleSubTicketShowById(val as string);
  },
  { immediate: true },
);

defineExpose({
  getData: triggerApi,
});
</script>

<template>
  <Panel class="panel" :title="t('子单信息')">
    <template #title-extra>
      <bk-button
        :disabled="!hasFailedTicket"
        style="margin-left: 21px"
        :loading="retryBtnLoading"
        @click="handleFailedTicketDebounce"
      >
        失败单据处理
      </bk-button>
      <copy-to-clipboard :content="ticketLinkArr.join('\n')" :success-msg="successMsg">
        <bk-button
          :disabled="!ticketLinkArr.length"
          v-bk-tooltips="{ content: t('CRP单据已全部审批完成'), disabled: ticketLinkArr.length }"
          style="margin-left: 12px"
        >
          复制CRP待审批链接
        </bk-button>
      </copy-to-clipboard>
    </template>
    <bk-loading :loading="isLoading">
      <bk-table
        :columns="columns"
        :pagination="null"
        :data="tableData"
        remote-pagination
        @page-limit-change="handlePageSizeChange"
        @page-value-change="handlePageChange"
        @column-sort="handleSort"
        @row-mouse-enter="handleMouseEnter"
        @row-mouse-leave="handleMouseLeave"
      />
    </bk-loading>
  </Panel>
  <SubTicketDetail ref="detailRef" />
</template>

<style lang="scss" scoped>
.panel {
  box-shadow: none;
  padding-bottom: 25px;
}

.cvm-status-container {
  display: flex;
  align-items: center;
}
</style>
