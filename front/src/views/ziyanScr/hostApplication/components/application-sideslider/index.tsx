import { defineComponent, ref, onMounted, watch } from 'vue';
import { Button, Form } from 'bkui-vue';
import AreaSelector from '@/views/ziyanScr/hostApplication/components/AreaSelector';
import ZoneSelector from '@/views/ziyanScr/hostApplication/components/ZoneSelector';
import { useTable } from '@/hooks/useTable/useTable';
import { Search } from 'bkui-vue/lib/icon';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import apiService from '@/api/scrApi';
import './index.scss';
import { expectedDeliveryTime } from '@/common/util';
const { FormItem } = Form;
export default defineComponent({
  name: 'Sideslider',
  props: {
    cpu: {
      type: String,
      default: '',
    },
    mem: {
      type: String,
      default: '',
    },
    region: {
      type: Array,
      default: () => [],
    },
    getform: {
      type: Boolean,
      default: false,
    },
    device: {
      type: Object,
      default: () => {},
    },
  },
  emits: ['oneApplication'],
  setup(props, { emit }) {
    const deviceTypeDisabled = ref(false);
    const deviceConfigDisabled = ref(false);
    const device = ref({
      filter: {
        require_type: props?.device?.filter.require_type ? props?.device?.filter.require_type : 1,
        region: [],
        zone: [],
        device_type: [],
        device_group: ['标准型'],
        cpu: '',
        mem: '',
        enable_capacity: true,
      },
      options: {
        require_types: [],
        regions: [],
        zones: [],
        device_groups: ['标准型', '高IO型', '大数据型', '计算型'],
        device_types: [],
        cpu: [],
        mem: [],
      },
      page: {
        limit: 50,
        start: 0,
        total: 0,
      },
    });
    const order = ref({
      loading: false,
      submitting: false,
      saving: false,
      model: {
        bkBizId: '',
        bkUsername: '',
        requireType: 1,
        enableNotice: false,
        expectTime: expectedDeliveryTime(),
        remark: '',
        follower: [] as any,
        suborders: [] as any,
      },
      rules: {
        bkBizId: [
          {
            required: true,
            message: '请选择业务',
            trigger: 'change',
          },
        ],
        requireType: [
          {
            required: true,
            message: '请选择需求类型',
            trigger: 'change',
          },
        ],
        expectTime: [
          {
            required: true,
            message: '请填写交付时间',
            trigger: 'change',
          },
        ],
        suborders: [
          {
            required: true,
            trigger: 'change',
          },
        ],
      },
      options: {
        requireTypes: [],
      },
    });
    const { columns: CVMApplicationcolumns } = useColumns('CVMApplication');
    const { CommonTable: CVMApplicationTable, getListData: CVMApplicationGetListData } = useTable({
      tableOptions: {
        columns: [
          ...CVMApplicationcolumns,
          {
            label: '操作',
            width: 120,
            render: ({ data }) => {
              return (
                <Button text theme='primary' onClick={() => OneClickApplication(data)}>
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
          payload: { ...requestListParams() },
        };
      },
    });
    const pageInfo = ref({
      limit: 10,
      start: 0,
      sort: '-capacity_flag',
    });
    const requestListParams = () => {
      const rules = [
        device.value.filter.region?.length && {
          field: 'region',
          operator: 'in',
          value: device.value.filter.region,
        },
        device.value.filter.zone?.length && { field: 'zone', operator: 'in', value: device.value.filter.zone },
        device.value.filter.require_type && {
          field: 'require_type',
          operator: 'equal',
          value: device.value.filter.require_type,
        },
        device.value.filter.device_type.length && {
          field: 'device_type',
          operator: 'in',
          value: device.value.filter.device_type,
        },
        device.value.filter.cpu && { field: 'cpu', operator: 'equal', value: device.value.filter.cpu },
        device.value.filter.mem && { field: 'mem', operator: 'equal', value: device.value.filter.mem },
        device.value.filter.device_group && {
          field: 'label.device_group',
          operator: typeof device.value.filter.device_group === 'string' ? '=' : 'in',
          value: device.value.filter.device_group,
        },
      ].filter(Boolean);
      return {
        filter: {
          condition: 'AND',
          rules,
        },
        page: pageInfo.value,
      };
    };

    const OneClickApplication = (data) => {
      emit('oneApplication', data, false);
    };
    // 一键申请侧边栏 改变实例族
    const handleDeviceGroupChange = () => {
      device.value.filter.cpu = '';
      device.value.filter.mem = '';
      device.value.filter.device_type = [];
      CVMapplicationDeviceTypes();
    };
    // 获取一键申请侧边栏地域
    const CVMapplicationDeviceTypes = async () => {
      const { info } = await apiService.getDeviceTypes(device.value.filter);
      device.value.options.device_types = info || [];
    };
    // 地域有数据时禁用cpu 和内存
    const handleCVMDeviceTypeChange = () => {
      device.value.filter.cpu = '';
      device.value.filter.mem = '';
      deviceConfigDisabled.value = device.value.filter.device_type.length > 0;
    };
    // cpu 和内存有数据时禁用地域
    const handleDeviceConfigChange = () => {
      device.value.filter.device_type = [];
      const { cpu, mem } = device.value.filter;
      deviceTypeDisabled.value = Boolean(cpu || mem);
    };
    // 一键申请提交按钮
    const filterDevices = () => {
      device.value.page.start = 0;
      loadResources();
    };
    const loadResources = () => {
      CVMApplicationGetListData();
    };
    // 一键申请取消按钮
    const CVMclearFilter = () => {
      device.value.filter = {
        require_type: 1,
        region: [],
        zone: [],
        device_type: [],
        device_group: ['标准型'],
        cpu: '',
        mem: '',
        enable_capacity: true,
      };
      deviceConfigDisabled.value = false;
      deviceTypeDisabled.value = false;
      filterDevices();
    };
    const getfetchOptionslist = async () => {
      const { info } = await apiService.getRequireTypes();
      order.value.options.requireTypes = info;
    };
    // 获取一键申请侧边栏cpu
    const loadRestrict = async () => {
      const { cpu, mem } = await apiService.getRestrict();
      device.value.options.cpu = cpu || [];
      device.value.options.mem = mem || [];
    };
    const getviewapplication = () => {
      if (props.getform && (props.cpu || props.mem)) {
        device.value.filter.cpu = props.cpu;
        device.value.filter.mem = props.mem;
      }
      device.value.filter.region = props.region;
      CVMApplicationGetListData();
    };
    watch(
      () => props.getform,
      () => {
        getviewapplication();
      },
    );
    onMounted(() => {
      getfetchOptionslist();
      CVMapplicationDeviceTypes();
      loadRestrict();
      getviewapplication();
    });
    return () => (
      <>
        <Form class={'scr-form-wrapper'}>
          <FormItem label='需求类型'>
            <bk-select class='bk-form-content' disabled v-model={device.value.filter.require_type} filterable>
              {order.value.options.requireTypes.map((item) => (
                <bk-option key={item.require_type} value={item.require_type} label={item.require_name}></bk-option>
              ))}
            </bk-select>
          </FormItem>

          <FormItem label='地域'>
            <AreaSelector
              ref='areaSelector'
              class='bk-form-content'
              v-model={device.value.filter.region}
              multiple
              clearable
              filterable
              params={{ resourceType: 'QCLOUDCVM' }}></AreaSelector>
          </FormItem>

          <FormItem label='园区'>
            <ZoneSelector
              ref='zoneSelector'
              multiple
              v-model={device.value.filter.zone}
              class='bk-form-content'
              params={{
                resourceType: 'QCLOUDCVM',
                region: device.value.filter.region,
              }}></ZoneSelector>
          </FormItem>

          <FormItem label='实例族'>
            <bk-select
              class='bk-form-content'
              v-model={device.value.filter.device_group}
              multiple
              clearable
              collapse-tags
              onChange={handleDeviceGroupChange}>
              {device.value.options.device_groups.map((item) => (
                <bk-option key={item} value={item} label={item}></bk-option>
              ))}
            </bk-select>
          </FormItem>

          <FormItem label='机型'>
            <bk-select
              class='bk-form-content'
              v-model={device.value.filter.device_type}
              clearable
              disabled={deviceTypeDisabled.value}
              multiple
              filterable
              onChange={handleCVMDeviceTypeChange}>
              {device.value.options.device_types.map((item) => (
                <bk-option key={item} value={item} label={item}></bk-option>
              ))}
            </bk-select>
          </FormItem>

          <FormItem label='CPU(核)'>
            <bk-select
              class='bk-form-content'
              v-model={device.value.filter.cpu}
              clearable
              disabled={deviceConfigDisabled.value}
              filterable
              onChange={handleDeviceConfigChange}>
              {device.value.options.cpu.map((item) => (
                <bk-option key={item} value={item} label={item}></bk-option>
              ))}
            </bk-select>
          </FormItem>

          <FormItem label='内存 (G)'>
            <bk-select
              class='bk-form-content'
              v-model={device.value.filter.mem}
              clearable
              disabled={deviceConfigDisabled.value}
              filterable
              onChange={handleDeviceConfigChange}>
              {device.value.options.mem.map((item) => (
                <bk-option key={item} value={item} label={item}></bk-option>
              ))}
            </bk-select>
          </FormItem>
          <Button class={'ml24 mr8'} theme='primary' onClick={filterDevices}>
            <Search></Search>
            查询
          </Button>
          <Button onClick={CVMclearFilter}>重置</Button>
        </Form>
        {/* <div style={{width: 100}}> */}
        <div class={'margin20'}>
          <CVMApplicationTable />
        </div>
      </>
    );
  },
});
