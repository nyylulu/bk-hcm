import { defineComponent, ref, onMounted } from 'vue';
import { useTable } from '@/hooks/useTable/useTable';
import { Search } from 'bkui-vue/lib/icon';
import { Dialog, Form } from 'bkui-vue';
import apiService from '@/api/scrApi';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import AreaSelector from '../hostApplication/components/AreaSelector';
import ZoneSelector from '../hostApplication/components/ZoneSelector';
import CreateDevice from './CreateDevice/index';
import './index.scss';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
const { FormItem } = Form;
export default defineComponent({
  name: 'AllhostInventoryManager',
  setup() {
    const { columns } = useColumns('cvmModel');
    const { selections, handleSelectionChange } = useSelection();
    const deviceGroups = ['标准型', '高IO型', '大数据型', '计算型'];
    const enableCapacitys = [
      { label: '是', value: true },
      { label: '否', value: false },
    ];
    const enableApplys = [
      { label: '是', value: true },
      { label: '否', value: false },
    ];
    const filter = ref({
      require_type: 1,
      region: [],
      zone: [],
      device_type: [],
      device_group: deviceGroups && [deviceGroups[0]],
      cpu: '',
      mem: '',
      disk: '',
      enableCapacity: '',
      enableApply: '',
    });
    const options = ref({
      require_types: [],
      device_groups: deviceGroups,
      device_types: [],
      regions: [],
      zones: [],
      cpu: [],
      mem: [],
      enableCapacitys,
      enableApplys,
    });
    const deviceConfigDisabled = ref(false);
    const deviceTypeDisabled = ref(false);
    const batchEditDialogVisible = ref(false);
    const createVisible = ref(false);
    const batchEditForm = ref({
      comment: '',
      enableCapacity: 0,
      enableApply: 0,
    });
    const whetherlist = ref([
      {
        value: 0,
        label: '保持不变',
      },
      {
        value: true,
        label: '是',
      },
      {
        value: false,
        label: '否',
      },
    ]);
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
        filter.value.enableCapacity && {
          field: 'enable_capacity',
          operator: 'equal',
          value: filter.value.enableCapacity,
        },
        filter.value.enableApply && { field: 'enable_apply', operator: 'equal', value: filter.value.enableApply },
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
      filter.value = {
        require_type: 1,
        region: [],
        zone: [],
        device_type: [],
        device_group: deviceGroups && [deviceGroups[0]],
        cpu: '',
        mem: '',
        disk: '',
        enableCapacity: '' as any,
        enableApply: '' as any,
      };
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
    const batchUpdates = () => {
      batchEditDialogVisible.value = true;
    };
    const createNewModel = () => {
      createVisible.value = true;
    };
    const triggerShow = (val: boolean) => {
      batchEditDialogVisible.value = val;
      batchEditForm.value = {
        comment: '',
        enableCapacity: 0,
        enableApply: 0,
      };
    };
    const handleConfirm = () => {
      const properties = serializeBatchEditForm();
      const ids = selections.value.map((row) => row.id);
      apiService.updateCvmDeviceTypeConfigs({
        ids,
        properties,
      });

      batchEditDialogVisible.value = false;
      selections.value = [];
      batchEditForm.value = {
        comment: '',
        enableCapacity: 0,
        enableApply: 0,
      };
      setTimeout(() => {
        getListData();
      }, 1000);
    };
    const serializeBatchEditForm = () => {
      const { comment, enableApply, enableCapacity } = batchEditForm.value;
      return {
        comment: comment.trim() !== '' ? comment : undefined,
        enable_apply: enableApply !== '' && enableApply !== 0 ? enableApply : undefined,
        enable_capacity: enableCapacity !== '' && enableCapacity !== 0 ? enableCapacity : undefined,
      };
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
        filter.value.enableCapacity !== '' && {
          field: 'enable_capacity',
          operator: 'equal',
          value: filter.value.enableCapacity,
        },
        filter.value.enableApply !== '' && {
          field: 'enable_apply',
          operator: 'equal',
          value: filter.value.enableApply,
        },
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
    const createHandleConfirm = async () => {
      await createRef.value.handleConfirm();
      createVisible.value = false;
    };
    onMounted(() => {
      loadRestrict();
      loadDeviceTypes();
      getfetchOptionslist();
    });

    const { CommonTable, getListData } = useTable({
      tableOptions: {
        columns,
        extra: {
          onSelect: (selections: any) => {
            handleSelectionChange(selections, () => true, false);
          },
          onSelectAll: (selections: any) => {
            handleSelectionChange(selections, () => true, true);
          },
        },
      },
      requestOption: {
        dataPath: 'data.info',
      },
      scrConfig: () => {
        return {
          url: '/api/v1/woa/config/findmany/config/cvm/device',
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
    const createRef = ref();
    const createTriggerShow = () => {
      createVisible.value = false;
      createRef.value.clearValidate();
    };
    return () => (
      <div class={'apply-list-container cvm-web-wrapper'}>
        <div class={'filter-container'}>
          <Form model={filter.value} formType='vertical' class={'scr-form-wrapper'}>
            <FormItem label='需求类型'>
              <bk-select v-model={filter.value.require_type}>
                {options.value.require_types.map((item) => (
                  <bk-option key={item.require_type} value={item.require_type} label={item.require_name}></bk-option>
                ))}
              </bk-select>
            </FormItem>
            <FormItem label='地域'>
              <AreaSelector
                ref='areaSelector'
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
                separateCampus={false}
                multiple
                params={{
                  resourceType: 'QCLOUDCVM',
                  region: filter.value.region,
                }}></ZoneSelector>
            </FormItem>
            <FormItem label='实例族'>
              <bk-select
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
            <FormItem label='可查询容量'>
              <bk-select
                v-model={filter.value.enableCapacity}
                clearable
                disabled={deviceConfigDisabled.value}
                filterable
                onChange={handleDeviceConfigChange}>
                {options.value.enableCapacitys.map((item) => (
                  <bk-option key={item.value} value={item.value} label={item.label}></bk-option>
                ))}
              </bk-select>
            </FormItem>
            <FormItem label='可申请'>
              <bk-select
                v-model={filter.value.enableApply}
                clearable
                disabled={deviceConfigDisabled.value}
                filterable
                onChange={handleDeviceConfigChange}>
                {options.value.enableApplys.map((item) => (
                  <bk-option key={item.value} value={item.value} label={item.label}></bk-option>
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
              重置
            </bk-button>
          </div>
        </div>
        <div class='btn-container oper-btn-pad'>
          <bk-button icon='bk-icon-refresh' disabled={!selections.value.length} onClick={batchUpdates}>
            批量更新
          </bk-button>
          <bk-button icon='bk-icon-refresh' onClick={createNewModel}>
            创建新机型
          </bk-button>
        </div>
        <CommonTable class={'filter-common-table'} />
        <Dialog
          class='common-dialog'
          close-icon={false}
          isShow={batchEditDialogVisible.value}
          title='批量更新'
          width={600}
          onConfirm={handleConfirm}
          onClosed={() => triggerShow(false)}>
          <bk-form v-loading='$isLoading(deviceTypeConfigs.updateRequestId)'>
            <bk-form-item label='可查询容量'>
              <bk-select v-model={batchEditForm.value.enableCapacity} style='width: 250px'>
                {whetherlist.value.map(({ label, value }) => {
                  return <bk-option key={value} label={label} value={value}></bk-option>;
                })}
              </bk-select>
            </bk-form-item>
            <bk-form-item label='可申请'>
              <bk-select v-model={batchEditForm.value.enableApply} style='width: 250px'>
                {whetherlist.value.map(({ label, value }) => {
                  return <bk-option key={value} label={label} value={value}></bk-option>;
                })}
              </bk-select>
            </bk-form-item>
            <bk-form-item label='备注'>
              <bk-input
                v-model={batchEditForm.value.comment}
                style='width: 250px'
                autosize
                type='textarea'
                maxlength={128}
              />
            </bk-form-item>
          </bk-form>
        </Dialog>
        <Dialog
          class='common-dialog'
          close-icon={false}
          isShow={createVisible.value}
          title='创建新机型'
          width={600}
          onConfirm={createHandleConfirm}
          onClosed={() => createTriggerShow()}>
          <CreateDevice onQueryList={loadResources} ref={createRef} />
        </Dialog>
      </div>
    );
  },
});
