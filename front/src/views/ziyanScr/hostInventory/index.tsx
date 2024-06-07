import { defineComponent, ref, onMounted } from 'vue';
import { useTable } from '@/hooks/useTable/useTable';
import { Search } from 'bkui-vue/lib/icon';
import apiService from '@/api/scrApi';
import { Button } from 'bkui-vue';
import { useRouter } from 'vue-router';
import AreaSelector from '../hostApplication/components/AreaSelector';
import ZoneSelector from '../hostApplication/components/ZoneSelector';
import './index.scss';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
export default defineComponent({
  name: 'AllhostInventoryManager',
  setup() {
    const { columns } = useColumns('hostInventor');
    const deviceGroups = ['标准型', '高IO型', '大数据型', '计算型'];
    const router = useRouter();
    const filter = ref({
      require_type: 1,
      region: [],
      zone: [],
      device_type: [],
      device_group: deviceGroups && [deviceGroups[0]],
      cpu: '',
      mem: '',
      disk: '',
      enable_capacity: true,
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
    const deviceTypeDisabled = ref(false);
    const page = ref({
      limit: 50,
      start: 0,
      sort: '-capacity_flag',
    });
    const queryrules = ref(
      [
        filter.value.region.length && { field: 'region', operator: 'in', value: filter.value.region },
        filter.value.zone.length && { field: 'zone', operator: 'in', value: filter.value.zone },
        filter.value.require_type && { field: 'require_type', operator: 'equal', value: filter.value.require_type },
        filter.value.device_group && { field: 'label.device_group', operator: 'in', value: filter.value.device_group },
        filter.value.device_type.length && { field: 'device_type', operator: 'in', value: filter.value.device_type },
        filter.value.cpu && { field: 'cpu', operator: 'equal', value: filter.value.cpu },
        filter.value.mem && { field: 'mem', operator: 'equal', value: filter.value.mem },
        filter.value.enable_capacity && {
          field: 'enable_capacity',
          operator: 'equal',
          value: filter.value.enable_capacity,
        },
      ].filter(Boolean),
    );
    const loadResources = () => {
      getListData();
    };
    const handleDeviceConfigChange = () => {
      filter.value.device_type = [];
      const { cpu, mem } = filter.value;
      deviceTypeDisabled.value = Boolean(cpu || mem);
    };
    const clearFilter = () => {
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
      queryrules.value = [
        filter.value.region.length && { field: 'region', operator: 'in', value: filter.value.region },
        filter.value.zone.length && { field: 'zone', operator: 'in', value: filter.value.zone },
        filter.value.require_type && { field: 'require_type', operator: 'equal', value: filter.value.require_type },
        filter.value.device_group && {
          field: 'label.device_group',
          operator: 'in',
          value: filter.value.device_group,
        },
        filter.value.device_type.length && { field: 'device_type', operator: 'in', value: filter.value.device_type },
        filter.value.cpu && { field: 'cpu', operator: 'equal', value: filter.value.cpu },
        filter.value.mem && { field: 'mem', operator: 'equal', value: filter.value.mem },
      ].filter(Boolean);

      page.value.start = 0;
      loadResources();
    };
    const handleDeviceTypeChange = () => {
      filter.value.cpu = '';
      filter.value.mem = '';
      deviceConfigDisabled.value = filter.value.device_type.length > 0;
    };
    const loadDeviceTypes = async () => {
      const { info } = await apiService.getDeviceTypes(filter.value);
      options.value.device_types = info || [];
    };
    const loadRestrict = async () => {
      const { cpu, mem } = await apiService.getRestrict();
      options.value.cpu = cpu || [];
      options.value.mem = mem || [];
    };
    const getfetchOptionslist = async () => {
      const { info } = await apiService.getRequireTypes();
      options.value.require_types = info;
    };
    const application = (row: any) => {
      router.push({
        path: '/ziyanScr/hostApplication',
        query: {
          ...row,
        },
      });
    };
    onMounted(() => {
      loadRestrict();
      loadDeviceTypes();
      getfetchOptionslist();
    });

    const { CommonTable, getListData } = useTable({
      tableOptions: {
        columns: [
          ...columns,
          {
            label: '操作',
            width: 120,
            render: ({ row }: { row: any }) => {
              return (
                <Button
                  text
                  theme='primary'
                  disabled={row.listenerNum > 0 || row.delete_protect}
                  onClick={() => application(row)}>
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
              rules: [...queryrules.value],
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
                  <bk-select class='tbkselect' v-model={filter.value.require_type}>
                    {options.value.require_types.map((item) => (
                      <bk-option
                        key={item.require_type}
                        value={item.require_type}
                        label={item.require_name}></bk-option>
                    ))}
                  </bk-select>
                </div>
                <div class='tabselect'>
                  <span>地域</span>
                  <AreaSelector
                    ref='areaSelector'
                    class='tbkselect'
                    v-model={filter.value.region}
                    multiple
                    clearable
                    filterable
                    params={{ resourceType: 'QCLOUDCVM' }}></AreaSelector>
                </div>
                <div class='tabselect'>
                  <span>园区</span>
                  <ZoneSelector
                    ref='zoneSelector'
                    v-model={filter.value.zone}
                    class='tbkselect'
                    multiple
                    params={{
                      resourceType: 'QCLOUDCVM',
                      region: filter.value.region,
                    }}></ZoneSelector>
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
                    disabled={deviceTypeDisabled.value}
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
                    disabled={deviceConfigDisabled.value}
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
                    disabled={deviceConfigDisabled.value}
                    filterable
                    onChange={handleDeviceConfigChange}>
                    {options.value.mem.map((item) => (
                      <bk-option key={item} value={item} label={item}></bk-option>
                    ))}
                  </bk-select>
                </div>
                <div class='tabselect'>
                  <bk-button icon='bk-icon-search' theme='primary' class='bkbutton' onClick={filterDevices}>
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
