import { defineComponent } from 'vue';
import cssModule from './index.module.scss';
import Panel from '@/components/panel';
import { Plus as PlusIcon } from 'bkui-vue/lib/icon';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { Button } from 'bkui-vue';

export default defineComponent({
  props: {
    isBiz: {
      type: Boolean,
      required: true,
    },
  },
  setup(props) {
    const { t } = useI18n();
    const router = useRouter();

    const handleToAdd = () => {
      router.push({
        path: '/business/resource-plan/add',
      });
    };

    return () => (
      <Panel class={cssModule['mb-16']}>
        {props.isBiz && (
          <Button theme='primary' onClick={handleToAdd} class={cssModule.button}>
            <PlusIcon class={cssModule['plus-icon']} />
            {t('新增')}
          </Button>
        )}
        <section>资源预测 -- 列表</section>
      </Panel>
    );
  },
});
