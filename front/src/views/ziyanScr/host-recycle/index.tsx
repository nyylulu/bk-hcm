import { defineComponent, ref, defineAsyncComponent, h, resolveComponent } from 'vue';
export default defineComponent({
  components: {
    HostRecycleTable: defineAsyncComponent(() => import('./host-recycle-table')),
  },
  setup() {
    const activeName = ref('HostRecycleTable');
    const pagePanel = ref([
      {
        name: 'HostRecycleTable',
        label: '主机回收',
      },
    ]);

    return () => (
      <bk-tab v-model:active={activeName.value} type='unborder-card'>
        {pagePanel.value.map(({ name, label }) => {
          return (
            <bk-tab-panel key={name} name={name} label={label}>
              {/* 注意 tsx 内置组件 component 不能直接使用 */}
              {h(resolveComponent(name))}
            </bk-tab-panel>
          );
        })}
      </bk-tab>
    );
  },
});
