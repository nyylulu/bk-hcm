import { getRecycleStatusOpts } from '@/api/host/recycle';
let recycleTaskStatus: any[] = [];

getRecycleStatusOpts().then((res: { data: { info: any[] } }) => {
  recycleTaskStatus = res.data?.info?.map((item: { description: any; status: any }) => ({
    label: item.description,
    value: item.status,
  }));
});

export const getRecycleTaskStatusLabel = (value: string) =>
  recycleTaskStatus?.find((status) => status.value === value)?.label || value || '-';
