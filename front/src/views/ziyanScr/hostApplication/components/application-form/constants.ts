export const CVM_RESOURCE_TYPES = {
  IDCDVM: 'IDC_DockerVM',
  IDCPM: 'IDC_物理机',
  QCLOUDCVM: '腾讯云_CVM',
  QCLOUDDVM: '腾讯云_DockerVM',
};

export const VerifyStatusMap = {
  PASS: '通过',
  FAILED: '未通过',
  NOT_INVOLVED: '不涉及'
};

export enum VerifyStatus {
  PASS = 'PASS',
  FAILED = 'FAILED',
  NOT_INVOLVED = 'NOT_INVOLVED',
}