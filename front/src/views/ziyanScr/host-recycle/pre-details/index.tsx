import { defineComponent, ref } from 'vue';
import { useTable } from '@/hooks/useTable/useTable';
import { Search } from 'bkui-vue/lib/icon';
import { Button, Sideslider } from 'bkui-vue';
import './index.scss';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
export default defineComponent({
  setup() {
    const { columns } = useColumns('ExecutionRecords');
    const PDcolumns = [
      {
        label: '单号',
        field: 'orderId',
      },
      {
        label: '子单号',
        field: 'suborderId',
      },
      {
        label: 'IP',
        field: 'ip',
        render: ({ row }) => {
          return (
            <div onClick={Detailslist(row)}>
              <span>{row.ip}</span>
            </div>
          );
        },
      },
      {
        label: '状态',
        field: 'status',
      },
      {
        label: '已执行/总数',
        field: 'mem',
        render: ({ row }) => {
          // return <w-name username={operator} />;
          return (
            <div>
              <span class={row.successNum > 0 ? 'c-success' : ''}>{row.successNum}</span>
              <span>/</span>
              <span>{row.totalNum}</span>
            </div>
          );
        },
      },
      {
        label: '更新时间',
        field: 'updateAt',
      },
      {
        label: '创建时间',
        field: 'createAt',
      },
    ];
    const deviceGroups = ['标准型', '高IO型', '大数据型', '计算型'];
    const application = () => {};
    const filter = ref({
      require_type: '',
      region: [],
      zone: [],
      device_type: [],
      device_group: '',
      cpu: '',
      mem: '',
      disk: '',
    });
    const options = ref({
      require_types: [],
      device_groups: deviceGroups,
      device_types: [],
      regions: [],
      zones: [],
      cpu: [],
      mem: [],
    });
    const querying = ref(false);
    const page = ref({
      limit: 50,
      start: 0,
      sort: '-capacity_flag',
    });
    const handleRequireTypeChange = () => {};
    const { CommonTable } = useTable({
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
                  disabled={data.listenerNum > 0 || data.delete_protect}
                  onClick={() => application()}>
                  详情
                </Button>
              );
            },
          },
        ],
      },
      requestOption: {
        type: 'load_balancers/with/delete_protection',
        sortOption: { sort: 'created_at', order: 'DESC' },
      },
      slotAllocation: () => {
        return {
          ScrSwitch: true,
          interface: {
            Parameters: {
              filter: {
                condition: 'AND',
                rules: [
                  {
                    field: 'require_type',
                    operator: 'equal',
                    value: 1,
                  },
                  {
                    field: 'label.device_group',
                    operator: 'in',
                    value: ['标准型'],
                  },
                ],
              },
              page: page.value,
            },
            filter: { simpleConditions: true, requestId: 'devices' },
            path: '/api/v1/woa/config/findmany/config/cvm/device/detail',
          },
        };
      },
    });
    const { CommonTable: ExecutionRecordsCommonTable } = useTable({
      tableOptions: {
        columns,
      },
      requestOption: {
        type: 'load_balancers/with/delete_protection',
        sortOption: { sort: 'created_at', order: 'DESC' },
      },
      slotAllocation: () => {
        return {
          ScrSwitch: false,
          interface: {
            Parameters: {
              filter: {
                condition: 'AND',
                rules: [
                  {
                    field: 'require_type',
                    operator: 'equal',
                    value: 1,
                  },
                  {
                    field: 'label.device_group',
                    operator: 'in',
                    value: ['标准型'],
                  },
                ],
              },
              page: page.value,
            },
            filter: { simpleConditions: true, requestId: 'devices' },
            path: '/api/v1/woa/config/findmany/config/cvm/device/detail',
          },
        };
      },
    });
    const openDetails = ref(false);
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const Detailslist = (row) => {
      openDetails.value = true;
    };
    return () => (
      <div class='common-card-wrap has-selection'>
        <CommonTable>
          {{
            tabselect: () => (
              <>
                <div class='tabselect'>
                  <span>单号</span>
                  <bk-select class='tbkselect' v-model={filter.value.require_type} onChange={handleRequireTypeChange}>
                    {options.value.require_types.map((item) => (
                      <bk-option
                        key={item.require_type}
                        value={item.require_name}
                        label={item.require_type}></bk-option>
                    ))}
                  </bk-select>
                </div>
                <div class='tabselect'>
                  <span>子单号</span>
                  <bk-select
                    class='tbkselect'
                    v-model={filter.value.region}
                    filterable
                    show-select-all
                    multiple-mode='tag'
                    collapse-tags>
                    {options.value.regions.map((item) => (
                      <bk-option
                        key={item.require_type}
                        value={item.require_name}
                        label={item.require_type}></bk-option>
                    ))}
                  </bk-select>
                </div>
                <div class='tabselect'>
                  <span>IP</span>
                  <bk-select
                    class='tbkselect'
                    v-model={filter.value.zone}
                    filterable
                    show-select-all
                    multiple-mode='tag'
                    collapse-tags>
                    {options.value.zones.map((item) => (
                      <bk-option
                        key={item.require_type}
                        value={item.require_name}
                        label={item.require_type}></bk-option>
                    ))}
                  </bk-select>
                </div>
                <div class='tabselect'>
                  <bk-button icon='bk-icon-search' theme='primary' class='bkbutton' loading={querying.value}>
                    <Search></Search>
                    查询
                  </bk-button>
                  <bk-button class='bkbutton' icon='bk-icon-refresh'>
                    清空
                  </bk-button>
                  <bk-button class='bkbutton' icon='bk-icon-refresh'>
                    复制所有主机IP
                  </bk-button>
                  <bk-button class='bkbutton' icon='bk-icon-refresh'>
                    复制失败主机IP
                  </bk-button>
                  <bk-button class='bkbutton' icon='bk-icon-refresh'>
                    导出全部
                  </bk-button>
                </div>
              </>
            ),
          }}
        </CommonTable>
        <Sideslider class='common-sideslider' width='700' isShow={openDetails.value} title='回收预检详情'>
          {{
            default: () => (
              <div class='common-sideslider-content'>
                <div>
                  IP : {} 云梯回收单号: {}
                </div>
                <ExecutionRecordsCommonTable></ExecutionRecordsCommonTable>
              </div>
            ),
          }}
        </Sideslider>
      </div>
    );
  },
});
