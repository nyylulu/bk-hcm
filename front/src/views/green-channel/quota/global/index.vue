<script setup lang="ts">
import { reactive, ref } from 'vue';
import { storeToRefs } from 'pinia';
import { Message } from 'bkui-vue';
import { Warn } from 'bkui-vue/lib/icon';
import { useGreenChannelQuotaStore, type IGlobalQuota } from '@/store/green-channel/quota';

const greenChannelQuotaStore = useGreenChannelQuotaStore();
const { globalQuotaConfig } = storeToRefs(greenChannelQuotaStore);

const formatValue = (value: number) => {
  if (isNaN(value) || value === undefined) {
    return '--';
  }
  return value.toLocaleString();
};

const editDialog = reactive<{ show: boolean; dataKey: keyof IGlobalQuota }>({
  show: false,
  dataKey: undefined,
});

const fieldNames = {
  ieg_quota: '总限额 (全平台)',
  biz_quota: '基础额度 (单业务/周)',
  audit_threshold: '需要人工审批核数/单',
};

const formRef = ref(null);

const formData = reactive<IGlobalQuota>({
  ieg_quota: 0,
  biz_quota: 0,
  audit_threshold: 0,
});

const closeDialog = () => {
  editDialog.show = false;
};

const handleEdit = (field: keyof IGlobalQuota) => {
  editDialog.show = true;
  editDialog.dataKey = field;
  formData[field] = globalQuotaConfig.value[field];
};

const handleEditConfirm = async () => {
  await formRef.value?.validate();
  await greenChannelQuotaStore.updateQuotaConfig({ [editDialog.dataKey]: formData[editDialog.dataKey] });
  Message({ theme: 'success', message: '修改成功' });
  closeDialog();
};
</script>

<template>
  <bk-alert closable class="info-alert">
    <template #title>全平台额度，是按自然月统计的当月数据。下月数据会自动计算。</template>
  </bk-alert>
  <div class="info-grid">
    <div class="row">
      <div class="cell head">{{ fieldNames.ieg_quota }}</div>
      <div class="cell">
        <span>{{ formatValue(globalQuotaConfig.ieg_quota) }} 核</span>
        <bk-button theme="primary" text @click="handleEdit('ieg_quota')">
          <i class="icon hcm-icon bkhcm-icon-bianji edit-icon"></i>
        </bk-button>
      </div>
    </div>
    <div class="row">
      <div class="cell head">{{ fieldNames.biz_quota }}</div>
      <div class="cell">
        <span>{{ formatValue(globalQuotaConfig.biz_quota) }} 核</span>
        <bk-button theme="primary" text @click="handleEdit('biz_quota')">
          <i class="icon hcm-icon bkhcm-icon-bianji edit-icon"></i>
        </bk-button>
      </div>
    </div>
    <div class="row">
      <div class="cell head">{{ fieldNames.audit_threshold }}</div>
      <div class="cell">
        <span>{{ formatValue(globalQuotaConfig.audit_threshold) }} 核</span>
        <bk-button theme="primary" text @click="handleEdit('audit_threshold')">
          <i class="icon hcm-icon bkhcm-icon-bianji edit-icon"></i>
        </bk-button>
      </div>
    </div>
    <div class="row">
      <div class="cell head">已交付 (全业务)</div>
      <div class="cell">{{ formatValue(globalQuotaConfig.sum_delivered_core) }} 核</div>
    </div>
    <div class="row">
      <div class="cell head">剩余额度 (全平台)</div>
      <div class="cell">{{ formatValue(globalQuotaConfig.ieg_quota - globalQuotaConfig.sum_delivered_core) }} 核</div>
    </div>
  </div>

  <bk-dialog
    :title="'额度管理配置'"
    :quick-close="false"
    :is-show="editDialog.show"
    :is-loading="greenChannelQuotaStore.updateQuotaConfigLoading"
    :confirm-text="'提交'"
    @confirm="handleEditConfirm"
    @closed="closeDialog"
  >
    <bk-form form-type="vertical" :model="formData" ref="formRef">
      <bk-form-item
        :label="fieldNames.ieg_quota"
        :required="true"
        property="ieg_quota"
        v-if="editDialog.dataKey === 'ieg_quota'"
      >
        <hcm-form-number suffix="核" v-model="formData.ieg_quota" :min="1" :max="999999" />
      </bk-form-item>
      <bk-form-item
        :label="fieldNames.biz_quota"
        :required="true"
        property="biz_quota"
        v-if="editDialog.dataKey === 'biz_quota'"
      >
        <hcm-form-number suffix="核" v-model="formData.biz_quota" :min="0" :max="999999" />
      </bk-form-item>
      <bk-form-item
        :label="fieldNames.audit_threshold"
        :required="true"
        property="biz_quota"
        v-if="editDialog.dataKey === 'audit_threshold'"
      >
        <hcm-form-number prefix="大于" suffix="核" v-model="formData.audit_threshold" :min="1" :max="999999" />
      </bk-form-item>
    </bk-form>
    <div class="form-tips">
      <Warn class="icon" />
      将调整{{ fieldNames[editDialog.dataKey] }}，请确认后提交
    </div>
  </bk-dialog>
</template>

<style lang="scss" scoped>
.info-alert {
  margin-bottom: 20px;
}
.info-grid {
  display: grid;
  grid-template-columns: 1fr;
  grid-gap: 0;
  .row {
    display: grid;
    grid-template-columns: 260px 1fr;
    gap: 0;
    .cell {
      display: flex;
      align-items: center;
      gap: 6px;
      font-size: 12px;
      color: #313238;
      padding: 12px 16px;
      border: 1px solid #dcdee5;
      height: 44px;
      overflow: hidden;
      margin-left: -1px;
      margin-top: -1px;
      &.head {
        background: #fafbfd;
      }
      .edit-icon {
        font-size: 14px;
      }
    }
  }
}
.form-tips {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  .icon {
    font-size: 14px;
    color: $danger-color;
  }
}
</style>
