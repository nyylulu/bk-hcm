import { watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { decodeValueByAtob, encodeValueByBtoa } from '@/utils';

export function useSaveSearchRules(
  queryKey: string,
  searchApi: (searchRulesStr?: string) => void,
  formModel: any,
  immediate = true,
) {
  const router = useRouter();
  const route = useRoute();

  const saveSearchRules = () => {
    router.replace({
      query: { ...route.query, [queryKey]: encodeValueByBtoa(formModel), _t: Date.now() },
    });
  };

  const clearSearchRules = () => {
    router.replace({ query: { ...route.query, [queryKey]: undefined, _t: Date.now() } });
  };

  const backfillSearchRules = (searchRulesStr: string) => {
    Object.assign(formModel, decodeValueByAtob(searchRulesStr));
  };

  watch(
    () => route.query,
    () => {
      const searchRulesStr = route.query[queryKey] as string;
      // 如果query中有搜索条件，则回填
      if (searchRulesStr) backfillSearchRules(searchRulesStr);
      // 请求数据
      searchApi(searchRulesStr);
    },
    { immediate },
  );

  return {
    saveSearchRules,
    clearSearchRules,
  };
}
