import { defineComponent, ref } from 'vue';
import cssModule from './index.module.scss';

import { Tab } from 'bkui-vue';

import { useI18n } from 'vue-i18n';
import HostApply from './apply';
import HostRecycle from './recycle';
import PublicCloudApplications from './components/public-cloud';
import { QueryRuleOPEnum, RulesItem } from '@/typings';

interface ApplicationsType {
  label: string;
  name: string;
  Component: any;
  rules?: RulesItem[];
}

export default defineComponent({
  setup() {
    const { t } = useI18n();
    const types = ref<ApplicationsType[]>([
      {
        label: t('主机申请'),
        name: 'host_apply',
        Component: HostApply,
      },
      {
        label: t('主机回收'),
        name: 'host_recycle',
        Component: HostRecycle,
      },
      {
        label: t('硬盘'),
        name: 'disk',
        Component: PublicCloudApplications,
        rules: [{ field: 'type', op: QueryRuleOPEnum.IN, value: ['create_disk'] }],
      },
      {
        label: t('VPC'),
        name: 'vpc',
        Component: PublicCloudApplications,
        rules: [{ field: 'type', op: QueryRuleOPEnum.IN, value: ['create_disk'] }],
      },
      {
        label: t('安全组'),
        name: 'security_group',
        Component: PublicCloudApplications,
        rules: [
          {
            field: 'type',
            op: QueryRuleOPEnum.IN,
            value: [
              'create_security_group',
              'create_security_group',
              'update_security_group',
              'delete_security_group',
              'associate_security_group',
              'disassociate_security_group',
              'create_security_group_rule',
              'update_security_group_rule',
              'delete_security_group_rule',
            ],
          },
        ],
      },
      {
        label: t('负载均衡'),
        name: 'load_balance',
        Component: PublicCloudApplications,
        rules: [{ field: 'type', op: QueryRuleOPEnum.IN, value: ['create_load_balancer'] }],
      },
    ]);
    const activeType = ref(types.value[0].name);

    return () => (
      <div class={cssModule.container}>
        <Tab v-model:active={activeType.value} type='card-grid'>
          {types.value.map(({ name, label, Component, rules }) => (
            <Tab.TabPanel key={name} name={name} label={label} renderDirective='if'>
              <Component rules={rules} />
            </Tab.TabPanel>
          ))}
        </Tab>
      </div>
    );
  },
});
