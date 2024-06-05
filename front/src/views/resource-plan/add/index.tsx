import { defineComponent } from 'vue';
import './index.scss';
import Basic from './basic';
import List from './list';
import Memo from './memo';
import Add from './add';

export default defineComponent({
  setup() {
    return () => (
      <section>
        <Basic></Basic>
        <List></List>
        <Memo></Memo>
        <Add></Add>
      </section>
    );
  },
});
