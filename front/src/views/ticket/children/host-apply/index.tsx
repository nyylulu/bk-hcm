import { defineComponent, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import cssModule from './index.module.scss';

import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import Applications from './applications';
import Device from './device';
import PublicCloudApplications from '../public-cloud';

import { useI18n } from 'vue-i18n';
import { QueryRuleOPEnum } from '@/typings';

export default defineComponent({
  setup() {
    const router = useRouter();
    const route = useRoute();
    const { t } = useI18n();

    const cloudTypes = ref([
      { label: t('自研云'), value: 'ziyan' },
      { label: t('公有云'), value: 'public' },
    ]);
    const activeCloudType = ref(route.query?.cloud_type || cloudTypes.value[0].value);

    const scenes = ref([
      { label: t('单据视角'), value: 'applications' },
      { label: t('设备视角'), value: 'device' },
    ]);
    const activeScene = ref(route.query?.scene_apply || scenes.value[0].value);

    const saveActiveCloudType = (val: string) => {
      activeScene.value = scenes.value[0].value;
      const { bizs, type } = route.query;
      router.replace({ query: { bizs, type, cloud_type: val } });
    };

    const saveActiveScene = (val: string) => {
      const { bizs, type, cloud_type } = route.query;
      router.replace({
        query: { bizs, type, cloud_type, scene_apply: val },
      });
    };

    return () => (
      <>
        <section class={cssModule['scene-wrapper']}>
          <BkRadioGroup
            v-model={activeCloudType.value}
            class={cssModule.mr24}
            onUpdate:modelValue={saveActiveCloudType}>
            {cloudTypes.value.map((item) => (
              <BkRadioButton class={cssModule['radio-button']} key={item.value} label={item.value}>
                {item.label}
              </BkRadioButton>
            ))}
          </BkRadioGroup>
          {activeCloudType.value === 'ziyan' && (
            <BkRadioGroup v-model={activeScene.value} onUpdate:modelValue={saveActiveScene}>
              {scenes.value.map((item) => (
                <BkRadioButton class={cssModule['radio-button']} key={item.value} label={item.value}>
                  {item.label}
                </BkRadioButton>
              ))}
            </BkRadioGroup>
          )}
        </section>
        <section class={cssModule['content-wrapper']}>
          {(function () {
            if (activeCloudType.value === 'public') {
              return (
                <PublicCloudApplications rules={[{ field: 'type', op: QueryRuleOPEnum.IN, value: ['create_cvm'] }]} />
              );
            }
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
