import { Message } from 'bkui-vue';
import routerAction from '@/router/utils/action';
import { GLOBAL_BIZS_KEY } from '@/common/constant';

export const applyClbSuccessHandler = (isBusinessPage: boolean, goBack: () => void, args: any) => {
  Message({ theme: 'success', message: '购买成功' });
  const { id, bk_biz_id: bkBizId } = args || {};
  if (isBusinessPage) {
    // 业务下购买CLB, 跳转至单据管理-负载均衡
    routerAction.redirect({
      path: '/business/applications/detail',
      query: {
        [GLOBAL_BIZS_KEY]: bkBizId,
        type: 'load_balancer',
        id,
      },
    });
  } else goBack();
};
