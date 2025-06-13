import { PropType, defineComponent, onUnmounted, ref, watch } from 'vue';
import './index.scss';
import { Tag } from 'bkui-vue';
import { InfoLine, Spinner } from 'bkui-vue/lib/icon';
export enum BpaasNodeType {
  approve = 1,
  excute = 2,
  condition = 3,
}
export enum BpaasNodeStatus {
  toApprove = 0,
  approved = 1,
  refused = 2,
  scfFailed = 3,
  scfSuccessed = 4,
  autoApproved = 8,
  autoRefused = 9,
  unnecessaryApprove = 10,
  scfApproved = 11,
  scfRefused = 12,
  ckafkaApproved = 14,
  ckafkaRefused = 15,
  asyncScfExcuting = 16,
  approving = 18,
  scfExcuting = 19,
}
export const BpassNodeStatusMap = {
  [BpaasNodeStatus.toApprove]: '待审批',
  [BpaasNodeStatus.approved]: '审批通过',
  [BpaasNodeStatus.refused]: '审批已拒绝',
  [BpaasNodeStatus.scfFailed]: 'scf执行失败',
  [BpaasNodeStatus.scfSuccessed]: 'scf执行成功',

  [BpaasNodeStatus.autoApproved]: '自动审批通过',
  [BpaasNodeStatus.autoRefused]: '自动审批拒绝',
  [BpaasNodeStatus.unnecessaryApprove]: '不需要执行的分支节点',
  [BpaasNodeStatus.scfApproved]: 'Scf审批通过',
  [BpaasNodeStatus.scfRefused]: 'Scf审批拒绝',
  [BpaasNodeStatus.ckafkaApproved]: 'Ckafka审批通过',
  [BpaasNodeStatus.ckafkaRefused]: 'Ckafka审批拒绝',
  [BpaasNodeStatus.asyncScfExcuting]: '异步Scf执行中',
  [BpaasNodeStatus.scfExcuting]: 'Scf执行中',

  [BpaasNodeStatus.approving]: '外部审批中',
};
export enum BpaasStatus {
  toApprove = 0,
  approved = 1,
  refused = 2,
  scfFailed = 3,
  scfSuccessed = 4,
}
export const BpaasStatusMap = {
  [BpaasStatus.toApprove]: '待审批',
  [BpaasStatus.approved]: '审批通过',
  [BpaasStatus.refused]: '拒绝',
  [BpaasStatus.scfFailed]: 'scf执行失败',
  [BpaasStatus.scfSuccessed]: 'scf执行成功',
};
export const BpaasEndStatus = [BpaasStatus.approved, BpaasStatus.refused];
export default defineComponent({
  props: {
    params: {
      required: true,
      type: Object,
    },
    bpaasPayload: {
      required: true,
      type: Object as PropType<{
        bpaas_sn: number;
        account_id: string;
        id: string;
        applicant: string;
      }>,
    },
    loading: {
      required: true,
      type: Boolean,
    },
    getBpaasDetail: {
      required: true,
      type: Function,
    },
    isGotoSecurityRuleShow: Boolean,
  },
  setup(props) {
    const renderStatus = (status: BpaasNodeStatus | BpaasStatus) => {
      let tagTheme = '';
      switch (status) {
        case BpaasStatus.toApprove:
        case BpaasNodeStatus.toApprove: {
          break;
        }
        case BpaasStatus.scfSuccessed:
        case BpaasStatus.approved:
        case BpaasNodeStatus.scfApproved:
        case BpaasNodeStatus.autoApproved:
        case BpaasNodeStatus.scfSuccessed:
        case BpaasNodeStatus.approved: {
          tagTheme = 'success';
          break;
        }
        case BpaasStatus.scfFailed:
        case BpaasStatus.refused:
        case BpaasNodeStatus.autoRefused:
        case BpaasNodeStatus.ckafkaRefused:
        case BpaasNodeStatus.scfRefused:
        case BpaasNodeStatus.refused:
        case BpaasNodeStatus.scfFailed: {
          tagTheme = 'warning';
          break;
        }
        case BpaasNodeStatus.asyncScfExcuting:
        case BpaasNodeStatus.scfExcuting:
        case BpaasNodeStatus.approving: {
          tagTheme = 'info';
          break;
        }
      }
      return (
        <Tag type='filled' theme={tagTheme as any} radius={'10px'}>
          {BpassNodeStatusMap[status] || '--'}
        </Tag>
      );
    };

    const isFirstLoading = ref(true);
    const bfsSortedNodes = ref([]);

    const intervalId = setInterval(() => {
      isFirstLoading.value = false;
      props?.getBpaasDetail(props.bpaasPayload.id, true, {
        bpaas_sn: props.bpaasPayload.bpaas_sn,
        account_id: props.bpaasPayload.account_id,
      });
    }, 30000);

    onUnmounted(() => {
      clearInterval(intervalId);
    });

    // BFS
    const sortNodes = () => {
      const res = [];
      // 记录流程节点与当前索引的映射，方便重排
      const map = new Map();
      let head = null;
      let tail = null;
      // 用于拓扑图节点去重
      const set = new Set();
      for (let i = 0; i < props.params.Nodes.length; i++) {
        const node = props.params.Nodes[i];
        map.set(+node.NodeId, i);
        if (node.PrevNode === '0') head = node;
        if (node.NextNode === '-1') tail = node;
      }
      // BFS
      const queue = [];
      queue.push(head);
      while (queue.length) {
        const tmpNode = queue.shift();
        res.push(tmpNode);
        if (tmpNode?.NextNode && tmpNode !== tail) {
          let start = 0;
          let end = 0;
          if (isNaN(tmpNode.NextNode)) {
            [start, end] = tmpNode.NextNode.split('-');
          } else {
            start = tmpNode.NextNode;
            end = tmpNode.NextNode;
          }
          for (let j = +start; j <= +end; j++) {
            if (!set.has(j)) {
              const idx = map.get(j);
              const tmp = props.params.Nodes[idx];
              queue.push(tmp);
              set.add(j);
            }
          }
        }
      }
      bfsSortedNodes.value = res;
    };

    watch(
      () => props.params.Nodes,
      (nodes) => {
        if (nodes?.length) sortNodes();
      },
      {
        deep: true,
        immediate: true,
      },
    );

    return () => (
      <div class={'bpass-apply-detail-container'}>
        <div class={'detail-card'}>
          <div class={'title'}>基本信息</div>

          <div class={'item'}>
            <label class={'label'}>单号:</label>
            <div>{props.bpaasPayload.bpaas_sn}</div>
          </div>
          <div class={'item'}>
            <label class={'label'}>类型:</label>
            <div>{props.params.BpaasName}</div>
          </div>
          <div class={'item'}>
            <label class={'label'}>操作人:</label>
            <div>{props.bpaasPayload.applicant}</div>
          </div>
          <div class={'item'}>
            <label class={'label'}>操作时间:</label>
            <div>{props.params.ModifyTime}</div>
          </div>
          {props.isGotoSecurityRuleShow && (
            <div class='item'>
              <label class='label'>安全组ID:</label>
              <div>{props.params.sg_cloud_id}</div>
            </div>
          )}
        </div>

        <div class={'detail-card'}>
          <div class={'title'}>
            申请状态
            <div class={'tip'}>
              {props.loading ? (
                <>
                  <Spinner />
                  &nbsp;正在同步最新进度中...
                </>
              ) : (
                <>
                  <InfoLine />
                  &nbsp;已是最新的进度数据。
                </>
              )}
            </div>
          </div>
          <div>{renderStatus(props.params.Status)}</div>
        </div>

        <div class={'detail-card'}>
          <div class={'title'}>进度：</div>

          {bfsSortedNodes.value.map(({ CreateTime, ScfName, SubStatus, NodeName, NodeType, ApprovedUin }: any) => (
            <div class={'item ml16 mb8'}>
              <span class={'label'}>{NodeName}</span>
              <span class={'content ml16'}>{renderStatus(SubStatus)}</span>
              {![BpaasNodeStatus.unnecessaryApprove].includes(SubStatus) ? (
                <span>
                  <span class={'content'}>
                    {NodeType === BpaasNodeType.approve ? (
                      <span>
                        <span class={'label'}>审批人：</span>
                        <span class={''}>{ApprovedUin?.join(',') || '--'}</span>
                      </span>
                    ) : (
                      <span>
                        <span class={'label'}>审批方式：</span>
                        <span>{ScfName || '--'}</span>
                      </span>
                    )}
                  </span>
                  <span>
                    {NodeType === BpaasNodeType.approve ? (
                      <span class={'label'}>审批时间：</span>
                    ) : (
                      <span class={'label'}>执行时间：</span>
                    )}
                    <span class={'content'}>{CreateTime || '--'}</span>
                  </span>
                </span>
              ) : null}
            </div>
          ))}
        </div>
      </div>
    );
  },
});
