import { defineComponent, ref, onMounted } from 'vue';
import './index.scss';
import http from '@/http';
import useFormModel from '@/hooks/useFormModel';
import { Button, Form, Message } from 'bkui-vue';
import apiService from '@/api/scrApi';
import CommonLocalTable from '@/components/CommonLocalTable';
import { removeEmptyFields } from '@/utils/scr/remove-query-fields';
import AreaSelector from '@/views/ziyanScr/hostApplication/components/AreaSelector';
import ZoneSelector from '@/views/ziyanScr/hostApplication/components/ZoneSelector';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
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
    const Modelform = ref({
      resource_type: props.formModelData.resource_type,
      spec: {
        region: [props.formModelData.spec.region],
        zone: [props.formModelData.spec.zone],
        device_type: [props.formModelData.spec.device_type],
      },
    });
    const tableColumns = ref([
      {
        type: 'selection',
        width: 30,
        minWidth: 30,
        align: 'center',
      },
      {
        label: '机型',
        field: 'device_type',
      },
      {
        label: '地域',
        field: 'region',
      },
      {
        label: '园区',
        field: 'zone',
      },
      {
        label: '数量',
        field: 'amount',
      },
      {
        label: '匹配数量',
        width: 250,
        render: ({ row }: any) => {
          return <bk-input size='mini' type='number' min={1} max={500} v-model={row.replicas}></bk-input>;
        },
      },
    ]);
    const domainList = ref([]);
    const { formModel, forceClear } = useFormModel({ ...Modelform.value });
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
    const getDomainList = () => {
      return http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/pool/findmany/recall/match/device`,
        removeEmptyFields({
          resource_type: formModel.resource_type,
          spec: {
            device_type: formModel.spec.device_type,
            region: formModel.spec.region,
            zone: formModel.spec.zone,
          },
        }),
      );
    };
    const onRegionChange = () => {
      formModel.spec.zone = [];
      loadDeviceTypes();
    };
    const onResourceTypeChange = () => {
      formModel.spec.region = [];
      formModel.spec.zone = [];
      loadDeviceTypes();
    };
    const onZoneChange = () => {
      formModel.spec.device_type = [];
      loadDeviceTypes();
    };
    const loadDeviceTypes = async () => {
      if (formModel.resource_type === 'QCLOUDCVM') {
        const { info } = await apiService.getDeviceTypes(formModel.spec);
        device_types.value = info || [];
      } else {
        const { info } = await apiService.getIDCPMDeviceTypes();
        device_types.value = info.map((item) => {
          return item.device_type;
        });
      }
    };
    onMounted(() => {
      loadDeviceTypes();
    });
    const isLoading = ref(false);
    const getListData = async () => {
      isLoading.value = true;
      try {
        const { data } = await getDomainList();
        domainList.value = data.info || [];
      } finally {
        isLoading.value = false;
      }
    };
    const loadResource = async () => {
      getListData();
    };
    const submitSelectedDevices = async () => {
      const {
        suborder_id,
        spec: { image_id, os_type },
      } = props.formModelData;
      const spec = selections.value.map((device) => {
        const { device_type, region, zone, replicas } = device;
        return {
          device_type,
          region,
          zone,
          replicas,
          image_id,
          os_type,
        };
      });
      await apiService.matchPools({
        suborder_id,
        spec,
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
          </Form>
        </div>
        <Button theme={'primary'} onClick={loadResource} class={'ml24 mr8'} loading={isLoading.value}>
          查询
        </Button>
        <Button
          onClick={() => {
            forceClear();
            getListData();
          }}>
          清空
        </Button>
        <Button theme='success' disabled={selections.value.length === 0} onClick={submitSelectedDevices} class={'ml24'}>
          手工匹配资源
        </Button>
        <div class={'table-container'}>
          {/* <CommonTable /> */}
          <CommonLocalTable
            loading={isLoading.value}
            hasSearch={false}
            tableOptions={{
              rowKey: 'domain',
              columns: tableColumns.value,
              extra: {
                onSelect: (selections: any) => {
                  handleSelectionChange(selections, () => true, false);
                },
                onSelectAll: (selections: any) => {
                  handleSelectionChange(selections, () => true, true);
                },
              },
            }}
            tableData={domainList.value}
          />
        </div>
      </div>
    );
  },
});
