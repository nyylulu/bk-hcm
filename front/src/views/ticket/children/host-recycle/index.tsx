import { defineComponent, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import cssModule from './index.module.scss';

import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import Applications from './applications';
import Device from './device';

import { useI18n } from 'vue-i18n';

export default defineComponent({
  setup() {
    const router = useRouter();
    const route = useRoute();
    const { t } = useI18n();

    const cloudTypes = ref([
      { label: t('自研云'), value: 'ziyan', disabled: false },
      // 本期只支持自研云, 无需保存云类型至url
      { label: t('公有云'), value: 'public', disabled: true },
    ]);
    const activeCloudType = ref(cloudTypes.value[0].value);

    const scenes = ref([
      { label: t('单据视角'), value: 'applications' },
      { label: t('设备视角'), value: 'device' },
    ]);
    const activeScene = ref(route.query?.scene_recycle || scenes.value[0].value);

    const saveActiveScene = (val: string) => {
      const { bizs, type } = route.query;
      router.replace({ query: { bizs, type, scene_recycle: val } });
    };

    return () => (
      <>
        <section class={cssModule['scene-wrapper']}>
          <BkRadioGroup v-model={activeCloudType.value} class={cssModule.mr24}>
            {cloudTypes.value.map(({ label, value, disabled }) => (
              <BkRadioButton
                class={cssModule['radio-button']}
                key={value}
                label={value}
                disabled={disabled}
                v-bk-tooltips={{
                  content: t('公有云无回收单据，主机回收后，请到回收站查看回收的主机'),
                  disabled: !disabled,
                }}>
                {label}
              </BkRadioButton>
            ))}
          </BkRadioGroup>
          {activeCloudType.value === 'ziyan' && (
            <BkRadioGroup v-model={activeScene.value} onUpdate:modelValue={saveActiveScene}>
              {scenes.value.map(({ label, value }) => (
                <BkRadioButton class={cssModule['radio-button']} key={value} label={value}>
                  {label}
                </BkRadioButton>
              ))}
            </BkRadioGroup>
          )}
        </section>
        <section class={cssModule['content-wrapper']}>
          {(function () {
            if (activeCloudType.value === 'ziyan') {
              if (activeScene.value === 'applications') {
                return <Applications />;
              }
              if (activeScene.value === 'device') {
                return <Device />;
              }
            }
          })()}
        </section>
      </>
    );
  },
});
