import { getDiskTypesName, getImageName } from '../../transform';
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
