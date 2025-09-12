<template>
  <bk-dialog v-model:is-show="model" title="屏蔽退还通知" width="640">
    <ul class="info">
      <li v-for="field in fields" :key="field.id" class="info-item">
        <span class="label">{{ field.name }}</span>
        <display-value
          :property="field"
          :value="details[field.id]"
          :display="field.meta?.display"
          v-bind="getDisplayCompProps(field)"
        />
      </li>
    </ul>
    <bk-alert>
      <template #title>
        注意：
        <div class="tips">
          1.提交屏蔽退还通知后，当前单据将不会发送退还提醒的通知。屏蔽一旦提交，不支持取消。
          <br />
          2.在机器申请后第31-121天内未退还，按超期机器的50%成本核算加收罚金，请参考
          <a href="https://iwiki.woa.com/p/4012608772" target="_blank" class="text-link">
            https://iwiki.woa.com/p/4012608772
          </a>
        </div>
      </template>
    </bk-alert>
    <template #footer>
      <div class="footer">
        <bk-checkbox v-model="isAgree" class="agree-checkbox">确认需要屏蔽该单据的退还通知</bk-checkbox>
        <modal-footer
          confirm-text="确认屏蔽"
          :disabled="!isAgree"
          :loading="rollingServerUsageStore.updateAppliedRecordsNoticeDisabledLoading"
          @confirm="handleConfirm"
          @closed="handleClosed"
        />
      </div>
    </template>
  </bk-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import usageOrderViewProperties from '@/model/rolling-server/usage-order.view';
import { RollingServerRecordItem, useRollingServerUsageStore } from '@/store';
import { ModelPropertyDisplay } from '@/model/typings';

import { Message } from 'bkui-vue';
import ModalFooter from '@/components/modal/modal-footer.vue';

const model = defineModel<boolean>();
const props = defineProps<{ details: RollingServerRecordItem }>();
const emit = defineEmits<{
  success: [];
}>();

const rollingServerUsageStore = useRollingServerUsageStore();

const fieldIds = ['suborder_id', 'roll_date', 'applied_core', 'delivered_core', 'returned_core', 'not_returned_core'];
const fields: ModelPropertyDisplay[] = fieldIds.map((id) => usageOrderViewProperties.find((view) => view.id === id));
const getDisplayCompProps = (field: ModelPropertyDisplay) => {
  const { id } = field;
  if (id === 'roll_date') {
    return { format: 'YYYY-MM-DD' };
  }
  return {};
};

const isAgree = ref(false);
const handleConfirm = async () => {
  await rollingServerUsageStore.updateAppliedRecordsNoticeDisabled([props.details.id]);
  Message({ theme: 'success', message: '提交成功' });
  handleClosed();
  emit('success');
};
const handleClosed = () => {
  model.value = false;
};
</script>

<style scoped lang="scss">
.info {
  margin-bottom: 12px;

  .info-item {
    display: flex;
    align-items: center;
    height: 32px;

    .label {
      width: 120px;
      text-align: right;
      margin-right: 10px;

      &::after {
        content: ':';
        margin: 0 10px 0 5px;
      }
    }
  }
}

.tips {
  word-break: break-all;
}

.footer {
  display: flex;
  align-items: center;

  .agree-checkbox {
    margin-right: auto;
  }
}
</style>
