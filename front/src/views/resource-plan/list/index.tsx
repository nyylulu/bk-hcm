import { defineComponent, ref, watch } from 'vue';
import Search from './search';
import Table from './table';
import cssModule from './index.module.scss';

export default defineComponent({
  setup() {
    const searchRef = ref(null);
    const searchData = ref();

    watch(
      () => searchRef.value?.searchData,
      (newVal) => {
        if (newVal) {
          searchData.value = { ...newVal };
        }
      },
      { deep: true },
    );
    return () => (
      <section class={cssModule.home}>
        <Search ref={searchRef}></Search>
        <Table searchData={searchData.value}></Table>
      </section>
    );
  },
});
