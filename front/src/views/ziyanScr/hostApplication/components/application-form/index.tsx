import { defineComponent, onMounted, ref, watch, nextTick } from 'vue';
import { Input, Button, Sideslider, Message, Popover } from 'bkui-vue';
import CommonCard from '@/components/CommonCard';
import BusinessSelector from '@/components/business-selector/index.vue';
import './index.scss';
import MemberSelect from '@/components/MemberSelect';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import AreaSelector from '../AreaSelector';
import ZoneSelector from '../ZoneSelector';
import DiskTypeSelect from '../DiskTypeSelect';
import AntiAffinityLevelSelect from '../AntiAffinityLevelSelect';
import { RightShape, DownShape } from 'bkui-vue/lib/icon';
import apiService from '@/api/scrApi';
import { useUserStore } from '@/store';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import http from '@/http';
import applicationSideslider from '../application-sideslider';
import { useRouter, useRoute } from 'vue-router';
import { timeFormatter, expectedDeliveryTime } from '@/common/util';
import { cloneDeep } from 'lodash';
import { convertKeysToSnakeCase } from '@/utils/scr/test';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;
export default defineComponent({
  components: {
    applicationSideslider,
  },
  setup() {
    const IDCPMformRef = ref();
    const QCLOUDCVMformRef = ref();
    const router = useRouter();
    const route = useRoute();
    const addResourceRequirements = ref(false);
    const isLoading = ref(false);
    const title = ref('增加资源需求');
    const CVMapplication = ref(false);
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
            message: '请填写期望交付时间',
            trigger: 'change',
          },
        ],
      },
      options: {
        requireTypes: [],
      },
    });
    const formRef = ref();
    const resourceFormRef = ref();
    const { columns: CloudHostcolumns } = useColumns('CloudHost');
    const { columns: PhysicalMachinecolumns } = useColumns('PhysicalMachine');
    const PhysicalMachineoperation = ref({
      label: '操作',
      width: 200,
      render: ({ row, index }) => {
        return (
          <div>
            <Button text theme='primary' onClick={() => clonelist(row, 'IDCPM')} class={'mr8'}>
              克隆
            </Button>
            <Button text theme='primary' onClick={() => modifylist(row, index, 'IDCPM')} class={'mr8'}>
              修改
            </Button>
            <Button text theme='primary' onClick={() => deletelist(index, 'IDCPM')} class={'mr8'}>
              删除
            </Button>
          </div>
        );
      },
    });
    const CloudHostoperation = ref({
      label: '操作',
      width: 200,
      render: ({ row, index }) => {
        return (
          <>
            <Button text theme='primary' onClick={() => clonelist(row, 'QCLOUDCVM')} class={'mr8'}>
              克隆
            </Button>
            <Button text theme='primary' onClick={() => modifylist(row, index, 'QCLOUDCVM')} class={'mr8'}>
              修改
            </Button>
            <Button text theme='primary' onClick={() => deletelist(index, 'QCLOUDCVM')} class={'mr8'}>
              删除
            </Button>
          </>
        );
      },
    });
    // 添加按钮侧边栏公共表单对象
    const resourceForm = ref({
      resourceType: 'QCLOUDCVM', // 主机类型
      remark: '', // 备注
      enable_disk_check: false,
      region: '', // 地域
      zone: '', // 园区
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
    // 机型列表
    const deviceTypes = ref([]);
    // 镜像列表
    const images = ref([]);
    // VPC列表
    const zoneTypes = ref([]);
    // 子网列表
    const subnetTypes = ref([]);

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
    const resourceFormRules = ref({});
    // 网络信息开关
    const NIswitch = ref(true);
    // QCLOUDCVM云地域变化
    const onQcloudRegionChange = () => {
      loadImages();
      loadVpcs();
      resourceForm.value.zone = '';
      QCLOUDCVMForm.value.spec.device_type = '';
      QCLOUDCVMForm.value.spec.vpc = '';
    };
    const onQcloudVpcChange = () => {
      QCLOUDCVMForm.value.spec.subnet = '';
      loadSubnets();
      onQcloudDeviceTypeChange();
    };
    const onQcloudSubnetChange = () => {
      onQcloudDeviceTypeChange();
    };
    // QCLOUDCVM可用区变化
    const onQcloudZoneChange = () => {
      if (resourceForm.value.resourceType === 'QCLOUDCVM') {
        QCLOUDCVMForm.value.spec.device_type = '';
        loadDeviceTypes();
        loadSubnets();
      }
    };
    const onQcloudAffinityChange = (val) => {
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
    const onResourceTypeChange = () => {
      resourceForm.value.region = '';
      resourceForm.value.zone = '';
    };
    // 监听物理机机型变化
    watch(
      () => pmForm.value.spec.device_type,
      () => {
        handleDeviceTypeChange();
      },
    );
    // 获取 QCLOUDCVM机型列表
    const loadDeviceTypes = async () => {
      const { zone, region } = resourceForm.value;

      const params = {
        region: [region],
        zone: zone !== 'cvm_separate_campus' ? [zone] : undefined,
      };

      const { info } = await apiService.getDeviceTypes(params);
      deviceTypes.value = info || [];
    };
    // 获取 QCLOUDCVM  VPC列表
    const loadVpcs = async () => {
      const { info } = await apiService.getVpcs(resourceForm.value.region);
      zoneTypes.value = info;
    };
    // 获取 QCLOUDCVM子网列表
    const loadSubnets = async () => {
      const { vpc } = QCLOUDCVMForm.value.spec;
      const { region, zone } = resourceForm.value;
      const { info } = await apiService.getSubnets({
        region,
        zone,
        vpc,
      });

      subnetTypes.value = info || [];
    };
    const clonelist = (row: any, resourceType: string) => {
      resourceType === 'QCLOUDCVM'
        ? cloudTableData.value.push(cloneDeep(row))
        : physicalTableData.value.push(cloneDeep(row));
    };
    const modifyindex = ref(0);
    const modifyresourceType = ref('');
    const modifylist = (row, index, resourceType) => {
      CVMapplication.value = false;
      resourceForm.value.resourceType = resourceType;
      modifyresourceType.value = resourceType;
      if (resourceType === 'QCLOUDCVM') {
        QCLOUDCVMForm.value.spec = cloudTableData.value[index].spec;
        resourceForm.value.region = cloudTableData.value[index].spec.region;
        resourceForm.value.zone = cloudTableData.value[index].spec.zone;
        QCLOUDCVMForm.value.spec.replicas = row.replicas;
        QCLOUDCVMForm.value.spec.anti_affinity_level = row.anti_affinity_level;
      } else {
        pmForm.value.spec = physicalTableData.value[index].spec;
        resourceForm.value.region = physicalTableData.value[index].spec.region;
        resourceForm.value.zone = physicalTableData.value[index].spec.zone;
        pmForm.value.spec.antiAffinityLevel = row.anti_affinity_level;
        pmForm.value.spec.replicas = row.replicas;
      }
      resourceForm.value.remark = row.remark;
      title.value = '修改资源需求';
      modifyindex.value = index;
      addResourceRequirements.value = true;
    };
    const deletelist = (index, resourceType) => {
      if (resourceType === 'QCLOUDCVM') {
        cloudTableData.value.splice(index, 1);
      } else {
        physicalTableData.value.splice(index, 1);
      }
    };
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
        order.value.model.suborders.forEach(({ resource_type, remark, replicas, spec }) => {
          resource_type === 'QCLOUDCVM'
            ? cloudTableData.value.push({
                remark,
                resource_type: 'QCLOUDCVM',
                replicas,
                spec,
              })
            : physicalTableData.value.push({
                remark,
                resource_type: 'IDCPM',
                replicas,
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
    const clickApplication = () => {
      CVMapplication.value = true;
    };
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
      };
      resourceForm.value.region = data.region;
      resourceForm.value.zone = data.zone;
    };
    const OneClickApplication = (row: any, val: boolean) => {
      CVMapplication.value = val;
      assignment(row);
      title.value = '增加资源需求';
      addResourceRequirements.value = true;
      onQcloudDeviceTypeChange();
    };
    const ARtriggerShow = (isShow: boolean) => {
      emptyform();
      addResourceRequirements.value = isShow;
      nextTick(() => {
        resourceFormRef.value.clearValidate();
        QCLOUDCVMformRef.value.clearValidate();
        IDCPMformRef.value.clearValidate();
      });
    };
    const CAtriggerShow = (isShow: boolean) => {
      CVMapplication.value = isShow;
    };
    const emptyform = () => {
      resourceForm.value = {
        resourceType: 'QCLOUDCVM',
        region: '', // 地域
        zone: '', // 园区
        remark: '',
        enable_disk_check: false,
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
        replicas: QCLOUDCVMForm.value.spec.replicas,
        spec: {
          region: resourceForm.value.region,
          zone: resourceForm.value.zone,
          ...QCLOUDCVMForm.value.spec,
        },
      };
    };
    const PMResourceForm = () => {
      return {
        resource_type: resourceForm.value.resourceType,
        remark: resourceForm.value.remark,
        anti_affinity_level: pmForm.value.spec.antiAffinityLevel,
        replicas: pmForm.value.spec.replicas,
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
        emptyform();
      } else {
        if (modifyresourceType.value === 'QCLOUDCVM') {
          if (modifyresourceType.value === resourceForm.value.resourceType) {
            cloudTableData.value[modifyindex.value] = cloudResourceForm();
          } else {
            cloudTableData.value.splice(modifyindex.value, 1);
            physicalTableData.value.push(PMResourceForm());
          }
          emptyform();
        } else {
          if (modifyresourceType.value === resourceForm.value.resourceType) {
            physicalTableData.value[modifyindex.value] = PMResourceForm();
          } else {
            physicalTableData.value.splice(modifyindex.value, 1);
            cloudTableData.value.push(cloudResourceForm());
          }
          emptyform();
        }
      }
      modifyindex.value = 0;
      cvmCapacity.value = [];
      addResourceRequirements.value = false;
      nextTick(() => {
        resourceFormRef.value.clearValidate();
        QCLOUDCVMformRef.value.clearValidate();
        IDCPMformRef.value.clearValidate();
      });
    };
    const handleSaveOrSubmit = async (type: 'save' | 'submit') => {
      await formRef.value.validate();
      const suborders = [...cloudTableData.value, ...physicalTableData.value].map((v) => {
        return convertKeysToSnakeCase(v);
      });
      if (!suborders.length) {
        return;
      }
      isLoading.value = true;
      try {
        const url = `${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/${
          type === 'submit' ? 'task/create/apply' : 'task/update/apply/ticket'
        }`;
        await http.post(url, {
          bk_biz_id: order.value.model.bkBizId === 'all' ? undefined : order.value.model.bkBizId,
          bk_username: useUserStore().username,
          require_type: order.value.model.requireType,
          expect_time: timeFormatter(order.value.model.expectTime),
          remark: order.value.model.remark,
          follower: order.value.model.follower,
          suborders,
        });
        Message({
          theme: 'success',
          message: '申请成功',
        });
        router.go(-1);
      } finally {
        isLoading.value = false;
      }
    };
    const cvmCapacity = ref([]);
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
        const { info } = await apiService.getCapacity(params);
        cvmCapacity.value = info || [];
      }
    };
    return () => (
      <div class='wid100'>
        <DetailHeader>新增申请</DetailHeader>
        <div class={'apply-form-wrapper'}>
          {/* 申请单据表单 */}
          <CommonCard class='mt15' title={() => '基本信息'} layout='grid'>
            <bk-form
              form-type='vertical'
              label-width='150'
              model={order.value.model}
              rules={order.value.rules}
              ref={formRef}>
              <div class='displayflex'>
                <bk-form-item label='所属业务' class='item-warp' required property='bkBizId'>
                  <BusinessSelector v-model={order.value.model.bkBizId} autoSelect authed />
                </bk-form-item>
                <bk-form-item label='需求类型' class='item-warp' required property='requireType'>
                  <bk-select class='item-warp-component' v-model={order.value.model.requireType}>
                    {order.value.options.requireTypes.map((item: { require_type: any; require_name: any }) => (
                      <bk-option
                        key={item.require_type}
                        value={item.require_type}
                        label={item.require_name}></bk-option>
                    ))}
                  </bk-select>
                </bk-form-item>
              </div>
              <div class='displayflex'>
                <bk-form-item label='期望交付时间' class='item-warp' required property='expectTime'>
                  <bk-date-picker
                    class='item-warp-component'
                    v-model={order.value.model.expectTime}
                    clearable
                    type='datetime'></bk-date-picker>
                </bk-form-item>
                <bk-form-item label='关注人' class='item-warp'>
                  <MemberSelect class='item-warp-component' multiple clearable v-model={order.value.model.follower} />
                </bk-form-item>
              </div>
            </bk-form>
          </CommonCard>
          <CommonCard class='mt15' title={() => '配置清单'}>
            <div>
              <Button
                class='mr16'
                theme='primary'
                onClick={() => {
                  addResourceRequirements.value = true;
                  title.value = '增加资源需求';
                  IDCPMlist();
                }}>
                添加
              </Button>
              <Button onClick={clickApplication}>一键申请</Button>
              <div class='mt15'>云主机</div>
              <div class='mt15'>
                <bk-table
                  align='left'
                  row-hover='auto'
                  columns={[...CloudHostcolumns, CloudHostoperation.value]}
                  data={cloudTableData.value}
                  show-overflow-tooltip
                  border={['outer', 'row', 'col']}
                />
              </div>
              {physicalTableData.value.length ? (
                <>
                  <div class='mt15'>物理机</div>
                  <div class='mt15'>
                    <bk-table
                      align='left'
                      row-hover='auto'
                      columns={[...PhysicalMachinecolumns, PhysicalMachineoperation.value]}
                      data={physicalTableData.value}
                      show-overflow-tooltip
                      border={['outer', 'row', 'col']}
                    />
                  </div>
                </>
              ) : (
                <></>
              )}
            </div>
          </CommonCard>
          <CommonCard title={() => '备注'}>
            <bk-form form-type='vertical' label-width='150' model={order.value.model}>
              <bk-form-item label='申请备注' class='item-warp'>
                <Input
                  type='textarea'
                  v-model={order.value.model.remark}
                  rows={3}
                  maxlength={255}
                  resize={false}
                  placeholder='请输入申请单备注'></Input>
              </bk-form-item>
              <bk-form-item class='item-warp' required property='bkBizId'>
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
                <Button
                  onClick={() => {
                    router.go(-1);
                  }}>
                  取消
                </Button>
              </bk-form-item>
            </bk-form>
          </CommonCard>

          {/* 增加资源需求 */}
          <Sideslider
            class='common-sideslider'
            width={1000}
            isShow={addResourceRequirements.value}
            title={title.value}
            onClosed={() => {
              ARtriggerShow(false);
            }}>
            {{
              default: () => (
                <div class='common-sideslider-content'>
                  <CommonCard title={() => '基本信息'}>
                    <bk-form
                      model={resourceForm.value}
                      class={'scr-form-wrapper'}
                      rules={resourceFormrules}
                      ref={resourceFormRef}>
                      <bk-form-item label='主机类型' required property='resourceType'>
                        <bk-select v-model={resourceForm.value.resourceType} onChange={onResourceTypeChange}>
                          {resourceTypes.value.map((resType: { value: any; label: any }) => (
                            <bk-option key={resType.value} value={resType.value} label={resType.label}></bk-option>
                          ))}
                        </bk-select>
                      </bk-form-item>

                      <bk-form-item class='mr16' label='云地域' required property='region'>
                        <AreaSelector
                          ref='areaSelector'
                          v-model={resourceForm.value.region}
                          params={{ resourceType: resourceForm.value.resourceType }}
                          onChange={onRegionChange}></AreaSelector>
                      </bk-form-item>
                      <bk-form-item label='可用区' required property='zone'>
                        <ZoneSelector
                          ref='zoneSelector'
                          v-model={resourceForm.value.zone}
                          params={{
                            resourceType: resourceForm.value.resourceType,
                            region: resourceForm.value.region,
                          }}
                          onChange={onQcloudZoneChange}
                        />
                      </bk-form-item>
                    </bk-form>
                  </CommonCard>

                  {resourceForm.value.resourceType === 'QCLOUDCVM' && (
                    <>
                      {NIswitch.value ? (
                        <>
                          <CommonCard
                            title={() => (
                              <>
                                <div class='displayflex'>
                                  <RightShape
                                    onClick={() => {
                                      NIswitch.value = false;
                                    }}
                                  />
                                  <span class='fontsize'>网络信息</span>
                                  <span class='fontweight'>VPC : {QCLOUDCVMForm.value.spec.vpc}</span>
                                  <span class='fontweight'>子网 : {QCLOUDCVMForm.value.spec.subnet}</span>
                                </div>
                              </>
                            )}
                            layout='grid'>
                            <></>
                          </CommonCard>
                        </>
                      ) : (
                        <>
                          <CommonCard
                            title={() => (
                              <>
                                <div class='displayflex'>
                                  <DownShape
                                    onClick={() => {
                                      NIswitch.value = true;
                                    }}
                                  />
                                  <span class='fontsize'>网络信息</span>
                                </div>
                              </>
                            )}
                            layout='grid'>
                            <>
                              <bk-form
                                form-type='vertical'
                                label-width='150'
                                model={QCLOUDCVMForm.value.spec}
                                rules={resourceFormRules.value}
                                ref='formRef'>
                                <bk-form-item label='VPC'>
                                  <div class='component-with-detail-container'>
                                    <bk-select
                                      class='item-warp-resourceType component-with-detail'
                                      disabled={resourceForm.value.zone === 'cvm_separate_campus'}
                                      v-model={QCLOUDCVMForm.value.spec.vpc}
                                      onChange={onQcloudVpcChange}>
                                      {zoneTypes.value.map((vpc) => (
                                        <bk-option
                                          key={vpc.vpc_id}
                                          value={vpc.vpc_id}
                                          label={`${vpc.vpc_id} | ${vpc.vpc_name}`}></bk-option>
                                      ))}
                                    </bk-select>
                                  </div>
                                </bk-form-item>
                                <bk-form-item label='子网'>
                                  <div class='component-with-detail-container'>
                                    <bk-select
                                      class='item-warp-resourceType component-with-detail'
                                      disabled={resourceForm.value.zone === 'cvm_separate_campus'}
                                      v-model={QCLOUDCVMForm.value.spec.subnet}
                                      onChange={onQcloudSubnetChange}>
                                      {subnetTypes.value.map((subnet) => (
                                        <bk-option
                                          key={subnet.subnet_id}
                                          value={subnet.subnet_id}
                                          label={`${subnet.subnet_id} | ${subnet.subnet_name}`}></bk-option>
                                      ))}
                                    </bk-select>
                                  </div>
                                </bk-form-item>
                              </bk-form>
                            </>
                          </CommonCard>
                        </>
                      )}
                    </>
                  )}

                  <CommonCard title={() => '实例配置'}>
                    <>
                      {resourceForm.value.resourceType !== 'IDCPM' ? (
                        <>
                          <bk-form
                            class={'scr-form-wrapper'}
                            model={QCLOUDCVMForm.value.spec}
                            rules={QCLOUDCVMformRules.value}
                            ref={QCLOUDCVMformRef}>
                            <bk-form-item label='机型' required property='device_type'>
                              <bk-select
                                v-model={QCLOUDCVMForm.value.spec.device_type}
                                disabled={resourceForm.value.zone === ''}
                                onChange={onQcloudDeviceTypeChange}
                                placeholder={resourceForm.value.zone === '' ? '请先选择可用区' : '请选择机型'}
                                filterable>
                                {deviceTypes.value.map((deviceType) => (
                                  <bk-option key={deviceType} label={deviceType} value={deviceType} />
                                ))}
                              </bk-select>
                            </bk-form-item>
                            <bk-form-item label='镜像' required property='image_id'>
                              <bk-select
                                v-model={QCLOUDCVMForm.value.spec.image_id}
                                disabled={resourceForm.value.region === ''}
                                filterable>
                                {images.value.map((item) => (
                                  <bk-option key={item.image_id} label={item.image_name} value={item.image_id} />
                                ))}
                              </bk-select>
                            </bk-form-item>
                            <bk-form-item label='数据盘'>
                              <div
                                style={{
                                  display: 'flex',
                                  alignItems: 'center',
                                }}>
                                <DiskTypeSelect v-model={QCLOUDCVMForm.value.spec.disk_type}></DiskTypeSelect>
                                <Input
                                  class={'ml8'}
                                  type='number'
                                  v-model={QCLOUDCVMForm.value.spec.disk_size}
                                  min={1}></Input>
                                <span class={'ml8'}>G</span>
                              </div>
                            </bk-form-item>
                            <bk-form-item label='需求数量' required property='replicas'>
                              <Input type='number' v-model={QCLOUDCVMForm.value.spec.replicas} min={1}></Input>
                            </bk-form-item>
                            <bk-form-item label='备注'>
                              <Input
                                type='textarea'
                                v-model={resourceForm.value.remark}
                                autosize
                                resize={false}></Input>
                            </bk-form-item>
                          </bk-form>
                          {resourceForm.value.resourceType === 'QCLOUDCVM' && (
                            <>
                              {cvmCapacity.value.length ? (
                                <>
                                  {cvmCapacity.value.map((item) => (
                                    <div class={'tooltips'}>
                                      <span>{item?.zone || ''}最大可申请量 </span>
                                      <span class={'volumetip'}>{item?.max_num || 0}</span>
                                      <Popover trigger='hover' theme='light' disableTeleport={true} arrow={false}>
                                        {{
                                          default: () => (
                                            <span>
                                              {item?.max_info.length && (
                                                <span class={'calculationDetails'}>( 计算明细 )</span>
                                              )}
                                            </span>
                                          ),
                                          content: () => (
                                            <div class={'content'}>
                                              {item?.max_info.length &&
                                                item?.max_info.map((val: { key: any; value: any }) => (
                                                  <div>
                                                    <span class={'application'}> {val.key}</span>
                                                    <span class={'volumetip'}> {val.value}</span>
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
                                  <span class={'volumetip'}>0</span>
                                </div>
                              )}
                            </>
                          )}
                        </>
                      ) : (
                        <>
                          <bk-form
                            model={pmForm.value.spec}
                            rules={pmForm.value.rules}
                            class={'scr-form-wrapper'}
                            ref={IDCPMformRef}>
                            <bk-form-item label='机型' required property='device_type'>
                              <bk-select
                                v-model={pmForm.value.spec.device_type}
                                default-first-option
                                class='width-300 mr16'
                                filterable>
                                {pmForm.value.options.deviceTypes.map((deviceType: { device_type: any }) => (
                                  <bk-option
                                    key={deviceType.device_type}
                                    value={deviceType.device_type}
                                    label={deviceType.device_type}></bk-option>
                                ))}
                              </bk-select>
                            </bk-form-item>
                            <bk-form-item label='RAID 类型'>
                              <span> {pmForm.value.spec.raid_type || '-'}</span>
                            </bk-form-item>

                            <bk-form-item label='操作系统' required property='os_type'>
                              <bk-select class='width-400' v-model={pmForm.value.spec.os_type}>
                                {pmForm.value.options.osTypes.map((osType) => (
                                  <bk-option key={osType} value={osType} label={osType}></bk-option>
                                ))}
                              </bk-select>
                            </bk-form-item>
                            <bk-form-item label='运营商'>
                              <bk-select class='width-300' v-model={pmForm.value.spec.isp}>
                                <bk-option key='无' value='' label='无'></bk-option>
                                {pmForm.value.options.isps.map((isp) => (
                                  <bk-option key={isp} value={isp} label={isp}></bk-option>
                                ))}
                              </bk-select>
                            </bk-form-item>
                            <div class='displayflex'>
                              <bk-form-item label='需求数量' required property='replicas'>
                                <Input
                                  class='item-warp-replicas mr16'
                                  type='number'
                                  v-model={pmForm.value.spec.replicas}
                                  min={1}></Input>
                              </bk-form-item>
                              <bk-form-item label='反亲和性' required property='antiAffinityLevel'>
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
                                class='width-300'
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
