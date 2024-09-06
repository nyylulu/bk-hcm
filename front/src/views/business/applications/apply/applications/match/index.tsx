import { defineComponent, ref } from 'vue';

import CommonSideslider from '@/components/common-sideslider';
import MatchPanel from '@/views/ziyanScr/hostApplication/components/match-panel';

export default defineComponent({
  props: { data: { type: Object, required: true } },
  setup(props, { expose }) {
    const isSidesliderShow = ref(false);

    const triggerShow = (v: boolean) => {
      isSidesliderShow.value = v;
    };

    expose({ triggerShow });

    return () => (
      <CommonSideslider v-model:isShow={isSidesliderShow.value} title='待匹配' width={1600} noFooter>
        <MatchPanel data={props.data} />
      </CommonSideslider>
    );
  },
});
