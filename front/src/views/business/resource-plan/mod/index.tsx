import Mod from '@/components/resource-plan/resource-manage/mod';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { computed, defineComponent } from 'vue';

export default defineComponent({
  setup() {
    const { getBizsId } = useWhereAmI();
    const currentGlobalBusinessId = computed(() => getBizsId());

    return () => <Mod currentGlobalBusinessId={currentGlobalBusinessId.value} />;
  },
});
