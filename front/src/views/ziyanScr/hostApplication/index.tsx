import { defineComponent, ref, onMounted, watch } from 'vue';
import { Tab, Input, Button, Sideslider } from 'bkui-vue';
import { BkTabPanel } from 'bkui-vue/lib/tab';
import CommonCard from '@/components/CommonCard';
import { useTable } from '@/hooks/useTable/useTable';
import './index.scss';
import MemberSelect from '@/components/MemberSelect';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import AreaSelector from './components/AreaSelector';
import ZoneSelector from './components/ZoneSelector';
// import CvmTypeSelector from './components/CvmTypeSelector';
// import ImageSelector from './components/ImageSelector';
import DiskTypeSelect from './components/DiskTypeSelect';
import AntiAffinityLevelSelect from './components/AntiAffinityLevelSelect';
import { RightShape, DownShape, Search } from 'bkui-vue/lib/icon';
import { useAccountStore } from '@/store';
import apiService from '@/api/scrApi';
export default defineComponent({
  setup() {
    const accountStore = useAccountStore();
    const addResourceRequirements = ref(false);
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
        expectTime: '',
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
    const { columns: CloudHostcolumns } = useColumns('CloudHost');
    const { columns: PhysicalMachinecolumns } = useColumns('PhysicalMachine');
    const { columns: DeviceQuerycolumns } = useColumns('DeviceQuerycolumns');
    const { columns: CVMApplicationcolumns } = useColumns('CVMApplication');
    const { CommonTable: DeviceQueryTable, getListData: DeviceQueryGetListData } = useTable({
      tableOptions: {
        columns: DeviceQuerycolumns,
      },
      requestOption: {
        dataPath: 'data.info',
      },
      scrConfig: () => {
        return {
          ScrSwitch: true,
          interface: {
            Parameters: {
              filter: undefined,
              page: { start: 0, limit: 10 },
            },
            filter: { simpleConditions: true, requestId: 'devices' },
            path: '/api/v1/woa/task/findmany/apply/device',
          },
        };
      },
    });
    const pageInfo = ref({
      limit: 10,
      start: 0,
      sort: '-capacity_flag',
    });
    const requestListParams = ref({
      filter: {
        condition: 'AND',
        rules: [
          {
            field: 'require_type',
            operator: 'equal',
            value: 1,
          },
          {
            field: 'label.device_group',
            operator: 'in',
            value: ['标准型'],
          },
        ],
      },
      page: pageInfo.value,
    });
    const { CommonTable: CVMApplicationTable, getListData: CVMApplicationGetListData } = useTable({
      tableOptions: {
        columns: [
          ...CVMApplicationcolumns,
          {
            label: '操作',
            width: 120,
            render: ({ row }) => {
              return (
                <Button text theme='primary' onClick={() => OneClickApplication(row)}>
                  一键申请
                </Button>
              );
            },
          },
        ],
      },
      requestOption: {
        dataPath: 'data.info',
      },
      scrConfig: () => {
        return {
          ScrSwitch: true,
          interface: {
            Parameters: {
              ...requestListParams.value,
            },
            path: '/api/v1/woa/config/findmany/config/cvm/device/detail',
          },
        };
      },
    });
    const PhysicalMachineoperation = ref({
      label: '操作',
      width: 120,
      render: () => {
        return (
          <>
            <Button text theme='primary'>
              克隆
            </Button>
            <Button text theme='primary'>
              修改
            </Button>
            <Button text theme='primary'>
              删除
            </Button>
          </>
        );
      },
    });
    const CloudHostoperation = ref({
      label: '操作',
      width: 120,
      render: () => {
        return (
          <>
            <Button text theme='primary'>
              克隆
            </Button>
            <Button text theme='primary'>
              修改
            </Button>
            <Button text theme='primary'>
              删除
            </Button>
          </>
        );
      },
    });
    // 添加按钮侧边栏公共表单对象
    const resourceForm = ref({
      resourceType: '', // 主机类型
      replicas: 1, // 需求数量
      remark: '', // 备注
    });
    // 侧边栏腾讯云CVM
    const QCLOUDCVMForm = ref({
      spec: {
        deviceType: '', // 机型
        region: '', // 地域
        zone: '', // 园区
        vpc: '', //  vpc
        subnet: '', //  子网
        imageId: '', // 镜像
        diskType: '', // 数据盘tyle
        diskSize: 0, // 数据盘size
        networkType: '万兆',
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
    // 反亲和性
    const antiAffinityLevel = ref('');
    // 侧边栏物理机CVM
    const pmForm = ref({
      spec: {
        deviceType: '', // 机型
        raidType: '', // RAID 类型
        osType: '', // 操作系统
        region: '', // 地域
        zone: '', // 园区
        isp: '', // 经营商
      },
      rules: {
        'spec.deviceType': [{ required: true, message: '请选择机型', trigger: 'blur' }],
        'spec.region': [{ required: true, message: '请选择地域', trigger: 'blur' }],
        replicas: [{ required: true, message: '请选择需求数量', trigger: 'blur' }],
        'spec.osType': [{ required: true, message: '请选择操作系统', trigger: 'blur' }],
        antiAffinityLevel: [{ required: true, message: '请选择反亲和性', trigger: 'blur' }],
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
    const dateRange = ref();
    const businessList = ref([]);
    // 网络信息开关
    const NIswitch = ref(true);
    watch(
      () => QCLOUDCVMForm.value.spec,
      () => {
        loadVpcs();
        loadSubnets();
        loadDeviceTypes();
        loadImages();
      },
    );
    // QCLOUDCVM云地域变化
    const onQcloudRegionChange = () => {
      loadImages();
      loadVpcs();
      QCLOUDCVMForm.value.spec.zone = '';
      QCLOUDCVMForm.value.spec.deviceType = '';
      QCLOUDCVMForm.value.spec.vpc = '';
    };
    const onQcloudVpcChange = () => {
      QCLOUDCVMForm.value.spec.subnet = '';
      loadSubnets();
    };
    // QCLOUDCVM可用区变化
    const onQcloudZoneChange = () => {
      QCLOUDCVMForm.value.spec.deviceType = '';
      loadDeviceTypes();
      loadSubnets();
    };
    // IDCPM云地域变化
    const onIdcpmRegionChange = () => {
      pmForm.value.spec.zone = '';
      pmForm.value.spec.deviceType = '';
    };
    // 获取QCLOUDCVM镜像列表
    const loadImages = async () => {
      const { info } = await apiService.getImages([QCLOUDCVMForm.value.spec.region]);
      images.value = info;
      if (QCLOUDCVMForm.value.spec.imageId === '') {
        QCLOUDCVMForm.value.spec.imageId = 'img-fjxtfi0n';
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
      pmForm.value.spec.raidType =
        pmForm.value.options.deviceTypes?.find((item) => item.device_type === pmForm.value.spec.deviceType)?.raid || '';
    };
    // 地域有数据时禁用cpu 和内存
    const handleCVMDeviceTypeChange = () => {
      device.value.filter.cpu = '';
      device.value.filter.mem = '';
      deviceConfigDisabled.value = device.value.filter.device_type.length > 0;
    };
    // 获取可用的IDCPM列表
    const IDCPMlist = () => {
      IDCPMDeviceTypes();
      IDCPMOsTypes();
      IDCPMIsps();
    };
    // 监听物理机机型变化
    watch(
      () => resourceForm.value.resourceType,
      () => {
        resourceForm.value.resourceType === 'IDCPM' && IDCPMlist();
      },
    );
    // 监听物理机机型变化
    watch(
      () => pmForm.value.spec.deviceType,
      () => {
        handleDeviceTypeChange();
      },
    );
    // 获取 QCLOUDCVM机型列表
    const loadDeviceTypes = async () => {
      const {
        spec: { zone, region },
      } = QCLOUDCVMForm.value;
      const params = {
        region: [region],
        zone: zone !== 'cvm_separate_campus' ? [zone] : undefined,
      };

      const { info } = await apiService.getDeviceTypes(params);
      deviceTypes.value = info || [];
    };
    // 获取 QCLOUDCVM  VPC列表
    const loadVpcs = async () => {
      const { info } = await apiService.getVpcs(QCLOUDCVMForm.value.spec.region);
      zoneTypes.value = info;
    };
    // 获取 QCLOUDCVM子网列表
    const loadSubnets = async () => {
      const { region, zone, vpc } = QCLOUDCVMForm.value.spec;
      const { info } = await apiService.getSubnets({
        region,
        zone,
        vpc,
      });

      subnetTypes.value = info || [];
    };
    const deviceTypeDisabled = ref(false);
    const deviceConfigDisabled = ref(false);
    const resourece = ref({
      filter: {
        filter: order.value.model.requireType,
        orderId: '',
        bkBizId: null,
        bkUsername: null,
        ip: null,
      },
      options: {
        requireTypes: [],
      },
      page: {
        start: 0,
      },
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
    const activeName = ref('HostApplication');
    const query = () => {
      resourece.value.page.start = 0;
      DeviceQueryGetListData();
    };
    const clearFilter = () => {
      resourece.value.filter = { bkBizId: '', orderId: '', bkUsername: [], requireType: '', ip: [] };
      DeviceQueryGetListData();
    };
    const getfetchOptionslist = async () => {
      const { info } = await apiService.getRequireTypes();
      order.value.options.requireTypes = info;
    };
    const getBusinessesList = async () => {
      const { data } = await accountStore.getBizListWithAuth();
      businessList.value = data;
    };
    // 一键申请按钮点击事件
    const clickApplication = () => {
      CVMapplication.value = true;
      CVMapplicationDeviceTypes();
      loadRestrict();
      CVMApplicationGetListData();
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
    // 获取一键申请侧边栏cpu
    const loadRestrict = async () => {
      const { cpu, mem } = await apiService.getRestrict();
      device.value.options.cpu = cpu || [];
      device.value.options.mem = mem || [];
    };
    // cpu 和内存有数据时禁用地域
    const handleDeviceConfigChange = () => {
      device.value.filter.device_type = [];
      const { cpu, mem } = device.value.filter;
      deviceTypeDisabled.value = Boolean(cpu || mem);
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
    // 一键申请提交按钮
    const filterDevices = () => {
      device.value.page.start = 0;
      loadResources();
    };

    // 提交接口
    const loadResources = () => {
      CVMApplicationGetListData();
    };
    const OneClickApplication = (row) => {
      CVMapplication.value = false;
      resourceForm.value.resourceType = 'QCLOUDCVM';
      QCLOUDCVMForm.value.spec = {
        deviceType: row.device_type, // 机型
        region: row.region, // 地域
        zone: row.zone, // 园区
        vpc: '', //  vpc
        subnet: '', //  子网
        imageId: 'img-fjxtfi0n', // 镜像
        diskType: 'CLOUD_PREMIUM', // 数据盘tyle
        diskSize: 0, // 数据盘size
        networkType: '万兆',
      };
      // QCLOUDCVMForm.value.spec.region = row.region;
      // QCLOUDCVMForm.value.spec.deviceType = row.device_type;
      // QCLOUDCVMForm.value.spec.zone = row.zone;
      // QCLOUDCVMForm.value.spec.imageId = 'img-fjxtfi0n';
      addResourceRequirements.value = true;
    };
    const ARtriggerShow = (isShow: boolean) => {
      addResourceRequirements.value = isShow;
    };
    const CAtriggerShow = (isShow: boolean) => {
      CVMapplication.value = isShow;
    };
    const handleSubmit = () => {
      if (resourceForm.value.resourceType === 'QCLOUDCVM') {
        cloudTableData.value.push({
          ...resourceForm.value,
          ...QCLOUDCVMForm.value,
        });
        resourceForm.value = {
          resourceType: '',
          replicas: 1,
          remark: '',
        };
        QCLOUDCVMForm.value = {
          spec: {
            deviceType: '', // 机型
            region: '', // 地域
            zone: '', // 园区
            vpc: '', //  vpc
            subnet: '', //  子网
            imageId: '', // 镜像
            diskType: '', // 数据盘tyle
            diskSize: 0, // 数据盘size
            networkType: '',
          },
        };
      } else {
        physicalTableData.value.push({
          ...resourceForm.value,
          ...pmForm.value.spec,
        });
        resourceForm.value = {
          resourceType: '',
          replicas: 1,
          remark: '',
        };
        pmForm.value.spec = {
          deviceType: '', // 机型
          raidType: '', // RAID 类型
          osType: '', // 操作系统
          region: '', // 地域
          zone: '', // 园区
          isp: '', // 经营商
        };
      }
      addResourceRequirements.value = false;
    };
    watch(
      () => activeName.value,
      (val: string) => {
        if (val === 'HostApplication') {
          getBusinessesList();
          getfetchOptionslist();
        }
      },
    );
    onMounted(() => {
      getBusinessesList();
      getfetchOptionslist();
    });

    return () => (
      <Tab v-model:active={activeName.value} type='unborder-card'>
        <BkTabPanel key='HostApplication' name='HostApplication' label='主机申请'>
          <div class='wid100'>
            <CommonCard class='mt15 ml16' title={() => '基本信息'} layout='grid'>
              <bk-form
                form-type='vertical'
                label-width='150'
                model={order.value.model}
                rules={order.value.rules}
                ref='formRef'>
                <div class='displayflex'>
                  <bk-form-item label='所属业务' class='item-warp' required property='bkBizId'>
                    <bk-select class='item-warp-component' v-model={order.value.model.bkBizId}>
                      {businessList.value.map((item) => (
                        <bk-option key={item.id} value={item.id} label={item.name}></bk-option>
                      ))}
                    </bk-select>
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
                  <bk-form-item label='关注人' class='item-warp' property='follower'>
                    <MemberSelect class='item-warp-component' multiple clearable v-model={order.value.model.follower} />
                    {/* <bk-select disabled class='item-warp-component' v-model={order.model.follower}>
              {accountList.map((item, index) => (
                <bk-option key={index} value={item.id} label={item.name}></bk-option>
              ))}
            </bk-select> */}
                  </bk-form-item>
                </div>
              </bk-form>
            </CommonCard>
            <CommonCard class='ml16 mt15' title={() => '配置清单'} layout='grid'>
              <div>
                <Button
                  class='mr16'
                  theme='primary'
                  onClick={() => {
                    addResourceRequirements.value = true;
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
                      />
                    </div>
                  </>
                ) : (
                  <></>
                )}
              </div>
            </CommonCard>
            <CommonCard class='ml16' title={() => '备注'} layout='grid'>
              <bk-form
                form-type='vertical'
                label-width='150'
                model={order.value.model}
                rules={order.value.rules}
                ref='formRef'>
                <bk-form-item label='申请的备注' class='item-warp' property='bkBizId'>
                  <Input
                    type='textarea'
                    v-model={order.value.model.remark}
                    rows={3}
                    maxlength={255}
                    resize={false}
                    placeholder='请输入申请单备注'></Input>
                </bk-form-item>
                <bk-form-item class='item-warp' required property='bkBizId'>
                  <Button class='mr16' theme='primary' onClick={() => {}}>
                    提交
                  </Button>
                  <Button onClick={() => {}}>取消</Button>
                </bk-form-item>
              </bk-form>
            </CommonCard>
            <Sideslider
              class='common-sideslider'
              width={640}
              isShow={addResourceRequirements.value}
              title='增加资源需求'
              onClosed={() => {
                ARtriggerShow(false);
              }}>
              {{
                default: () => (
                  <div class='common-sideslider-content'>
                    <CommonCard title={() => '基本信息'} layout='grid'>
                      <bk-form
                        form-type='vertical'
                        label-width='150'
                        model={resourceForm.value}
                        rules={resourceFormRules.value}
                        ref='formRef'>
                        <div class='displayflex'>
                          <bk-form-item label='主机类型' required>
                            <bk-select class='item-warp-resourceType' v-model={resourceForm.value.resourceType}>
                              {resourceTypes.value.map((resType: { value: any; label: any }) => (
                                <bk-option key={resType.value} value={resType.value} label={resType.label}></bk-option>
                              ))}
                            </bk-select>
                          </bk-form-item>
                        </div>
                        {resourceForm.value.resourceType === 'QCLOUDCVM' ? (
                          <div class='displayflex'>
                            <bk-form-item class='mr16' label='云地域' required>
                              <AreaSelector
                                ref='areaSelector'
                                v-model={QCLOUDCVMForm.value.spec.region}
                                class='item-warp-qcloudRegionId'
                                params={{ resourceType: resourceForm.value.resourceType }}
                                onChange={onQcloudRegionChange}></AreaSelector>
                            </bk-form-item>
                            <bk-form-item label='可用区' property='zone'>
                              <ZoneSelector
                                ref='zoneSelector'
                                v-model={QCLOUDCVMForm.value.spec.zone}
                                class='item-warp-qcloudZoneId'
                                params={{
                                  resourceType: resourceForm.value.resourceType,
                                  region: QCLOUDCVMForm.value.spec.region,
                                }}
                                onChange={onQcloudZoneChange}
                              />
                            </bk-form-item>
                          </div>
                        ) : (
                          <div class='displayflex'>
                            <bk-form-item class='mr16' label='云地域' required>
                              <AreaSelector
                                ref='areaSelector'
                                v-model={pmForm.value.spec.region}
                                class='item-warp-qcloudRegionId'
                                params={{ resourceType: resourceForm.value.resourceType }}
                                onChange={onIdcpmRegionChange}></AreaSelector>
                            </bk-form-item>
                            <bk-form-item label='可用区' property='zone'>
                              <ZoneSelector
                                ref='zoneSelector'
                                v-model={pmForm.value.spec.zone}
                                class='item-warp-qcloudZoneId'
                                params={{
                                  resourceType: resourceForm.value.resourceType,
                                  region: pmForm.value.spec.region,
                                }}
                              />
                            </bk-form-item>
                          </div>
                        )}
                      </bk-form>
                    </CommonCard>
                    {resourceForm.value.resourceType === 'QCLOUDCVM' && (
                      <>
                        {NIswitch.value ? (
                          <>
                            <CommonCard
                              class='mt15'
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
                              class='mt15'
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
                                        disabled={QCLOUDCVMForm.value.spec.zone === 'cvm_separate_campus'}
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
                                        disabled={QCLOUDCVMForm.value.spec.zone === 'cvm_separate_campus'}
                                        v-model={QCLOUDCVMForm.value.spec.subnet}>
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

                    <CommonCard class='mt15' title={() => '实例配置'} layout='grid'>
                      <>
                        {resourceForm.value.resourceType !== 'IDCPM' ? (
                          <>
                            <bk-form
                              form-type='vertical'
                              label-width='150'
                              model={QCLOUDCVMForm.value.spec}
                              rules={resourceFormRules.value}
                              ref='formRef'>
                              <bk-form-item label='机型' required>
                                <bk-select
                                  class='item-warp-resourceType'
                                  v-model={QCLOUDCVMForm.value.spec.deviceType}
                                  disabled={QCLOUDCVMForm.value.spec.zone === ''}
                                  placeholder={QCLOUDCVMForm.value.spec.zone === '' ? '请先选择可用区' : '请选择机型'}
                                  filterable>
                                  {deviceTypes.value.map((deviceType) => (
                                    <bk-option key={deviceType} label={deviceType} value={deviceType} />
                                  ))}
                                </bk-select>
                              </bk-form-item>
                              <bk-form-item label='镜像' required>
                                <bk-select
                                  class='item-warp-resourceType'
                                  v-model={QCLOUDCVMForm.value.spec.imageId}
                                  disabled={QCLOUDCVMForm.value.spec.region === ''}
                                  filterable>
                                  {images.value.map((item) => (
                                    <bk-option key={item.image_id} label={item.image_name} value={item.image_id} />
                                  ))}
                                </bk-select>
                                {/* <ImageSelector
                                  class='item-warp-imageId'
                                  ref='imageSelector'
                                  v-model={resourceForm.value.selector.imageId}
                                  area={resourceForm.value.selector.qcloudRegionId}
                                /> */}
                              </bk-form-item>
                              <bk-form-item label='数据盘' required>
                                <div class='displayflex'>
                                  <DiskTypeSelect v-model={QCLOUDCVMForm.value.spec.diskType}></DiskTypeSelect>
                                  <Input
                                    type='number'
                                    class='item-warp-dataDisk-diskSize'
                                    v-model={QCLOUDCVMForm.value.spec.diskSize}
                                    min={1}></Input>
                                  <span>G</span>
                                </div>
                              </bk-form-item>
                              <bk-form-item label='需求数量' required property='resourceType'>
                                <Input
                                  class='item-warp-replicas'
                                  type='number'
                                  v-model={resourceForm.value.replicas}
                                  min={1}></Input>
                              </bk-form-item>
                              <bk-form-item label='备注' property='remark'>
                                <Input
                                  type='textarea'
                                  v-model={resourceForm.value.remark}
                                  rows={3}
                                  maxlength={255}
                                  resize={false}
                                  placeholder='请输入备注'></Input>
                              </bk-form-item>
                            </bk-form>
                          </>
                        ) : (
                          <>
                            <bk-form
                              form-type='vertical'
                              label-width='150'
                              model={pmForm.value.spec}
                              rules={resourceFormRules.value}
                              ref='formRef'>
                              <div class='displayflex'>
                                <bk-form-item label='机型' required>
                                  <bk-select
                                    v-model={pmForm.value.spec.deviceType}
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
                                <bk-form-item label='RAID 类型' prop='spec.raidType'>
                                  <span> {pmForm.value.spec.raidType || '-'}</span>
                                </bk-form-item>
                              </div>
                              <bk-form-item label='操作系统' required>
                                <bk-select class='width-400' v-model={pmForm.value.spec.osType}>
                                  {pmForm.value.options.osTypes.map((osType) => (
                                    <bk-option key={osType} value={osType} label={osType}></bk-option>
                                  ))}
                                </bk-select>
                              </bk-form-item>
                              <bk-form-item label='经营商' required>
                                <bk-select class='width-300' v-model={pmForm.value.spec.isp}>
                                  {pmForm.value.options.isps.map((isp) => (
                                    <bk-option key={isp} value={isp} label={isp}></bk-option>
                                  ))}
                                </bk-select>
                              </bk-form-item>
                              <div class='displayflex'>
                                <bk-form-item label='需求数量' required>
                                  <Input
                                    class='item-warp-replicas mr16'
                                    type='number'
                                    v-model={resourceForm.value.replicas}
                                    min={1}></Input>
                                </bk-form-item>
                                <bk-form-item label='反亲和性' required>
                                  <AntiAffinityLevelSelect
                                    v-model={antiAffinityLevel.value}
                                    params={{
                                      resourceType: resourceForm.value.resourceType,
                                      hasZone: pmForm.value.spec.zone !== '',
                                    }}></AntiAffinityLevelSelect>
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
            <Sideslider
              class='common-sideslider'
              width={1100}
              isShow={CVMapplication.value}
              title='CVM一键申请'
              onClosed={() => {
                CAtriggerShow(false);
              }}>
              {{
                default: () => (
                  <div class='common-sideslider-content'>
                    <CVMApplicationTable>
                      {{
                        tabselect: () => (
                          <>
                            <div class='tabselect'>
                              <span class='label'>需求类型</span>
                              <bk-select
                                class='tbkselect'
                                disabled
                                v-model={device.value.filter.require_type}
                                filterable>
                                {order.value.options.requireTypes.map((item) => (
                                  <bk-option
                                    key={item.require_type}
                                    value={item.require_type}
                                    label={item.require_name}></bk-option>
                                ))}
                              </bk-select>
                            </div>
                            <div class='tabselect'>
                              <span class='label'>地域</span>
                              <AreaSelector
                                ref='areaSelector'
                                class='tbkselect'
                                v-model={device.value.filter.region}
                                multiple
                                clearable
                                filterable
                                params={{ resourceType: 'QCLOUDCVM' }}></AreaSelector>
                            </div>
                            <div class='tabselect'>
                              <span class='label'>园区</span>
                              <ZoneSelector
                                ref='zoneSelector'
                                v-model={device.value.filter.zone}
                                class='tbkselect'
                                params={{
                                  resourceType: 'QCLOUDCVM',
                                  region: device.value.filter.region,
                                }}></ZoneSelector>
                            </div>
                            <div class='tabselect'>
                              <span class='label'>实例族</span>
                              <bk-select
                                class='tbkselect'
                                v-model={device.value.filter.device_group}
                                multiple
                                clearable
                                collapse-tags
                                onChange={handleDeviceGroupChange}>
                                {device.value.options.device_groups.map((item) => (
                                  <bk-option key={item} value={item} label={item}></bk-option>
                                ))}
                              </bk-select>
                            </div>
                            <div class='tabselect'>
                              <span class='label'>机型</span>
                              <bk-select
                                class='tbkselect'
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
                            </div>
                            <div class='tabselect'>
                              <span class='label'>CPU(核)</span>
                              <bk-select
                                class='tbkselect'
                                v-model={device.value.filter.cpu}
                                clearable
                                disabled={deviceConfigDisabled.value}
                                filterable
                                onChange={handleDeviceConfigChange}>
                                {device.value.options.cpu.map((item) => (
                                  <bk-option key={item} value={item} label={item}></bk-option>
                                ))}
                              </bk-select>
                            </div>
                            <div class='tabselect'>
                              <span class='label'>内存 (G)</span>
                              <bk-select
                                class='tbkselect'
                                v-model={device.value.filter.mem}
                                clearable
                                disabled={deviceConfigDisabled.value}
                                filterable
                                onChange={handleDeviceConfigChange}>
                                {device.value.options.mem.map((item) => (
                                  <bk-option key={item} value={item} label={item}></bk-option>
                                ))}
                              </bk-select>
                            </div>
                            <div class='tabselect'>
                              <bk-button icon='bk-icon-search' theme='primary' class='bkbutton' onClick={filterDevices}>
                                <Search></Search>
                                查询
                              </bk-button>
                              <bk-button icon='bk-icon-refresh' onClick={CVMclearFilter}>
                                清空
                              </bk-button>
                            </div>
                          </>
                        ),
                      }}
                    </CVMApplicationTable>
                  </div>
                ),
              }}
            </Sideslider>
          </div>
        </BkTabPanel>
        <BkTabPanel key='DeviceQuery' name='DeviceQuery' label='设备查询'>
          <div>
            <DeviceQueryTable>
              {{
                tabselect: () => (
                  <>
                    <bk-form label-width='110' class='bill-filter-form' model={resourece}>
                      <bk-form-item label='业务'>
                        <bk-select v-model={resourece.value.filter.bkBizId} multiple clearable placeholder='请选择业务'>
                          {/* {bussinessList.map(({ key, value }) => {
                            return <bk-option key={key} label={value} value={key}></bk-option>;
                          })} */}
                        </bk-select>
                      </bk-form-item>
                      <bk-form-item label='需求类型'>
                        <bk-select class='tbkselect' v-model={resourece.value.filter.requireType} filterable>
                          {resourece.value.options.requireTypes.map((item) => (
                            <bk-option
                              key={item.require_type}
                              value={item.require_name}
                              label={item.require_type}></bk-option>
                          ))}
                        </bk-select>
                      </bk-form-item>
                      <bk-form-item label='单号'>
                        <bk-input
                          v-model={resourece.value.filter.orderId}
                          clearable
                          type='number'
                          placeholder='请输入单号'></bk-input>
                      </bk-form-item>
                      <bk-form-item label='申请人'>
                        <bk-select v-model={resourece.value.filter.bkUsername} multiple clearable>
                          {/* {recycleMen.map(({ key, value }) => {
                            return <bk-option key={key} label={value} value={key}></bk-option>;
                          })} */}
                        </bk-select>
                      </bk-form-item>
                      <bk-form-item label='交付时间'>
                        <bk-date-picker v-model={dateRange.value} type='daterange' />
                      </bk-form-item>
                      <bk-form-item label='内网 IP'>
                        <bk-input v-model={resourece.value.filter.ip} clearable></bk-input>
                      </bk-form-item>
                      <bk-form-item class='bill-form-btn' label-width='20'>
                        <bk-button theme='primary' native-type='submit' onClick={query}>
                          查询
                        </bk-button>
                        <bk-button onClick={clearFilter}>清空</bk-button>
                        <bk-button onClick={clearFilter}>导出</bk-button>
                        <bk-button>导出全部</bk-button>
                        <bk-button>复制</bk-button>
                      </bk-form-item>
                    </bk-form>
                  </>
                ),
              }}
            </DeviceQueryTable>
          </div>
        </BkTabPanel>
      </Tab>
    );
  },
});
