import { defineComponent, ref, onMounted, computed } from 'vue';
import './index.scss';
import useFormModel from '@/hooks/useFormModel';
import { Button, Form, Input, Message } from 'bkui-vue';
import apiService from '@/api/scrApi';
import { useTable } from '@/hooks/useTable/useTable';
import { removeEmptyFields } from '@/utils/scr/remove-query-fields';
import AreaSelector from '@/views/ziyanScr/hostApplication/components/AreaSelector';
import ZoneSelector from '@/views/ziyanScr/hostApplication/components/ZoneSelector';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import DiskTypeSelect from '../../../DiskTypeSelect';
import http from '@/http';
import { useUserStore } from '@/store';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

const { FormItem } = Form;
export default defineComponent({
  props: {
    formModelData: {
      type: Object,
    },
    handleClose: Function,
  },
  setup(props) {
    const { selections, handleSelectionChange } = useSelection();
    const userStore = useUserStore();
    const { formModel, resetForm } = useFormModel({
      resource_type: props.formModelData.resource_type,
      ips: '',
      spec: {
        region: [props.formModelData.spec.region],
        zone: [props.formModelData.spec.zone],
        device_type: [props.formModelData.spec.device_type],
        disk_type: [props.formModelData.spec.disk_type],
      },
    });
    const options = ref([
      {
        value: 'IDCPM',
        label: 'IDC_物理机',
      },
      {
        value: 'QCLOUDCVM',
        label: '腾讯云_CVM',
      },
    ]);
    const device_types = ref([]);
    const { CommonTable, getListData, isLoading } = useTable({
      tableOptions: {
        columns: [
          {
            type: 'selection',
            width: 32,
          },
          {
            field: 'match_tag',
            label: '星标',
            render: ({ data }: any) => {
              if (data.matchTag) {
                return <i class='hcm-icon bkhcm-icon-collect' color='gold'></i>;
              }
              return <span>-</span>;
            },
          },
          {
            field: 'asset_id',
            label: '固资号',
            fixed: true,
          },
          {
            field: 'ip',
            label: '内网 IP',
          },
          {
            field: 'device_type',
            label: '机型',
          },
          {
            field: 'outer_ip',
            label: '外网IP',
          },
          {
            field: 'isp',
            label: '外网运营商',
          },
          {
            field: 'os_type',
            label: '操作系统',
          },
          {
            field: 'equipment',
            label: '机架号',
          },
          {
            field: 'zone',
            label: '园区',
          },
          {
            field: 'module',
            label: '模块',
          },
          {
            field: 'hardMemo',
            label: '硬件描述',
            showOverflowTooltip: true,
          },
          {
            field: 'raid_type',
            label: 'RAID 类型',
          },
          {
            field: 'idc_unit',
            label: 'IDC 单元',
          },
          {
            field: 'idc_logic_area',
            label: '逻辑区域',
          },
          {
            field: 'input_time',
            label: '入库时间',
          },
          {
            prop: 'match_score',
            label: '匹配得分',
          },
        ],
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
      scrConfig: () => ({
        url: '/api/v1/woa/task/findmany/apply/match/device',
        payload: removeEmptyFields({
          resource_type: formModel.resource_type,
          ips: ipArray.value,
          spec: {
            device_type: formModel.spec.device_type,
            region: formModel.spec.region,
            zone: formModel.spec.zone,
            disk_type: formModel.spec.disk_type,
          },
        }),
      }),
    });
    const ipArray = computed(() => {
      const ipv4 = /^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$/;
      const ips = [];
      formModel.ips
        .split(/\r?\n/)
        .map((ip) => ip.trim())
        .filter((ip) => ip.length > 0)
        .forEach((item) => {
          if (ipv4.test(item)) {
            ips.push(item);
          }
        });
      return ips;
    });
    const onRegionChange = () => {
      formModel.spec.zone = [];
    };
    const onResourceTypeChange = () => {
      formModel.spec.region = [];
      formModel.spec.zone = [];
    };
    const onZoneChange = () => {
      formModel.spec.device_type = [];
    };
    const loadDeviceTypes = async () => {
      const { info } = await apiService.getDeviceTypes(formModel.spec);
      device_types.value = info || [];
    };
    onMounted(() => {
      loadDeviceTypes();
    });
    const loadResource = async () => {
      getListData();
    };
    const submitSelectedDevices = async () => {
      const { suborder_id } = props.formModelData;
      const device = selections.value.map((device) => {
        const { bk_host_id, asset_id, ip } = device;
        return {
          bk_host_id,
          asset_id,
          ip,
        };
      });
      await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/task/commit/apply/match`, {
        suborder_id,
        operator: userStore.username,
        device,
      });
      Message({
        message: '匹配成功',
        theme: 'success',
      });
      props.handleClose();
    };
    return () => (
      <div class={'apply-list-container'}>
        <div class={'filter-container'}>
          <Form model={formModel} class={'scr-form-wrapper'}>
            <FormItem label='资源类型'>
              <bk-select v-model={formModel.resource_type} onChange={onResourceTypeChange}>
                {options.value.map((opt) => (
                  <bk-option key={opt.value} value={opt.value} label={opt.label} />
                ))}
              </bk-select>
            </FormItem>
            <FormItem label='地域'>
              <AreaSelector
                ref='areaSelector'
                multiple
                v-model={formModel.spec.region}
                params={{ resourceType: formModel.resource_type }}
                onChange={onRegionChange}></AreaSelector>
            </FormItem>
            <FormItem label='园区'>
              <ZoneSelector
                ref='zoneSelector'
                multiple
                v-model={formModel.spec.zone}
                params={{
                  resourceType: formModel.resource_type,
                  region: formModel.spec.region,
                }}
                onChange={onZoneChange}
              />
            </FormItem>
            <FormItem label='机型'>
              <bk-select class='tbkselect' v-model={formModel.spec.device_type} clearable multiple>
                {device_types.value.map((item) => (
                  <bk-option key={item} value={item} label={item}></bk-option>
                ))}
              </bk-select>
            </FormItem>
            <FormItem label='数据盘类型'>
              <DiskTypeSelect v-model={formModel.spec.disk_type} />
            </FormItem>
            <FormItem label='内网IP'>
              <Input type='textarea' v-model={formModel.ips} />
            </FormItem>
          </Form>
        </div>
        <Button theme={'primary'} onClick={loadResource} class={'ml24 mr8'} loading={isLoading.value}>
          查询
        </Button>
        <Button
          onClick={() => {
            resetForm();
            getListData();
          }}>
          清空
        </Button>
        <Button theme='success' disabled={selections.value.length === 0} onClick={submitSelectedDevices} class={'ml24'}>
          手工匹配资源
        </Button>
        <div class={'table-container'}>
          <CommonTable />
        </div>
      </div>
    );
  },
});
