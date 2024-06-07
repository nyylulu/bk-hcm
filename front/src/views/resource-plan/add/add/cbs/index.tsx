import { defineComponent, ref } from 'vue';
import Panel from '@/components/panel';
import { useI18n } from 'vue-i18n';
import cssModule from './index.module.scss';

export default defineComponent({
  setup() {
    const { t } = useI18n();

    const basic = ref({});

    return () => (
      <Panel title={t('CBS云磁盘信息')}>
        <bk-form form-type='vertical' model={basic.value} class={cssModule.home}>
          <bk-form-item label={t('云盘类型')} property='name' required class={cssModule['span-line']}>
            <bk-select clearable>
              <bk-option></bk-option>
            </bk-select>
          </bk-form-item>
          <bk-form-item label={t('云磁盘容量/块')} property='name' required class={cssModule['span-half-line']}>
            <bk-input type='number' suffix='GB' clearable />
          </bk-form-item>
          <bk-form-item label={t('云盘总量')} property='name'>
            <span class={cssModule.number}>7854GB</span>
          </bk-form-item>
          <bk-form-item label={t('所需数量')} property='name' required class={cssModule['span-line']}>
            <bk-input type='number' clearable />
          </bk-form-item>
          <bk-form-item label={t('单实例磁盘IO')} property='name' class={cssModule['span-line']}>
            <bk-input type='number' clearable />
          </bk-form-item>
        </bk-form>
      </Panel>
    );
  },
});
