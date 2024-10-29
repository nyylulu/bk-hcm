import { RulesItem, QueryRuleOPEnum } from '@/typings';
import type { ResourceTypeEnum } from '@/common/resource-constant';

export type ModelPropertyType =
  | 'string'
  | 'datetime'
  | 'enum'
  | 'number'
  | 'account'
  | 'user'
  | 'array'
  | 'bool'
  | 'cert'
  | 'ca'
  | 'region'
  | 'bizs'
  | 'business';

export type ModelPropertyMeta = {
  display: PropertyDisplayConfig;
  search: PropertySearchConfig;
  column: PropertyColumnConfig;
  form: PropertyFormConfig;
};

// 模型的基础字段，与业务场景无关
export type ModelProperty = {
  id: string;
  name: string;
  type: ModelPropertyType;
  resource?: ResourceTypeEnum;
  option?: Record<string, any>;
  meta?: ModelPropertyMeta;
  unit?: string;
  index?: number;
};

export type PropertyColumnConfig = {
  sort?: boolean;
  align?: 'left' | 'center' | 'right';
  defaultHidden?: boolean;
};

export type PropertyFormConfig = {
  rules?: object;
};

export type PropertySearchConfig = {
  op: QueryRuleOPEnum;
  filterRules: (value: any) => RulesItem;
};

export type PropertyDisplayConfig = {
  appearance: string;
};

// 与列展示场景相关，联合列的配置属性
export type ModelPropertyColumn = ModelProperty & PropertyColumnConfig;

// 与表单场景相关，联合表单的配置属性
export type ModelPropertyForm = ModelProperty & PropertyFormConfig;

// 与展示场景相关，联合展示的配置属性
export type ModelPropertyDisplay = ModelProperty & PropertyDisplayConfig;

// 与搜索场景相关，联合搜索的配置属性
export type ModelPropertySearch = ModelProperty & PropertySearchConfig;
