import { defineComponent, ref } from 'vue';
import { Tab } from 'bkui-vue';
import { BkTabPanel } from 'bkui-vue/lib/tab';
import { useVerify } from '@/hooks';
import ErrorPage from '@/views/error-pages/403';

import './index.scss';

import DeviceQuery from './components/device-query';
import ApplicationList from './components/application-list';
import { HostApplicationTabEnum } from './constants';
export default defineComponent({
  setup() {
    const { authVerifyData } = useVerify();
    if (!authVerifyData.value.permissionAction.biz_access) return () => <ErrorPage urlKeyId='biz_access' />;

    const activeName = ref('ApplicationList');
    const tabs = [
      {
        key: HostApplicationTabEnum.HostApplicationList,
        label: '申请单据',
        component: () => <ApplicationList />,
      },
      {
        key: HostApplicationTabEnum.DeviceQuery,
        label: '设备查询',
        component: () => <DeviceQuery />,
      },
    ];

    return () => (
      <div class={'host-application-container'}>
        <Tab v-model:active={activeName.value} type='card-grid' class='tab-wrapper'>
          {tabs.map(({ key, label, component }) => (
            <BkTabPanel key={key} label={label} name={key} renderDirective='if'>
              {component()}
            </BkTabPanel>
          ))}
        </Tab>
      </div>
    );
  },
});
