<script setup lang="ts">
import Panel from '@/components/panel/panel.vue';
import { computed, h, ref, useTemplateRef, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { Button, Message } from 'bkui-vue';
import { Spinner } from 'bkui-vue/lib/icon';
import { timeFormatter } from '@/common/util';
import { useTable } from '@/hooks/useResourcePlanTable';
import { useI18n } from 'vue-i18n';
import StatusLoading from '@/assets/image/status_loading.png';
import StatusSuccess from '@/assets/image/success-account.png';
import StatusFailure from '@/assets/image/failed-account.png';
import ResultDefault from '@/assets/image/result-default.svg';
import CopyToClipboard from '@/components/copy-to-clipboard/index.vue';
import { IPageQuery } from '@/typings';
import Stage from './stage.vue';
import SubTicketDetail from './sub-ticket-detail.vue';
import { useResSubTicketStore, SubTicketItem, STATUS_ENUM, STAGE_ENUM } from '@/store/ticket/res-sub-ticket';
import { GLOBAL_BIZS_KEY } from '@/common/constant';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const store = useResSubTicketStore();
const detailRef = useTemplateRef('detailRef');
const bizId = computed(() => Number(route.query[GLOBAL_BIZS_KEY]));

// 表格
const hoverIndex = ref(-1);
const columns: any[] = [
  {
    label: '子单号',
    field: 'id',
    render: ({ data }: { data: SubTicketItem }) => {
      return h(
        Button,
        {
          text: true,
          theme: 'primary',
          onClick: () => {
            router.replace({ query: { ...route.query, subId: data.id } });
            detailRef.value.open(data);
          },
        },
        data.id,
      );
    },
  },
  {
    label: '子单状态',
    field: 'status',
    render({ cell }: { cell: string }) {
      let icon = ResultDefault;
      const txt = STATUS_ENUM[cell] || '---';
      switch (cell) {
        case 'init':
        case 'auditing':
          icon = StatusLoading;
          break;
        case 'done':
          icon = StatusSuccess;
          break;
        case 'rejected':
        case 'failed':
          icon = StatusFailure;
          break;
        case 'invalid':
          icon = ResultDefault;
          break;
        default:
          icon = ResultDefault;
      }
      return h('div', { style: { display: 'flex', alignItems: 'center' } }, [
        icon === StatusLoading
          ? h(Spinner, { fill: '#3A84FF', class: 'mr6', width: 14, height: 14 })
          : h('img', { src: icon, class: 'mr6', width: 14, height: 14 }),
        txt,
      ]);
    },
  },
  {
    label: '审批步骤',
    field: 'stage',
    render({ cell, data, index }: { cell: string; index: number; data: SubTicketItem }) {
      if (cell === 'crp_audit') {
        return h(Stage, {
          showExpeditingBtn: data.status === 'auditing' || data.status === 'init',
          text: STAGE_ENUM[cell],
          ticketData: data,
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
    render: ({ data }: { cell: number; data: SubTicketItem }) => {
      const updatedCore = Number(data.updated_info.cvm.cpu_core) || 0;
      const originalCore = Number(data.original_info.cvm.cpu_core) || 0;
      const compare = updatedCore - originalCore;
      let value = 0;
      let prefix = '';
      let color = '';

      switch (data.sub_ticket_type) {
        case 'add':
        case 'transfer':
          prefix = '+';
          color = '#00B545'; // 绿色
          value = originalCore;
          break;
        case 'cancel':
          prefix = '-';
          color = '#EA3636'; // 红色
          value = originalCore;
          break;
        case 'adjust':
          prefix = compare >= 0 ? '+' : '-';
          color = '#EA3636'; // 红色
          value = Math.abs(compare);
          break;
      }

      return h('span', { style: { color } }, isNaN(value) ? '--' : `${prefix}${value}`);
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
  return store.getList(
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

// 数据
const ticketLinkArr = computed(() => {
  return tableData.value.reduce((acc, cur) => {
    if (cur.status === 'auditing') acc.push(`${cur.crp_url}\n`);
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
const handelFailedTicket = () => {
  store
    .retryTickets(route.query?.id as string, bizId.value)
    .then(() => {
      Message({ theme: 'success', message: '重试成功' });
    })
    .catch(() => {
      Message({ theme: 'error', message: `重试失败` });
    })
    .finally(() => {
      triggerApi();
    });
};

const handleMouseEnter = (e: any, row: any, index: number) => {
  hoverIndex.value = index;
};
const handleMouseLeave = () => {
  hoverIndex.value = -1;
};

const handleSubTicketShowById = async (id: string) => {
  const res = await getData({ limit: 500 });
  const subTicketItem = res.data.details.find((item) => item.id === id);
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
      <bk-button :disabled="!hasFailedTicket" style="margin-left: 21px" @click="handelFailedTicket">
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
