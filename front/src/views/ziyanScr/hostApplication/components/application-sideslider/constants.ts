import { RequirementType } from '@/store/config/requirement';

export const deviceGroups = ['标准型', '高IO型', '大数据型', '计算型'];
export const deviceGroupsInGreenChannel = ['标准型'];

type RequirementMap = {
  [K in RequirementType | 'default']?: string[];
};
export const deviceGroupsMap: RequirementMap = {
  [RequirementType.GreenChannel]: deviceGroupsInGreenChannel,
  default: deviceGroups,
};
