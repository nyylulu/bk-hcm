import { defineComponent, ref } from 'vue';
import Search from './search';
import Table from './table';
import cssModule from './index.module.scss';
import { useI18n } from 'vue-i18n';

import type { IListTicketsParam } from '@/typings/resourcePlan';

export default defineComponent({
  setup() {
    const { t } = useI18n();

    const tableRef = ref(null);

    const handleSearch = (searchModel: Partial<IListTicketsParam>) => {
      tableRef.value.searchTableData(searchModel);
    };

    return () => (
      <>
        <span class={cssModule.header}>{t('资源预测')}</span>
        <section class={cssModule.home}>
          <Search onSearch={handleSearch}></Search>
          <Table ref={tableRef}></Table>
        </section>
      </>
    );
  },
});
