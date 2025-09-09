<template>
  <div class="tab-container">
    <bk-tab type="card-grid" v-model:active="applyType" class="header-tab" @update:active="saveActiveType">
      <bk-tab-panel v-for="(item, index) in tabList" :name="item.name" :label="item.label" :key="index">
        <component
          v-if="item.name === applyType"
          :is="item.Component"
          :rules="item.rules"
          v-bind="item.props"
        ></component>
      </bk-tab-panel>
    </bk-tab>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { ApplicationsType } from './typings';
import { QueryRuleOPEnum } from '@/typings';
import HostApply from './children/host-apply';
import PublicCloud from './children/public-cloud';
import HostRecycle from './children/host-recycle';
import ResourcePlanList from './children/resource-plan/list/list-biz.vue';

const router = useRouter();
const route = useRoute();
const { t } = useI18n();

const applyType = ref(route.query?.type || 'all');

const saveActiveType = (val: string) => {
  router.replace({ query: { ...route.query, type: val } });
};
const tabList = ref<ApplicationsType[]>([
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
    Component: PublicCloud,
    rules: [{ field: 'type', op: QueryRuleOPEnum.IN, value: ['create_disk'] }],
  },
  {
    label: t('VPC'),
    name: 'vpc',
    Component: PublicCloud,
    rules: [{ field: 'type', op: QueryRuleOPEnum.IN, value: ['create_vpc'] }],
  },
  {
    label: t('安全组'),
    name: 'security_group',
    Component: PublicCloud,
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
    name: 'load_balancer',
    Component: PublicCloud,
    rules: [{ field: 'type', op: QueryRuleOPEnum.IN, value: ['create_load_balancer'] }],
  },
  {
    label: t('资源预测'),
    name: 'resource_plan',
    Component: ResourcePlanList,
    rules: [],
  },
]);
</script>

<style lang="scss" scoped>
.tab-container {
  height: 100%;
  padding: 24px;
}
</style>
