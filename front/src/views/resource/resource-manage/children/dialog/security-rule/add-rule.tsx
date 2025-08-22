import { PropType, defineComponent, reactive, ref } from 'vue';
import StepDialog from '@/components/step-dialog/step-dialog';
import './add-rule.scss';
import { GLOBAL_BIZS_KEY, VendorEnum } from '@/common/constant';
import AddRuleTable from '../../components/security/add-rule/AddRuleTable';
import { useAccountStore } from '@/store';
import routerAction from '@/router/utils/action';

export type SecurityRule = {
  name: string;
  priority: number;
  ethertype: string;
  sourceAddress: string;
  source_port_range: string;
  targetAddress: string;
  protocol: string;
  destination_port_range: string;
  port: number | string;
  access: string;
  action: string;
  memo: string;
  cloud_service_id: string;
  cloud_service_group_id: string;
};

export enum IP_CIDR {
  IPV4_ALL = '0.0.0.0/0',
  IPV6_ALL = '::/0',
}

export default defineComponent({
  components: {
    StepDialog,
  },

  props: {
    title: {
      type: String,
    },
    isShow: {
      type: Boolean,
    },
    vendor: {
      type: String,
    },
    loading: {
      type: Boolean,
    },
    dialogWidth: {
      type: String,
    },
    activeType: {
      type: String as PropType<'ingress' | 'egress'>,
    },
    relatedSecurityGroups: {
      type: Array as PropType<any>,
    },
    isEdit: {
      type: Boolean as PropType<boolean>,
    },
    templateData: {
      type: Object as PropType<{ ipList: Array<string>; ipGroupList: Array<string> }>,
    },
    id: String,
  },

  emits: ['update:isShow', 'submit'],

  setup(props, { emit }) {
    const accountStore = useAccountStore();

    const instance = ref();
    const isSubmitLoading = ref(false);
    const steps = [
      {
        component: () => (
          <AddRuleTable
            ref={instance}
            vendor={props.vendor as VendorEnum}
            templateData={props.templateData}
            relatedSecurityGroups={props.relatedSecurityGroups}
            id={props.id}
            activeType={props.activeType}
            isEdit={props.isEdit}
          />
        ),
      },
    ];

    const handleClose = () => {
      emit('update:isShow', false);
    };

    const handleConfirm = async () => {
      isSubmitLoading.value = true;
      try {
        await instance.value.handleSubmit();
        emit('submit');
        emit('update:isShow', false);
      } catch (error: any) {
        // 弹出bpaas结果确认弹窗
        if (error.code === 2000012) {
          showBpaasResultConfirmDialog(error);
        }
      } finally {
        isSubmitLoading.value = false;
      }
    };

    const bpaasResultConfirmState = reactive({ isShow: false, isHidden: false, errorId: '' });
    const showBpaasResultConfirmDialog = (error: any) => {
      bpaasResultConfirmState.isHidden = false;
      bpaasResultConfirmState.isShow = true;
      bpaasResultConfirmState.errorId = error.message;
    };
    const handleBpaasResultConfirm = () => {
      routerAction.open({
        path: '/business/ticket/detail',
        query: {
          [GLOBAL_BIZS_KEY]: accountStore.bizs,
          type: 'security_group',
          id: bpaasResultConfirmState.errorId,
          source: 'bpaas',
        },
      });
      handleClose();
    };
    const handleBpaasResultClosed = () => {
      bpaasResultConfirmState.isHidden = true;
      bpaasResultConfirmState.isShow = false;
      bpaasResultConfirmState.errorId = '';
    };

    return {
      steps,
      handleClose,
      handleConfirm,
      isSubmitLoading,
      bpaasResultConfirmState,
      handleBpaasResultConfirm,
      handleBpaasResultClosed,
    };
  },

  render() {
    return (
      <>
        <step-dialog
          renderType={'if'}
          dialogWidth={this.dialogWidth}
          title={this.title}
          loading={this.isSubmitLoading}
          isShow={this.isShow}
          steps={this.steps}
          onConfirm={this.handleConfirm}
          onCancel={this.handleClose}></step-dialog>
        {!this.bpaasResultConfirmState.isHidden && (
          <bk-dialog
            v-model:isShow={this.bpaasResultConfirmState.isShow}
            title='结果确认'
            confirmText='查看审计流程'
            onConfirm={this.handleBpaasResultConfirm}
            onClosed={this.handleBpaasResultClosed}>
            <span>
              当前配置已提交，查看审批流程关注进度。在审批通过后，返回安全组规则列表，点击安全组“同步”按钮以获取最新规则数据
            </span>
          </bk-dialog>
        )}
      </>
    );
  },
});
