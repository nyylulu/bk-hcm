// 服务管理 资源预测 详情
import { defineComponent } from 'vue';
import Basic from '@/components/resource-plan/resource-manage/detail/basic/index';
import Table from '@/components/resource-plan/resource-manage/detail/list/index';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import cssModule from './index.module.scss';
import { useI18n } from 'vue-i18n';

export default defineComponent({
  setup() {
    const { t } = useI18n();

    return () => (
      <>
        <DetailHeader>{t('资源预测详情')}</DetailHeader>
        <section class={cssModule['resource-forecast-details-section']}>
          <Basic class={cssModule['mb-16']} isBiz={false}></Basic>
          <Table isBiz={false}></Table>
        </section>
      </>
    );
  },
});
