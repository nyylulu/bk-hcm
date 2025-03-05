<script setup lang="ts">
import { computed, ref, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';

import CopyToClipboard from '@/components/copy-to-clipboard/index.vue';

const props = withDefaults(
  defineProps<{
    checkPermission?: boolean; // TODO：临时字段，目前CRP的人不能到HCM审批，不应该引导用户给CRP的人添加权限
    processors: string[];
    processorsWithBizAccess: string[];
    processorsWithoutBizAccess: string[];
    copyText: string;
    ticketLink: string;
    defaultShow?: boolean;
  }>(),
  {
    // 兼容空数据，防止页面崩溃
    checkPermission: true,
    processors: () => [],
    processorsWithBizAccess: () => [],
    processorsWithoutBizAccess: () => [],
    copyText: '',
    ticketLink: '',
    defaultShow: false,
  },
);

const { t } = useI18n();
const isShow = ref(false);
const hasNoBizAccess = computed(() => props.processors.length !== props.processorsWithBizAccess.length);

watchEffect(() => {
  isShow.value = hasNoBizAccess.value || props.defaultShow;
});
</script>

<template>
  <bk-popover
    :is-show="isShow"
    trigger="click"
    theme="light"
    placement="bottom-start"
    :offset="8"
    :allow-html="true"
    @after-show="isShow = true"
    @after-hidden="isShow = false"
  >
    <template #default>
      <div class="expediting-btn-wrap" :class="{ active: isShow }">
        <bk-button theme="primary" text>
          <i class="hcm-icon bkhcm-icon-notice" style="margin-right: 0; font-size: 16px"></i>
          {{ t('催单') }}
        </bk-button>
      </div>
    </template>
    <template #content>
      <div class="expediting-content">
        <!-- case1：有审批人 -->
        <template v-if="processors.length">
          <div class="mb4">
            {{ t('当前审批人为') }}
            <!-- 都有权限 -->
            <!-- TODO：或无需引导权限申请 -->
            <template v-if="!hasNoBizAccess || !checkPermission">
              <display-value
                v-for="processor in processors"
                class="mr4"
                :key="processor"
                :value="processor"
                :property="{ type: 'user' }"
                :display="{ appearance: 'wxwork-link' }"
              />
            </template>
            <!-- 存在无权限审批人 -->
            <template v-else>
              <template v-if="processorsWithBizAccess.length">
                <display-value
                  v-for="processor in processorsWithBizAccess"
                  class="mr4"
                  :key="processor"
                  :value="processor"
                  :property="{ type: 'user' }"
                  :display="{ appearance: 'wxwork-link' }"
                />
                {{ t('有权限，') }}
              </template>
              <template v-if="processorsWithoutBizAccess.length">
                <display-value
                  v-for="processor in processorsWithoutBizAccess"
                  class="mr4"
                  :key="processor"
                  :value="processor"
                  :property="{ type: 'user' }"
                  :display="{ appearance: 'wxwork-link' }"
                />
                {{ t('无权限。') }}
              </template>
            </template>
          </div>

          <!-- 都有权限 -->
          <!-- TODO：或无需引导权限申请 -->
          <div v-if="!hasNoBizAccess || !checkPermission">
            <copy-to-clipboard :content="ticketLink">
              <bk-button theme="primary" text>{{ copyText }}</bk-button>
            </copy-to-clipboard>
            {{ t('去企业微信上联系审批人催单') }}
          </div>
          <!-- 存在无权限审批人 -->
          <div v-else>
            <p class="mb4">{{ t('审批人需要具备「DB数据生产环境」的「业务访问」权限，您可以：') }}</p>
            <p class="mb4">
              1. {{ t('企业微信线下沟通，并') }}
              <copy-to-clipboard :content="ticketLink">
                <bk-button theme="primary" text>{{ copyText }}</bk-button>
              </copy-to-clipboard>
              {{ t('给审批人，需要提供 HCM 相关单据截图') }}
            </p>
            <p>2. {{ t('在权限中心给审批人添加「业务访问」权限后，复制 HCM 链接给审批人到 HCM 进行审批') }}</p>
          </div>
        </template>
        <!-- case2：没有审批人 -->
        <template v-else>{{ t('审批人当前为空，请等待系统自动刷新') }}</template>
      </div>
    </template>
  </bk-popover>
</template>

<style scoped lang="scss">
.expediting-btn-wrap {
  margin-left: 8px;
  padding: 0 8px;
  font-size: 14px;

  &.active {
    background: #e1ecff;
    border-radius: 2px;
  }
}
</style>
