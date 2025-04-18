export enum MoaRequestScene {
  sg_delete = 'sg_delete',
  cvm_start = 'cvm_start',
  cvm_stop = 'cvm_stop',
  cvm_reset = 'cvm_reset',
  cvm_reboot = 'cvm_reboot',
}

export interface IProps {
  scene: MoaRequestScene;
  resIds: string[]; // 操作影响资源IDs
  verifyText?: string;
  successText?: string;
  failText?: string;
  theme?: 'primary' | 'success' | 'warning' | 'danger';
  showVerifyResult?: boolean;
  boundary?: string | HTMLElement;
  disableTeleport?: boolean;
}

export interface IExposes {
  verifyResult: IMoaVerifyResult;
  resetVerifyResult: () => void;
}

export interface IMoaVerifyResult {
  session_id: string;
  status: 'pending' | 'finish' | 'error';
  button_type: 'confirm' | 'cancel';
  errorMessage?: string; // 当 status 为 error 时，会有该字段
}
