import { defineComponent, ref, computed } from 'vue';
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

    // 在search中，模块名是 `which_stages__module_name` 的格式，这里需要提取出 module_name
    const moduleNameList = computed(() => {
      const names = moduleNames.value.map((item) => item.split('__')[1]).filter(Boolean);
      return [...new Set(names)];
    });

    return () => (
      <section class={cssModule.home}>
        <Search v-model:moduleNames={moduleNames.value}></Search>
        <Table moduleNames={moduleNameList.value}></Table>
      </section>
    );
  },
});
