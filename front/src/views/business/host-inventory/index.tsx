import { defineComponent, ref, onMounted, computed, reactive } from 'vue';
import { useTable } from '@/hooks/useTable/useTable';
import apiService from '@/api/scrApi';
import { Button } from 'bkui-vue';
import AreaSelector from '@/views/ziyanScr/hostApplication/components/AreaSelector';
import ZoneSelector from '@/views/ziyanScr/hostApplication/components/ZoneSelector';
import { useI18n } from 'vue-i18n';
import cssModule from './index.module.scss';
import GridFilterComp from '@/components/grid-filter-comp';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import DevicetypeSelector from '@/views/ziyanScr/components/devicetype-selector/index.vue';
import { useConfigRequirementStore, type IRequirementObsProject } from '@/store/config/requirement';
import ResourceDemandsResult from './resource-demands-result.vue';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { ResourceDemandResultStatusCode } from '@/typings/resourcePlan';
import routerAction from '@/router/utils/action';
import { GLOBAL_BIZS_KEY } from '@/common/constant';

export default defineComponent({
  name: 'BusinessHostInventory',
  setup() {
    const { columns } = useColumns('hostInventor');
    const deviceGroups = ['标准型', '高IO型', '大数据型', '计算型'];
    const { t } = useI18n();
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

    const whereAmI = useWhereAmI();
    const configRequirementStore = useConfigRequirementStore();
    const requirementObsProjectMap = ref<IRequirementObsProject>({});

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
    const loadRestrict = async () => {
      const { cpu, mem } = await apiService.getRestrict();
      options.value.cpu = cpu || [];
      options.value.mem = mem || [];
    };
    const getOptions = async () => {
      const [{ info }, obsProjectMap] = await Promise.all([
        apiService.getRequireTypes(),
        configRequirementStore.getRequirementObsProject(),
      ]);
      options.value.require_types = info;
      requirementObsProjectMap.value = obsProjectMap;
    };
    const handleApply = (row: any) => {
      routerAction.open({
        path: '/business/service/service-apply/cvm',
        query: {
          ...row,
          from: 'businessCvmInventory',
          [GLOBAL_BIZS_KEY]: whereAmI.getBizsId(),
        },
      });
    };
    onMounted(() => {
      loadRestrict();
      getOptions();
    });

    const demandStatus = reactive<Record<number, { code: number; text: string }>>({});
    const updateResourceDemands = (planStatus: { code: number; text: string }, row: any) => {
      demandStatus[row.id] = planStatus;
    };

    const { CommonTable, getListData, isLoading } = useTable({
      tableOptions: {
        columns: [
          ...columns,
          {
            label: '预测情况',
            width: 150,
            render: ({ row }: { row: any }) => (
              <ResourceDemandsResult
                data={row}
                obsProjectMap={requirementObsProjectMap.value}
                bizId={whereAmI.getBizsId()}
                onUpdate={updateResourceDemands}
              />
            ),
          },
          {
            label: '操作',
            width: 180,
            showOverflowTooltip: false,
            render: ({ row }: { row: any }) => {
              return (
                <div class={cssModule['operation-button-group']}>
                  <Button
                    text
                    theme='primary'
                    disabled={
                      row.listenerNum > 0 ||
                      row.delete_protect ||
                      [
                        undefined,
                        ResourceDemandResultStatusCode.BGNone,
                        ResourceDemandResultStatusCode.BIZNone,
                      ].includes(demandStatus[row.id]?.code)
                    }
                    onClick={() => handleApply(row)}>
                    一键申请
                  </Button>
                  {/* 滚服禁用，增加提示 滚服由BG统一提交预测 */}
                  <Button
                    text
                    theme='primary'
                    disabled={row.listenerNum > 0 || row.delete_protect || row.require_type === 6}
                    v-bk-tooltips={{ content: '滚服由BG统一提交预测', disabled: row.require_type !== 6 }}
                    onClick={() => {
                      routerAction.open({
                        path: '/business/resource-plan/add',
                        query: {
                          [GLOBAL_BIZS_KEY]: whereAmI.getBizsId(),
                          action: 'add',
                          payload: encodeURIComponent(
                            JSON.stringify({
                              obs_project: requirementObsProjectMap.value[row.require_type],
                              region_id: row.region,
                              zone_id: row.zone,
                              cvm: {
                                device_type: row.device_type,
                              },
                            }),
                          ),
                        },
                      });
                    }}>
                    增加预测
                  </Button>
                </div>
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

    const cvmDevicetypeParams = computed(() => {
      const { region, zone, device_group, cpu, mem, disk, enable_capacity } = filter.value;
      return { region, zone, device_group, cpu, mem, disk, enable_capacity };
    });

    return () => (
      <div class={cssModule.page}>
        <GridFilterComp
          rules={[
            {
              // 小额绿通和春保资源池不显示
              title: t('需求类型'),
              content: (
                <bk-select v-model={filter.value.require_type}>
                  {options.value.require_types
                    .filter((item) => ![7, 8].includes(item.require_type))
                    .map((item) => (
                      <bk-option
                        key={item.require_type}
                        value={item.require_type}
                        label={item.require_name}></bk-option>
                    ))}
                </bk-select>
              ),
            },
            {
              title: t('地域'),
              content: (
                <AreaSelector
                  ref='areaSelector'
                  v-model={filter.value.region}
                  multiple
                  clearable
                  filterable
                  params={{ resourceType: 'QCLOUDCVM' }}></AreaSelector>
              ),
            },
            {
              title: t('园区'),
              content: (
                <ZoneSelector
                  ref='zoneSelector'
                  v-model={filter.value.zone}
                  separateCampus={false}
                  multiple
                  params={{
                    resourceType: 'QCLOUDCVM',
                    region: filter.value.region,
                  }}></ZoneSelector>
              ),
            },
            {
              title: t('实例族'),
              content: (
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
              ),
            },
            {
              title: t('机型'),
              content: (
                <DevicetypeSelector
                  v-model={filter.value.device_type}
                  resourceType='cvm'
                  params={cvmDevicetypeParams.value}
                  multiple
                  disabled={deviceTypeDisabled.value}
                  onChange={handleDeviceTypeChange}
                />
              ),
            },
            {
              title: t('CPU(核)'),
              content: (
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
              ),
            },
            {
              title: t('内存(G)'),
              content: (
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
              ),
            },
          ]}
          onSearch={filterDevices}
          onReset={clearFilter}
          loading={isLoading.value}
          col={5}
          class={cssModule.filter}
        />
        <section class={cssModule.table}>
          <CommonTable style={{ height: 'calc(100% - 48px)' }} />
        </section>
      </div>
    );
  },
});
