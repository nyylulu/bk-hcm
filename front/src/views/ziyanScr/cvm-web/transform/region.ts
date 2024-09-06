import { ref } from 'vue';
import { getRegions } from '@/api/host/config-management';
export const useRegions = () => {
  const regionList = ref([]);
  const fetchRegions = async () => {
    const [qcloud, idc] = await Promise.all(['qcloud', 'idc'].map((item) => getRegions(item, {})));
    regionList.value = [...(qcloud?.info || []), ...idc?.info?.map((item) => ({ region_cn: item, region: item }))];
  };
  fetchRegions();
  const findRegion = (someValue) => {
    return regionList.value.find((item) => {
      return Object.values(item).some((value) => value === someValue);
    });
  };
  const getRegionCn = (someValue) => {
    return findRegion(someValue)?.region_cn || someValue;
  };
  return { getRegionCn };
};
