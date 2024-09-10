import { QueryRuleOPEnum, type RulesItem } from '@/typings';
import { VendorEnum } from '@/common/constant';

export const resourceCond: RulesItem[] = [{ op: QueryRuleOPEnum.NEQ, field: 'vendor', value: VendorEnum.ZIYAN }];
