import { defineComponent, ref, defineAsyncComponent, h, resolveComponent } from 'vue';
export default defineComponent({
  components: {
    HostRecycleTable: defineAsyncComponent(() => import('./host-recycle-table')),
    DeviceQueryTable: defineAsyncComponent(() => import('./device-query-table')),
  },
  setup() {
    const activeName = ref('HostRecycleTable');
    const recycleSubBizId = ref({});
    const getDetailPage = (paramObj) => {
      activeName.value = 'HostRecycleTable';
      recycleSubBizId.value = paramObj;
    };
    const pagePanel = ref([
      {
        name: 'HostRecycleTable',
        label: '主机回收',
        tranferProps: { subBizBillNum: recycleSubBizId },
      },
      {
        name: 'DeviceQueryTable',
        label: '设备查询',
        tranferProps: { onGoBillDetailPage: getDetailPage },
      },
    ]);

    return () => (
      <bk-tab v-model:active={activeName.value} type='unborder-card'>
        {pagePanel.value.map(({ name, label, tranferProps }) => {
          return (
            <bk-tab-panel key={name} name={name} label={label}>
              {/* 注意 tsx 内置组件 component 不能直接使用 */}
              {h(resolveComponent(name), tranferProps)}
            </bk-tab-panel>
          );
        })}
      </bk-tab>
    );
  },
});
