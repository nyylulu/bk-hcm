// 服务管理 资源预测 详情

import { defineComponent } from 'vue';
import Table from '@/components/resource-plan/resource-manage/detail/list/index';
import Basic from '@/components/resource-plan/resource-manage/detail/basic/index';
import cssModule from './index.module.scss';

export default defineComponent({
  setup() {
    return () => (
      <>
        <section class={cssModule.home}>资源管理 资源预测 详情</section>
        <Basic isBiz={false}></Basic>
        <Table isBiz={false}></Table>
      </>
    );
  },
});
