import { SCR_RESOURCE_TYPE_NAME } from '@/constants';

export const getResourceTypeName = (value: string) => {
  return SCR_RESOURCE_TYPE_NAME[value as keyof typeof SCR_RESOURCE_TYPE_NAME] || value;
};
