import { defineComponent } from 'vue';
import List from './list';

export default defineComponent({
  setup() {
    return () => (
      <section>
        <List></List>
      </section>
    );
  },
});
