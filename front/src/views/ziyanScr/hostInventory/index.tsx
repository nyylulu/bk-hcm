import { defineComponent, ref, onMounted } from 'vue';
import { useTable } from '@/hooks/useTable/useTable';
import { Search } from 'bkui-vue/lib/icon';
// import http from '@/http';
// import components
import { Button } from 'bkui-vue';
import './index.scss';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
export default defineComponent({
  name: 'AllhostInventoryManager',
  setup() {
    // const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;
    const { columns } = useColumns('hostInventor');
    const deviceGroups = ['标准型', '高IO型', '大数据型', '计算型'];
    const application = () => {};
    const defaultFilter = () => ({
      require_type: 1,
      region: [],
      zone: [],
      device_type: [],
      device_group: deviceGroups || [deviceGroups[0]],
      cpu: '',
      mem: '',
      enable_capacity: true,
    });
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
    const deviceConfigDisabled = ref(false);
    const querying = ref(false);
    const deviceTypeDisabled = ref(false);
    const page = ref({
      limit: 50,
      start: 0,
      sort: '-capacity_flag',
    });
    const loadResources = () => {
      querying.value = true;
      getListData();
    };
    const handleRequireTypeChange = () => {};
    // const handleZoneChange = () => {};
    const handleDeviceConfigChange = () => {
      filter.value.device_type = [];
      const { cpu, mem } = filter.value;
      deviceConfigDisabled.value = Boolean(cpu || mem);
    };
    const clearFilter = () => {
      filter.value = Object.assign({}, filter.value, defaultFilter());
      deviceConfigDisabled.value = false;
      deviceTypeDisabled.value = false;
      filterDevices();
    };
    const handleDeviceGroupChange = () => {
      filter.value.cpu = '';
      filter.value.mem = '';
      filter.value.device_type = [];
      loadDeviceTypes();
    };
    const filterDevices = () => {
      page.value.start = 0;
      loadResources();
    };
    const handleDeviceTypeChange = () => {
      filter.value.cpu = '';
      filter.value.mem = '';
      deviceConfigDisabled.value = filter.value.device_type.length > 0;
    };
    const loadRequireTypes = async () => {
      //   const [res] = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/find/config/requirement`);
      //   options.value.require_types = res?.data?.info || [];
    };
    const loadDeviceTypes = async () => {
      //   const [res] = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/find/config/requirement`, filter.value);
      //   options.value.device_types = res?.data?.info || [];
    };
    const loadRestrict = async () => {
      //   const [res] = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/find/config/requirement`);
      //   const { cpu, mem } = res?.data || {};
      //   options.value.cpu = cpu || [];
      //   options.value.mem = mem || [];
    };
    onMounted(() => {
      loadRequireTypes();
      loadRestrict();
      loadDeviceTypes();
    });
    const { CommonTable, getListData } = useTable({
      tableOptions: {
        columns: [
          ...columns,
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
                  一键申请
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
          url: '/api/v1/woa/config/findmany/config/cvm/device/detail',
          payload: {
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
        };
      },
    });
    return () => (
      <div class='common-card-wrap has-selection'>
        <CommonTable>
          {{
            tabselect: () => (
              <>
                <div class='tabselect'>
                  <span>需求类型</span>
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
                  <span>地域</span>
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
                  <span>园区</span>
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
                  <span>实例族</span>
                  <bk-select
                    class='tbkselect'
                    v-model={filter.value.device_group}
                    multiple
                    clearable
                    collapse-tags
                    onChange={handleDeviceGroupChange}>
                    {options.value.device_groups.map((item) => (
                      <bk-option key={item} value={item} label={item}></bk-option>
                    ))}
                  </bk-select>
                </div>
                <div class='tabselect'>
                  <span>机型</span>
                  <bk-select
                    class='tbkselect'
                    v-model={filter.value.device_type}
                    clearable
                    multiple
                    filterable
                    onChange={handleDeviceTypeChange}>
                    {options.value.device_types.map((item) => (
                      <bk-option key={item} value={item} label={item}></bk-option>
                    ))}
                  </bk-select>
                </div>
                <div class='tabselect'>
                  <span>CPU(核)</span>
                  <bk-select
                    class='tbkselect'
                    v-model={filter.value.cpu}
                    clearable
                    filterable
                    onChange={handleDeviceConfigChange}>
                    {options.value.cpu.map((item) => (
                      <bk-option key={item} value={item} label={item}></bk-option>
                    ))}
                  </bk-select>
                </div>
                <div class='tabselect'>
                  <span>内存(G)</span>
                  <bk-select
                    class='tbkselect'
                    v-model={filter.value.mem}
                    clearable
                    filterable
                    onChange={handleDeviceConfigChange}>
                    {options.value.mem.map((item) => (
                      <bk-option key={item} value={item} label={item}></bk-option>
                    ))}
                  </bk-select>
                </div>
                <div class='tabselect'>
                  <bk-button
                    icon='bk-icon-search'
                    theme='primary'
                    class='bkbutton'
                    loading={querying.value}
                    onClick={filterDevices}>
                    <Search></Search>
                    查询
                  </bk-button>
                  <bk-button icon='bk-icon-refresh' onClick={clearFilter}>
                    清空
                  </bk-button>
                </div>
              </>
            ),
          }}
        </CommonTable>
      </div>
    );
  },
});
