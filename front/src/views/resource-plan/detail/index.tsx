import { defineComponent } from 'vue';
import Approval from './approval';
import Basic from './basic';
import List from '../list';

export default defineComponent({
  setup() {
    return () => (
      <section>
        <Approval></Approval>
        <Basic></Basic>
        <List></List>
      </section>
    );
  },
});
