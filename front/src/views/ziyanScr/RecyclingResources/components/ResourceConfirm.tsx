import { defineComponent, onMounted, ref, watch } from 'vue';
import './index.scss';
import apiService from '../../../../api/scrApi';
export default defineComponent({
  name: 'ResourceConfirm',
  props: {
    recycleForm: {
      type: Object,
      default: () => ({}),
    },
    updateConfirm: {
      type: Function,
      default: () => {},
    },
    returnPlan: {
      type: Object,
      default: () => ({
        cvm: 'IMMEDIATE',
        pm: 'IMMEDIATE',
        skipConfirm: '',
      }),
    },
  },
  emits: ['updateConfirm'],
  setup(props, { emit }) {
    const hostList = ref([]);
    const tableList = ref([]);
    const suborderId = ref();
    const drawer = ref(false);
    const drawerTitle = ref('');
    const getReturnPlanName = (returnPlan: string, resourceType: string) => {
      if (returnPlan === 'IMMEDIATE') {
        return resourceType === 'IDCPM' ? '立即销毁(隔离2小时)' : '立即销毁';
      }
      if (returnPlan === 'DELAY') {
        let label = '延迟销毁';
        if (resourceType === 'IDCPM') {
          label += '(隔离1天)';
        } else if (resourceType === 'QCLOUDCVM') {
          label += '(隔离7天)';
        }
        return label;
      }
      return '';
    };
    const getResourceTypeName = (resourceType: string | number) => {
      const resourceTypeMap = {
        QCLOUDCVM: '腾讯云虚拟机',
        IDCPM: 'IDC物理机',
        OTHERS: '其他',
      };
      return resourceTypeMap[resourceType];
    };
    const columns = ref([
      {
        type: 'selection',
        width: 32,
        minWidth: 32,
        onlyShowOnList: true,
      },
      {
        label: '业务',
        field: 'bk_biz_name',
      },
      {
        label: '资源类型',
        field: 'resource_type',
        render: (row: { resource_type: any }) => {
          return <span>{getResourceTypeName(row.resource_type)}</span>;
        },
      },
      {
        label: '回收类型',
        field: 'recycle_type',
      },
      {
        label: '回收选项',
        render: (row: { return_plan: any; resource_type: any }) => {
          return <span>{getReturnPlanName(row.return_plan, row.resource_type)}</span>;
        },
      },
      {
        label: '资源总数',
        field: 'total_num',
        render: (row: { total_num: any }) => {
          return (
            <>
              <span>{row.total_num}</span>&nbsp;&nbsp;&nbsp;&nbsp;
              <bk-button type='text' size='small' onClick={handleRowClick(scope.row)}>
                详情
              </bk-button>
              ;
            </>
          );
        },
      },
      {
        label: '回收成本',
        field: 'cost_concerned',
      },
      {
        label: '备注',
        field: 'remark',
      },
    ]);
    const page = ref({
      start: 0,
      limit: 10,
      total: 0,
    });
    const handleRowClick = (row: { bk_biz_name: string; suborder_id: any }) => {
      drawer.value = true;
      drawerTitle.value = row.bk_biz_name;
      suborderId.value = [row.suborder_id];
      getList(true);
    };
    /** 获取资源回收单据预览列表 */
    const getPreRecycleList = async () => {
      emit('updateConfirm', []);

      const {
        returnPlan,
        recycleForm: { ips, remark },
      } = props;
      const { info } = await apiService.getPreRecycleList({ ips, remark, returnPlan });
      tableList.value = info || [];
    };
    const getList = async (enableCount = false) => {
      const data = await apiService.getRecycleHosts({
        suborderId,
        page: page.value,
      });
      if (enableCount) page.value.total = data?.count;
      hostList.value = data?.info || [];
    };
    watch(
      () => props.recycleForm,
      () => {
        getPreRecycleList();
      },
    );
    onMounted(() => {
      getPreRecycleList();
    });
    return () => (
      <div class='div-ResourceSelect'>
        <bk-table align='left' row-hover='auto' columns={columns.value} data={hostList.value} show-overflow-tooltip />
        <bk-pagination
          style='float: right;margin: 20px 0;'
          v-model={page.value.start}
          count={page.value.total}
          limit={page.value.limit}
        />
      </div>
    );
  },
});
