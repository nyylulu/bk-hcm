import { defineComponent } from 'vue';
import './index.scss';
import WName from '@/components/w-name';
import { useI18n } from 'vue-i18n';

export default defineComponent({
  name: 'BcsSelectTips',
  props: {
    desc: { type: String, required: true },
  },
  setup(props) {
    const { t } = useI18n();

    return () => (
      <div class='bcs-select-tips text-desc'>
        <span class='text-danger'>{t('注意：')}</span>
        <span>{props.desc}</span>
        <span>
          {t('，请与')}
          <WName name='BCS' alias={t('BCS蓝鲸容器助手')} />
          {t('确认。')}
        </span>
      </div>
    );
  },
});
