import { getBusinessNameById } from '../../transform';
export const bkBizId = Object.freeze({
  name: 'bk_biz_id',
  cn: '业务',
  type: Number,
  transformer: getBusinessNameById,
});
