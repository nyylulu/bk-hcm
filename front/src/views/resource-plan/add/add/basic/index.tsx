import { defineComponent, ref } from 'vue';
import Panel from '@/components/panel';
import { useI18n } from 'vue-i18n';
import cssModule from './index.module.scss';

export default defineComponent({
  setup() {
    const { t } = useI18n();

    const basic = ref({});

    return () => (
      <Panel title={t('基础信息')}>
        <bk-form form-type='vertical' model={basic.value} class={cssModule.home}>
          <bk-form-item label={t('资源类型')} property='name' required>
            <bk-radio-group>
              <bk-radio-button label='CVM' />
              <bk-radio-button label='CBS' />
            </bk-radio-group>
          </bk-form-item>
          <bk-form-item label={t('项目类型')} property='name' required>
            <bk-select clearable>
              <bk-option></bk-option>
            </bk-select>
          </bk-form-item>
          <bk-form-item label={t('云地域')} property='name' required>
            <bk-select clearable>
              <bk-option></bk-option>
            </bk-select>
          </bk-form-item>
          <bk-form-item label={t('可用区')} property='name'>
            <bk-select clearable>
              <bk-option></bk-option>
            </bk-select>
          </bk-form-item>
          <bk-form-item label={t('期望到货日期')} property='name' required>
            <bk-input clearable />
          </bk-form-item>
          <bk-form-item label={t('变更原因')} property='name'>
            <bk-select clearable>
              <bk-option></bk-option>
            </bk-select>
          </bk-form-item>
          <bk-form-item label={t('需求备注')} property='name' class={cssModule['span-2']}>
            <bk-input type='textarea' clearable />
          </bk-form-item>
        </bk-form>
      </Panel>
    );
  },
});
