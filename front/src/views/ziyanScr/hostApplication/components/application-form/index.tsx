import { defineComponent, onMounted, ref, watch, nextTick, computed, reactive, useTemplateRef } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import './index.scss';
import { Input, Button, Sideslider, Message, Dropdown, Radio, Form, Alert, Tag } from 'bkui-vue';
import { Plus } from 'bkui-vue/lib/icon';
import CommonCard from '@/components/CommonCard';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import BusinessSelector from '@/components/business-selector/index.vue';
import AreaSelector from '../AreaSelector';
import ZoneTagSelector from '@/components/zone-tag-selector/index.vue';
import DiskTypeSelect from '../DiskTypeSelect';
import NetworkInfoCollapsePanel from '../network-info-collapse-panel/index.vue';
import AntiAffinityLevelSelect from '../AntiAffinityLevelSelect';
import DevicetypeSelector from '@/views/ziyanScr/components/devicetype-selector/index.vue';
import applicationSideslider from '../application-sideslider/index.vue';
import WName from '@/components/w-name';
import HostApplyTipsAlert from './host-apply-tips-alert/index.vue';
import HostApplySpringPoolTips from './spring-pool/tips.vue';
import CvmMaxCapacity from '@/views/ziyanScr/components/cvm-max-capacity/index.vue';
import ReqTypeValue from '@/components/display-value/req-type-value.vue';
import { MENU_SERVICE_HOST_APPLICATION } from '@/constants/menu-symbol';
import { useAccountStore, useUserStore } from '@/store';
import usePlanStore from '@/store/usePlanStore';
import useCvmChargeType from '@/views/ziyanScr/hooks/use-cvm-charge-type';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import useFormModel from '@/hooks/useFormModel';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import apiService from '@/api/scrApi';
import { VendorEnum, CLOUD_CVM_DISKTYPE, GLOBAL_BIZS_KEY } from '@/common/constant';
import { VerifyStatus, VerifyStatusMap } from './constants';
import { ChargeType } from '@/typings/plan';
import { cloneDeep } from 'lodash';
import { timeFormatter, expectedDeliveryTime } from '@/common/util';
import http from '@/http';

// 滚服项目
import RollingServerTipsAlert from '@/views/ziyanScr/rolling-server/tips-alert/index.vue';
import InheritPackageFormItem, {
  type RollingServerHost,
} from '@/views/ziyanScr/rolling-server/inherit-package-form-item/index.vue';
import RollingServerCpuCoreLimits from '@/views/ziyanScr/rolling-server/cpu-core-limits/index.vue';
import { DeviceType, CvmDeviceType } from '@/views/ziyanScr/components/devicetype-selector/types';
// 小额绿通
import GreenChannelTipsAlert from './green-channel/tips-alert.vue';
import GreenChannelCpuCoreLimits from './green-channel/cpu-core-limits.vue';
// 预测
import usePlanDeviceType from '@/views/ziyanScr/hostApplication/plan/usePlanDeviceType';
import PlanLinkAlert from '../../plan/plan-link-alert.vue';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;
const { DropdownMenu, DropdownItem } = Dropdown;
const { Group: RadioGroup, Button: RadioButton } = Radio;
export default defineComponent({
  components: {
    applicationSideslider,
  },
  props: {
    isbusiness: {
      type: Boolean,
      default: false,
    },
  },
  setup(props) {
    const accountStore = useAccountStore();
    const IDCPMformRef = ref();
    const QCLOUDCVMformRef = ref();

    const networkInfoFormRef = ref();
    const networkInfoPanelRef = useTemplateRef<typeof NetworkInfoCollapsePanel>('network-info-panel');

    const router = useRouter();
    const route = useRoute();
    const addResourceRequirements = ref(false);
    const isLoading = ref(false);
    const title = ref('增加资源需求');
    const CVMapplication = ref(false);
    const { getBizsId, whereAmI } = useWhereAmI();
    const planStore = usePlanStore();
    const isNeedVerfiy = ref(false);
    const isVerifyFailed = ref(false);
    const {
      formModel: cpuAmount,
      setFormValues: setCpuAmount,
      resetForm: resetCpuAmount,
    } = useFormModel({
      prepaid: 0,
      postpaid: 0,
    });
    const order = ref({
      loading: false,
      submitting: false,
      saving: false,
      model: {
        bkBizId: undefined as number,
        bkUsername: '',
        requireType: 1,
        enableNotice: false,
        expectTime: expectedDeliveryTime(),
        remark: '',
        follower: [] as any,
        suborders: [] as any,
      },
      rules: {
        bkBizId: [{ required: true, message: '请选择业务', trigger: 'change' }],
        requireType: [{ required: true, message: '请选择需求类型', trigger: 'change' }],
        expectTime: [{ required: true, message: '请填写期望交付时间', trigger: 'change' }],
      },
    });

    const deviceTypeSelectorRef = useTemplateRef<typeof DevicetypeSelector>('device-type-selector');
    const selectedChargeType = computed(() => resourceForm.value.charge_type);
    const {
      isPlanedDeviceTypeLoading,
      availableDeviceTypeSet,
      computedAvailableDeviceTypeSet,
      hasPlanedDeviceType,
      getPlanedDeviceType,
    } = usePlanDeviceType(deviceTypeSelectorRef, selectedChargeType);

    // 机型排序
    const deviceTypeCompareFn = (a: DeviceType, b: DeviceType) => {
      // 非滚服、非小额绿通，走预测
      if (!isSpecialRequirement.value) {
        const set = computedAvailableDeviceTypeSet.value;
        return Number(set.has(b.device_type)) - Number(set.has(a.device_type));
      }
      // 滚服、小额绿通
      const aDeviceTypeClass = (a as CvmDeviceType).device_type_class;
      const bDeviceTypeClass = (b as CvmDeviceType).device_type_class;
      if (aDeviceTypeClass === 'CommonType' && bDeviceTypeClass === 'SpecialType') return -1;
      if (aDeviceTypeClass === 'SpecialType' && bDeviceTypeClass === 'CommonType') return 1;
      return 0;
    };
    // 机型选项禁用
    const deviceTypeOptionDisabledCallback = (option: DeviceType) => {
      // 非滚服、非小额绿通
      if (!isSpecialRequirement.value) {
        return !computedAvailableDeviceTypeSet.value.has(option.device_type);
      }
      // 滚服、小额绿通
      const { device_type_class, device_group } = option as CvmDeviceType;
      return (
        'SpecialType' === device_type_class ||
        (isRollingServer.value && device_group !== rollingServerHost?.device_group)
      );
    };
    const deviceTypeOptionDisabledTipsCallback = (option: DeviceType) => {
      // 非滚服、非小额绿通
      if (!isSpecialRequirement.value) return '当前机型不在有效预测范围内';
      // 滚服、小额绿通
      const { device_type_class, device_group } = option as CvmDeviceType;
      if (device_type_class === 'SpecialType') return '专用机型不允许选择';
      if (isRollingServer.value && device_group !== rollingServerHost?.device_group) return '机型族不匹配';
    };

    const formRef = ref();
    const IDCPMIndex = ref(-1);
    const QCLOUDCVMIndex = ref(-1);
    const resourceFormRef = ref();
    const dropdownMenuShowState = reactive({
      idc: false,
      cvm: false,
    });
    const { columns: CloudHostcolumns, generateColumnsSettings } = useColumns('CloudHost');
    let cloudHostSetting = generateColumnsSettings(CloudHostcolumns);
    const { columns: PhysicalMachinecolumns } = useColumns('PhysicalMachine');
    const { cvmChargeTypes, cvmChargeTypeNames, cvmChargeTypeTips, getMonthName } = useCvmChargeType();
    const cloudTableColumns = ref([]);

    // 特殊需求类型（滚服项目、小额绿通）-状态
    const isRollingServer = computed(() => order.value.model.requireType === 6);
    const isGreenChannel = computed(() => order.value.model.requireType === 7);
    const isSpringPool = computed(() => order.value.model.requireType === 8);
    const isRollingServerLike = computed(() => isRollingServer.value || isSpringPool.value);
    const isSpecialRequirement = computed(() => isRollingServer.value || isGreenChannel.value);

    // 选择的cvm机型
    const selectedCvmDeviceType = ref<CvmDeviceType>(null);
    // 需求类型-常规项目
    const isNormalRequireType = computed(() => order.value.model.requireType === 1);

    // 是否为默认包4年
    const isDefaultFourYears = computed(
      () =>
        isNormalRequireType.value &&
        selectedCvmDeviceType.value?.device_type_class === 'SpecialType' &&
        resourceForm.value.charge_type === cvmChargeTypes.PREPAID,
    );

    // 购买时长的禁用状态
    const chargeMonthDisabledState = computed(() => {
      if (isRollingServer.value || isDefaultFourYears.value) {
        return {
          disabled: true,
          tips: isRollingServer.value
            ? '继承原有套餐包年包月时长，此处的购买时长为剩余时长'
            : '专用机型只能选择4年套餐',
        };
      }
      return {
        disabled: false,
      };
    });

    // 滚服继承套餐的机器
    let rollingServerHost: RollingServerHost = null;

    const PhysicalMachineoperation = ref({
      label: '操作',
      width: 100,
      render: ({ row, index }: any) => {
        return (
          <div class='operation-column'>
            <Button text theme='primary' class='mr10' onClick={() => clonelist(row, 'IDCPM')}>
              克隆
            </Button>
            <Dropdown
              trigger='manual'
              isShow={IDCPMIndex.value === index && dropdownMenuShowState.idc}
              popoverOptions={{
                renderType: 'shown',
                onAfterHidden: () => {
                  IDCPMIndex.value = -1;
                  dropdownMenuShowState.idc = false;
                },
                forceClickoutside: true,
              }}>
              {{
                default: () => (
                  <div
                    class={`more-action${IDCPMIndex.value === index ? ' current-operate-row' : ''}`}
                    onClick={() => {
                      IDCPMIndex.value = index;
                      dropdownMenuShowState.idc = true;
                    }}>
                    <i class='hcm-icon bkhcm-icon-more-fill' />
                  </div>
                ),
                content: () => (
                  <DropdownMenu>
                    <DropdownItem
                      key='retry'
                      onClick={() => {
                        modifylist(row, index, 'IDCPM');
                        dropdownMenuShowState.idc = false;
                      }}>
                      修改
                    </DropdownItem>
                    <DropdownItem
                      key='stop'
                      onClick={() => {
                        deletelist(index, 'IDCPM');
                        dropdownMenuShowState.idc = false;
                      }}>
                      删除
                    </DropdownItem>
                  </DropdownMenu>
                ),
              }}
            </Dropdown>
          </div>
        );
      },
    });
    const CloudHostoperation = ref({
      label: '操作',
      fixed: 'right',
      width: 120,
      render: ({ row, index }: any) => {
        return (
          <div class='operation-column'>
            <Button text theme='primary' class='mr10' onClick={() => clonelist(row, 'QCLOUDCVM')}>
              克隆
            </Button>
            <Dropdown
              trigger='manual'
              isShow={QCLOUDCVMIndex.value === index && dropdownMenuShowState.cvm}
              popoverOptions={{
                renderType: 'shown',
                onAfterHidden: () => {
                  QCLOUDCVMIndex.value = -1;
                  dropdownMenuShowState.cvm = false;
                },
                forceClickoutside: true,
              }}>
              {{
                default: () => (
                  <div
                    class={`more-action${QCLOUDCVMIndex.value === index ? ' current-operate-row' : ''}`}
                    onClick={() => {
                      QCLOUDCVMIndex.value = index;
                      dropdownMenuShowState.cvm = true;
                    }}>
                    <i class='hcm-icon bkhcm-icon-more-fill' />
                  </div>
                ),
                content: () => (
                  <DropdownMenu>
                    <DropdownItem
                      key='retry'
                      onClick={() => {
                        modifylist(row, index, 'QCLOUDCVM');
                        dropdownMenuShowState.cvm = false;
                      }}>
                      修改
                    </DropdownItem>
                    <DropdownItem
                      key='stop'
                      onClick={() => {
                        deletelist(index, 'QCLOUDCVM');
                        dropdownMenuShowState.cvm = false;
                      }}>
                      删除
                    </DropdownItem>
                  </DropdownMenu>
                ),
              }}
            </Dropdown>
          </div>
        );
      },
    });
    const CVMVerifyColumns = [
      {
        field: 'verify_result',
        label: '预检状态',
        minWidth: 90,
        isDefaultShow: true,
        render({ cell }: { cell: VerifyStatus }) {
          return <span class={`status-${cell}`}>{VerifyStatusMap[cell] || '待校验'}</span>;
        },
        isHidden: isSpecialRequirement.value,
      },
      {
        field: 'reason',
        label: '预检信息',
        minWidth: 150,
        isDefaultShow: true,
        render({ cell }: { cell: string }) {
          return <span v-bk-tooltips={{ content: cell, disabled: !cell?.length }}>{cell || '--'}</span>;
        },
        isHidden: isSpecialRequirement.value,
        showOverflowTooltip: false,
      },
    ];
    // 添加按钮侧边栏公共表单对象
    const resourceForm = ref({
      resourceType: 'QCLOUDCVM', // 主机类型
      remark: '', // 备注
      enable_disk_check: false,
      region: '', // 地域
      zone: '', // 园区
      charge_type: cvmChargeTypes.PREPAID,
      charge_months: 36, // 计费时长
      bk_asset_id: '', // 继承套餐的机器代表固资号
    });
    // 侧边栏腾讯云CVM
    const QCLOUDCVMForm = ref({
      spec: {
        device_type: '', // 机型
        replicas: 1, // 需求数量
        anti_affinity_level: 'ANTI_NONE',
        vpc: '', //  vpc
        subnet: '', //  子网
        image_id: '', // 镜像
        disk_type: 'CLOUD_PREMIUM', // 数据盘tyle
        disk_size: 0, // 数据盘size
        network_type: 'TENTHOUSAND',
        inherit_instance_id: '', // 继承套餐的机器代表实例ID
        cpu: undefined,
      },
    });
    // 主机类型列表
    const resourceTypes = ref([
      {
        value: 'IDCPM',
        label: 'IDC_物理机',
      },

      {
        value: 'QCLOUDCVM',
        label: '腾讯云_CVM',
      },
    ]);

    // 镜像列表
    const images = ref([]);

    // 侧边栏物理机CVM
    const pmForm = ref({
      spec: {
        device_type: '', // 机型
        raid_type: '', // RAID 类型
        os_type: '', // 操作系统
        isp: '', // 经营商
        anti_affinity_level: '',
        network_type: 'TENTHOUSAND',
        replicas: 1, // 需求数量
      },
      rules: {
        device_type: [{ required: true, message: '请选择机型', trigger: 'change' }],
        region: [{ required: true, message: '请选择地域', trigger: 'change' }],
        replicas: [{ required: true, message: '请输入需求数量', trigger: 'blur' }],
        os_type: [{ required: true, message: '请选择操作系统', trigger: 'change' }],
        anti_affinity_level: [{ required: true, message: '请选择反亲和性', trigger: 'change' }],
      },
      options: {
        deviceTypes: [],
        osTypes: [],
        raidType: [],
        regions: [],
        zones: [],
        isps: [],
      },
    });
    // 云组件table
    const cloudTableData = ref([]);
    // 物理机table
    const physicalTableData = ref([]);

    // 云地域变更时, 清空zone, vpc, subnet, device_type
    const handleRegionChange = () => {
      resourceForm.value.zone = '';
      handleZoneChange();

      if (resourceForm.value.resourceType === 'QCLOUDCVM') {
        QCLOUDCVMForm.value.spec.vpc = '';
      } else {
        pmForm.value.spec.device_type = '';
      }
    };
    // 可用区变更时, 清空subnet, device_type
    const handleZoneChange = () => {
      if (resourceForm.value.resourceType === 'QCLOUDCVM') {
        QCLOUDCVMForm.value.spec.subnet = '';
        QCLOUDCVMForm.value.spec.device_type = '';
        selectedCvmDeviceType.value = null;
      }
    };

    const onQcloudAffinityChange = (val: any) => {
      pmForm.value.spec.anti_affinity_level = val;
    };

    // 获取QCLOUDCVM镜像列表
    const loadImages = async () => {
      const { info } = await apiService.getImages([resourceForm.value.region]);
      images.value = info;
      if (QCLOUDCVMForm.value.spec.image_id === '') {
        QCLOUDCVMForm.value.spec.image_id = 'img-fjxtfi0n';
      }
    };
    // 获取 IDCPM机型列表
    const IDCPMOsTypes = async () => {
      const { info } = await apiService.getOsTypes();
      pmForm.value.options.osTypes = info || [];
    };
    // 获取 IDCPM机型列表
    const IDCPMIsps = async () => {
      const { info } = await apiService.getIsps();
      pmForm.value.options.isps = info || [];
    };
    // 获取 IDCPM机型列表
    const IDCPMDeviceTypes = async () => {
      const { info } = await apiService.getIDCPMDeviceTypes();
      pmForm.value.options.deviceTypes = info || [];
    };

    // RAID 类型
    const handleDeviceTypeChange = () => {
      pmForm.value.spec.raid_type =
        pmForm.value.options.deviceTypes?.find((item) => item.device_type === pmForm.value.spec.device_type)?.raid ||
        '';
    };
    // 获取可用的IDCPM列表
    const IDCPMlist = () => {
      IDCPMDeviceTypes();
      IDCPMOsTypes();
      IDCPMIsps();
    };
    const onResourceTypeChange = (resourceType: string) => {
      resourceForm.value.region = '';
      resourceForm.value.zone = '';
      const { osTypes, deviceTypes, isps } = pmForm.value.options;
      if (resourceType === 'IDCPM' && osTypes.length === 0 && deviceTypes.length === 0 && isps.length === 0) {
        IDCPMlist();
      }
    };
    // 监听物理机机型变化
    watch(
      () => pmForm.value.spec.device_type,
      () => {
        handleDeviceTypeChange();
      },
    );

    // 获取 QCLOUDCVM机型列表
    const cvmDevicetypeParams = computed(() => {
      const { region, zone } = resourceForm.value;
      return {
        region,
        zone: zone !== 'cvm_separate_campus' ? zone : undefined,
        require_type: order.value.model.requireType,
      };
    });

    const clonelist = (originRow: any, resourceType: string) => {
      const cloneRow = cloneDeep(originRow);

      if (resourceType === 'QCLOUDCVM') {
        // 克隆后，需要重新进行需求预检
        Object.assign(cloneRow, { verify_result: '', reason: '' });
        cloudTableData.value.push(cloneRow);
      } else {
        physicalTableData.value.push(cloneRow);
      }
    };

    const modifyindex = ref(0);
    const modifyresourceType = ref('');
    const modifylist = (originRow: any, index: number, resourceType: string) => {
      const cloneRow = cloneDeep(originRow);

      CVMapplication.value = false;
      resourceForm.value.resourceType = resourceType;
      modifyresourceType.value = resourceType;

      const { anti_affinity_level, bk_asset_id, remark, replicas, spec } = cloneRow;
      const { region, zone, charge_type, charge_months } = spec;

      Object.assign(resourceForm.value, { bk_asset_id, region, zone, remark });

      if (resourceType === 'QCLOUDCVM') {
        QCLOUDCVMForm.value.spec = { ...spec, anti_affinity_level, replicas: +replicas };
        Object.assign(resourceForm.value, { charge_type, charge_months });
      } else {
        pmForm.value.spec = { ...spec, anti_affinity_level, replicas: +replicas };
      }

      title.value = '修改资源需求';
      modifyindex.value = index;
      addResourceRequirements.value = true;
    };
    const deletelist = (index: number, resourceType: string) => {
      if (resourceType === 'QCLOUDCVM') {
        cloudTableData.value.splice(index, 1);
      } else {
        physicalTableData.value.splice(index, 1);
      }
    };
    watch(
      () => resourceForm.value.zone,
      () => {
        loadImages();
      },
    );
    watch(
      () => resourceForm.value.charge_type,
      (chargeType) => {
        if (chargeType === cvmChargeTypes.PREPAID) {
          resourceForm.value.charge_months = 36;
        } else {
          resourceForm.value.charge_months = undefined;
        }
      },
    );
    const unReapply = async () => {
      if (route?.query?.order_id) {
        const data = await apiService.getOrderDetail(+route?.query?.order_id);

        const {
          bk_biz_id: bkBizId,
          bk_username: bkUsername,
          require_type: requireType,
          enable_notice: enableNotice,
          expect_time: expectTime,
          remark,
          follower,
          suborders,
        } = data ?? {};

        order.value.model = {
          bkBizId,
          bkUsername,
          requireType,
          enableNotice,
          expectTime,
          remark,
          follower: follower || [],
          suborders,
        };

        suborders.forEach(({ resource_type, remark, replicas, spec, applied_core }: any) => {
          const data = { resource_type, remark, replicas: +replicas, spec, applied_core };

          resource_type === 'QCLOUDCVM' ? cloudTableData.value.push(data) : physicalTableData.value.push(data);
        });
      }
      if (route?.query?.id) {
        assignment(route?.query);

        // 来源为业务下CVM库存一键或资源预测申请时，需要回填需求类型
        if (route.query.from === 'businessCvmInventory' || route.query.from === 'businessResourcePlan') {
          order.value.model.requireType = Number(route.query.require_type);
        }

        addResourceRequirements.value = true;
      }
    };
    onMounted(() => {
      unReapply();
    });
    watch(
      () => order.value.model.requireType,
      () => {
        device.value.filter.require_type = order.value.model.requireType;
      },
    );
    const device = ref({
      filter: {
        require_type: 1,
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
        sort: '-capacity_flag',
      },
    });

    // 一键申请按钮点击事件
    const handleApplication = () => {
      CVMapplication.value = true;
    };
    watch(
      () => resourceForm.value.region,
      () => {
        if (resourceForm.value.resourceType === 'QCLOUDCVM') {
          loadImages();
        }
      },
    );

    const computedBiz = computed(() => {
      return whereAmI.value === Senarios.business ? getBizsId() : order.value.model.bkBizId;
    });

    watch(
      [
        () => computedBiz.value,
        () => order.value.model.requireType,
        () => resourceForm.value.region,
        () => resourceForm.value.zone,
      ],
      async ([bk_biz_id, require_type, region, zone]) => {
        // 查询有效预测范围内的机型
        if (
          !bk_biz_id ||
          !require_type ||
          !region ||
          !zone ||
          resourceForm.value.resourceType !== 'QCLOUDCVM' ||
          isSpecialRequirement.value
        )
          return;

        await getPlanedDeviceType(!isSpringPool.value ? bk_biz_id : 931, require_type, region, zone);

        if (availableDeviceTypeSet.prepaid.size === 0) {
          resourceForm.value.charge_type = cvmChargeTypes.POSTPAID_BY_HOUR;
        }
      },
      { deep: true },
    );

    const assignment = (data: any) => {
      resourceForm.value.resourceType = 'QCLOUDCVM';
      QCLOUDCVMForm.value.spec = {
        device_type: data.device_type, // 机型
        vpc: '', //  vpc
        subnet: '', //  子网
        replicas: 1,
        anti_affinity_level: 'ANTI_NONE',
        image_id: 'img-fjxtfi0n', // 镜像
        disk_type: 'CLOUD_PREMIUM', // 数据盘tyle
        disk_size: 0, // 数据盘size
        network_type: 'TENTHOUSAND',
        inherit_instance_id: '',
        cpu: data.cpu,
      };
      resourceForm.value.region = data.region;
      resourceForm.value.zone = data.zone;
      resourceForm.value.charge_months = 36;
    };
    const isOneClickApplication = ref(false);
    const OneClickApplication = (row: any, val: boolean) => {
      isOneClickApplication.value = true;
      CVMapplication.value = val;
      assignment(row);
      title.value = '增加资源需求';
      addResourceRequirements.value = true;
    };
    const ARtriggerShow = (isShow: boolean) => {
      emptyForm();
      addResourceRequirements.value = isShow;
      CVMapplication.value = isOneClickApplication.value;
      isOneClickApplication.value = false;
      nextTick(() => {
        resourceFormRef.value?.clearValidate();
        QCLOUDCVMformRef.value?.clearValidate();
        IDCPMformRef.value?.clearValidate();
        networkInfoFormRef.value?.clearValidate();
      });
    };
    const CAtriggerShow = (isShow: boolean) => {
      CVMapplication.value = isShow;
    };
    const emptyForm = () => {
      resourceForm.value = {
        resourceType: 'QCLOUDCVM',
        region: '', // 地域
        zone: '', // 园区
        remark: '',
        enable_disk_check: false,
        charge_type: cvmChargeTypes.PREPAID,
        charge_months: 36,
        bk_asset_id: resourceForm.value.bk_asset_id, // 继承套餐的机器固资号不用清除
      };
      QCLOUDCVMForm.value = {
        spec: {
          device_type: '', // 机型
          replicas: 1,
          vpc: '', //  vpc
          subnet: '', //  子网
          anti_affinity_level: 'ANTI_NONE',
          image_id: '', // 镜像
          disk_type: 'CLOUD_PREMIUM', // 数据盘tyle
          disk_size: 0, // 数据盘size
          network_type: 'TENTHOUSAND',
          inherit_instance_id: QCLOUDCVMForm.value.spec.inherit_instance_id, // 继承套餐的机器实例id不用清除
          cpu: undefined,
        },
      };
      pmForm.value.spec = {
        device_type: '', // 机型
        raid_type: '', // RAID 类型
        os_type: '', // 操作系统
        anti_affinity_level: '',
        replicas: 1,
        isp: '', // 运营商
        network_type: 'TENTHOUSAND',
      };
    };
    const cloudResourceForm = () => {
      const {
        resourceType: resource_type,
        remark,
        enable_disk_check,
        region,
        zone,
        charge_type,
        charge_months,
        bk_asset_id,
      } = resourceForm.value;

      return {
        bk_asset_id,
        resource_type,
        remark,
        enable_disk_check,
        anti_affinity_level: QCLOUDCVMForm.value.spec.anti_affinity_level,
        replicas: +QCLOUDCVMForm.value.spec.replicas,
        spec: {
          ...QCLOUDCVMForm.value.spec,
          // bk-input[number]粘贴与当前值相同的字符数值最终得到的是字符串值，这里统一转换为数字
          disk_size: Number(QCLOUDCVMForm.value.spec.disk_size),
          region,
          zone,
          charge_type,
          charge_months,
        },
      };
    };
    const PMResourceForm = () => {
      return {
        resource_type: resourceForm.value.resourceType,
        remark: resourceForm.value.remark,
        anti_affinity_level: pmForm.value.spec.anti_affinity_level,
        replicas: +pmForm.value.spec.replicas,
        spec: {
          region: resourceForm.value.region,
          zone: resourceForm.value.zone,
          ...pmForm.value.spec,
        },
      };
    };
    const QCLOUDCVMformRules = ref({
      device_type: [{ required: true, message: '请选择机型', trigger: 'change' }],
      image_id: [{ required: true, message: '请选择镜像', trigger: 'change' }],
      replicas: [
        { required: true, message: '请输入需求数量', trigger: 'blur' },
        // 临时规则双十一后可能需要去除
        {
          validator: (value: number) => !(isRollingServerLike.value && value > 100),
          message: '注意：因云接口限制，单次的机器数最大值为100，超过后请手动克隆为多条配置',
          trigger: 'blur',
        },
      ],
      disk_size: [
        {
          trigger: 'change',
          message: '数据盘大小范围在0-16000GB之间，数值必须是10的整数倍',
          validator: (val: number) => {
            return val >= 0 && val <= 16000 && val % 10 === 0;
          },
        },
      ],
    });
    const resourceFormrules = ref({
      resourceType: [{ required: true, message: '请选择主机类型', trigger: 'change' }],
      region: [{ required: true, message: '请选择云地域', trigger: 'change' }],
      zone: [{ required: true, message: '请选择可用区', trigger: 'change' }],
    });
    const handleSubmit = async () => {
      await resourceFormRef.value.validate();
      if (resourceForm.value.resourceType === 'QCLOUDCVM') {
        await QCLOUDCVMformRef.value.validate();
      } else {
        await IDCPMformRef.value.validate();
      }

      // 需要注意当主机类型为物理机时不会存在networkInfoFormRef
      try {
        await networkInfoFormRef.value?.validate();
      } catch (error) {
        networkInfoPanelRef.value?.handleToggle(true);
        return Promise.reject(error);
      }

      if (title.value === '增加资源需求') {
        if (resourceForm.value.resourceType === 'QCLOUDCVM') {
          cloudTableData.value.push(cloudResourceForm());
        } else {
          physicalTableData.value.push(PMResourceForm());
        }
        emptyForm();
      } else {
        if (modifyresourceType.value === 'QCLOUDCVM') {
          if (modifyresourceType.value === resourceForm.value.resourceType) {
            cloudTableData.value[modifyindex.value] = cloudResourceForm();
          } else {
            cloudTableData.value.splice(modifyindex.value, 1);
            physicalTableData.value.push(PMResourceForm());
          }
          emptyForm();
        } else {
          if (modifyresourceType.value === resourceForm.value.resourceType) {
            physicalTableData.value[modifyindex.value] = PMResourceForm();
          } else {
            physicalTableData.value.splice(modifyindex.value, 1);
            cloudTableData.value.push(cloudResourceForm());
          }
          emptyForm();
        }
      }
      modifyindex.value = 0;
      addResourceRequirements.value = false;
      nextTick(() => {
        resourceFormRef.value?.clearValidate();
        QCLOUDCVMformRef.value?.clearValidate();
        IDCPMformRef.value?.clearValidate();
      });
    };
    const isUncommit = computed(() => {
      return route?.query?.order_id && +route?.query?.unsubmitted === 1;
    });
    const handleSaveOrSubmit = async (type: 'save' | 'submit') => {
      await formRef.value.validate();
      const suborders = [...cloudTableData.value, ...physicalTableData.value];
      isLoading.value = true;

      try {
        const basePath = `${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa`;
        const taskPath = type === 'submit' ? 'task/create/apply' : 'task/update/apply/ticket';
        let url = null;
        let bk_biz_id;
        if (props.isbusiness) {
          bk_biz_id = accountStore.bizs;
          url = `${basePath}/bizs/${accountStore.bizs}/${taskPath}`;
        } else {
          bk_biz_id = order.value.model.bkBizId === 'all' ? undefined : order.value.model.bkBizId;
          url = `${basePath}/${taskPath}`;
        }
        await http.post(url, {
          order_id: isUncommit.value ? +route?.query.order_id : undefined,
          bk_biz_id,
          bk_username: useUserStore().username,
          require_type: order.value.model.requireType,
          expect_time: timeFormatter(order.value.model.expectTime),
          remark: order.value.model.remark,
          follower: order.value.model.follower,
          suborders,
        });
        const message = `${type === 'submit' ? '申请成功' : '保存成功'}`;
        Message({
          theme: 'success',
          message,
        });
        // 合代码之后完善跳转路由
        if (props.isbusiness) {
          router.replace({
            name: 'ApplicationsManage',
            query: { [GLOBAL_BIZS_KEY]: bk_biz_id, type: 'host_apply' },
          });
        } else {
          router.replace({ name: MENU_SERVICE_HOST_APPLICATION });
        }
      } finally {
        isLoading.value = false;
      }
    };

    const handleVerify = async () => {
      await formRef.value.validate();
      const suborders = cloudTableData.value;
      isLoading.value = true;
      try {
        const { data } = await planStore.verify_resource_demand({
          bk_biz_id: !isSpringPool.value ? +computedBiz.value : 931,
          require_type: order.value.model.requireType,
          suborders,
        });
        for (let i = 0; i < cloudTableData.value.length; i++) {
          cloudTableData.value[i].verify_result = data.verifications[i].verify_result;
          cloudTableData.value[i].reason = data.verifications[i].reason;
        }
        isNeedVerfiy.value = data.verifications.reduce((acc, cur) => {
          acc ||= cur.verify_result !== 'PASS';
          return acc;
        }, false);
        Message({
          theme: isNeedVerfiy.value ? 'warning' : 'success',
          message: isNeedVerfiy.value ? '校验不通过' : '校验通过',
        });
      } finally {
        isLoading.value = false;
      }
    };

    const handleCancel = () => {
      if (props.isbusiness) {
        router.push({
          name: 'hostBusinessList',
        });
      } else {
        router.go(-1);
      }
    };

    // vpc变更时，置空subnet
    const handleVpcChange = () => {
      QCLOUDCVMForm.value.spec.subnet = '';
    };

    const cvmMaxCapacityQueryParams = computed(() => {
      const { requireType: require_type } = order.value.model;
      const { device_type, vpc, subnet } = QCLOUDCVMForm.value.spec;
      const { region, zone, charge_type } = resourceForm.value;
      return { require_type, region, zone, device_type, vpc, subnet, charge_type };
    });

    watch(
      isSpecialRequirement,
      (val) => {
        if (!val) {
          resourceForm.value.bk_asset_id = '';
          QCLOUDCVMForm.value.spec.inherit_instance_id = '';
          cloudTableColumns.value = [...CloudHostcolumns, ...CVMVerifyColumns, CloudHostoperation.value];
          cloudHostSetting = generateColumnsSettings(cloudTableColumns.value);
          return;
        }
        // 滚服项目、小额绿通, 默认为云主机, 禁用选择
        resourceForm.value.resourceType = 'QCLOUDCVM';
        // 动态更新云主机列字段
        cloudTableColumns.value = [...CloudHostcolumns, CloudHostoperation.value];
        cloudHostSetting = generateColumnsSettings(cloudTableColumns.value);
      },
      {
        immediate: true,
      },
    );

    // 清空配置清单
    const clearResRequirements = () => {
      cloudTableData.value = [];
      physicalTableData.value = [];
    };

    // 需求核数
    const replicasCpuCores = computed(() =>
      cloudTableData.value.reduce((prev, curr) => {
        const { replicas, spec, applied_core } = curr;
        // 如果 applied_core(接口值) 有值，则优先使用；若无值，则使用前端计算值
        if (applied_core !== undefined) return prev + applied_core;

        return prev + replicas * spec.cpu;
      }, 0),
    );

    // 滚服项目、小额绿通 - cpu需求限额
    const rollingServerCpuCoreLimitsRef = useTemplateRef<typeof RollingServerCpuCoreLimits>(
      'rolling-server-cpu-core-limits',
    );
    const availableCpuCoreQuota = computed(() => {
      let val = 0;
      if (isRollingServerLike.value) val = rollingServerCpuCoreLimitsRef.value?.availableCpuCoreQuota ?? val;
      if (isGreenChannel.value) val = greenChannelCpuCoreLimitsRef.value?.availableCpuCoreQuota ?? val;
      return val;
    });
    const isCpuCoreExceeded = computed(() => replicasCpuCores.value > availableCpuCoreQuota.value);

    // 小额绿通
    const greenChannelCpuCoreLimitsRef = useTemplateRef<typeof GreenChannelCpuCoreLimits>(
      'green-channel-cpu-core-limits',
    );

    watch(
      () => cloudTableData.value,
      async (val) => {
        resetCpuAmount();
        for (const item of val) {
          const { replicas, spec, applied_core } = item;
          const { cpu, charge_type } = spec;

          // 如果 applied_core(接口值) 有值，则优先使用；若无值，则使用前端计算值
          if (ChargeType.POSTPAID_BY_HOUR === charge_type) {
            const postpaid =
              applied_core !== undefined ? cpuAmount.postpaid + applied_core : cpuAmount.postpaid + cpu * replicas;
            setCpuAmount({ ...cpuAmount, postpaid });
          }
          if (ChargeType.PREPAID === charge_type) {
            const prepaid =
              applied_core !== undefined ? cpuAmount.prepaid + applied_core : cpuAmount.prepaid + cpu * replicas;
            setCpuAmount({ ...cpuAmount, prepaid });
          }
        }
        isNeedVerfiy.value = val.reduce((acc, cur) => {
          acc ||= cur.verify_result !== 'PASS';
          return acc;
        }, false);
        isVerifyFailed.value = val.reduce((acc, cur) => {
          acc ||= cur.verify_result === 'FAILED';
          return acc;
        }, false);
      },
      {
        deep: true,
      },
    );

    return () => (
      <div class='host-application-form-wrapper'>
        {!props.isbusiness && <DetailHeader useRouterAction>新增申请</DetailHeader>}
        <div class={props.isbusiness ? '' : 'apply-form-wrapper'}>
          {/* 申请单据表单 */}
          <bk-form
            form-type='vertical'
            label-width='150'
            model={order.value.model}
            rules={order.value.rules}
            ref={formRef}>
            <CommonCard title={() => '基本信息'} class='mb12'>
              <div class='basic-top-row'>
                {!props.isbusiness && (
                  <bk-form-item label='所属业务' required property='bkBizId' class='mr24'>
                    <BusinessSelector
                      class='item-warp-component'
                      v-model={order.value.model.bkBizId}
                      autoSelect
                      authed
                      url-key='scr_apply_host_bizs'
                      apiMethod={apiService.getCvmApplyAuthBizList}
                      base64Encode
                    />
                  </bk-form-item>
                )}

                <bk-form-item label='需求类型' required property='requireType'>
                  <hcm-form-req-type
                    appearance='card'
                    v-model={order.value.model.requireType}
                    // 春保资源池不显示
                    filter={(list: any) => list.filter((item: any) => item.require_type !== 8)}
                    onChange={() => {
                      // 手动更改时，需要清空已保存的需求
                      clearResRequirements();
                    }}
                  />
                </bk-form-item>
                <div class='alert-content'>
                  {(function () {
                    if (isRollingServer.value) return <RollingServerTipsAlert />;
                    if (isGreenChannel.value) return <GreenChannelTipsAlert />;
                    if (isSpringPool.value) return <HostApplySpringPoolTips />;
                    return <HostApplyTipsAlert requireType={order.value.model.requireType} />;
                  })()}
                </div>
              </div>
              <div class='flex-row align-content-center'>
                <bk-form-item
                  label='期望交付时间'
                  required
                  property='expectTime'
                  class='mr24'
                  description={() => (
                    <span>
                      期望申领时间默认为当月，在资源预测额度充足时，过单后会立即申领。如希望审批时按指定时间过单后生产，请联系
                      <WName name={'ICR'} />
                      (IEG资源服务助手)确认{' '}
                    </span>
                  )}>
                  <bk-date-picker
                    class='item-warp-component'
                    v-model={order.value.model.expectTime}
                    clearable
                    type='datetime'></bk-date-picker>
                </bk-form-item>
                <bk-form-item label='关注人'>
                  <hcm-form-user class='item-warp-component' v-model={order.value.model.follower} />
                </bk-form-item>
              </div>
            </CommonCard>
            <CommonCard
              title={() => (
                <div class='flex-row align-items-center'>
                  <span class='mr5'>配置清单</span>
                  <i
                    class={'hcm-icon bkhcm-icon-info-line'}
                    v-bk-tooltips={{
                      content: (
                        <div>
                          <div>自研云主机购买，经过以下步骤后交付给业务</div>
                          <div>1.提交参数后，云梯生产主机</div>
                          <div>2.资源平台对系统初始化，包括GSE agent安装，磁盘格式化等</div>
                          <div>3.转交到业务</div>
                        </div>
                      ),
                    }}></i>
                </div>
              )}
              class='mb12 config-ticket-card'>
              <div class='mb12 tools-wrapper'>
                <Button
                  class='button'
                  theme='primary'
                  outline
                  onClick={() => {
                    addResourceRequirements.value = true;
                    title.value = '增加资源需求';
                    IDCPMlist();
                  }}
                  disabled={(isSpecialRequirement.value || isSpringPool.value) && availableCpuCoreQuota.value <= 0}
                  v-bk-tooltips={{
                    content: `已超过${
                      // eslint-disable-next-line no-nested-ternary
                      isRollingServerLike.value ? (isRollingServer.value ? '滚服项目' : '春保资源池') : '小额绿通'
                    }的CPU可用额度，不允许添加`,
                    disabled: !((isSpecialRequirement.value || isSpringPool.value) && availableCpuCoreQuota.value <= 0),
                  }}>
                  <Plus class={'prefix-icon'} />
                  添加
                </Button>
                <Button
                  class='button'
                  onClick={handleApplication}
                  disabled={isRollingServer.value || (isGreenChannel.value && availableCpuCoreQuota.value <= 0)}
                  v-bk-tooltips={{
                    content: isRollingServer.value
                      ? '滚服项目暂不支持一键申请'
                      : '已超过小额绿通的CPU可用额度，不允许添加',
                    disabled: !(isRollingServer.value || (isGreenChannel.value && availableCpuCoreQuota.value <= 0)),
                  }}>
                  一键申请
                </Button>
                {/* 滚服项目-cpu需求限额，春保资源池复用滚服 */}
                {isRollingServerLike.value && (
                  <RollingServerCpuCoreLimits
                    ref='rolling-server-cpu-core-limits'
                    bizId={computedBiz.value}
                    replicasCpuCores={replicasCpuCores.value}
                  />
                )}
                {/* 小额绿通-cpu需求限额 */}
                {isGreenChannel.value && (
                  <GreenChannelCpuCoreLimits
                    ref='green-channel-cpu-core-limits'
                    replicasCpuCores={replicasCpuCores.value}
                    bizId={computedBiz.value}
                  />
                )}
              </div>
              <bk-form-item label='云主机'>
                <p class={'statistics'}>
                  <span class={'label'}>包年包月CPU总数：</span>
                  <span class={'value'}>{cpuAmount.prepaid}核</span>
                  <span class={'ml24 label'}>按量计费CPU总数：</span>
                  <span class={'value'}>{cpuAmount.postpaid}核</span>
                </p>
                <bk-table
                  align='left'
                  row-hover='auto'
                  columns={cloudTableColumns.value}
                  settings={cloudHostSetting.value}
                  data={cloudTableData.value}
                  show-overflow-tooltip
                />
              </bk-form-item>
              {physicalTableData.value.length > 0 && (
                <bk-form-item label='物理机'>
                  <bk-table
                    align='left'
                    row-hover='auto'
                    columns={[...PhysicalMachinecolumns, PhysicalMachineoperation.value]}
                    data={physicalTableData.value}
                    show-overflow-tooltip
                  />
                </bk-form-item>
              )}
            </CommonCard>

            <CommonCard title={() => '备注'}>
              <bk-form-item label='申请备注'>
                <Input
                  type='textarea'
                  v-model={order.value.model.remark}
                  rows={3}
                  maxlength={255}
                  resize={false}
                  placeholder='请输入申请单备注'></Input>
              </bk-form-item>
            </CommonCard>

            {!isSpecialRequirement.value && isVerifyFailed.value && (
              <CommonCard title={() => '需求预检'}>
                <Alert theme='danger' showIcon={false} class={'mb24'}>
                  资源需求超过资源预测的剩余额度，请查看预检信息的报错明细，处理建议：
                  <br />
                  1.调整所需的资源，修改机型或者调整需求数量
                  <br />
                  2.增加资源预测报备后再重试，去
                  <Button
                    theme='primary'
                    text
                    onClick={() => window.open(`#/business/resource-plan?bizs=${computedBiz.value}`, '_blank')}>
                    查看资源预测
                  </Button>
                </Alert>
              </CommonCard>
            )}

            <bk-form-item class={'mt16 form-button-row'}>
              {!isSpecialRequirement.value ? (
                // 非滚服、非小额绿通
                <>
                  {!!cloudTableData.value.length && isNeedVerfiy.value ? (
                    <Button class='mr16' theme='primary' loading={isLoading.value} onClick={handleVerify}>
                      需求校验
                    </Button>
                  ) : (
                    <>
                      <Button
                        class='mr16'
                        theme='primary'
                        loading={isLoading.value}
                        disabled={!physicalTableData.value.length && !cloudTableData.value.length}
                        v-bk-tooltips={{
                          content: '资源需求不能为空',
                          disabled: physicalTableData.value.length || cloudTableData.value.length,
                        }}
                        onClick={() => handleSaveOrSubmit('submit')}>
                        提交
                      </Button>
                      <Button
                        class={'mr16'}
                        loading={isLoading.value}
                        disabled={!physicalTableData.value.length && !cloudTableData.value.length}
                        v-bk-tooltips={{
                          content: '资源需求不能为空',
                          disabled: physicalTableData.value.length || cloudTableData.value.length,
                        }}
                        onClick={() => handleSaveOrSubmit('save')}>
                        保存
                      </Button>
                    </>
                  )}
                </>
              ) : (
                // 滚服、小额绿通
                <>
                  <Button
                    class='mr16'
                    theme='primary'
                    loading={isLoading.value}
                    disabled={
                      (!physicalTableData.value.length && !cloudTableData.value.length) ||
                      // todo：如果是滚服项目、小额绿通，且需求核数超过限额，暂不允许提交，后续与资源预测交互同步。
                      isCpuCoreExceeded.value
                    }
                    v-bk-tooltips={(function () {
                      let disabled = true;
                      let content = '';
                      if (!physicalTableData.value.length && !cloudTableData.value.length) {
                        content = '资源需求不能为空';
                        disabled = Boolean(physicalTableData.value.length || cloudTableData.value.length);
                      }
                      if (isCpuCoreExceeded.value) {
                        const name = isRollingServer.value ? '滚服项目' : '小额绿通';
                        content = `当前所需的CPU总核数超过${name}CPU限额，请调整后再重试。`;
                        disabled = !isCpuCoreExceeded.value;
                      }
                      return { content, disabled };
                    })()}
                    onClick={() => handleSaveOrSubmit('submit')}>
                    提交
                  </Button>
                  {/* 滚服、小额绿通不支持保存 */}
                  <Button
                    class={'mr16'}
                    loading={isLoading.value}
                    disabled={true}
                    v-bk-tooltips={{ content: `${isRollingServer.value ? '滚服项目' : '小额绿通'}暂不支持保存` }}>
                    保存
                  </Button>
                </>
              )}

              <Button onClick={handleCancel}>取消</Button>
            </bk-form-item>
          </bk-form>

          {/* 增加资源需求 */}
          <Sideslider
            class='add-resource-requirements-sideslider'
            width={1200}
            isShow={addResourceRequirements.value}
            title={title.value}
            onClosed={() => {
              ARtriggerShow(false);
            }}>
            {{
              default: () => (
                <div class={'sideslider-layer'}>
                  <CommonCard title={() => '基本信息'}>
                    <bk-form
                      model={resourceForm.value}
                      rules={resourceFormrules}
                      ref={resourceFormRef}
                      form-type='vertical'>
                      <bk-form-item label='主机类型' required property='resourceType'>
                        <bk-select
                          class={'selection-box'}
                          v-model={resourceForm.value.resourceType}
                          onChange={onResourceTypeChange}
                          // 滚服项目、小额绿通只支持云主机
                          disabled={isSpecialRequirement.value || isRollingServerLike.value}
                          v-bk-tooltips={{
                            content: `${
                              // eslint-disable-next-line no-nested-ternary
                              isRollingServerLike.value
                                ? isRollingServer.value
                                  ? '滚服项目'
                                  : '春保资源池'
                                : '小额绿通'
                            }只支持云主机`,
                            disabled: !(isSpecialRequirement.value || isRollingServerLike.value),
                          }}>
                          {resourceTypes.value.map((resType: { value: any; label: any }) => (
                            <bk-option key={resType.value} value={resType.value} label={resType.label}></bk-option>
                          ))}
                        </bk-select>
                      </bk-form-item>
                      <bk-form-item label='云地域' required property='region'>
                        <AreaSelector
                          class={'selection-box'}
                          v-model={resourceForm.value.region}
                          params={{ resourceType: resourceForm.value.resourceType }}
                          onChange={handleRegionChange}></AreaSelector>
                      </bk-form-item>
                      <bk-form-item label='可用区' required property='zone'>
                        <ZoneTagSelector
                          class={'selection-box'}
                          key={resourceForm.value.region}
                          style={{ width: '760px' }}
                          v-model={resourceForm.value.zone}
                          vendor={VendorEnum.ZIYAN}
                          region={resourceForm.value.region}
                          resourceType={resourceForm.value.resourceType}
                          separateCampus={true}
                          emptyText='请先选择云地域'
                          minWidth={184}
                          maxWidth={184}
                          autoExpand={'selected'}
                          onChange={handleZoneChange}
                        />
                      </bk-form-item>
                      {/* 预测指引 */}
                      {!isSpecialRequirement.value &&
                        resourceForm.value.zone &&
                        resourceForm.value.resourceType === 'QCLOUDCVM' &&
                        !hasPlanedDeviceType.value &&
                        !isPlanedDeviceTypeLoading.value && (
                          <PlanLinkAlert
                            bkBizId={computedBiz.value}
                            showSuggestions={!isSpringPool.value}
                            style='margin: -12px 0 24px'
                          />
                        )}
                      {resourceForm.value.resourceType === 'QCLOUDCVM' && (
                        <>
                          {/* 滚服项目 - 继承套餐 */}
                          {isRollingServer.value && (
                            <InheritPackageFormItem
                              v-model={resourceForm.value.bk_asset_id}
                              bizs={order.value.model.bkBizId}
                              region={resourceForm.value.region}
                              onValidateSuccess={(host) => {
                                const { instance_charge_type: chargeType, charge_months: chargeMonths } = host;
                                resourceForm.value.charge_type = chargeType;
                                resourceForm.value.charge_months =
                                  chargeType === cvmChargeTypes.PREPAID ? chargeMonths : undefined;
                                QCLOUDCVMForm.value.spec.inherit_instance_id = host.bk_cloud_inst_id;
                                // 机型族与上次数据不一致时需要清除机型选择
                                if (
                                  rollingServerHost &&
                                  host.device_group !== rollingServerHost.device_group &&
                                  QCLOUDCVMForm.value.spec.device_type
                                ) {
                                  QCLOUDCVMForm.value.spec.device_type = '';
                                }
                                rollingServerHost = host;
                              }}
                              onValidateFailed={() => {
                                // 恢复默认值
                                resourceForm.value.charge_type = cvmChargeTypes.PREPAID;
                                resourceForm.value.charge_months = 36;
                                if (QCLOUDCVMForm.value.spec.device_type) {
                                  QCLOUDCVMForm.value.spec.device_type = '';
                                }
                                rollingServerHost = null;
                              }}
                            />
                          )}
                          <bk-form-item label='计费模式' required property='charge_type'>
                            {isSpecialRequirement.value ? (
                              // 滚服项目、小额绿通
                              <RadioGroup
                                v-model={resourceForm.value.charge_type}
                                type='card'
                                style={{ width: '260px' }}
                                // 滚服不支持选择计费模式
                                disabled={isRollingServer.value}
                                v-bk-tooltips={{
                                  content: '继承原有套餐，计费模式不可选',
                                  disabled: !isRollingServer.value,
                                }}>
                                <RadioButton label={cvmChargeTypes.PREPAID}>
                                  {cvmChargeTypeNames[cvmChargeTypes.PREPAID]}
                                </RadioButton>
                                <RadioButton label={cvmChargeTypes.POSTPAID_BY_HOUR}>
                                  {cvmChargeTypeNames[cvmChargeTypes.POSTPAID_BY_HOUR]}
                                </RadioButton>
                              </RadioGroup>
                            ) : (
                              // 其他需求类型
                              <RadioGroup
                                v-model={resourceForm.value.charge_type}
                                type='card'
                                style={{ width: '260px' }}>
                                <RadioButton
                                  label={cvmChargeTypes.PREPAID}
                                  disabled={availableDeviceTypeSet.prepaid.size === 0}
                                  v-bk-tooltips={{
                                    content: '当前地域无有效的预测需求，请提预测单后再按量申请',
                                    disabled: !resourceForm.value.zone || availableDeviceTypeSet.prepaid.size > 0,
                                  }}>
                                  {cvmChargeTypeNames[cvmChargeTypes.PREPAID]}
                                </RadioButton>
                                <RadioButton
                                  label={cvmChargeTypes.POSTPAID_BY_HOUR}
                                  disabled={availableDeviceTypeSet.postpaid.size === 0}
                                  v-bk-tooltips={{
                                    content: '当前地域无有效的预测需求，请提预测单后再按量申请',
                                    disabled: !resourceForm.value.zone || availableDeviceTypeSet.postpaid.size > 0,
                                  }}>
                                  {cvmChargeTypeNames[cvmChargeTypes.POSTPAID_BY_HOUR]}
                                </RadioButton>
                              </RadioGroup>
                            )}
                            <bk-alert theme='info' class='form-item-tips'>
                              {{
                                title: () => (
                                  <>
                                    {cvmChargeTypeTips[resourceForm.value.charge_type]}
                                    <bk-link
                                      href='https://crp.woa.com/crp-outside/yunti/news/20'
                                      theme='primary'
                                      target='_blank'>
                                      计费模式说明
                                    </bk-link>
                                  </>
                                ),
                              }}
                            </bk-alert>
                          </bk-form-item>
                          {resourceForm.value.charge_type === cvmChargeTypes.PREPAID && (
                            <bk-form-item label='购买时长' required property='charge_months'>
                              <bk-select
                                v-model={resourceForm.value.charge_months}
                                filterable={false}
                                clearable={false}
                                style={{ width: '260px' }}
                                disabled={chargeMonthDisabledState.value.disabled}
                                v-bk-tooltips={{
                                  content: chargeMonthDisabledState.value.tips,
                                  disabled: !chargeMonthDisabledState.value.disabled,
                                }}>
                                {(function () {
                                  const options = isRollingServer.value
                                    ? Array.from({ length: 48 }, (v, i) => i + 1)
                                    : [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 24, 36, 48];
                                  return options.map((option) => (
                                    <bk-option key={option} value={option} name={getMonthName(option)} />
                                  ));
                                })()}
                              </bk-select>
                            </bk-form-item>
                          )}
                        </>
                      )}
                    </bk-form>
                  </CommonCard>

                  {resourceForm.value.resourceType === 'QCLOUDCVM' && (
                    <Form
                      model={QCLOUDCVMForm.value.spec}
                      formType='vertical'
                      class='mt15'
                      ref={networkInfoFormRef}
                      rules={{
                        subnet: [
                          {
                            validator: (value: string) => (QCLOUDCVMForm.value.spec.vpc ? !!value : true),
                            message: '选择 VPC 后必须选择子网',
                            trigger: 'change',
                          },
                        ],
                      }}>
                      <NetworkInfoCollapsePanel
                        ref={'network-info-panel'}
                        class='network-info-collapse-panel'
                        v-model:vpc={QCLOUDCVMForm.value.spec.vpc}
                        v-model:subnet={QCLOUDCVMForm.value.spec.subnet}
                        region={resourceForm.value.region}
                        zone={resourceForm.value.zone}
                        vpcProperty={'vpc'}
                        subnetProperty={'subnet'}
                        disabledVpc={resourceForm.value.zone === 'cvm_separate_campus'}
                        disabledSubnet={resourceForm.value.zone === 'cvm_separate_campus'}
                        onChangeVpc={handleVpcChange}
                      />
                    </Form>
                  )}

                  <CommonCard
                    title={() => '实例配置'}
                    class={resourceForm.value.resourceType === 'IDCPM' ? 'idcpm-card' : 'not-idcpm-card'}>
                    <>
                      {resourceForm.value.resourceType !== 'IDCPM' ? (
                        <>
                          <bk-form
                            model={QCLOUDCVMForm.value.spec}
                            rules={QCLOUDCVMformRules.value}
                            ref={QCLOUDCVMformRef}
                            form-type='vertical'>
                            <bk-form-item label='机型' required property='device_type'>
                              <DevicetypeSelector
                                ref='device-type-selector'
                                class='commonCard-form-select'
                                v-model={QCLOUDCVMForm.value.spec.device_type}
                                resourceType='cvm'
                                params={cvmDevicetypeParams.value}
                                disabled={resourceForm.value.zone === ''}
                                isLoading={isPlanedDeviceTypeLoading.value}
                                placeholder={resourceForm.value.zone === '' ? '请先选择可用区' : '请选择机型'}
                                sort={deviceTypeCompareFn}
                                optionDisabled={deviceTypeOptionDisabledCallback}
                                optionDisabledTipsContent={deviceTypeOptionDisabledTipsCallback}
                                onChange={(result) => {
                                  QCLOUDCVMForm.value.spec.cpu = (result as CvmDeviceType)?.cpu_amount;
                                  selectedCvmDeviceType.value = result as CvmDeviceType;
                                  if (isDefaultFourYears.value) {
                                    // 专用机型默认包4年
                                    resourceForm.value.charge_months = 48;
                                  }
                                }}>
                                {{
                                  option: (option: CvmDeviceType) => {
                                    const { device_type, device_type_class, device_group } = option;
                                    const isSpecialType = device_type_class === 'SpecialType';
                                    return (
                                      <>
                                        <span>{device_type}</span>
                                        <Tag class='ml12' theme={isSpecialType ? 'danger' : 'success'} size='small'>
                                          {isSpecialType ? '专用机型' : '通用机型'}
                                        </Tag>
                                        {device_group && (
                                          <Tag class='ml12' size='small'>
                                            {device_group}
                                          </Tag>
                                        )}
                                      </>
                                    );
                                  },
                                }}
                              </DevicetypeSelector>
                              {selectedCvmDeviceType.value && (
                                <>
                                  <div class='form-item-tips'>{`所选机型为${selectedCvmDeviceType.value.device_type}，CPU为${selectedCvmDeviceType.value.cpu_amount}核，内存为${selectedCvmDeviceType.value.ram_amount}G`}</div>
                                </>
                              )}
                            </bk-form-item>
                            <bk-form-item label='镜像' required property='image_id'>
                              <bk-select
                                class={'commonCard-form-select'}
                                v-model={QCLOUDCVMForm.value.spec.image_id}
                                disabled={resourceForm.value.region === ''}
                                filterable>
                                {images.value.map((item) => (
                                  <bk-option key={item.image_id} label={item.image_name} value={item.image_id} />
                                ))}
                              </bk-select>
                            </bk-form-item>
                            <bk-form-item label='数据盘' property='disk_size'>
                              <div
                                style={{
                                  display: 'flex',
                                  alignItems: 'center',
                                }}>
                                <DiskTypeSelect
                                  style={'width:360px'}
                                  v-model={QCLOUDCVMForm.value.spec.disk_type}></DiskTypeSelect>
                                <Input
                                  class={'ml8'}
                                  type='number'
                                  style={'width:210px'}
                                  prefix='大小'
                                  suffix='GB'
                                  v-model={QCLOUDCVMForm.value.spec.disk_size}
                                  min={0}
                                  max={16000}></Input>
                                <i
                                  class={'hcm-icon bkhcm-icon-question-circle-fill ml5'}
                                  v-bk-tooltips={'最大为 16T(16000 G)，且必须为 10 的倍数'}></i>
                              </div>
                              {[CLOUD_CVM_DISKTYPE.SSD].includes(QCLOUDCVMForm.value.spec.disk_type) && (
                                <bk-alert theme='warning' class='form-item-tips' style='width:600px'>
                                  <>SSD 云硬盘的运营成本约为高性能云盘的 4 倍，请合理评估使用。</>
                                </bk-alert>
                              )}
                            </bk-form-item>
                            <bk-form-item label='需求数量' required property='replicas'>
                              <hcm-form-number
                                class='commonCard-form-select'
                                v-model={QCLOUDCVMForm.value.spec.replicas}
                                min={1}
                                max={1000}
                              />
                              {resourceForm.value.resourceType === 'QCLOUDCVM' && (
                                <CvmMaxCapacity params={cvmMaxCapacityQueryParams.value} />
                              )}
                            </bk-form-item>
                            <bk-form-item label='备注'>
                              <Input
                                type='textarea'
                                class={'commonCard-form-select'}
                                rows={3}
                                maxlength={255}
                                v-model={resourceForm.value.remark}
                                autosize
                                resize={false}></Input>
                            </bk-form-item>
                          </bk-form>
                        </>
                      ) : (
                        <>
                          <bk-form
                            model={pmForm.value.spec}
                            rules={pmForm.value.rules}
                            class={'example-form'}
                            ref={IDCPMformRef}
                            form-type='vertical'>
                            <bk-form-item label='机型' required property='device_type'>
                              <bk-select
                                v-model={pmForm.value.spec.device_type}
                                default-first-option
                                class='select-model'
                                filterable>
                                {pmForm.value.options.deviceTypes.map((deviceType: { device_type: any }) => (
                                  <bk-option
                                    key={deviceType.device_type}
                                    value={deviceType.device_type}
                                    label={deviceType.device_type}></bk-option>
                                ))}
                              </bk-select>
                            </bk-form-item>
                            <bk-form-item label='RAID 类型' class={'form-item-raid'}>
                              <div class={'raidText'}> {pmForm.value.spec.raid_type || '-'}</div>
                            </bk-form-item>
                            <bk-form-item label='操作系统' required property='os_type'>
                              <bk-select class={'commonCard-form-select'} v-model={pmForm.value.spec.os_type}>
                                {pmForm.value.options.osTypes.map((osType) => (
                                  <bk-option key={osType} value={osType} label={osType}></bk-option>
                                ))}
                              </bk-select>
                            </bk-form-item>
                            <bk-form-item label='运营商'>
                              <bk-select class={'commonCard-form-select'} v-model={pmForm.value.spec.isp}>
                                <bk-option key='无' value='' label='无'></bk-option>
                                {pmForm.value.options.isps.map((isp) => (
                                  <bk-option key={isp} value={isp} label={isp}></bk-option>
                                ))}
                              </bk-select>
                            </bk-form-item>
                            <div class='commonCard-form'>
                              <bk-form-item label='需求数量' required property='replicas'>
                                <hcm-form-number class='input-demand' v-model={pmForm.value.spec.replicas} min={1} />
                              </bk-form-item>
                              <bk-form-item
                                label='反亲和性'
                                required
                                class={'select-Affinity'}
                                property='anti_affinity_level'>
                                <AntiAffinityLevelSelect
                                  v-model={pmForm.value.spec.anti_affinity_level}
                                  params={{
                                    resourceType: resourceForm.value.resourceType,
                                    hasZone: resourceForm.value.zone !== '',
                                  }}
                                  onAffinitychange={onQcloudAffinityChange}></AntiAffinityLevelSelect>
                              </bk-form-item>
                            </div>
                            <bk-form-item label='备注' property='remark'>
                              <Input
                                class={'commonCard-form-select'}
                                type='textarea'
                                v-model={resourceForm.value.remark}
                                rows={3}
                                maxlength={255}
                                resize={false}
                                placeholder='请输入备注'></Input>
                            </bk-form-item>
                          </bk-form>
                        </>
                      )}
                    </>
                  </CommonCard>
                </div>
              ),
              footer: () => (
                <>
                  <Button
                    theme='primary'
                    onClick={handleSubmit}
                    v-bk-tooltips={{
                      content: '当前地域无资源预测，提预测单后再按量申请',
                      disabled: !(
                        !isSpecialRequirement.value &&
                        !!resourceForm.value.zone &&
                        resourceForm.value.resourceType === 'QCLOUDCVM' &&
                        !hasPlanedDeviceType.value
                      ),
                    }}
                    disabled={
                      !isSpecialRequirement.value &&
                      !!resourceForm.value.zone &&
                      resourceForm.value.resourceType === 'QCLOUDCVM' &&
                      !hasPlanedDeviceType.value
                    }>
                    保存需求
                  </Button>
                  <Button class='ml16' onClick={() => ARtriggerShow(false)}>
                    取消
                  </Button>
                </>
              ),
            }}
          </Sideslider>

          {/* CVM一键申请 */}
          <Sideslider
            class='common-sideslider cvm-apply-sideslider'
            width={960}
            isShow={CVMapplication.value}
            title='一键申请主机'
            onClosed={() => {
              CAtriggerShow(false);
            }}>
            {{
              header: () => (
                <div class='custom-header-content'>
                  一键申请主机
                  <ReqTypeValue
                    value={device.value.filter.require_type}
                    display={{ appearance: 'tag' }}
                    {...{ theme: 'info' }}
                  />
                </div>
              ),
              default: () => (
                <applicationSideslider
                  isShow={CVMapplication.value}
                  requireType={device.value.filter.require_type}
                  bizId={computedBiz.value}
                  onApply={OneClickApplication}
                />
              ),
            }}
          </Sideslider>
        </div>
      </div>
    );
  },
});
