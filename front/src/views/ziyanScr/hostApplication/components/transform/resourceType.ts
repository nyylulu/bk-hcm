import { IDCDVM, IDCPM, QCLOUDCVM, QCLOUDDVM } from '@/constants/scr.ts';

const resourceTypes = [
  {
    value: IDCDVM,
    label: 'IDC_DockerVM',
  },
  {
    value: IDCPM,
    label: 'IDC_物理机',
  },
  {
    value: QCLOUDCVM,
    label: '腾讯云_CVM',
  },
  {
    value: QCLOUDDVM,
    label: '腾讯云_DockerVM',
  },
];

export const getResourceTypeName = (value) => {
  return resourceTypes.find((item) => item.value === value)?.label || value;
};
