import { defineComponent } from 'vue';
import Search from './search';
import Table from './table';
import './index.scss';

export default defineComponent({
  setup() {
    return () => (
      <section class='plan-list-home'>
        <Search></Search>
        <Table></Table>
      </section>
    );
  },
});
