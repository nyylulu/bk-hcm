import { Ref } from 'vue';
import { ACCOUNT_TYPES, COMMON_TYPES } from '../apply-list/constants';

import AccountApplyDetail from './account-apply-detail';
import ApplyDetail from '@/views/service/my-apply/components/apply-detail/index.vue';
import BpassApplyDetail from '../my-apply/components/bpass-apply-detail';

export const applyContentRender = (
  currentApplyData: Ref<any>,
  curApplyKey: Ref<string>,
  applyDetailProps: any,
  bpaasProps: any,
) => {
  if (currentApplyData.value.source === 'bpaas') {
    return <BpassApplyDetail params={currentApplyData.value} key={curApplyKey.value} {...bpaasProps} />;
  }
  return (
    <>
      {ACCOUNT_TYPES.includes(currentApplyData.value.type) && <AccountApplyDetail detail={currentApplyData.value} />}

      {COMMON_TYPES.includes(currentApplyData.value.type) && (
        <ApplyDetail params={currentApplyData.value} key={curApplyKey.value} {...applyDetailProps} />
      )}
    </>
  );
};
