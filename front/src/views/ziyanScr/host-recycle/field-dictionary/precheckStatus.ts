import { getPrecheckStatusList } from '@/api/host/recycle';
let precheckStatus: any[] = [];

getPrecheckStatusList().then((res: { data: { info: any[] } }) => {
  precheckStatus = res.data?.info?.map((item: { description: any; status: any }) => ({
    label: item.description,
    value: item.status,
  }));
});

export const getPrecheckStatusLabel = (value: string) =>
  precheckStatus?.find((status) => status.value === value)?.label || value || '-';
