import { TENTHOUSAND, ONETHOUSAND } from '@/constants/scr';

export const getNetworkTypeCn = (type) => {
  const types = new Map();

  types.set(TENTHOUSAND, '万兆');
  types.set(ONETHOUSAND, '千兆');

  return types.get(type) || type;
};
