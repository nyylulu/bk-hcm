<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import Panel from '@/components/panel';
import { timeFormatter } from '@/common/util';
import { TicketBaseInfo } from '@/typings/resourcePlan';
// import CopyToClipboard from '@/components/copy-to-clipboard/index.vue';
import EllipseFoldText from '@/components/ellipse-fold-text.vue';

interface ListItem {
  label?: string;
  value?: string;
  fullCol?: boolean;
  slot?: string;
}
interface Props {
  baseInfo?: TicketBaseInfo;
  dataList?: ListItem[];
}
const props = defineProps<Props>();
const baseList = computed(() => [
  {
    label: '需求类型',
    value: props.baseInfo?.type_name,
  },
  {
    label: '业务名称',
    value: props.baseInfo?.bk_biz_name,
  },
  {
    label: '部门',
    value: props.baseInfo?.virtual_dept_name,
  },
  {
    label: '运营产品',
    value: props.baseInfo?.op_product_name,
  },
  {
    label: '规划产品',
    value: props.baseInfo?.plan_product_name,
  },
  {
    label: '提单人',
    value: props.baseInfo?.applicant,
  },
  {
    label: '提单时间',
    value: timeFormatter(props.baseInfo?.submitted_at, 'YYYY-MM-DD'),
  },
  {
    label: t('预测说明'),
    value: props.baseInfo?.remark,
    fullCol: true,
  },
]);

const { t } = useI18n();
const displayList = computed(() => {
  if (props.baseInfo) {
    return baseList.value;
  }
  return props.dataList || [];
});
</script>

<template>
  <Panel class="panel" :title="t('基本信息')">
    <ul class="home">
      <li v-for="(item, index) in displayList" :class="{ 'full-col': item.fullCol }" :key="index">
        <span class="label">{{ t(item.label) }}：</span>
        <!-- <span class="value" :title="item.value">{{ item.value || '--' }}</span> -->
        <ellipse-fold-text :text="item.value || '--'"></ellipse-fold-text>
      </li>
    </ul>
  </Panel>
</template>

<style lang="scss" scoped>
.home {
  display: grid;
  grid-template-columns: 500px 500px;
  padding: 0 56px;

  li {
    line-height: 40px;
    display: flex;

    &.full-col {
      grid-column: 1 / -1;
      max-width: 80%;
    }
  }

  .label {
    width: 70px;
    flex-shrink: 0;
    text-align: right;
  }

  .value {
    color: #313238;
    display: inline-block;
    max-width: 80%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  :global(.span-2) {
    grid-column-start: span 2;
  }
}

.col {
  display: flex;
  line-height: 40px;
  padding: 0 56px 8px;
}
</style>
