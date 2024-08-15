// 资源池 - 任务执行阶段
export const SCR_POOL_PHASE_MAP = {
  INIT: '未执行',
  RUNNING: '执行中',
  SUCCESS: '成功',
  FAILED: '失败',
};

// 资源池 - 资源池下架任务执行详情 - 执行状态
export const SCR_RECALL_DETAIL_STATUS_MAP = {
  RETURNED: '已归还',
  REINSTALLING: '系统重装中',
  REINSTALL_FAILED: '系统重装失败',
  CONF_CHECKING: '配置检查中',
  CONF_CHECK_FAILED: '配置检查失败',
  DONE: '完成',
  TERMINATE: '终止',
};

/**
 * @const {String} IDCDVM IDC 富容器
 */
export const IDCDVM = 'IDCDVM';

/**
 * @const {String} IDCPM IDC 物理机
 */
export const IDCPM = 'IDCPM';

/**
 * @const {String} QCLOUDCVM 腾讯云云虚拟机
 */
export const QCLOUDCVM = 'QCLOUDCVM';

/**
 * @const {String} QCLOUDDVM 腾讯云富容器
 */
export const QCLOUDDVM = 'QCLOUDDVM';

/**
 * @const {String} TENTHOUSAND 万兆
 */
export const TENTHOUSAND = 'TENTHOUSAND';

/**
 * @const {String} ONETHOUSAND 千兆
 */
export const ONETHOUSAND = 'ONETHOUSAND';
