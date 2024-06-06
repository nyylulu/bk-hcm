import { defineComponent } from 'vue';
import Search from './search';
import Table from './table';
import cssModule from './index.module.scss';

export default defineComponent({
  setup() {
    return () => (
      <section class={cssModule.home}>
        <Search></Search>
        <Table></Table>
      </section>
    );
  },
});
