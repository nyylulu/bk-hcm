import { defineComponent, ref, onBeforeMount } from 'vue';
import { useZiyanScrStore } from '@/store/ziyanScr';

import Panel from '@/components/panel';

import type { IRecycleArea } from '@/typings/ziyanScr';

type RecycleArea = {
  start_time: string;
  end_time: string;
  which_stages: number;
  children: IRecycleArea[];
};

export default defineComponent({
  setup() {
    const ziyanScrStore = useZiyanScrStore();

    const recycleAreaGroups = ref<RecycleArea[]>([]);
    const loading = ref(false);

    const handleInitRecycleArea = async () => {
      try {
        loading.value = true;
        const data = await ziyanScrStore.getRecycleAreas({
          count: false,
          start: 0,
          limit: 500,
        });
        recycleAreaGroups.value =
          data?.data?.detail?.reduce((acc, cur) => {
            let currentRecycleAreaGroup = acc.find(
              (recycleAreaGroup) => recycleAreaGroup.which_stages === cur.which_stages,
            );

            if (!currentRecycleAreaGroup) {
              currentRecycleAreaGroup = {
                start_time: cur.start_time,
                end_time: cur.end_time,
                which_stages: cur.which_stages,
                children: [],
              };

              acc.push(currentRecycleAreaGroup);
            }

            currentRecycleAreaGroup.children.push(cur);

            return acc;
          }, [] as RecycleArea[]) || [];
      } catch (error) {
        recycleAreaGroups.value = [];
      } finally {
        loading.value = false;
      }
    };

    onBeforeMount(handleInitRecycleArea);

    return () => (
      <>
        <Panel>
          <bk-checkbox>全选所有区间</bk-checkbox>
        </Panel>

        <Panel>
          <bk-checkbox>全选所有区间</bk-checkbox>
        </Panel>
      </>
    );
  },
});
