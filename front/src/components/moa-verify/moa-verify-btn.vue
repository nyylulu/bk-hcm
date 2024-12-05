<script setup lang="ts">
import { reactive, ref, useAttrs } from 'vue';
import { useI18n } from 'vue-i18n';
import { merge } from 'lodash';

import Cookies from 'js-cookie';
import i18n, { Locale } from '@/language/i18n';
import { useUserStore } from '@/store';
import useTimeoutPoll from '@/hooks/use-timeout-poll';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import http from '@/http';

import type { IExposes, IMoaVerifyResult, IPromptPayloadTypes, IProps } from './typings';
import type { IQueryResData } from '@/typings';

import MoaVerifyResult from './moa-verify-result.vue';

defineOptions({ name: 'moa-verify-btn' });
const props = withDefaults(defineProps<IProps>(), {
  channel: 'moa',
  verifyText: 'MOA校验',
  theme: 'primary',
  showVerifyResult: true,
});
const attrs = useAttrs();

const userStore = useUserStore();
const { t } = useI18n();
const { getBusinessApiPath } = useWhereAmI();

const languageMap: { [key in Locale]: string } = {
  // 针对 moa 只识别 zh 而言，需要转换一下
  'zh-cn': 'zh',
  en: 'en',
};
const defaultPromptPayload: IPromptPayloadTypes = {
  zh: {
    title: '',
    desc: '',
    navigator: '导航栏',
    buttons: [
      { desc: '确定', button_type: 'confirm' },
      { desc: '取消', button_type: 'cancel' },
    ],
  },
  en: {
    title: '',
    desc: '',
    navigator: 'navigator',
    buttons: [
      { desc: 'Allow', button_type: 'confirm' },
      { desc: 'Do Not Allow', button_type: 'cancel' },
    ],
  },
};

let session_id: string;
const loading = ref(false);
const handleClick = async () => {
  const bluekingLanguage = (Cookies.get('blueking_language') || i18n.global.locale.value) as Locale;
  const { username } = userStore;
  const { channel, promptPayload } = props;

  const language = languageMap[bluekingLanguage];

  const res: IQueryResData<{ session_id: string }> = await http.post(`/api/v1/web/${getBusinessApiPath()}moa/request`, {
    username,
    channel,
    language,
    prompt_payload: JSON.stringify(merge(defaultPromptPayload, promptPayload)),
  });

  session_id = res.data?.session_id;

  Object.assign(verifyResult, getDefaultVerifyResult());
  loading.value = true;
  verifyMoa();

  // 轮询 moa 验证结果
  verifyTask.resume();
};

const getDefaultVerifyResult = (): IMoaVerifyResult => ({
  session_id: '',
  status: undefined,
  button_type: undefined,
});
const verifyResult = reactive<IMoaVerifyResult>(getDefaultVerifyResult());
const verifyMoa = async () => {
  try {
    const { username } = userStore;
    const res: IQueryResData<IMoaVerifyResult> = await http.post(`/api/v1/web/${getBusinessApiPath()}moa/verify`, {
      username,
      session_id,
    });
    const { status } = res.data ?? {};

    if (status === 'finish') {
      verifyTask.pause();
      Object.assign(verifyResult, res.data);
      loading.value = false;
    }
  } catch (error) {
    verifyTask.pause();
    Object.assign(verifyResult, { status: 'error', errorMessage: (error as any).message });
    loading.value = false;
  }
};

const verifyTask = useTimeoutPoll(() => {
  verifyMoa();
}, 10 * 1000);

defineExpose<IExposes>({ verifyResult });
</script>

<template>
  <div class="wrapper">
    <bk-button
      :theme="theme"
      :outline="verifyResult.button_type === 'confirm'"
      :loading="loading"
      v-bind="attrs"
      :class="attrs.class"
      @click="handleClick"
    >
      {{ props.verifyText }}
    </bk-button>
    <teleport v-if="loading" :to="boundary" :disabled="disableTeleport || !boundary" defer>
      <bk-alert class="loading-message" theme="warning" closable>
        <span>
          {{ t('请打开手机MOA，对操作内容确认。超时3分钟未确认，需重新校验。如有疑问，点击使用指引（') }}
        </span>
        <bk-link theme="primary" target="_blank" href="https://iwiki.woa.com/p/4013134908">
          https://iwiki.woa.com/p/4013134908
        </bk-link>
        <span>{{ t('）') }}</span>
      </bk-alert>
    </teleport>
    <moa-verify-result v-if="showVerifyResult" :verify-result="verifyResult" />
  </div>
</template>

<style scoped lang="scss">
.wrapper {
  display: flex;
  align-items: center;

  :deep(.error-message) {
    max-width: 1000px;
  }
}

.loading-message {
  :deep(.bk-alert-wraper) {
    align-items: center;
  }
}
</style>
