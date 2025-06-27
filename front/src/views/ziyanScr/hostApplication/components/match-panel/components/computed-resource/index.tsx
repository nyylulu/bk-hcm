import { defineComponent, ref, computed, reactive, watchEffect } from 'vue';
import './index.scss';
import http from '@/http';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { removeEmptyFields } from '@/utils/scr/remove-query-fields';

import { Button, Form, Message } from 'bkui-vue';
import CommonLocalTable from '@/components/CommonLocalTable';
import QcloudRegionSelector from '@/views/ziyanScr/components/qcloud-resource/region-selector.vue';
import QcloudZoneSelector from '@/views/ziyanScr/components/qcloud-resource/zone-selector.vue';
import DevicetypeSelector from '@/views/ziyanScr/components/devicetype-selector/index.vue';
import QcloudZoneValue from '@/views/ziyanScr/components/qcloud-resource/zone-value.vue';
import QcloudRegionValue from '@/views/ziyanScr/components/qcloud-resource/region-value.vue';

const { FormItem } = Form;

export default defineComponent({
  props: {
    formModelData: {
      type: Object,
    },
    handleClose: Function,
  },
  setup(props) {
    const { getBusinessApiPath } = useWhereAmI();
    const { selections, handleSelectionChange } = useSelection();

    const options = ref([
      { value: 'IDCPM', label: 'IDC_物理机' },
      { value: 'QCLOUDCVM', label: '腾讯云_CVM' },
    ]);

    const formModel = reactive({
      resource_type: '',
      spec: { bk_cloud_regions: [], bk_cloud_zones: [], device_type: [] },
    });
    watchEffect(() => {
      const { resource_type: resourceType, spec } = props.formModelData ?? {};
      const { region, zone, device_type: deviceType } = spec;
      Object.assign(formModel, {
        resource_type: resourceType,
        spec: {
          bk_cloud_regions: [region],
          bk_cloud_zones: [zone],
          device_type: [deviceType],
        },
      });
    });

    const tableColumns = ref([
      { type: 'selection', width: 30, minWidth: 30, isDefaultShow: true },
      { label: '机型', field: 'device_type' },
      {
        label: '地域',
        field: 'bk_cloud_region',
        render: ({ data }: any) => <QcloudRegionValue value={data.bk_cloud_region} />,
      },
      {
        label: '园区',
        field: 'bk_cloud_zone',
        render: ({ data }: any) => <QcloudZoneValue value={data.bk_cloud_zone} />,
      },
      { label: '数量', field: 'amount' },
      {
        label: '匹配数量',
        width: 250,
        render: ({ row }: any) => {
          return <bk-input size='mini' type='number' min={1} max={500} v-model={row.replicas}></bk-input>;
        },
      },
    ]);
    const deviceList = ref([]);
    const isLoading = ref(false);
    const getDeviceList = async () => {
      isLoading.value = true;
      try {
        const res = await http.post(
          `/api/v1/woa/${getBusinessApiPath()}pool/findmany/recall/match/device`,
          removeEmptyFields({ ...formModel }),
        );
        deviceList.value = res.data?.info || [];
      } finally {
        isLoading.value = false;
      }
    };
    const handleReset = () => {
      Object.assign(formModel, {
        resource_type: props.formModelData.resource_type,
        spec: { bk_cloud_regions: [], bk_cloud_zones: [], device_type: [] },
      });

      getDeviceList();
    };

    const onRegionChange = () => {
      formModel.spec.bk_cloud_zones = [];
    };
    const onResourceTypeChange = () => {
      Object.assign(formModel, { spec: { bk_cloud_regions: [], bk_cloud_zones: [], device_type: [] } });
    };
    const onZoneChange = () => {
      formModel.spec.device_type = [];
    };

    const cvmDevicetypeParams = computed(() => {
      const { bk_cloud_regions: region, bk_cloud_zones: zone } = formModel.spec || {};
      return { region, zone };
    });

    const submitSelectedDevices = async () => {
      const {
        suborder_id,
        spec: { image_id, os_type },
      } = props.formModelData;
      const spec = selections.value.map((device) => {
        const { device_type, bk_cloud_region, bk_cloud_zone, replicas } = device;
        return {
          device_type,
          bk_cloud_region,
          bk_cloud_zone,
          replicas,
          image_id,
          os_type,
        };
      });
      await http.post(
        `/api/v1/woa/${getBusinessApiPath()}task/commit/apply/pool/match`,
        { suborder_id, spec },
        { removeEmptyFields: true },
      );
      Message({ message: '匹配成功', theme: 'success' });
      props.handleClose();
    };
    return () => (
      <div class={'apply-list-container'}>
        <div class={'filter-container'}>
          <Form model={formModel} class={'scr-form-wrapper'}>
            <FormItem label='资源类型' property='resource_type'>
              <bk-select v-model={formModel.resource_type} onChange={onResourceTypeChange}>
                {options.value.map((opt) => (
                  <bk-option key={opt.value} value={opt.value} label={opt.label} />
                ))}
              </bk-select>
            </FormItem>
            <FormItem label='地域' property='bk_cloud_regions'>
              <QcloudRegionSelector v-model={formModel.spec.bk_cloud_regions} onChange={onRegionChange} />
            </FormItem>
            <FormItem label='园区' property='bk_cloud_zones'>
              <QcloudZoneSelector
                v-model={formModel.spec.bk_cloud_zones}
                region={formModel.spec.bk_cloud_regions}
                onChange={onZoneChange}
              />
            </FormItem>
            <FormItem label='机型' property='device_type'>
              <DevicetypeSelector
                class='tbkselect'
                v-model={formModel.spec.device_type}
                resourceType={formModel.resource_type === 'QCLOUDCVM' ? 'cvm' : 'idcpm'}
                params={cvmDevicetypeParams.value}
                multiple
              />
            </FormItem>
          </Form>
        </div>
        <Button theme={'primary'} class={'ml24 mr8'} loading={isLoading.value} onClick={getDeviceList}>
          查询
        </Button>
        <Button disabled={isLoading.value} onClick={handleReset}>
          重置
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
            tableData={deviceList.value}
          />
        </div>
      </div>
    );
  },
});
