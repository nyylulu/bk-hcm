import { ref, computed } from 'vue';
import { getZones } from '@/api/host/config-management';
export const useZones = () => {
  const zoneList = ref([]);

  const requestParams = computed(() => {
    return (vendor) => {
      const fixAttr = {
        vendor,
      };
      if (vendor === 'qcloud') {
        return {
          ...fixAttr,
          params: {
            region: [],
          },
        };
      }
      if (vendor === 'idc') {
        return { ...fixAttr, params: { cmdb_region_name: [] } };
      }
      return fixAttr;
    };
  });
  const fetchZones = async () => {
    const [qcloud, idc] = await Promise.all(['qcloud', 'idc'].map((item) => getZones(requestParams.value(item), {})));
    zoneList.value = [
      ...(qcloud?.data?.info || []),
      ...idc?.data?.info?.map((item) => ({ zone_cn: item, zone: item })),
      { zone_cn: '分Campus', zone: 'cvm_separate_campus' }, // 分Campus 是特殊的，不是真正的Zone
    ];
  };
  fetchZones();
  const getZoneCn = (zoneId) => {
    const zone = zoneList.value.find((zone) => zone.zone === zoneId);
    const zoneLabel = zone?.zone_cn || zoneId || '--';
    const cmdbZoneName = zone?.cmdb_zone_name ? `(${zone.cmdb_zone_name})` : '';

    return `${zoneLabel}${cmdbZoneName}`;
  };
  return { getZoneCn };
};
