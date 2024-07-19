import { defineComponent, ref } from 'vue';
import Search from './search';
import Table from './table';

import cssModule from './index.module.scss';
import { useVerify } from '@/hooks';
import ErrorPage from '@/views/error-pages/403';

export default defineComponent({
  setup() {
    const { authVerifyData } = useVerify();
    if (!authVerifyData.value.permissionAction.service_resource_dissolve_find)
      return () => <ErrorPage urlKeyId='biz_ziyan_resource_dissolve' />;

    const moduleNames = ref<string[]>([]);

    return () => (
      <section class={cssModule.home}>
        <Search v-model:moduleNames={moduleNames.value}></Search>
        <Table moduleNames={moduleNames.value}></Table>
      </section>
    );
  },
});
