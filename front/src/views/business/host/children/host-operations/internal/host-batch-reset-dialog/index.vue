<script setup lang="ts">
import { reactive, ref, useTemplateRef, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';

import { Message } from 'bkui-vue';
import description from './description.vue';
import firstStep from './first-step/index.vue';
import secondStep from './second-step/index.vue';
import thirdStep from './third-step/index.vue';

import { useCvmResetStore } from '@/store/cvm/reset';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { getPrivateIPs, getPublicIPs } from '@/utils';
import routerAction from '@/router/utils/action';
import type { CvmListRestStatusData, ITableModel } from './typings';
import { MENU_BUSINESS_TASK_MANAGEMENT_DETAILS } from '@/constants/menu-symbol';
import { ResourceTypeEnum } from '@/common/constant';

interface Exposes {
  show: (ids: string[]) => Promise<void>;
}

defineOptions({ name: 'host-batch-reset-dialog' });

const { t } = useI18n();
const { getBizsId } = useWhereAmI();
const cvmResetStore = useCvmResetStore();

const isShow = ref(false);
const listData = reactive<CvmListRestStatusData>({ reset: [], unReset: [], count: 0 });
const show = async (ids: string[]) => {
  try {
    // 获取主机重装状态list
    const list = (await cvmResetStore.getCvmListResetStatus({ ids })).map((item) => ({
      ...item,
      private_ip_address: getPrivateIPs(item),
      public_ip_address: getPublicIPs(item),
    }));

    // 组装主机重装状态data
    Object.assign(listData, {
      reset: list.filter((item) => item.reset_status === 0),
      unReset: list.filter((item) => item.reset_status !== 0),
      count: list.length,
    });

    isShow.value = true;
  } catch (error) {
    console.error(error);
  }
};
const hide = () => {
  isShow.value = false;
  state.curStep = 1;
};

const state = reactive({
  steps: [{ title: t('确认主机') }, { title: t('设置参数') }, { title: '重装确认' }],
  curStep: 1,
});

const secondStepRef = useTemplateRef('second-step-ref');
const thirdStepRef = useTemplateRef('third-step-ref');
const thirdStepInitialList = ref<ITableModel[]>();

// 判断是否可点击下一步、提交
const isNextStepDisabled = ref(true);
const nextStepDisabledTooltips = reactive<{ content?: string; disabled: boolean }>({ disabled: true });
watchEffect(() => {
  let disabled = true;
  let content = '';

  if (state.curStep === 1) {
    // 若当前为第一步, 检查是否有可重装的主机
    disabled = listData.reset.length === 0;
    content = t('没有可重装的主机');
  } else if (state.curStep === 2) {
    // 若当前为第二步，检查镜像参数、密码是否为空
    disabled = secondStepRef.value?.validateEmpty();
    content = t('镜像参数或密码不能为空');
  } else {
    // 若当前为第三步，检查是否同意协议
    disabled = thirdStepRef.value?.validateEmpty();
    content = t('请阅读并同意重装协议');
  }

  isNextStepDisabled.value = disabled;
  Object.assign(nextStepDisabledTooltips, { disabled: !disabled, content });
});

const handleNext = async () => {
  if (state.curStep === 1) {
    // 若当前为第一步，直接跳转第二步
    state.curStep += 1;
  } else if (state.curStep === 2) {
    // 若当前为第二步：1、赋值第三步的初始列表 2、校验密码格式
    thirdStepInitialList.value = secondStepRef.value.formModel.hosts;
    await secondStepRef.value.validateForm();
    state.curStep += 1;
  } else {
    // 若当前为第三步, 提交重装
    const { pwd, pwd_confirm } = secondStepRef.value?.formModel || {};
    const hosts = thirdStepRef.value?.getSubmitHosts();
    const params = { hosts, pwd, pwd_confirm };
    const res = await cvmResetStore.cvmBatchResetAsync(params);
    Message({ theme: 'success', message: t('提交成功') });
    hide();
    // 跳转至新任务详情页
    routerAction.redirect({
      name: MENU_BUSINESS_TASK_MANAGEMENT_DETAILS,
      params: { resourceType: ResourceTypeEnum.CVM, id: res.task_management_id },
      query: { bizs: getBizsId() },
    });
  }
};

const handlePrev = () => {
  state.curStep -= 1;
};

defineExpose<Exposes>({ show });
</script>

<template>
  <bk-dialog
    class="i-dialog"
    v-model:is-show="isShow"
    :title="t('批量重装系统')"
    :quick-close="false"
    render-directive="if"
    width="1280"
    @hidden="hide"
  >
    <description />

    <bk-steps class="i-steps" :steps="state.steps" :cur-step="state.curStep" />

    <!-- 1.确认机器 -->
    <first-step v-show="state.curStep === 1" :list-data="listData" />
    <!-- 2.参数设置 -->
    <second-step ref="second-step-ref" v-show="state.curStep === 2" :list="listData.reset" />
    <!-- 3.信息确认 -->
    <third-step ref="third-step-ref" v-show="state.curStep === 3" :list="thirdStepInitialList" />

    <template #footer>
      <div class="i-footer-wrap">
        <bk-button v-if="state.curStep !== 1" class="button" @click="handlePrev">{{ t('上一步') }}</bk-button>
        <bk-button
          class="button ml-auto"
          theme="primary"
          :disabled="isNextStepDisabled"
          :loading="cvmResetStore.isCvmBatchResetAsyncLoading"
          @click="handleNext"
          v-bk-tooltips="nextStepDisabledTooltips"
        >
          {{ state.curStep === 3 ? t('提交') : t('下一步') }}
        </bk-button>
        <bk-button class="button ml8" @click="hide">{{ t('取消') }}</bk-button>
      </div>
    </template>
  </bk-dialog>
</template>

<style scoped lang="scss">
.i-dialog {
  .i-steps {
    margin: 24px auto 16px;
    width: 60%;
  }

  .i-footer-wrap {
    display: flex;
    align-items: center;
    .button {
      min-width: 88px;
    }
    .ml-auto {
      margin-left: auto;
    }
  }
}
</style>
