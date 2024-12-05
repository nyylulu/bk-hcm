// 业务需要传入 title、desc
interface IPromptPayloadType {
  title: string;
  desc: string;
  navigator?: string;
  footer?: string;
  icon_url?: string;
  buttons?: {
    desc: string;
    button_type: string;
  }[];
}

export interface IPromptPayloadTypes {
  zh: IPromptPayloadType;
  en: IPromptPayloadType;
  [key: string]: IPromptPayloadType;
}

export interface IProps {
  channel?: 'moa' | 'sms'; // 二次验证通道
  promptPayload: IPromptPayloadTypes; // 二次验证弹窗内容
  verifyText?: string;
  theme?: 'primary' | 'success' | 'warning' | 'danger';
  showVerifyResult?: boolean;
  boundary?: string | HTMLElement;
  disableTeleport?: boolean;
}

export interface IExposes {
  verifyResult: IMoaVerifyResult;
}

export interface IMoaVerifyResult {
  session_id: string;
  status: 'pending' | 'finish' | 'error';
  button_type: 'confirm' | 'cancel';
  errorMessage?: string; // 当 status 为 error 时，会有该字段
}
