import { defineComponent, ref, computed, onMounted } from 'vue';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { getPrecheckIpList, getPrecheckList } from '@/api/host/recycle';
import { exportTableToExcel } from '@/utils';
import { useTable } from '@/hooks/useTable/useTable';
import { Search } from 'bkui-vue/lib/icon';
import { Button } from 'bkui-vue';
import ExecuteRecord from '../execute-record';
import './index.scss';
export default defineComponent({
  components: {
    ExecuteRecord,
  },
  props: {
    dataInfo: {
      type: Object,
      default: () => {
        return {};
      },
    },
  },
  setup(props) {
    const defaultFilter = () => ({
      order_id: [],
      suborder_id: props.dataInfo?.suborder_id || [],
      ip: props.dataInfo?.ip || [],
    });
    const filter = ref(defaultFilter());
    const page = ref({
      limit: 50,
      start: 0,
      enable_count: false,
    });
    const requestParams = computed(() => {
      return (data) => {
        return {
          ...data,
          page: page.value,
        };
      };
    });
    const getDyParams = ref(requestParams.value(props.dataInfo));
    const querying = ref(false);
    const { columns } = useColumns('pdExecutecolumns');
    const PDcolumns = [...columns];
    PDcolumns.splice(2, 0, {
      label: 'IP',
      field: 'ip',
      render: ({ row }) => {
        return (
          <Button
            text
            theme='primary'
            disabled={row.listener_num > 0 || row.delete_protect}
            onClick={() => application(row)}>
            {row.ip}
          </Button>
        );
      },
    });
    const { CommonTable, getListData } = useTable({
      tableOptions: {
        columns: [
          ...PDcolumns,
          {
            label: '操作',
            width: 120,
            render: ({ data }: { data: any }) => {
              return (
                <Button
                  text
                  theme='primary'
                  disabled={data.listener_num > 0 || data.delete_protect}
                  onClick={() => application(data)}>
                  详情
                </Button>
              );
            },
          },
        ],
      },
      requestOption: {
        dataPath: 'data.info',
      },
      scrConfig: () => {
        return {
          payload: {
            ...getDyParams.value,
          },
          url: '/api/v1/woa/task/findmany/recycle/detect',
        };
      },
    });
    const getPdList = (enableCount = false) => {
      page.value.enable_count = enableCount;
      page.value = enableCount ? Object.assign(page.value, { limit: 0 }) : page.value;
      const params = {
        ...filter.value,
        order_id: filter.value.order_id.map((item) => Number(item)),
      };
      getDyParams.value = requestParams.value(params);
      getListData();
      getIpList();
    };
    const failIpList = ref([]);
    const allIpList = ref([]);
    const getIpList = async () => {
      const params = {
        ...requestParams.value(filter.value),
        order_id: filter.value.order_id.map((item) => Number(item)),
        page: { start: 0, limit: 500 },
      };
      const [failIpData, allIpData] = await Promise.all([
        getPrecheckIpList(Object.assign(params, { status: ['FAILED'] }), {}),
        getPrecheckIpList(params, {}),
      ]);
      failIpList.value = failIpData?.info || [];
      allIpList.value = allIpData?.info || [];
    };
    const filterOrders = () => {
      page.value.start = 0;
      getPdList();
    };
    const clearFilter = () => {
      filter.value = defaultFilter();
      filterOrders();
    };
    const exportToExcel = () => {
      getPrecheckList(Object.assign(requestParams.value(filter.value), { page: { start: 0, limit: 500 } }), {})
        .then((res) => {
          const totalList = res.data?.info || [];
          exportTableToExcel(totalList, columns, '预检详情列表');
        })
        .finally(() => {});
    };
    const openDetails = ref(false);
    const transferData = ref({});
    const application = (row) => {
      openDetails.value = true;
      transferData.value = {
        suborderId: row.suborder_id,
        ip: row.ip,
        page: {
          start: 0,
          limit: 10,
          enable_count: true,
        },
      };
    };
    onMounted(() => {
      getIpList();
    });
    return () => (
      <div class='common-card-wrap has-selection'>
        <CommonTable>
          {{
            tabselect: () => (
              <div class='precheck-operation'>
                <div class='precheck-input'>
                  <span>单号</span>
                  <bk-tag-input
                    class='tag-input-width'
                    v-model={filter.value.order_id}
                    placeholder='请输入单号'
                    allow-create
                    has-delete-icon
                    allow-auto-match
                  />
                </div>
                <div class='precheck-input'>
                  <span>子单号</span>
                  <bk-tag-input
                    class='tag-input-width'
                    v-model={filter.value.suborder_id}
                    placeholder='请输入子单号'
                    allow-create
                    has-delete-icon
                    allow-auto-match
                  />
                </div>
                <div class='precheck-input'>
                  <span>IP</span>
                  <bk-tag-input
                    class='tag-input-width'
                    v-model={filter.value.ip}
                    placeholder='请输入IP'
                    allow-create
                    has-delete-icon
                    allow-auto-match
                  />
                </div>
                <div class='precheck-input'>
                  <bk-button theme='primary' onClick={filterOrders} loading={querying.value}>
                    <Search></Search>
                    查询
                  </bk-button>
                  <bk-button onClick={clearFilter}>清空</bk-button>
                  <bk-button disabled={allIpList.value.length === 0} v-clipboard={allIpList.value.join('\n')}>
                    复制所有主机IP <span>({allIpList.value.length})</span>
                  </bk-button>
                  <bk-button disabled={failIpList.value.length === 0} v-clipboard={failIpList.value.join('\n')}>
                    复制失败主机IP <span>({failIpList.value.length})</span>
                  </bk-button>
                  <bk-button onClick={exportToExcel}>导出全部</bk-button>
                </div>
              </div>
            ),
          }}
        </CommonTable>
        <execute-record v-model={openDetails.value} dataInfo={transferData.value} />
      </div>
    );
  },
});
