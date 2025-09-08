import http from '@/http';
import { IPageQuery, IQueryResData } from '@/typings';
import { IPlanTicketAudit, IPlanTicketCrpAudit } from '@/typings/resourcePlan';
import { getEntirePath } from '@/utils';
import { defineStore } from 'pinia';

export const useResSubTicketStore = defineStore('resSubTicketStore', () => {
  // 业务接口
  const getListByBiz = (bk_biz_id: number, params: SubTicketParam): Promise<SubTicketsResult> => {
    return http.post(getEntirePath(`bizs/${bk_biz_id}/plans/resources/sub_tickets/list`), params);
  };
  const getAuditByBiz = (bk_biz_id: number, sub_ticket_id: string): Promise<SubTicketAuditResult> => {
    return http.get(getEntirePath(`bizs/${bk_biz_id}/plans/resources/sub_tickets/${sub_ticket_id}/audit`));
  };
  const getDetailByBiz = (bk_biz_id: number, sub_ticket_id: string): Promise<SubTicketDetailResult> => {
    return http.get(getEntirePath(`bizs/${bk_biz_id}/plans/resources/sub_tickets/${sub_ticket_id}`));
  };
  const retryTicketsByBiz = (bk_biz_id: number, ticket_id: string): Promise<ActionResult> => {
    return http.post(getEntirePath(`bizs/${bk_biz_id}/plans/resources/tickets/${ticket_id}/retry`));
  };
  const approveAdminNodeByBiz = (
    bk_biz_id: number,
    sub_ticket_id: string,
    params: ApproveAdminNodeParams,
  ): Promise<ActionResult> => {
    return http.post(
      getEntirePath(`bizs/${bk_biz_id}/plans/resources/sub_tickets/${sub_ticket_id}/approve_admin_node`),
      params,
    );
  };

  // 管理员接口
  const getList = (params: SubTicketParam): Promise<SubTicketsResult> => {
    return http.post(getEntirePath(`plans/resources/sub_tickets/list`), params);
  };
  const getAudit = (sub_ticket_id: string): Promise<SubTicketAuditResult> => {
    return http.get(getEntirePath(`plans/resources/sub_tickets/${sub_ticket_id}/audit`));
  };
  const getDetail = (sub_ticket_id: string): Promise<SubTicketDetailResult> => {
    return http.get(getEntirePath(`plans/resources/sub_tickets/${sub_ticket_id}`));
  };
  const retryTickets = (ticket_id: string): Promise<ActionResult> => {
    return http.post(getEntirePath(`plans/resources/tickets/${ticket_id}/retry`));
  };
  const approveAdminNode = (sub_ticket_id: string, params: ApproveAdminNodeParams): Promise<ActionResult> => {
    return http.post(getEntirePath(`plans/resources/sub_tickets/${sub_ticket_id}/approve_admin_node`), params);
  };

  // 获取审批额度
  const getTransferQuotaConfigs = (): Promise<TransferQuotasConfigsResult> => {
    return http.get(getEntirePath(`plans/resources/transfer_quotas/configs`));
  };
  // 获取业务下剩余额度
  const getTransferQuotaSummaryByBiz = (
    bk_biz_id: number,
    params: TransferQuotasParams,
  ): Promise<TransferQuotasResult> => {
    return http.post(getEntirePath(`bizs/${bk_biz_id}/plans/resources/transfer_quotas/summary`), params);
  };
  // 获取资源哎剩余额度
  const getTransferQuotaSummary = (params: TransferQuotasParams): Promise<TransferQuotasResult> => {
    return http.post(getEntirePath(`plans/resources/transfer_quotas/summary`), params);
  };

  return {
    getListByBiz,
    getAuditByBiz,
    getDetailByBiz,
    retryTicketsByBiz,
    approveAdminNodeByBiz,
    getTransferQuotaSummary,
    getList,
    getAudit,
    getDetail,
    retryTickets,
    approveAdminNode,
    getTransferQuotaConfigs,
    getTransferQuotaSummaryByBiz,
  };
});

// 子单列表
export interface SubTicketParam {
  ticket_id?: string;
  statuses?: string[];
  sub_ticket_types?: string[];
  page?: IPageQuery;
}

export const STATUS_ENUM: Record<string, string> = {
  init: '待审批',
  auditing: '审批中',
  rejected: '审批拒绝',
  failed: '失败',
  done: '成功',
  invalid: '已失效',
};
export const STAGE_ENUM: Record<string, string> = {
  admin_audit: '部门审批',
  crp_audit: '公司审批',
};
export interface SubTicketItem {
  id: string;
  // 获取STATUS_ENUM的key
  status: keyof typeof STATUS_ENUM;
  status_name: string;
  stage: keyof typeof STAGE_ENUM;
  sub_ticket_type: string;
  ticket_type_name: string;
  crp_sn: string;
  crp_url: string;
  original_info: {
    cvm: {
      cpu_core: number | null;
      memory: number | null;
    };
  };
  updated_info: {
    cvm: {
      cpu_core: number;
      memory: number;
    };
  };
  submitted_at: string;
  created_at: string;
  updated_at: string;
}
export type SubTicketsResult = { details: SubTicketItem[]; data: { details: SubTicketItem[] } };

// 部门审核
export interface ApproveAdminNodeParams {
  approval: boolean;
  use_transfer_pool: boolean;
}

// 审批流
export interface AdminAudit {
  status: string;
  status_name?: string;
  current_steps: {
    name: string;
    processors: string[];
    processors_auth: Record<string, boolean>;
  }[];
  logs: {
    name: string;
    operator: string;
    operate_at: string;
    message: string;
  }[];
}

export interface Log {
  operator: string;
  operate_at: string;
  message: string;
  name: string;
}

export interface SubTicketAudit extends IPlanTicketAudit {
  id: string;
  admin_audit: AdminAudit;
  crp_audit: IPlanTicketCrpAudit;
}

export type SubTicketAuditResult = IQueryResData<SubTicketAudit>;

// 子单详情类型
export interface CvmInfo {
  res_mode: string;
  device_type: string;
  device_class: string;
  device_family: string;
  core_type: string;
  os?: number;
  cpu_core: number;
  memory: number;
  technical_class?: string;
}

export interface CbsInfo {
  disk_type: string;
  disk_io: number;
  disk_size: number;
}

export interface TicketInfo {
  obs_project: string;
  expect_time: string;
  region_id: string;
  zone_id: string;
  demand_res_types: string[];
  cvm: CvmInfo;
  cbs: CbsInfo;
}

export interface SubTicketDemand {
  demand_class: string;
  original_info: TicketInfo;
  updated_info: TicketInfo;
}

export interface BaseInfo {
  type: string;
  type_name: string;
  bk_biz_id: number;
  op_product_id: number;
  plan_product_id: number;
  virtual_dept_id: number;
  submitted_at: string;
}

export interface StatusInfo {
  status: string;
  status_name: string;
  stage: string;
  admin_audit_status: string;
  crp_sn: string;
  crp_url: string;
  message: string;
}

export interface SubTicketDetail {
  id: string;
  base_info: BaseInfo;
  status_info: StatusInfo;
  demands: SubTicketDemand[];
}
export type SubTicketDetailResult = IQueryResData<SubTicketDetail>;

// 操作类请求统一回调（重试、审批）
export type ActionResult = IQueryResData<null>;

export interface TransferQuotasParams {
  bk_biz_id?: number[];
  year?: number;
  applied_type?: string[];
  sub_ticket_id?: string[];
  technical_class?: string[];
  obs_project: string[];
}

export interface TransferQuotas {
  used_quota: number;
  remain_quota: number;
}

export type TransferQuotasResult = IQueryResData<TransferQuotas>;

export interface TransferQuotasConfigs {
  quota: number;
  audit_quota: number;
}

export type TransferQuotasConfigsResult = IQueryResData<TransferQuotasConfigs>;
