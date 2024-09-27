import { Input, Button, Sideslider, Message, Popover, Dropdown, Radio, Form, Alert } from 'bkui-vue';
import { defineComponent, onMounted, ref, watch, nextTick, computed, reactive, useTemplateRef } from 'vue';
import { VendorEnum, CLOUD_CVM_DISKTYPE } from '@/common/constant';
import CommonCard from '@/components/CommonCard';
import BusinessSelector from '@/components/business-selector/index.vue';
import './index.scss';
import { useAccountStore, useUserStore } from '@/store';
import MemberSelect from '@/components/MemberSelect';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import AreaSelector from '../AreaSelector';
import ZoneTagSelector from '@/components/zone-tag-selector/index.vue';
import { getZoneCn } from '@/views/ziyanScr/cvm-web/transform';
import { Spinner } from 'bkui-vue/lib/icon';
import DiskTypeSelect from '../DiskTypeSelect';
import AntiAffinityLevelSelect from '../AntiAffinityLevelSelect';
import NetworkInfoPanel from '../network-info-panel/index.vue';
import DevicetypeSelector from '@/views/ziyanScr/components/devicetype-selector/index.vue';
import apiService from '@/api/scrApi';

import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import useCvmChargeType from '@/views/ziyanScr/hooks/use-cvm-charge-type';
import http from '@/http';
import applicationSideslider from '../application-sideslider';
import { useRouter, useRoute } from 'vue-router';
import { timeFormatter, expectedDeliveryTime } from '@/common/util';
import { cloneDeep } from 'lodash';
import { VerifyStatus, VerifyStatusMap } from './constants';
import usePlanStore from '@/store/usePlanStore';
import WName from '@/components/w-name';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { ChargeType } from '@/typings/plan';
import useFormModel from '@/hooks/useFormModel';
// 滚服项目
import RollingServerTipsAlert from '@/views/ziyanScr/rolling-server/tips-alert/index.vue';
import InheritPackageFormItem from '@/views/ziyanScr/rolling-server/inherit-package-form-item/index.vue';
import CpuCorsLimits from '@/views/ziyanScr/rolling-server/cpu-cors-limits/index.vue';
import { CvmDeviceType, IdcpmDeviceType } from '@/views/ziyanScr/components/devicetype-selector/types';
import success from 'bkui-vue/lib/icon/success';

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
    const router = useRouter();
    const route = useRoute();
    const addResourceRequirements = ref(false);
    const isLoading = ref(false);
    const title = ref('增加资源需求');
    const CVMapplication = ref(false);
    const { getBizsId, whereAmI } = useWhereAmI();
    const planStore = usePlanStore();
    const availablePrepaidSet = ref(new Set());
    const availablePostpaidSet = ref(new Set());
    const isNeedVerfiy = ref(true);
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
        bkBizId: [{ required: true, message: '请选择业务', trigger: 'change' }],
        requireType: [{ required: true, message: '请选择需求类型', trigger: 'change' }],
        expectTime: [{ required: true, message: '请填写期望交付时间', trigger: 'change' }],
      },
      options: {
        requireTypes: [],
      },
    });
    const computedAvailableSet = computed(() =>
      resourceForm.value.charge_type === ChargeType.PREPAID ? availablePrepaidSet.value : availablePostpaidSet.value,
    );
    const formRef = ref();
    const IDCPMIndex = ref(-1);
    const QCLOUDCVMIndex = ref(-1);
    const resourceFormRef = ref();
    const dropdownMenuShowState = reactive({
      idc: false,
      cvm: false,
    });
    const { columns: CloudHostcolumns } = useColumns('CloudHost');
    const { columns: PhysicalMachinecolumns } = useColumns('PhysicalMachine');
    const { cvmChargeTypes, cvmChargeTypeNames, cvmChargeTypeTips, getMonthName } = useCvmChargeType();
    const cloudTableColumns = ref([]);
    // 滚服项目-状态
    const isRollingServer = computed(() => order.value.model.requireType === 6);
    const PhysicalMachineoperation = ref({
      label: '操作',
      width: 200,
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
      width: 200,
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
        label: '预检校验状态',
        width: 110,
        render({ cell }: { cell: VerifyStatus }) {
          return <span class={`status-${cell}`}>{VerifyStatusMap[cell] || '待校验'}</span>;
        },
        isHidden: isRollingServer.value,
      },
      {
        field: 'reason',
        label: '预检校验原因',
        width: 120,
        render({ cell }: { cell: string }) {
          return cell || '--';
        },
        isHidden: isRollingServer.value,
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
        antiAffinityLevel: '',
        network_type: 'TENTHOUSAND',
        replicas: 1, // 需求数量
      },
      rules: {
        device_type: [{ required: true, message: '请选择机型', trigger: 'change' }],
        region: [{ required: true, message: '请选择地域', trigger: 'change' }],
        replicas: [{ required: true, message: '请输入需求数量', trigger: 'blur' }],
        os_type: [{ required: true, message: '请选择操作系统', trigger: 'change' }],
        antiAffinityLevel: [{ required: true, message: '请选择反亲和性', trigger: 'change' }],
      },
      options: {
        deviceTypes: [],
        osTypes: [],
        raidType: [],
        antiAffinityLevels: [],
        regions: [],
        zones: [],
        isps: [],
      },
    });
    // 云组件table
    const cloudTableData = ref([]);
    // 物理机table
    const physicalTableData = ref([]);
    // QCLOUDCVM云地域变化
    const onQcloudRegionChange = () => {
      resourceForm.value.zone = '';
      QCLOUDCVMForm.value.spec.device_type = '';
      QCLOUDCVMForm.value.spec.vpc = '';
    };

    // QCLOUDCVM可用区变化
    const onQcloudZoneChange = () => {
      if (resourceForm.value.resourceType === 'QCLOUDCVM') {
        QCLOUDCVMForm.value.spec.device_type = '';
      }
    };
    const onQcloudAffinityChange = (val: any) => {
      pmForm.value.spec.antiAffinityLevel = val;
    };
    // IDCPM云地域变化
    const onRegionChange = () => {
      if (resourceForm.value.resourceType === 'QCLOUDCVM') {
        onQcloudRegionChange();
      } else {
        resourceForm.value.zone = '';
        pmForm.value.spec.device_type = '';
      }
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
    watch(
      () => QCLOUDCVMForm.value.spec.device_type,
      () => {
        onQcloudDeviceTypeChange();
      },
    );
    // 获取 QCLOUDCVM机型列表
    const cvmDevicetypeParams = computed(() => {
      const { region, zone } = resourceForm.value;
      return {
        region,
        zone: zone !== 'cvm_separate_campus' ? zone : undefined,
      };
    });

    const clonelist = (row: any, resourceType: string) => {
      resourceType === 'QCLOUDCVM'
        ? cloudTableData.value.push(cloneDeep(row))
        : physicalTableData.value.push(cloneDeep(row));
    };
    const modifyindex = ref(0);
    const modifyresourceType = ref('');
    const modifylist = (row: any, index: number, resourceType: string) => {
      CVMapplication.value = false;
      resourceForm.value.resourceType = resourceType;
      modifyresourceType.value = resourceType;
      if (resourceType === 'QCLOUDCVM') {
        QCLOUDCVMForm.value.spec = cloudTableData.value[index].spec;
        resourceForm.value.region = cloudTableData.value[index].spec.region;
        resourceForm.value.zone = cloudTableData.value[index].spec.zone;
        resourceForm.value.charge_type = cloudTableData.value[index].spec.charge_type;
        resourceForm.value.charge_months = cloudTableData.value[index].spec.charge_months;
        QCLOUDCVMForm.value.spec.replicas = +row.replicas;
        QCLOUDCVMForm.value.spec.anti_affinity_level = row.anti_affinity_level;
      } else {
        pmForm.value.spec = physicalTableData.value[index].spec;
        resourceForm.value.region = physicalTableData.value[index].spec.region;
        resourceForm.value.zone = physicalTableData.value[index].spec.zone;
        pmForm.value.spec.antiAffinityLevel = row.anti_affinity_level;
        pmForm.value.spec.replicas = +row.replicas;
      }
      resourceForm.value.remark = row.remark;
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
    const defaultUserlist = ref([]);
    const unReapply = async () => {
      if (route?.query?.order_id) {
        const data = await apiService.getOrderDetail(+route?.query?.order_id);
        order.value.model = {
          bkBizId: data.bk_biz_id,
          bkUsername: data.bk_username,
          requireType: data.require_type,
          enableNotice: data.enable_notice,
          expectTime: data.expect_time,
          remark: data.remark,
          follower: data.follower,
          suborders: data.suborders,
        };
        defaultUserlist.value = order.value.model.follower.map((element: any) => ({
          username: element,
          display_name: element,
        }));

        order.value.model.suborders.forEach(({ resource_type, remark, replicas, spec }: any) => {
          resource_type === 'QCLOUDCVM'
            ? cloudTableData.value.push({
                remark,
                resource_type: 'QCLOUDCVM',
                replicas: +replicas,
                spec,
              })
            : physicalTableData.value.push({
                remark,
                resource_type: 'IDCPM',
                replicas: +replicas,
                spec,
              });
        });
      }
      if (route?.query?.id) {
        assignment(route?.query);
        addResourceRequirements.value = true;
      }
    };
    const getfetchOptionslist = async () => {
      const { info } = await apiService.getRequireTypes();
      order.value.options.requireTypes = info;
    };
    onMounted(() => {
      getfetchOptionslist();
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
        if (!bk_biz_id || !require_type || !region || !zone) return;
        const { data } = await planStore.list_config_cvm_charge_type_device_type({
          bk_biz_id,
          require_type,
          region,
          zone,
        });
        const { info } = data;
        for (const item of info) {
          const { charge_type, device_types } = item;
          let set = availablePostpaidSet.value;
          if (charge_type === ChargeType.PREPAID) {
            set = availablePrepaidSet.value;
          }
          for (const device of device_types) {
            const { device_type, available } = device;
            if (available) set.add(device_type);
          }
        }
        if (availablePrepaidSet.value.size === 0) resourceForm.value.charge_type = cvmChargeTypes.POSTPAID_BY_HOUR;
      },
      {
        deep: true,
      },
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
      onQcloudDeviceTypeChange();
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
        antiAffinityLevel: '',
        replicas: 1,
        isp: '', // 运营商
        network_type: 'TENTHOUSAND',
      };
    };
    const cloudResourceForm = () => {
      return {
        resource_type: resourceForm.value.resourceType,
        remark: resourceForm.value.remark,
        enable_disk_check: resourceForm.value.enable_disk_check,
        anti_affinity_level: QCLOUDCVMForm.value.spec.anti_affinity_level,
        replicas: +QCLOUDCVMForm.value.spec.replicas,
        spec: {
          ...QCLOUDCVMForm.value.spec,
          region: resourceForm.value.region,
          zone: resourceForm.value.zone,
          charge_type: resourceForm.value.charge_type,
          charge_months: resourceForm.value.charge_months,
        },
      };
    };
    const PMResourceForm = () => {
      return {
        resource_type: resourceForm.value.resourceType,
        remark: resourceForm.value.remark,
        anti_affinity_level: pmForm.value.spec.antiAffinityLevel,
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
      replicas: [{ required: true, message: '请输入需求数量', trigger: 'blur' }],
      disk_size: [
        {
          trigger: 'change',
          message: '数据盘大小范围在0-16000GB之间，数值必须是10的整数倍',
          validator: (val: number) => {
            return /^(0|[1-9]\d{0,3}0{1,3})$/.test(String(val)) && val <= 16000;
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
      cvmCapacity.value = [];
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
        const path = props.isbusiness ? '/business/applications' : '/service/hostApplication';
        router.push({
          path,
        });
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
          bk_biz_id: accountStore.bizs,
          require_type: 1,
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
          message: isNeedVerfiy.value ?  '校验不通过' : '校验通过'
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
    const cvmCapacity = ref([]);
    const loading = ref(false);
    const onQcloudDeviceTypeChange = async () => {
      const { device_type, vpc, subnet } = QCLOUDCVMForm.value.spec;
      const { region, zone } = resourceForm.value;
      const params = {
        require_type: 1,
        region,
        zone,
        device_type,
        vpc,
        subnet,
      };
      if (params.device_type) {
        loading.value = true;
        const { info } = await apiService.getCapacity(params);
        cvmCapacity.value = info || [];
        loading.value = false;
      }
    };
    watch(
      isRollingServer,
      (val) => {
        if (!val) {
          resourceForm.value.bk_asset_id = '';
          QCLOUDCVMForm.value.spec.inherit_instance_id = '';
          cloudTableColumns.value = [...CloudHostcolumns, ...CVMVerifyColumns, CloudHostoperation.value];
          return;
        }
        // 滚服项目, 默认为云主机, 禁用选择
        resourceForm.value.resourceType = 'QCLOUDCVM';
        // 动态更新云主机列字段
        cloudTableColumns.value = [...CloudHostcolumns, CloudHostoperation.value];
      },
      {
        immediate: true,
      },
    );

    // 滚服项目-cpu需求限额
    const cpuCorsLimitsRef = useTemplateRef<typeof CpuCorsLimits>('cpu-cors-limits');

    // 非滚服-机型排序逻辑
    const handleSortDemands = (a: CvmDeviceType, b: CvmDeviceType) => {
      const set = computedAvailableSet.value;
      return Number(set.has(b.device_type)) - Number(set.has(a.device_type));
    };

    watch(
      () => cloudTableData.value,
      async (val) => {
        resetCpuAmount();
        for (const item of val) {
          const { cpu, charge_type } = item.spec;
          if (ChargeType.POSTPAID_BY_HOUR === charge_type) {
            setCpuAmount({
              ...cpuAmount,
              postpaid: cpuAmount.postpaid + cpu,
            });
          }
          if (ChargeType.PREPAID === charge_type) {
            setCpuAmount({
              ...cpuAmount,
              prepaid: cpuAmount.prepaid + cpu,
            });
          }
        }
        isNeedVerfiy.value = val.reduce((acc, cur) => {
          acc ||= cur.verify_result !== 'PASS';
          return acc;
        }, false)
      },
      {
        deep: true,
      },
    );

    return () => (
      <div class='host-application-form-wrapper'>
        {!props.isbusiness && <DetailHeader backRouteName='主机申领'>新增申请</DetailHeader>}
        <div class={props.isbusiness ? '' : 'apply-form-wrapper'}>
          {/* 申请单据表单 */}
          <bk-form
            form-type='vertical'
            label-width='150'
            model={order.value.model}
            rules={order.value.rules}
            ref={formRef}>
            <CommonCard title={() => '基本信息'} class='mb12'>
              <div class='flex-row align-content-center'>
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
                  <bk-select
                    class='item-warp-component'
                    v-model={order.value.model.requireType}
                    onChange={() => {
                      // 手动更改时，需要清空已保存的需求
                      cloudTableData.value = [];
                      physicalTableData.value = [];
                    }}>
                    {order.value.options.requireTypes.map((item: { require_type: any; require_name: any }) => (
                      <bk-option
                        key={item.require_type}
                        value={item.require_type}
                        label={item.require_name}></bk-option>
                    ))}
                  </bk-select>
                </bk-form-item>
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
                  <MemberSelect
                    class='item-warp-component'
                    multiple
                    clearable
                    v-model={order.value.model.follower}
                    defaultUserlist={defaultUserlist.value}
                  />
                </bk-form-item>
              </div>
              {/* 滚服项目-tips */}
              {isRollingServer.value && <RollingServerTipsAlert />}
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
                  onClick={() => {
                    addResourceRequirements.value = true;
                    title.value = '增加资源需求';
                    IDCPMlist();
                  }}>
                  添加
                </Button>
                <Button
                  class='button'
                  onClick={handleApplication}
                  disabled={isRollingServer.value}
                  v-bk-tooltips={{ content: '滚服项目暂不支持一键申请', disabled: !isRollingServer.value }}>
                  一键申请
                </Button>
                {/* 滚服项目-cpu需求限额 */}
                {isRollingServer.value && <CpuCorsLimits ref='cpu-cors-limits' cloudTableData={cloudTableData.value} />}
              </div>
              <bk-form-item label='云主机'>
                {!isRollingServer.value && (
                  <p class={'statistics'}>
                    <span class={'label'}>包年包月CPU总数：</span>
                    <span class={'value'}>{cpuAmount.prepaid}核</span>
                    <span class={'ml24 label'}>按量计费CPU总数：</span>
                    <span class={'value'}>{cpuAmount.postpaid}核</span>
                  </p>
                )}
                <bk-table
                  align='left'
                  row-hover='auto'
                  columns={cloudTableColumns.value}
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

              {!isRollingServer.value && isNeedVerfiy.value && (
                <Alert theme='danger' showIcon={false} class={'mb24'}>
                  <p class={'status-FAILED'}>
                    前包年包月计费模式的资源需求超过资源预测的额度，请调整后重试，
                    <Button theme='primary' text>
                      查看资源预测
                    </Button>
                  </p>
                  <p class={'status-FAILED'}>
                    资源需求中有使用按量计费模式，长期使用成本较高，建议提预测单13周后转包年包月，
                    <Button theme='primary' text>
                      去创建提预测单
                    </Button>
                  </p>
                </Alert>
              )}

              <bk-form-item>
                {/* 非滚服 */}
                {!isRollingServer.value && (
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
                          disabled={!physicalTableData.value.length && !cloudTableData.value.length}
                          loading={isLoading.value}
                          v-bk-tooltips={{
                            content: '资源需求不能为空',
                            disabled: physicalTableData.value.length || cloudTableData.value.length,
                          }}
                          onClick={() => {
                            handleSaveOrSubmit('submit');
                          }}>
                          提交
                        </Button>
                        <Button
                          loading={isLoading.value}
                          disabled={!physicalTableData.value.length && !cloudTableData.value.length}
                          v-bk-tooltips={{
                            content: '资源需求不能为空',
                            disabled: physicalTableData.value.length || cloudTableData.value.length,
                          }}
                          onClick={() => {
                            handleSaveOrSubmit('save');
                          }}
                          class={'mr16'}>
                          保存
                        </Button>
                      </>
                    )}
                  </>
                )}

                {/* 滚服 */}
                {isRollingServer.value && (
                  <>
                    <Button
                      class='mr16'
                      theme='primary'
                      disabled={
                        (!physicalTableData.value.length && !cloudTableData.value.length) ||
                        // todo：如果是滚服项目，且需求核数超过限额，暂不允许提交，后续与资源预测交互同步。
                        cpuCorsLimitsRef.value?.isReplicasCpuCorsExceedsLimit
                      }
                      loading={isLoading.value}
                      v-bk-tooltips={(function () {
                        let disabled = true;
                        let content = '';
                        if (!physicalTableData.value.length && !cloudTableData.value.length) {
                          content = '资源需求不能为空';
                          disabled = Boolean(physicalTableData.value.length || cloudTableData.value.length);
                        }
                        if (cpuCorsLimitsRef.value?.isReplicasCpuCorsExceedsLimit) {
                          content = '当前所需的CPU总核数超过滚服CPU限额，请调整后再重试。';
                          disabled = !cpuCorsLimitsRef.value?.isReplicasCpuCorsExceedsLimit;
                        }
                        return { content, disabled };
                      })()}
                      onClick={() => {
                        handleSaveOrSubmit('submit');
                      }}>
                      提交
                    </Button>
                    <Button
                      loading={isLoading.value}
                      // 滚服项目暂不支持保存
                      disabled={
                        (!physicalTableData.value.length && !cloudTableData.value.length) || isRollingServer.value
                      }
                      v-bk-tooltips={(function () {
                        let disabled = true;
                        let content = '';
                        if (!physicalTableData.value.length && !cloudTableData.value.length) {
                          content = '资源需求不能为空';
                          disabled = Boolean(physicalTableData.value.length || cloudTableData.value.length);
                        }
                        if (isRollingServer.value) {
                          content = '滚服项目暂不支持保存';
                          disabled = !isRollingServer.value;
                        }
                        return { content, disabled };
                      })()}
                      onClick={() => {
                        handleSaveOrSubmit('save');
                      }}
                      class={'mr16'}>
                      保存
                    </Button>
                  </>
                )}
                <Button
                  onClick={() => {
                    handleCancel();
                  }}>
                  取消
                </Button>
              </bk-form-item>
            </CommonCard>
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
                          disabled={isRollingServer.value}
                          v-bk-tooltips={{ content: '滚服项目只允许云主机', disabled: !isRollingServer.value }}>
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
                          onChange={onRegionChange}></AreaSelector>
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
                          onChange={onQcloudZoneChange}
                        />
                      </bk-form-item>
                      {resourceForm.value.zone &&
                        !availablePostpaidSet.value.size &&
                        !availablePrepaidSet.value.size && (
                          <Alert class={'mb8'} theme='warning'>
                            当前地域无资源预测，提预测单后再按量申请，
                            {availablePostpaidSet.value.size}
                            {availablePrepaidSet.value.size}
                            <Button theme='primary' text>
                              去创建提预测单
                            </Button>
                          </Alert>
                        )}
                      {resourceForm.value.resourceType === 'QCLOUDCVM' && (
                        <>
                          {/* 滚服项目 - 继承套餐 */}
                          {isRollingServer.value && (
                            <InheritPackageFormItem
                              v-model={resourceForm.value.bk_asset_id}
                              bizs={order.value.model.bkBizId}
                              onValidateSuccess={(host) => {
                                resourceForm.value.charge_type = host.instance_charge_type;
                                resourceForm.value.charge_months = host.charge_months;
                                QCLOUDCVMForm.value.spec.inherit_instance_id = host.bk_cloud_inst_id;
                              }}
                              onValidateFailed={() => {
                                // 恢复默认值
                                resourceForm.value.charge_type = cvmChargeTypes.PREPAID;
                                resourceForm.value.charge_months = 36;
                              }}
                            />
                          )}
                          <bk-form-item label='计费模式' required property='charge_type'>
                            <RadioGroup
                              v-model={resourceForm.value.charge_type}
                              type='card'
                              style={{ width: '260px' }}
                              disabled={isRollingServer.value}
                              v-bk-tooltips={{
                                content: '继承原有套餐，计费模式不可选',
                                disabled: !isRollingServer.value,
                              }}>
                              <RadioButton
                                label={cvmChargeTypes.PREPAID}
                                disabled={availablePrepaidSet.value.size === 0}
                                v-bk-tooltips={{
                                  content: '当前地域无有效的预测需求，请提预测单后再按量申请',
                                  disabled:
                                    isRollingServer.value ||
                                    !resourceForm.value.zone ||
                                    availablePrepaidSet.value.size > 0,
                                }}>
                                {cvmChargeTypeNames[cvmChargeTypes.PREPAID]}
                              </RadioButton>
                              <RadioButton
                                label={cvmChargeTypes.POSTPAID_BY_HOUR}
                                disabled={availablePostpaidSet.value.size === 0}
                                v-bk-tooltips={{
                                  content: '当前地域无有效的预测需求，请提预测单后再按量申请',
                                  disabled:
                                    isRollingServer.value ||
                                    !resourceForm.value.zone ||
                                    availablePostpaidSet.value.size > 0,
                                }}>
                                {cvmChargeTypeNames[cvmChargeTypes.POSTPAID_BY_HOUR]}
                              </RadioButton>
                            </RadioGroup>
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
                                disabled={isRollingServer.value}
                                v-bk-tooltips={{
                                  content: '继承原有套餐包年包月时长，此处的购买时长为剩余时长',
                                  disabled: !isRollingServer.value,
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
                    <Form model={QCLOUDCVMForm.value.spec} formType='vertical' class='mt15'>
                      <NetworkInfoPanel
                        v-model:vpc={QCLOUDCVMForm.value.spec.vpc}
                        v-model:subnet={QCLOUDCVMForm.value.spec.subnet}
                        region={resourceForm.value.region}
                        zone={resourceForm.value.zone}
                        disabledVpc={resourceForm.value.zone === 'cvm_separate_campus'}
                        disabledSubnet={resourceForm.value.zone === 'cvm_separate_campus'}
                        onChangeVpc={onQcloudDeviceTypeChange}
                        onChangeSubnet={onQcloudDeviceTypeChange}
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
                                class='commonCard-form-select'
                                v-model={QCLOUDCVMForm.value.spec.device_type}
                                resourceType='cvm'
                                params={cvmDevicetypeParams.value}
                                disabled={resourceForm.value.zone === ''}
                                optionDisabled={
                                  !isRollingServer.value
                                    ? (v) => !computedAvailableSet.value.has(v.device_type)
                                    : undefined
                                }
                                optionDisabledTipsContent={
                                  !isRollingServer.value ? () => '当前机型不在有效预测范围内' : undefined
                                }
                                placeholder={resourceForm.value.zone === '' ? '请先选择可用区' : '请选择机型'}
                                sort={(a, b) => {
                                  if (!isRollingServer.value) return handleSortDemands(a, b);
                                  const aDeviceTypeClass = (a as CvmDeviceType).device_type_class;
                                  const bDeviceTypeClass = (b as CvmDeviceType).device_type_class;
                                  if (aDeviceTypeClass === 'CommonType' && bDeviceTypeClass === 'SpecialType')
                                    return -1;
                                  if (aDeviceTypeClass === 'SpecialType' && bDeviceTypeClass === 'CommonType') return 1;
                                  return 0;
                                }}
                                onChange={(result) => {
                                  QCLOUDCVMForm.value.spec.cpu = (result as CvmDeviceType).cpu_amount;
                                }}
                              />
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
                              <Input
                                type='number'
                                class='commonCard-form-select'
                                v-model={QCLOUDCVMForm.value.spec.replicas}
                                min={1}></Input>
                              <div class={'request-quantity-container'}>
                                {resourceForm.value.resourceType === 'QCLOUDCVM' && (
                                  <>
                                    {cvmCapacity.value.length ? (
                                      <>
                                        {cvmCapacity.value.map((item) => (
                                          <div class={'tooltips'}>
                                            <span class={'request-quantity-text'}>
                                              {getZoneCn(item?.zone)}最大可申请量
                                            </span>
                                            <span class={'max-request-hint'}>{item?.max_num || 0}</span>
                                            {loading.value ? <Spinner class={'mr10'} /> : <></>}
                                            <Popover trigger='hover' theme='light' disableTeleport={true} arrow={false}>
                                              {{
                                                default: () => (
                                                  <span>
                                                    {item?.max_info.length && (
                                                      <span class={'calculation-details'}>( 查看明细 )</span>
                                                    )}
                                                  </span>
                                                ),
                                                content: () => (
                                                  <div class={'content'}>
                                                    {item?.max_info.length &&
                                                      item?.max_info.map((val: { key: any; value: any }) => (
                                                        <div>
                                                          <span class={'application'}> {val.key}</span>
                                                          <span class={'max-request-hint'}> {val.value}</span>
                                                        </div>
                                                      ))}
                                                  </div>
                                                ),
                                              }}
                                            </Popover>
                                          </div>
                                        ))}
                                      </>
                                    ) : (
                                      <div class={'tooltips'}>
                                        <span>最大可申请量 </span>
                                        <span class={'max-request-hint'}>0</span>
                                        {loading.value ? <Spinner class={'mr10'} /> : <></>}
                                      </div>
                                    )}
                                  </>
                                )}
                              </div>
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
                                <Input
                                  class={'input-demand'}
                                  type='number'
                                  v-model={pmForm.value.spec.replicas}
                                  min={1}></Input>
                              </bk-form-item>
                              <bk-form-item
                                label='反亲和性'
                                required
                                class={'select-Affinity'}
                                property='antiAffinityLevel'>
                                <AntiAffinityLevelSelect
                                  v-model={pmForm.value.spec.antiAffinityLevel}
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
                  <Button theme='primary' onClick={handleSubmit}>
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
            class='common-sideslider'
            width={1100}
            isShow={CVMapplication.value}
            title='CVM一键申请'
            onClosed={() => {
              CAtriggerShow(false);
            }}>
            <applicationSideslider device={device.value} onOneApplication={OneClickApplication} />
          </Sideslider>
        </div>
      </div>
    );
  },
});
