import { defineComponent, onMounted, ref, watch } from 'vue';
import { Input, Button, Sideslider, Message } from 'bkui-vue';
import CommonCard from '@/components/CommonCard';
import './index.scss';
import MemberSelect from '@/components/MemberSelect';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import AreaSelector from '../AreaSelector';
import ZoneSelector from '../ZoneSelector';
import DiskTypeSelect from '../DiskTypeSelect';
import AntiAffinityLevelSelect from '../AntiAffinityLevelSelect';
import { RightShape, DownShape } from 'bkui-vue/lib/icon';
import apiService from '@/api/scrApi';
import { useAccountStore, useUserStore } from '@/store';
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
    const router = useRouter();
    const route = useRoute();
    const addResourceRequirements = ref(false);
    const isLoading = ref(false);
    const title = ref('增加资源需求');
    const CVMapplication = ref(false);
    const accountStore = useAccountStore();
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
      resourceType: '', // 主机类型
      replicas: 1, // 需求数量
      remark: '', // 备注
      anti_affinity_level: 'ANTI_NONE',
      enable_disk_check: false,
    });
    // 侧边栏腾讯云CVM
    const QCLOUDCVMForm = ref({
      spec: {
        device_type: '', // 机型
        region: '', // 地域
        zone: '', // 园区
        vpc: '', //  vpc
        subnet: '', //  子网
        image_id: '', // 镜像
        disk_type: '', // 数据盘tyle
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
        region: '', // 地域
        zone: '', // 园区
        isp: '', // 经营商
        network_type: 'TENTHOUSAND',
      },
      rules: {
        'spec.device_type': [{ required: true, message: '请选择机型', trigger: 'blur' }],
        'spec.region': [{ required: true, message: '请选择地域', trigger: 'blur' }],
        replicas: [{ required: true, message: '请选择需求数量', trigger: 'blur' }],
        'spec.os_type': [{ required: true, message: '请选择操作系统', trigger: 'blur' }],
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
    // const dateRange = ref();
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
      QCLOUDCVMForm.value.spec.device_type = '';
      QCLOUDCVMForm.value.spec.vpc = '';
    };
    const onQcloudVpcChange = () => {
      QCLOUDCVMForm.value.spec.subnet = '';
      loadSubnets();
    };
    // QCLOUDCVM可用区变化
    const onQcloudZoneChange = () => {
      QCLOUDCVMForm.value.spec.device_type = '';
      loadDeviceTypes();
      loadSubnets();
    };
    // IDCPM云地域变化
    const onIdcpmRegionChange = () => {
      pmForm.value.spec.zone = '';
      pmForm.value.spec.device_type = '';
    };
    // 获取QCLOUDCVM镜像列表
    const loadImages = async () => {
      const { info } = await apiService.getImages([QCLOUDCVMForm.value.spec.region]);
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
    // 监听物理机机型变化
    watch(
      () => resourceForm.value.resourceType,
      () => {
        resourceForm.value.resourceType === 'IDCPM' && IDCPMlist();
      },
    );
    // 监听物理机机型变化
    watch(
      () => pmForm.value.spec.device_type,
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
    const getBusinessesList = async () => {
      const { data } = await accountStore.getBizListWithAuth();
      businessList.value = data;
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
      if (resourceType === 'QCLOUDCVM') {
        QCLOUDCVMForm.value.spec = cloudTableData.value[index].spec;
      } else {
        pmForm.value.spec = physicalTableData.value[index].spec;
      }
      resourceForm.value.resourceType = resourceType;
      modifyresourceType.value = resourceType;
      resourceForm.value.remark = row.remark;
      resourceForm.value.replicas = row.replicas;
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
      getBusinessesList();
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
    const assignment = (data) => {
      resourceForm.value.resourceType = 'QCLOUDCVM';
      QCLOUDCVMForm.value.spec = {
        device_type: data.device_type, // 机型
        region: data.region, // 地域
        zone: data.zone, // 园区
        vpc: '', //  vpc
        subnet: '', //  子网
        image_id: 'img-fjxtfi0n', // 镜像
        disk_type: 'CLOUD_PREMIUM', // 数据盘tyle
        disk_size: 0, // 数据盘size
        network_type: 'TENTHOUSAND',
      };
    };
    const OneClickApplication = (row, val) => {
      CVMapplication.value = val;

      assignment(row);
      title.value = '增加资源需求';
      addResourceRequirements.value = true;
    };
    const ARtriggerShow = (isShow: boolean) => {
      emptyform();
      addResourceRequirements.value = isShow;
    };
    const CAtriggerShow = (isShow: boolean) => {
      CVMapplication.value = isShow;
    };
    const emptyform = () => {
      resourceForm.value = {
        resourceType: '',
        replicas: 1,
        remark: '',
        anti_affinity_level: 'ANTI_NONE',
        enable_disk_check: false,
      };
      QCLOUDCVMForm.value = {
        spec: {
          device_type: '', // 机型
          region: '', // 地域
          zone: '', // 园区
          vpc: '', //  vpc
          subnet: '', //  子网
          image_id: '', // 镜像
          disk_type: '', // 数据盘tyle
          disk_size: 0, // 数据盘size
          network_type: 'TENTHOUSAND',
        },
      };
      pmForm.value.spec = {
        device_type: '', // 机型
        raid_type: '', // RAID 类型
        os_type: '', // 操作系统
        region: '', // 地域
        zone: '', // 园区
        isp: '', // 运营商
        network_type: 'TENTHOUSAND',
      };
    };
    const cloudResourceForm = () => {
      return {
        resource_type: resourceForm.value.resourceType,
        remark: resourceForm.value.remark,
        enable_disk_check: resourceForm.value.enable_disk_check,
        anti_affinity_level: resourceForm.value.anti_affinity_level,
        replicas: resourceForm.value.replicas,
        ...QCLOUDCVMForm.value,
      };
    };
    const PMResourceForm = () => {
      return {
        resource_type: resourceForm.value.resourceType,
        remark: resourceForm.value.remark,
        anti_affinity_level: resourceForm.value.anti_affinity_level,
        replicas: resourceForm.value.replicas,
        spec: {
          ...pmForm.value.spec,
        },
      };
    };
    const handleSubmit = () => {
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
      addResourceRequirements.value = false;
    };
    const handleSaveOrSubmit = async (type: 'save' | 'submit') => {
      isLoading.value = true;
      try {
        const url = `${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/${
          type === 'submit' ? 'task/create/apply' : 'task/update/apply/ticket'
        }`;
        await http.post(url, {
          bk_biz_id: order.value.model.bkBizId,
          bk_username: useUserStore().username,
          require_type: order.value.model.requireType,
          // enable_notice: order.value.model.enableNotice,
          expect_time: timeFormatter(order.value.model.expectTime),
          remark: order.value.model.remark,
          follower: order.value.model.follower,
          suborders: [...cloudTableData.value, ...physicalTableData.value].map((v) => {
            return convertKeysToSnakeCase(v);
          }),
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
            <bk-form
              form-type='vertical'
              label-width='150'
              model={order.value.model}
              rules={order.value.rules}
              ref='formRef'>
              <bk-form-item label='申请备注' class='item-warp' property='bkBizId'>
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
                  loading={isLoading.value}
                  onClick={() => {
                    handleSaveOrSubmit('submit');
                  }}>
                  提交
                </Button>
                <Button
                  loading={isLoading.value}
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
                      rules={resourceFormRules.value}
                      ref='formRef'>
                      <bk-form-item label='主机类型' required>
                        <bk-select v-model={resourceForm.value.resourceType}>
                          {resourceTypes.value.map((resType: { value: any; label: any }) => (
                            <bk-option key={resType.value} value={resType.value} label={resType.label}></bk-option>
                          ))}
                        </bk-select>
                      </bk-form-item>
                      {resourceForm.value.resourceType === 'QCLOUDCVM' ? (
                        <>
                          <bk-form-item class='mr16' label='云地域' required>
                            <AreaSelector
                              ref='areaSelector'
                              v-model={QCLOUDCVMForm.value.spec.region}
                              params={{ resourceType: resourceForm.value.resourceType }}
                              onChange={onQcloudRegionChange}></AreaSelector>
                          </bk-form-item>
                          <bk-form-item label='可用区' property='zone'>
                            <ZoneSelector
                              ref='zoneSelector'
                              v-model={QCLOUDCVMForm.value.spec.zone}
                              params={{
                                resourceType: resourceForm.value.resourceType,
                                region: QCLOUDCVMForm.value.spec.region,
                              }}
                              onChange={onQcloudZoneChange}
                            />
                          </bk-form-item>
                        </>
                      ) : (
                        <>
                          <bk-form-item class='mr16' label='云地域' required>
                            <AreaSelector
                              ref='areaSelector'
                              v-model={pmForm.value.spec.region}
                              params={{ resourceType: resourceForm.value.resourceType }}
                              onChange={onIdcpmRegionChange}></AreaSelector>
                          </bk-form-item>
                          <bk-form-item label='可用区' property='zone'>
                            <ZoneSelector
                              ref='zoneSelector'
                              v-model={pmForm.value.spec.zone}
                              params={{
                                resourceType: resourceForm.value.resourceType,
                                region: pmForm.value.spec.region,
                              }}
                            />
                          </bk-form-item>
                        </>
                      )}
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

                  <CommonCard title={() => '实例配置'}>
                    <>
                      {resourceForm.value.resourceType !== 'IDCPM' ? (
                        <>
                          <bk-form
                            class={'scr-form-wrapper'}
                            model={QCLOUDCVMForm.value.spec}
                            rules={resourceFormRules.value}
                            ref='formRef'>
                            <bk-form-item label='机型' required>
                              <bk-select
                                v-model={QCLOUDCVMForm.value.spec.device_type}
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
                                v-model={QCLOUDCVMForm.value.spec.image_id}
                                disabled={QCLOUDCVMForm.value.spec.region === ''}
                                filterable>
                                {images.value.map((item) => (
                                  <bk-option key={item.image_id} label={item.image_name} value={item.image_id} />
                                ))}
                              </bk-select>
                            </bk-form-item>
                            <bk-form-item label='数据盘' required>
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
                            <bk-form-item label='需求数量' required property='resourceType'>
                              <Input type='number' v-model={resourceForm.value.replicas} min={1}></Input>
                            </bk-form-item>
                            <bk-form-item label='备注' property='remark'>
                              <Input
                                type='textarea'
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
                            rules={resourceFormRules.value}
                            class={'scr-form-wrapper'}
                            ref='formRef'>
                            <div>
                              <bk-form-item label='机型' required>
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
                              <bk-form-item label='RAID 类型' prop='spec.raid_type'>
                                <span> {pmForm.value.spec.raid_type || '-'}</span>
                              </bk-form-item>
                            </div>
                            <bk-form-item label='操作系统' required>
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
                              <bk-form-item label='需求数量' required>
                                <Input
                                  class='item-warp-replicas mr16'
                                  type='number'
                                  v-model={resourceForm.value.replicas}
                                  min={1}></Input>
                              </bk-form-item>
                              <bk-form-item label='反亲和性' required>
                                <AntiAffinityLevelSelect
                                  v-model={resourceForm.value.anti_affinity_level}
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

          {/* CVM一键申请 */}
          <Sideslider
            class='common-sideslider'
            width={1100}
            isShow={CVMapplication.value}
            title='CVM一键申请'
            onClosed={() => {
              CAtriggerShow(false);
            }}>
            <applicationSideslider onOneApplication={OneClickApplication} />
          </Sideslider>
        </div>
      </div>
    );
  },
});
