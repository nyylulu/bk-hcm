import { defineComponent } from 'vue';
import Panel from '@/components/panel';

export default defineComponent({
  setup() {
    return () => <Panel title='基础信息'>233</Panel>;
  },
});
