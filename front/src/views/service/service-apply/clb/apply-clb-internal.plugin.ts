import { Message } from 'bkui-vue';
import routerAction from '@/router/utils/action';
import { ApplyClbModel } from '@/api/load_balancers/apply-clb/types';
import { GLOBAL_BIZS_KEY } from '@/common/constant';

export const applyClbSuccessHandler = (isBusinessPage: boolean, goBack: () => void, formModel: ApplyClbModel) => {
  Message({ theme: 'success', message: '购买成功' });
  if (isBusinessPage) {
    // 业务下购买CLB, 跳转至单据管理-负载均衡
    routerAction.redirect({
      name: 'ApplicationsManage',
      query: {
        [GLOBAL_BIZS_KEY]: formModel.bk_biz_id,
        type: 'load_balance',
      },
    });
  } else goBack();
};
