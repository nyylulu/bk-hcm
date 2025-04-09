import { h } from 'vue';

import WName from '@/components/w-name';

export const getWNameVNodeList = (nameList: string[]) => {
  return nameList.map((name, index) => {
    if (index < nameList.length - 1) return [h(WName, { name }), ', '];
    return h(WName, { name, class: 'mr4' });
  });
};
