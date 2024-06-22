import { getDiskTypesName, getImageName, getNetworkTypeCn } from '../../transform';
import { getRegionCn, getZoneCn } from '@/views/ziyanScr/cvm-web/transform';
export const imageId = Object.freeze({
  name: 'image_id',
  cn: '镜像',
  transformer: getImageName,
});

export const diskType = Object.freeze({
  name: 'disk_type',
  cn: '数据盘类型',
  transformer: getDiskTypesName,
});

export const diskSize = Object.freeze({
  name: 'disk_size',
  cn: '数据盘大小',
  suffix: 'G',
});

export const vpc = Object.freeze({
  name: 'vpc',
  cn: '私有网络',
});

export const subnet = Object.freeze({
  name: 'subnet',
  cn: '私有子网',
  type: String,
});

export const module = Object.freeze({
  name: 'module',
  cn: '模块',
  type: String,
});

export const replicas = Object.freeze({
  name: 'replicas',
  cn: '需求数量',
});

export const region = Object.freeze({
  name: 'region',
  cn: '地域',
  transformer: getRegionCn,
});

export const zone = Object.freeze({
  name: 'zone',
  cn: '园区',
  transformer: getZoneCn,
});

export const deviceType = Object.freeze({
  name: 'device_type',
  cn: '机型',
});

export const networkType = Object.freeze({
  name: 'network_type',
  cn: '网络类型',
  transformer: getNetworkTypeCn,
});
