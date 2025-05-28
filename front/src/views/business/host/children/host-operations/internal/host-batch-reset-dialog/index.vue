<script setup lang="ts">
import { computed, reactive, ref, useTemplateRef } from 'vue';
import { useI18n } from 'vue-i18n';

import { Message } from 'bkui-vue';
import description from './description.vue';
import firstStep from './first-step/index.vue';
import secondStep from './second-step/index.vue';
import thirdStep from './third-step/index.vue';
import moaVerifyBtn from '@/components/moa-verify/moa-verify-btn.vue';
import { MoaRequestScene } from '@/components/moa-verify/typings';

import { useCvmOperateStore } from '@/store/cvm-operate';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { getPrivateIPs, getPublicIPs } from '@/utils';
import routerAction from '@/router/utils/action';
import type { CvmListRestDataView, ICvmOperateTableView } from '../../typings';
import { MENU_BUSINESS_TASK_MANAGEMENT_DETAILS } from '@/constants/menu-symbol';
import { ResourceTypeEnum } from '@/common/constant';

interface Exposes {
  show: (ids: string[]) => Promise<void>;
}

defineOptions({ name: 'host-batch-reset-dialog' });

const { t } = useI18n();
const { getBizsId } = useWhereAmI();
const cvmOperateStore = useCvmOperateStore();

const isShow = ref(false);
const listData = reactive<CvmListRestDataView>({ reset: [], unReset: [], count: 0 });
const nonIdleCvmList = computed(() => listData.reset.filter((item) => item.operate_status === 2));
const show = async (ids: string[]) => {
  isShow.value = true;

  // 获取主机重装状态list
  const list = (await cvmOperateStore.getCvmListOperateStatus({ ids, operate_type: 'reset' })).map((item) => ({
    ...item,
    private_ip_address: getPrivateIPs(item),
    public_ip_address: getPublicIPs(item),
  }));

  // 组装主机重装状态data
  Object.assign(listData, {
    reset: list.filter((item) => [0, 2].includes(item.operate_status)), // 0：可重装 2：非空闲机
    unReset: list.filter((item) => ![0, 2].includes(item.operate_status)),
    count: list.length,
  });
};
const hide = () => {
  isShow.value = false;
  Object.assign(state, { curStep: 1, isAgreeFirstStep: false });
  Object.assign(listData, { reset: [], unReset: [], count: 0 });
};

// 删除可重装机器
const handleDelete = (index: number) => {
  listData.reset.splice(index, 1);
};

const state = reactive({
  steps: [{ title: t('确认主机') }, { title: t('设置参数') }, { title: '重装确认' }],
  curStep: 1,
  isAgreeFirstStep: false,
});

const secondStepRef = useTemplateRef('second-step-ref');
const thirdStepRef = useTemplateRef<InstanceType<typeof thirdStep>>('third-step-ref');
const thirdStepInitialList = ref<ICvmOperateTableView[]>();

// 判断是否可点击下一步、提交
const nextStepDisabledOptions = computed(() => {
  let disabled = true;
  let content = '';

  if (state.curStep === 1) {
    // 若当前为第一步, 检查是否有可重装的主机以及是否勾选影响须知
    disabled = listData.reset.length === 0 || (nonIdleCvmList.value.length > 0 && !state.isAgreeFirstStep);
    content = listData.reset.length === 0 ? t('没有可重装的主机') : t('机器处于非空闲机模块，请确认左侧的影响须知');
  } else if (state.curStep === 2) {
    // 若当前为第二步，检查镜像参数、密码是否为空
    disabled = secondStepRef.value?.isParamsHasEmptyValue;
    content = t('镜像参数或密码不能为空');
  } else {
    // 若当前为第三步，检查是否同意协议以及MOA校验是否通过
    const isAgree = thirdStepRef.value?.formModel.agree;
    disabled = !isAgree || moaVerifyResult.value?.button_type !== 'confirm';
    content = isAgree ? t('未通过MOA校验') : t('请阅读并同意重装协议');
  }
  return { disabled, tooltips: { disabled: !disabled, content } };
});

const submitHosts = computed(() => thirdStepRef.value?.submitHosts || []);

const hostIds = computed(() => submitHosts.value.map((item) => item.id));

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
    const hosts = submitHosts.value;
    const { session_id } = moaVerifyResult.value;
    const params = { hosts, pwd, pwd_confirm, session_id };

    try {
      const res = await cvmOperateStore.cvmBatchResetAsync(params);
      Message({ theme: 'success', message: t('提交成功') });
      hide();
      // 跳转至新任务详情页
      routerAction.redirect({
        name: MENU_BUSINESS_TASK_MANAGEMENT_DETAILS,
        params: { resourceType: ResourceTypeEnum.CVM, id: res.task_management_id },
        query: { bizs: getBizsId() },
      });
    } catch (error: any) {
      if (error.code === 2000019) {
        // MOA校验过期
        Message({ theme: 'error', message: t('MOA校验过期，请重新发起校验后操作') });
        moaVerifyRef.value?.resetVerifyResult();
      } else {
        Message({ theme: 'error', message: error.message });
      }
    }
  }
};

const handlePrev = () => {
  state.curStep -= 1;
};

const moaVerifyRef = useTemplateRef('moa-verify');
const moaVerifyResult = computed(() => moaVerifyRef.value?.verifyResult);

const footerRef = useTemplateRef<HTMLElement>('footer');

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
    <first-step
      v-show="state.curStep === 1"
      :list-data="listData"
      :non-idle-cvm-list="nonIdleCvmList"
      :loading="cvmOperateStore.isCvmListOperateStatusLoading"
      @delete="handleDelete"
    />
    <!-- 2.参数设置 -->
    <second-step ref="second-step-ref" v-show="state.curStep === 2" :list="listData.reset" />
    <!-- 3.信息确认 -->
    <third-step ref="third-step-ref" v-show="state.curStep === 3" :list="thirdStepInitialList" />

    <template #footer>
      <div class="i-footer-wrap" ref="footer">
        <!-- 非空闲机操作影响须知 -->
        <bk-checkbox
          v-if="state.curStep === 1 && nonIdleCvmList.length > 0"
          v-model="state.isAgreeFirstStep"
          class="idle-cvm-agree"
        >
          {{ t('我已确认所选主机正确，并确认非空闲机模块重装不会对业务造成影响') }}
        </bk-checkbox>
        <moa-verify-btn
          v-if="state.curStep === 3"
          class="button moa-verify-btn"
          ref="moa-verify"
          :scene="MoaRequestScene.cvm_reset"
          :res-ids="hostIds"
          :boundary="footerRef"
          :success-text="t('校验成功，请点击右侧“重装”按钮，5分钟内有效。')"
        />
        <bk-button v-if="state.curStep !== 1" class="button" @click="handlePrev">{{ t('上一步') }}</bk-button>
        <bk-button
          class="button"
          theme="primary"
          :disabled="nextStepDisabledOptions.disabled"
          :loading="cvmOperateStore.isCvmBatchResetAsyncLoading"
          @click="handleNext"
          v-bk-tooltips="nextStepDisabledOptions.tooltips"
        >
          {{ state.curStep === 3 ? t('提交') : t('下一步') }}
        </bk-button>
        <bk-button class="button" @click="hide">{{ t('取消') }}</bk-button>
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
    position: relative;
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 8px;

    .button {
      min-width: 88px;
    }

    .idle-cvm-agree {
      margin-right: auto;
    }

    .moa-verify-btn {
      margin-right: auto;

      :deep(.error-message) {
        max-width: 800px;
      }
    }

    :deep(.loading-message) {
      position: absolute;
      left: -24px;
      top: -48px;
      width: calc(100% + 48px);
    }
  }

  :deep(.bk-dialog-content) {
    margin-bottom: 40px;
  }
}
</style>
