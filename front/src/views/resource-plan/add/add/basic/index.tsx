import { defineComponent } from 'vue';
import Panel from '@/components/panel';
import { useI18n } from 'vue-i18n';

export default defineComponent({
  setup() {
    const { t } = useI18n();

    return () => <Panel title={t('基础信息')}>233</Panel>;
  },
});
