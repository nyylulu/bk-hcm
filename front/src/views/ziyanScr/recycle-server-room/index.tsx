import { defineComponent, ref } from 'vue';
import Search from './search';
import Table from './table';

import cssModule from './index.module.scss';

export default defineComponent({
  setup() {
    const moduleNames = ref<string[]>([]);

    return () => (
      <section class={cssModule.home}>
        <Search v-model:moduleNames={moduleNames.value}></Search>
        <Table moduleNames={moduleNames.value}></Table>
      </section>
    );
  },
});
