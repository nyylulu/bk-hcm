import { defineComponent } from 'vue';
import cssModule from './index.module.scss';
import Basic from './basic';
import List from './list';
import Memo from './memo';
import Add from './add';

export default defineComponent({
  setup() {
    return () => (
      <section class={cssModule.home}>
        <Basic></Basic>
        <List></List>
        <Memo></Memo>
        <Add></Add>
      </section>
    );
  },
});
