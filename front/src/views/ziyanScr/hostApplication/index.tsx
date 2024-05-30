import { defineComponent, ref } from 'vue';
import CommonCard from '@/components/CommonCard';
import { Input, Button } from 'bkui-vue';
import { useTable } from '@/hooks/useTable/useTable';
import CommonSideslider from '@/components/common-sideslider';
import './index.scss';
import MemberSelect from '@/components/MemberSelect';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import AreaSelector from './components/AreaSelector';
import ZoneSelector from './components/ZoneSelector';
import CvmTypeSelector from './components/CvmTypeSelector';
import ImageSelector from './components/ImageSelector';
import AntiAffinityLevelSelect from './components/AntiAffinityLevelSelect';
import { RightShape, DownShape } from 'bkui-vue/lib/icon';
import { useziyanScrStore } from '@/store';
export default defineComponent({
  name: 'AllhostInventoryManager',
  setup() {
    const addResourceRequirements = ref(false);
    const order = {
      loading: false,
      submitting: false,
      saving: false,
      model: {
        bkBizId: '',
        bkUsername: 'mjw',
        requireType: '',
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
        requireTypes: [] as any,
      },
    };
    const { columns: CloudHostcolumns } = useColumns('CloudHost');
    const { columns: PhysicalMachinecolumns } = useColumns('PhysicalMachine');
    const { CommonTable: CloudHostTable } = useTable({
      tableOptions: {
        columns: [
          ...CloudHostcolumns,
          {
            label: '操作',
            width: 120,
            render: () => {
              return (
                <>
                  <Button text theme='primary'>
                    克隆
                  </Button>
                  <Button text theme='primary'>
                    编辑
                  </Button>
                </>
              );
            },
          },
        ],
      },
      requestOption: {
        type: 'load_balancers/with/delete_protection',
        sortOption: { sort: 'created_at', order: 'DESC' },
      },
      slotAllocation: () => {
        return {
          ScrSwitch: false,
          interface: {
            Parameters: {
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
              page: [],
            },
            filter: { simpleConditions: true, requestId: 'devices' },
            path: '/api/v1/woa/config/findmany/config/cvm/device/detail',
          },
        };
      },
    });
    const { CommonTable: PhysicalMachineTable } = useTable({
      tableOptions: {
        columns: [
          ...PhysicalMachinecolumns,
          {
            label: '操作',
            width: 120,
            render: () => {
              return (
                <>
                  <Button text theme='primary'>
                    克隆
                  </Button>
                  <Button text theme='primary'>
                    编辑
                  </Button>
                </>
              );
            },
          },
        ],
      },
      requestOption: {
        type: 'load_balancers/with/delete_protection',
        sortOption: { sort: 'created_at', order: 'DESC' },
      },
      slotAllocation: () => {
        return {
          ScrSwitch: false,
          interface: {
            Parameters: {
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
              page: [],
            },
            filter: { simpleConditions: true, requestId: 'devices' },
            path: '/api/v1/woa/config/findmany/config/cvm/device/detail',
          },
        };
      },
    });
    const resourceForm = ref({
      resourceType: '',
      selector: {
        qcloudZoneId: '',
        deviceClass: '',
        qcloudRegionId: '',
        vpcId: '',
        subnetId: '',
        imageId: '',
        dataDisk: [{ diskType: 'CLOUD_PREMIUM', diskSize: 0 }],
      },
      replicas: 1,
      remark: '',
    });
    const pmForm = ref({
      model: {
        replicas: 0,
        antiAffinityLevel: '',
        remark: '',
        spec: {
          deviceType: '',
          raidType: '',
          osType: '',
          region: '',
          zone: '',
          isp: '',
        },
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
    const subnets = ref([]);
    const resourceFormRules = ref({});
    const vpcs = ref([]);
    // 物理机表格开关
    const PMswitch = ref(false);
    // 网络信息开关
    const NIswitch = ref(true);
    const onQcloudRegionChange = () => {
      loadVpcs();
      resourceForm.value.selector.qcloudZoneId = '';
      resourceForm.value.selector.deviceClass = '';
    };
    const onQcloudZoneChange = () => {
      resourceForm.value.selector.vpcId = '';
      resourceForm.value.selector.subnetId = '';
      loadSubnets();
    };
    const loadSubnets = () => {
      const { qcloudRegionId, qcloudZoneId, vpcId } = resourceForm.value.selector;

      useziyanScrStore()
        .listSubnet({
          region: qcloudRegionId,
          zone: qcloudZoneId,
          vpcId,
        })
        .then(({ data }: any) => {
          subnets.value =
            data?.filter(
              ({ subnetName }: any) => !subnetName.includes('tenc_docker_') && !subnetName.includes('tenc_tke_'),
            ) || [];
        });
    };
    const loadVpcs = () => {
      useziyanScrStore()
        .listVpc(resourceForm.value.selector.qcloudRegionId)
        .then(({ data }: any) => {
          vpcs.value = data;
        });
    };
    return () => (
      <div class='wid100'>
        <CommonCard class='mt15 ml16' title={() => '基本信息'} layout='grid'>
          <bk-form form-type='vertical' label-width='150' model={order.model} rules={order.rules} ref='formRef'>
            <div class='displayflex'>
              <bk-form-item label='所属业务' class='item-warp' required property='bkBizId'>
                <bk-select class='item-warp-component' v-model={order.model.bkBizId}>
                  {/* {accountList.map((item, index) => (
                <bk-option key={index} value={item.id} label={item.name}></bk-option>
              ))} */}
                </bk-select>
              </bk-form-item>
              <bk-form-item label='需求类型' class='item-warp' required property='requireType'>
                <bk-select class='item-warp-component' v-model={order.model.requireType}>
                  {order.options.requireTypes.map((item: { id: any; name: any }, index: any) => (
                    <bk-option key={index} value={item.id} label={item.name}></bk-option>
                  ))}
                </bk-select>
              </bk-form-item>
            </div>
            <div class='displayflex'>
              <bk-form-item label='期望交付时间' class='item-warp' required property='expectTime'>
                <bk-date-picker
                  class='item-warp-component'
                  v-model={order.model.expectTime}
                  clearable
                  type='datetime'></bk-date-picker>
              </bk-form-item>
              <bk-form-item label='关注人' class='item-warp' property='follower'>
                <MemberSelect class='item-warp-component' v-model={order.model.follower} />
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
            <Button
              onClick={() => {
                addResourceRequirements.value = true;
              }}>
              一键申请
            </Button>
            <div class='mt16'>云主机</div>
            <div class='mt16'>
              <CloudHostTable></CloudHostTable>
            </div>
            {PMswitch.value ? (
              <>
                <div class='mt16'>物理机</div>
                <div class='mt16'>
                  <PhysicalMachineTable></PhysicalMachineTable>
                </div>
              </>
            ) : (
              <></>
            )}
          </div>
        </CommonCard>
        <CommonCard class='ml16' title={() => '备注'} layout='grid'>
          <bk-form form-type='vertical' label-width='150' model={order.model} rules={order.rules} ref='formRef'>
            <bk-form-item label='申请的备注' class='item-warp' property='bkBizId'>
              <Input
                type='textarea'
                v-model={order.model.remark}
                rows={3}
                maxlength={255}
                resize={false}
                placeholder='请输入申请单备注'></Input>
            </bk-form-item>
            <bk-form-item class='item-warp' required property='bkBizId'>
              <Button
                class='mr16'
                theme='primary'
                onClick={() => {
                  addResourceRequirements.value = true;
                }}>
                提交
              </Button>
              <Button
                onClick={() => {
                  addResourceRequirements.value = true;
                }}>
                取消
              </Button>
            </bk-form-item>
          </bk-form>
        </CommonCard>
        <CommonSideslider v-model:isShow={addResourceRequirements.value} title='增加资源需求' width={640}>
          <CommonCard title={() => '基本信息'} layout='grid'>
            <bk-form
              form-type='vertical'
              label-width='150'
              model={resourceForm.value}
              rules={resourceFormRules.value}
              ref='formRef'>
              <div class='displayflex'>
                <bk-form-item label='主机类型' required property='resourceType'>
                  <bk-select class='item-warp-resourceType' v-model={resourceForm.value.resourceType}>
                    {resourceTypes.value.map((resType) => (
                      <bk-option key={resType.value} value={resType.value} label={resType.label}></bk-option>
                    ))}
                  </bk-select>
                </bk-form-item>
              </div>
              <div class='displayflex'>
                <bk-form-item class='mr16' label='云地域' required property='qcloudRegionId'>
                  <AreaSelector
                    ref='areaSelector'
                    v-model={resourceForm.value.selector.qcloudRegionId}
                    class='item-warp-qcloudRegionId'
                    onChange={onQcloudRegionChange}></AreaSelector>
                </bk-form-item>
                <bk-form-item label='可用区' property='zone'>
                  <ZoneSelector
                    ref='zoneSelector'
                    v-model={resourceForm.value.selector.qcloudZoneId}
                    class='item-warp-qcloudZoneId'
                    area={resourceForm.value.selector.qcloudRegionId}
                    onChange={onQcloudZoneChange}
                  />
                </bk-form-item>
              </div>
            </bk-form>
          </CommonCard>
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
                      <span class='fontweight'>VPC : 系统自动匹配</span>
                      <span class='fontweight'>子网 : 系统自动匹配</span>
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
                    model={resourceForm.value}
                    rules={resourceFormRules.value}
                    ref='formRef'>
                    <bk-form-item label='VPC' required property='resourceType'>
                      <div class='component-with-detail-container'>
                        <bk-select
                          class='item-warp-resourceType component-with-detail'
                          v-model={resourceForm.value.resourceType}>
                          {/* {resourceTypes.value.map((resType) => (
                            <bk-option key={resType.value} value={resType.value} label={resType.label}></bk-option>
                          ))} */}
                        </bk-select>
                        <Button text style={{ marginRight: '-50px' }} theme='primary'>
                          预览
                        </Button>
                      </div>
                    </bk-form-item>
                    <bk-form-item label='子网' required property='qcloudZoneId'>
                      <div class='component-with-detail-container'>
                        <bk-select
                          class='item-warp-resourceType component-with-detail'
                          v-model={resourceForm.value.selector.qcloudZoneId}>
                          {/* {accountList.map((item, index) => (
                <bk-option key={index} value={item.id} label={item.name}></bk-option>
              ))} */}
                        </bk-select>
                        <Button text style={{ marginRight: '-50px' }} theme='primary'>
                          预览
                        </Button>
                      </div>
                    </bk-form-item>
                  </bk-form>
                </>
              </CommonCard>
            </>
          )}
          <CommonCard class='mt15' title={() => '实例配置'} layout='grid'>
            <>
              {resourceForm.value.resourceType !== 'IDCPM' ? (
                <>
                  <bk-form
                    form-type='vertical'
                    label-width='150'
                    model={resourceForm.value}
                    rules={resourceFormRules.value}
                    ref='formRef'>
                    <bk-form-item label='机型' required property='resourceType'>
                      <CvmTypeSelector
                        class='item-warp-resourceType'
                        v-model={resourceForm.value.selector.deviceClass}
                        area={resourceForm.value.selector.qcloudRegionId}
                        zone={resourceForm.value.selector.qcloudZoneId}
                      />
                    </bk-form-item>
                    <bk-form-item class='mr16' label='镜像' required property='imageId'>
                      <ImageSelector
                        class='item-warp-imageId'
                        ref='imageSelector'
                        v-model={resourceForm.value.selector.imageId}
                        area={resourceForm.value.selector.qcloudRegionId}
                      />
                    </bk-form-item>
                    <bk-form-item label='数据盘' required property='dataDisk'>
                      <div class='displayflex'>
                        <bk-select
                          v-model={resourceForm.value.selector.dataDisk[0].diskType}
                          class='item-warp-dataDisk-diskType'>
                          {resourceTypes.value.map((resType) => (
                            <bk-option key={resType.value} value={resType.value} label={resType.label}></bk-option>
                          ))}
                        </bk-select>
                        <Input
                          type='number'
                          class='item-warp-dataDisk-diskSize'
                          v-model={resourceForm.value.selector.dataDisk[0].diskSize}
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
                    model={pmForm.value.model}
                    rules={resourceFormRules.value}
                    ref='formRef'>
                    <div class='displayflex'>
                      <bk-form-item label='机型' required property='resourceType'>
                        <bk-select
                          v-model={pmForm.value.model.spec.deviceType}
                          default-first-option
                          class='width-300 mr16'
                          filterable>
                          {pmForm.value.options.deviceTypes.map((deviceType: { device_type: any }) => (
                            <bk-option key={deviceType.device_type} value={deviceType.device_type}></bk-option>
                          ))}
                        </bk-select>
                      </bk-form-item>
                      <bk-form-item label='RAID 类型' prop='spec.raidType'>
                        <span> {pmForm.value.model.spec.raidType || '-'}</span>
                      </bk-form-item>
                    </div>
                    <bk-form-item label='操作系统' required property='qcloudRegionId'>
                      <bk-select class='width-400' v-model={pmForm.value.model.spec.osType}>
                        {pmForm.value.options.osTypes.map((osType) => (
                          <bk-option key={osType} value={osType}></bk-option>
                        ))}
                      </bk-select>
                    </bk-form-item>
                    <bk-form-item label='经营商' required property='resourceType'>
                      <bk-select class='width-300' v-model={pmForm.value.model.spec.isp}>
                        {pmForm.value.options.isps.map((isp) => (
                          <bk-option key={isp} value={isp} label={isp}></bk-option>
                        ))}
                      </bk-select>
                    </bk-form-item>
                    <div class='displayflex'>
                      <bk-form-item label='需求数量' required property='resourceType'>
                        <Input
                          class='item-warp-replicas mr16'
                          type='number'
                          v-model={pmForm.value.model.replicas}
                          min={1}></Input>
                      </bk-form-item>
                      <bk-form-item label='反亲和性' required property='resourceType'>
                        <AntiAffinityLevelSelect
                          v-model={pmForm.value.model.antiAffinityLevel}></AntiAffinityLevelSelect>
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
        </CommonSideslider>
      </div>
    );
  },
});
