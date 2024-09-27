// 资源管理 (业务)下 资源预测 详情
import { defineComponent } from 'vue';
import Table from '@/components/resource-plan/resource-manage/detail/list/index';
import Basic from '@/components/resource-plan/resource-manage/detail/basic/index';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import cssModule from './index.module.scss';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { Button } from 'bkui-vue';

export default defineComponent({
  setup() {
    const router = useRouter();
    const route = useRoute();
    const { t } = useI18n();

    const handleAdjust = () => {
      router.push({
        path: '/business/service/resource-plan-mod',
        query: {
          planIds: route.query.crpDemandId,
        },
      });
    };

    return () => (
      <>
        <DetailHeader>{t('资源预测详情')}</DetailHeader>
        <section class={cssModule['resource-forecast-details-section']}>
          <Button class={cssModule.button} onClick={handleAdjust}>
            {t('调整预测')}
          </Button>
          <Basic class={cssModule['mb-16']} isBiz={true}></Basic>
          <Table isBiz={true}></Table>
        </section>
      </>
    );
  },
});
