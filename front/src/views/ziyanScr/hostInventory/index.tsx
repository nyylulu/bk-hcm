import { defineComponent, ref, onMounted } from 'vue';
import { useTable } from '@/hooks/useTable/useTable';
import { Search } from 'bkui-vue/lib/icon';
import apiService from '@/api/scrApi';
import { Button, Form } from 'bkui-vue';
import { useRouter } from 'vue-router';
import AreaSelector from '../hostApplication/components/AreaSelector';
import ZoneSelector from '../hostApplication/components/ZoneSelector';
import './index.scss';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
const { FormItem } = Form;
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
    });
    const queryrules = ref(
      [
        filter.value.region.length && { field: 'region', operator: 'in', value: filter.value.region },
        filter.value.zone.length && { field: 'zone', operator: 'in', value: filter.value.zone },
        filter.value.require_type && { field: 'require_type', operator: 'equal', value: filter.value.require_type },
        filter.value.device_group.length && {
          field: 'label.device_group',
          operator: 'in',
          value: filter.value.device_group,
        },
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
    const emptyform = () => {
      filter.value = {
        require_type: 1,
        region: [],
        zone: [],
        device_type: [],
        device_group: deviceGroups && [deviceGroups[0]],
        cpu: '',
        mem: '',
        disk: '',
        enable_capacity: true,
      };
    };
    const handleDeviceConfigChange = () => {
      filter.value.device_type = [];
      const { cpu, mem } = filter.value;
      deviceTypeDisabled.value = Boolean(cpu || mem);
    };
    const clearFilter = () => {
      emptyform();
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
        filter.value.device_group.length && {
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
        path: '/ziyanScr/hostApplication/apply',
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
        sortOption: {
          sort: 'capacity_flag',
          order: 'DESC',
        },
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
      <div class={'apply-list-container cvm-web-wrapper'}>
        <div class={'filter-container'}>
          <Form model={filter.value} formType='vertical' class={'scr-form-wrapper'}>
            <FormItem label='需求类型'>
              <bk-select class='tbkselect' v-model={filter.value.require_type}>
                {options.value.require_types.map((item) => (
                  <bk-option key={item.require_type} value={item.require_type} label={item.require_name}></bk-option>
                ))}
              </bk-select>
            </FormItem>
            <FormItem label='地域'>
              <AreaSelector
                ref='areaSelector'
                class='tbkselect'
                v-model={filter.value.region}
                multiple
                clearable
                filterable
                params={{ resourceType: 'QCLOUDCVM' }}></AreaSelector>
            </FormItem>
            <FormItem label='园区'>
              <ZoneSelector
                ref='zoneSelector'
                v-model={filter.value.zone}
                class='tbkselect'
                separateCampus={false}
                multiple
                params={{
                  resourceType: 'QCLOUDCVM',
                  region: filter.value.region,
                }}></ZoneSelector>
            </FormItem>
            <FormItem label='实例族'>
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
            </FormItem>
            <FormItem label='机型'>
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
            </FormItem>
            <FormItem label='CPU(核)'>
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
            </FormItem>
            <FormItem label='内存(G)'>
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
            </FormItem>
          </Form>
          <div class='btn-container'>
            <bk-button icon='bk-icon-search' theme='primary' onClick={filterDevices}>
              <Search></Search>
              查询
            </bk-button>
            <bk-button icon='bk-icon-refresh' onClick={clearFilter}>
              清空
            </bk-button>
          </div>
        </div>
        <CommonTable class={'filter-CommonTable'}></CommonTable>
      </div>
    );
  },
});
