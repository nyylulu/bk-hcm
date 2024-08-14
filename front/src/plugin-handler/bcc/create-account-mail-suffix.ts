import { ref } from 'vue';
export const PluginHandlerMailbox = {
  suffixText: '@tencent.com' as any,
  isMailValid: ref(false),
  emailRules: [
    {
      trigger: 'change',
      message: '账号邮箱不能为空',
      validator: (val: string) => {
        const isValid = val.trim() !== '';
        PluginHandlerMailbox.isMailValid.value = isValid;
        return isValid;
      },
    },
  ],
};

export type PluginHandlerMailbox = typeof PluginHandlerMailbox;
