import { cloneDeep } from 'lodash-es';
import {
  ZiyanSecurityGroupRule,
  ZiyanSourceTypeArr,
  ZiyanSourceAddressType,
  ZiyanTemplatePortArr,
  ZiyanTemplatePort,
} from '.';

export const ziyanHandler = (data: ZiyanSecurityGroupRule & { sourceAddress: ZiyanSourceAddressType }) => {
  // 协议选择参数模板端口/端口组
  if (ZiyanTemplatePortArr.includes(data.protocol as ZiyanTemplatePort)) {
    delete data.protocol;
    delete data.port;
  }
  // 仅保留选中的源地址类型
  ZiyanSourceTypeArr.forEach((type) => data.sourceAddress !== type && delete data[type]);
  return data;
};

export const ZiyanPreHandler = (data: ZiyanSecurityGroupRule & { sourceAddress: ZiyanSourceAddressType }) => {
  const res = cloneDeep(data);
  // 源地址类型

  ZiyanSourceTypeArr.forEach((type) => res[type] && (res.sourceAddress = type));

  // 协议为参数模板时，给协议、端口赋值
  ZiyanTemplatePortArr.forEach((type) => res[type] && (res.protocol = type));
  return res;
};
