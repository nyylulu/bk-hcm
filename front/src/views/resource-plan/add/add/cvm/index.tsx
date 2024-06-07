import { defineComponent, ref } from 'vue';
import Panel from '@/components/panel';
import { useI18n } from 'vue-i18n';
import cssModule from './index.module.scss';

export default defineComponent({
  setup() {
    const { t } = useI18n();

    const basic = ref({});

    return () => (
      <Panel title={t('CVM云主机信息')}>
        <bk-form form-type='vertical' model={basic.value} class={cssModule.home}>
          <bk-form-item label={t('资源模式')} property='name' required class={cssModule['span-6']}>
            <bk-radio-group>
              <bk-radio-button label='按机型' />
              <bk-radio-button label='按机型族' />
            </bk-radio-group>
          </bk-form-item>
          <bk-form-item label={t('机型类型')} property='name' required class={cssModule['span-3']}>
            <bk-select clearable>
              <bk-option></bk-option>
            </bk-select>
          </bk-form-item>
          <bk-form-item label={t('机型规格')} property='name' required class={cssModule['span-3']}>
            <bk-select clearable>
              <bk-option></bk-option>
            </bk-select>
          </bk-form-item>
          <bk-form-item label={t('实例数量')} property='name' class={cssModule['span-2']}>
            <bk-input type='number' suffix={t('台')} clearable />
          </bk-form-item>
          <bk-form-item label={t('CPU总核数')} property='name'>
            <span class={cssModule.number}>7854核</span>
          </bk-form-item>
          <bk-form-item label={t('内存总量')} property='name'>
            <span class={cssModule.number}>533456</span>
          </bk-form-item>
        </bk-form>
      </Panel>
    );
  },
});
