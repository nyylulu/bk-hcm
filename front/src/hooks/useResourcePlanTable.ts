import { ref } from 'vue';
import type { IPageQuery } from '@/typings';

export const useTable = <T>(callBack: (arg: IPageQuery) => Promise<T>, key = 'details') => {
  // 查询列表相关状态
  const isLoading = ref(false);
  const tableData = ref<T>();
  const pagination = ref({
    current: 1,
    limit: 10,
    count: 0,
  });
  const sort = ref();
  const order = ref();

  // 更新数据
  const triggerApi = () => {
    isLoading.value = true;

    // 拉取数据
    Promise.all([
      callBack({
        count: false,
        start: (pagination.value.current - 1) * pagination.value.limit,
        limit: pagination.value.limit,
        sort: sort.value,
        order: order.value,
      }),
      callBack({
        count: true,
      }),
    ])
      .then(([listResult, countResult]: [any, any]) => {
        tableData.value = listResult?.data?.[key] || ([] as T);
        pagination.value.count = countResult?.data?.count || 0;
      })
      .finally(() => {
        isLoading.value = false;
      });
  };

  // 页码变化发生的事件
  const handlePageChange = (current: number) => {
    pagination.value.current = current;
    triggerApi();
  };

  // 条数变化发生的事件
  const handlePageSizeChange = (limit: number) => {
    pagination.value.limit = limit;
    triggerApi();
  };

  // 排序变化发生的事件
  const handleSort = ({ column, type }: { column: { field: string }; type: string }) => {
    pagination.value.current = 1;
    sort.value = column.field;
    order.value = type === 'desc' ? 'DESC' : 'ASC';
    triggerApi();
  };

  const resetPagination = () => {
    pagination.value = {
      current: 1,
      limit: 10,
      count: 0,
    };
  };

  return {
    tableData,
    pagination,
    isLoading,
    handlePageChange,
    handlePageSizeChange,
    handleSort,
    resetPagination,
    triggerApi,
  };
};
