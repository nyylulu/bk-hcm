import { defineComponent, ref, computed, watch, onMounted } from 'vue';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import { useTable } from '@/hooks/useTable/useTable';
import { getRequireTypes } from '@/api/host/task';
import { getRestrict } from '@/api/host/cvm';
import MemberSelect from '@/components/MemberSelect';
import AreaSelector from '@/views/ziyanScr/hostApplication/components/AreaSelector';
import ZoneSelector from '@/views/ziyanScr/hostApplication/components/ZoneSelector';
import { HelpFill, Search } from 'bkui-vue/lib/icon';
import { Button, Form, Select, Sideslider } from 'bkui-vue';
import DevicetypeSelector from '@/views/ziyanScr/components/devicetype-selector/index.vue';
import { ICvmDeviceDetailItem } from '@/typings/ziyanScr';
const { FormItem } = Form;
// import { statusList } from './transform';
// import './index.scss';

export default defineComponent({
  components: {
    MemberSelect,
    AreaSelector,
    ZoneSelector,
  },
  props: {
    modelValue: {
      type: Boolean,
      default: false,
    },
    title: {
      type: String,
      default: '快速生产',
    },
    actionText: {
      type: String,
      default: '快速生产',
    },
  },
  emits: ['update:modelValue', 'oneKeyApply'],
  setup(props, { attrs, emit }) {
    const instanceList = ['标准型', '高IO型', '大数据型', '计算型'];
    const isDisplay = ref(false);
    watch(
      () => props.modelValue,
      (val) => {
        isDisplay.value = val;
      },
      {
        immediate: true,
      },
    );
    const updateShowValue = () => {
      emit('update:modelValue', false);
    };
    const deviceTypeDisabled = ref(false);
    const defaultFilterForm = () => ({
      require_type: 1,
      region: [],
      zone: [],
      device_type: [],
      device_group: [instanceList[0]],
      cpu: '',
      mem: '',
      enable_capacity: true,
    });
    const filterForm = ref(defaultFilterForm());
    const pageInfo = ref({
      start: 0,
      limit: 10,
    });
    const defaultFilter = () => ({
      condition: 'AND',
      rules: [
        {
          field: 'require_type',
          operator: 'equal',
          value: filterForm.value.require_type,
        },
        {
          field: 'enable_capacity',
          operator: 'equal',
          value: filterForm.value.enable_capacity,
        },
        {
          field: 'label.device_group',
          operator: 'in',
          value: filterForm.value.device_group,
        },
      ],
    });
    const requestListParams = ref({
      filter: defaultFilter(),
      page: pageInfo.value,
    });
    const paramTableRules = computed(() => {
      const rules = [];
      ['region', 'zone', 'device_type', 'device_group'].map((item) => {
        if (Array.isArray(filterForm.value[item]) && filterForm.value[item].length) {
          rules.push({
            field: item === 'device_group' ? 'label.device_group' : item,
            operator: 'in',
            value: filterForm.value[item],
          });
        }
        return null;
      });
      ['require_type', 'cpu', 'mem'].map((item) => {
        if (String(filterForm.value[item])) {
          rules.push({
            field: item,
            operator: 'equal',
            value: filterForm.value[item],
          });
        }
        return null;
      });
      return rules;
    });
    const loadOrders = () => {
      let filter = {};
      if (paramTableRules.value.length) {
        filter = {
          condition: 'AND',
          rules: paramTableRules.value,
        };
      }
      const params = {
        filter,
        page: pageInfo.value,
      };
      requestListParams.value = { ...params };
      getListData();
    };
    const filterOrders = () => {
      pageInfo.value.start = 0;
      loadOrders();
    };
    const clearFilter = () => {
      filterForm.value = defaultFilterForm();
      deviceConfigDisabled.value = false;
      deviceTypeDisabled.value = false;
      filterOrders();
    };
    const oneKeyApply = (row: ICvmDeviceDetailItem) => {
      emit('oneKeyApply', row);
    };
    const { columns } = useColumns('cvmFastProduceQuery');
    // columns.splice();
    const operationList = [
      {
        label: '操作',
        render: ({ row }) => {
          return (
            <Button disabled={row.capacity_flag === 0} theme='primary' text onClick={() => oneKeyApply(row)}>
              {props.actionText}
            </Button>
          );
        },
      },
    ];
    const tableColumns = [...columns, ...operationList];
    const { CommonTable, getListData } = useTable({
      tableOptions: {
        columns: tableColumns,
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
            ...requestListParams.value,
          },
        };
      },
    });
    // 需求类型
    const requireTypeList = ref([]);
    const fetchRequireType = async () => {
      const res = await getRequireTypes();
      requireTypeList.value = res.data.info.map((item) => ({
        label: item.require_name,
        value: item.require_type,
      }));
    };
    const cpuList = ref([]);
    const memList = ref([]);
    const fetchCpuOrMem = async () => {
      const res = await getRestrict();
      const { cpu, mem } = res?.data || {};
      cpuList.value = cpu || [];
      memList.value = mem || [];
    };
    // CVM机型
    const cvmDevicetypeParams = computed(() => {
      const { require_type, region, zone, device_group, cpu, mem, enable_capacity } = filterForm.value;
      return { require_type, region, zone, device_group, cpu, mem, enable_capacity };
    });

    const handleDeviceGroupChange = () => {
      filterForm.value.cpu = '';
      filterForm.value.mem = '';
      filterForm.value.device_type = [];
    };
    const deviceConfigDisabled = ref(false);
    const handleDeviceTypeChange = () => {
      filterForm.value.cpu = '';
      filterForm.value.mem = '';
      deviceConfigDisabled.value = filterForm.value.device_type.length > 0;
    };
    const handleDeviceConfigChange = () => {
      filterForm.value.device_type = [];

      const { cpu, mem } = filterForm.value;

      deviceTypeDisabled.value = Boolean(cpu || mem);
    };
    onMounted(() => {
      fetchRequireType();
      fetchCpuOrMem();
    });
    return () => (
      <Sideslider
        class='common-sideslider'
        v-bind={attrs}
        width='1080'
        v-model:isShow={isDisplay.value}
        title={props.title}
        before-close={updateShowValue}>
        {{
          default: () => (
            <div class='apply-list-container common-sideslider-content'>
              <div class={'filter-container'}>
                <Form formType='vertical' class='scr-form-wrapper' model={filterForm}>
                  <FormItem label='需求类型'>
                    <Select v-model={filterForm.value.require_type} clearable placeholder='请选择'>
                      {requireTypeList.value.map(({ label, value }) => {
                        return <Select.Option key={value} name={label} id={value} />;
                      })}
                    </Select>
                  </FormItem>
                  <FormItem label='地域'>
                    <area-selector multiple v-model={filterForm.value.region} params={{ resourceType: 'QCLOUDCVM' }} />
                  </FormItem>
                  <FormItem label='园区'>
                    <zone-selector
                      multiple
                      v-model={filterForm.value.zone}
                      params={{ resourceType: 'QCLOUDCVM', region: filterForm.value.region }}
                    />
                  </FormItem>
                  <FormItem label='实例族'>
                    <Select
                      v-model={filterForm.value.device_group}
                      multiple
                      clearable
                      placeholder='请选择'
                      onChange={handleDeviceGroupChange}>
                      {instanceList.map((item) => {
                        return <Select.Option key={item} name={item} id={item} />;
                      })}
                    </Select>
                    <div
                      class='tool-pos'
                      v-bk-tooltips={{
                        theme: 'light',
                        content: (
                          <div>
                            实例族相关概念请
                            <a
                              class='link-type'
                              href='https://cloud.tencent.com/document/product/213/11518'
                              target='_blank'>
                              查看文档
                            </a>
                          </div>
                        ),
                      }}>
                      <HelpFill />
                    </div>
                  </FormItem>
                  <FormItem label='机型'>
                    <DevicetypeSelector
                      v-model={filterForm.value.device_type}
                      resourceType='cvm'
                      params={cvmDevicetypeParams.value}
                      multiple
                      disabled={deviceTypeDisabled.value}
                      onChange={handleDeviceTypeChange}
                    />
                  </FormItem>
                  <FormItem label='CPU(核)'>
                    <Select
                      v-model={filterForm.value.cpu}
                      clearable
                      placeholder='请选择'
                      disabled={deviceConfigDisabled.value}
                      onChange={handleDeviceConfigChange}>
                      {cpuList.value.map((item) => {
                        return <Select.Option key={item} name={item} id={item} />;
                      })}
                    </Select>
                  </FormItem>
                  <FormItem label='内存(G)'>
                    <Select
                      v-model={filterForm.value.mem}
                      clearable
                      placeholder='请选择'
                      disabled={deviceConfigDisabled.value}
                      onChange={handleDeviceConfigChange}>
                      {memList.value.map((item) => {
                        return <Select.Option key={item} name={item} id={item} />;
                      })}
                    </Select>
                  </FormItem>
                </Form>
                <div class='btn-container'>
                  <Button theme='primary' onClick={filterOrders}>
                    <Search />
                    查询
                  </Button>
                  <Button onClick={() => clearFilter()}>重置</Button>
                </div>
              </div>
              <CommonTable />
            </div>
          ),
        }}
      </Sideslider>
    );
  },
});
