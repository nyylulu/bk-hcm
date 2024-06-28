import { defineComponent, onMounted, ref, watch, nextTick } from 'vue';
import { Input, Button, Sideslider } from 'bkui-vue';
import CommonCard from '@/components/CommonCard';
import './index.scss';
import ZoneSelector from '../ZoneSelector';
import { RightShape, DownShape } from 'bkui-vue/lib/icon';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import { useRouter, useRoute } from 'vue-router';
import WName from '@/components/w-name';
import apiService from '@/api/scrApi';
import applicationSideslider from '../application-sideslider/index';
import { getBusinessNameById } from '@/views/ziyanScr/host-recycle/field-dictionary';
import { getTypeCn } from '@/views/ziyanScr/cvm-produce/transform';
import { getResourceTypeName } from '../transform';
import { getRegionCn, getZoneCn } from '@/views/ziyanScr/cvm-web/transform';
import { getDiskTypesName, getImageName } from '@/components/property-list/transform';
export default defineComponent({
  components: {
    applicationSideslider,
  },
  setup() {
    const router = useRouter();
    const route = useRoute();
    const cvmOneKeyApplyVisible = ref(false);
    // 机型列表
    const deviceTypes = ref([]);
    // VPC列表
    const zoneTypes = ref([]);
    // 子网列表
    const subnetTypes = ref([]);
    // 网络信息开关
    const NIswitch = ref(true);
    const handleVpcChange = () => {
      loadSubnets();
    };
    const onZoneChange = () => {
      order.value.model.spec.device_type = '';
      loadDeviceTypes();
      loadVpcs();
    };
    const loadDeviceTypes = async () => {
      const {
        spec: { zone },
      } = order.value.model;
      const {
        spec: { region },
      } = rawOrder.value;
      const params = {
        region: [region],
        zone: zone !== 'cvm_separate_campus' ? [zone] : undefined,
      };
      if (rawOrder.value.resource_type === 'QCLOUDCVM') {
        const { info } = await apiService.getDeviceTypes(params);
        deviceTypes.value = info || [];
      } else {
        const { info } = await apiService.getIDCPMDeviceTypes();
        deviceTypes.value = info.map((item) => {
          return item.device_type;
        });
      }
    };
    const loadVpcs = async () => {
      const { info } = await apiService.getVpcs(rawOrder.value.spec.region);
      zoneTypes.value = info;
    };

    const loadSubnets = async () => {
      const { zone, vpc } = order.value.model.spec;
      const { info } = await apiService.getSubnets({
        region: rawOrder.value.spec.region,
        zone,
        vpc,
      });

      subnetTypes.value = info || [];
    };
    const ARtriggerShow = (isShow: boolean) => {
      cvmOneKeyApplyVisible.value = isShow;
    };

    onMounted(() => {
      getOrders();
    });
    const originalDocumentslist = ref([]);
    const nameList = ref([
      '子单据 ID',
      '业务',
      '需求类型',
      '期望交付时间',
      '提单人',
      '资源类型',
      '地域',
      '数据盘类型',
      '镜像',
      '备注',
    ]);
    const getOrders = async () => {
      if (route?.query?.suborder_id) {
        originalDocumentslist.value = [];
        const { info } = await apiService.getOrders({
          bk_biz_id: [+route?.query?.bk_biz_id],
          suborder_id: [route?.query?.suborder_id],
        });
        rawOrder.value = info[0] || { spec: {} };
        const {
          suborder_id,
          bk_biz_id,
          require_type,
          expect_time,
          bk_username,
          resource_type,
          remark,
          spec: { region, disk_type, image_id },
        } = rawOrder.value;
        const informationList = [
          suborder_id,
          getBusinessNameById(bk_biz_id),
          getTypeCn(require_type),
          expect_time,
          bk_username,
          getResourceTypeName(resource_type),
          getRegionCn(region),
          getDiskTypesName(disk_type),
          getImageName(image_id),
          remark,
        ];
        informationList.forEach((item, index) => {
          originalDocumentslist.value.push({
            name: nameList.value[index],
            value: item,
          });
        });
      }
    };
    const order = ref({
      model: {
        spec: {
          zone: '',
          device_type: '',
          region: '',
          replicas: '',
          vpc: '',
          subnet: '',
        },
      },
    });
    const rawOrder = ref({
      resource_type: '',
      total_num: 0,
      success_num: 0,
      spec: {
        region: '',
        zone: '',
        device_type: '',
      },
    });
    const applyCpu = ref('');
    const applyMem = ref('');
    const applyRegion = ref([]);
    const deviceGroup = ref('');
    const handleSearchAvailable = async () => {
      const { zone, device_type } = order.value.model.spec;
      const { region } = rawOrder.value.spec;
      const filter = {
        region: [region],
        device_group: deviceGroup.value,
        device_type: [device_type],
        zone: zone !== 'cvm_separate_campus' ? [zone] : undefined,
      };
      const page = {
        start: 0,
        limit: 50,
      };

      if (zone !== 'cvm_separate_campus') {
        filter.zone = [zone];
      }
      const { info } = await apiService.getAvailDevices({ filter, page });
      applyCpu.value = String(info[0]?.cpu || '');
      applyMem.value = String(info[0]?.mem || '');
      applyRegion.value = region ? [region] : [];
      cvmOneKeyApplyVisible.value = true;
    };
    const CVMapplication = (data, val) => {
      order.value.model.spec.device_type = data.device_type;
      order.value.model.spec.zone = data.zone;
      cvmOneKeyApplyVisible.value = val;
    };
    watch(
      () => order.value.model.spec.zone,
      () => {
        loadDeviceTypes();
        loadVpcs();
      },
    );
    const handleSubmit = async () => {
      await formRef.value.validate();
      const { device_type, zone, replicas, subnet, vpc } = order.value.model.spec;
      const {
        suborder_id,
        bk_username,
        spec: { region, image_id, disk_size, disk_type, network_type },
      } = rawOrder.value;
      const params = {
        replicas: Number(replicas) + Number(rawOrder.value.success_num),
        bk_username,
        suborder_id,
        spec: {
          region,
          device_type,
          zone,
          subnet,
          vpc,
          image_id,
          disk_size,
          disk_type,
          network_type,
        },
      };
      const { code } = await apiService.modifyOrder(params);
      if (code === 0) {
        router.push({
          path: '/ziyanScr/hostApplication',
        });
      }
      nextTick(() => {
        formRef.value.clearValidate();
      });
    };
    const formRef = ref();
    const rules = ref({
      zone: [{ required: true, message: '请选择园区', trigger: 'change' }],
      device_type: [{ required: true, message: '请选择机型', trigger: 'change' }],
      replicas: [{ required: true, message: '输入申请数量', trigger: 'blur' }],
    });
    return () => (
      <div class='wid100'>
        <DetailHeader>修改申请</DetailHeader>
        <div class={'apply-form-wrapper'}>
          {/* 原始单据 */}
          <CommonCard class='mt15' title={() => '原始单据'}>
            <div class={'display'}>
              {originalDocumentslist.value.map(
                (item, index) =>
                  index <= 4 && (
                    <div style={'width:250px'}>
                      <span style={'width:60px'}>{item.name}:</span>
                      {item.name === '提单人' ? <WName name={item.value} /> : <span>{item.value}</span>}
                    </div>
                  ),
              )}
            </div>
            <div class={'display'}>
              {originalDocumentslist.value.map(
                (item, index) =>
                  index > 4 && (
                    <div style={'width:250px'}>
                      <span style={'width:60px'}>{item.name}:</span>
                      {item.name === '提单人' ? <WName name={item.value} /> : <span>{item.value}</span>}
                    </div>
                  ),
              )}
            </div>
          </CommonCard>
          {/* 推荐修改 */}
          <CommonCard class='mt15' title={() => '推荐修改'}>
            <bk-form model={order.value.model.spec} rules={rules.value} ref={formRef} class={'form'}>
              <bk-form-item label='园区' required property='zone'>
                <ZoneSelector
                  ref='zoneSelector'
                  class={'w200px'}
                  v-model={order.value.model.spec.zone}
                  params={{
                    resourceType: rawOrder.value.resource_type,
                    region: rawOrder.value.spec.region,
                  }}
                  onChange={onZoneChange}
                />
                <div>原始值：{getZoneCn(rawOrder.value.spec.zone)}</div>
              </bk-form-item>

              <bk-form-item label='机型' required property='device_type'>
                <div class='component-with-preview'>
                  <bk-select
                    class={'w200px'}
                    v-model={order.value.model.spec.device_type}
                    disabled={order.value.model.spec.zone === ''}
                    placeholder={order.value.model.spec.zone === '' ? '请先选择园区' : '请选择机型'}
                    filterable>
                    {deviceTypes.value.map((device_type) => (
                      <bk-option key={device_type} label={device_type} value={device_type} />
                    ))}
                  </bk-select>
                  <Button class='preview-btn' onClick={handleSearchAvailable}>
                    查询可替代资源库存
                  </Button>
                </div>
                <div>原始值：{rawOrder.value.spec.device_type}</div>
              </bk-form-item>

              <bk-form-item label='剩余可申请数量' required property='replicas'>
                <Input
                  class={'w200px'}
                  type='number'
                  v-model={order.value.model.spec.replicas}
                  min={1}
                  max={rawOrder.value.total_num - rawOrder.value.success_num}></Input>
                <div>
                  原始值：{rawOrder.value.total_num}, 已交付数量：{rawOrder.value.success_num}, 可申请数量：
                  {rawOrder.value.total_num - rawOrder.value.success_num}
                  (已交付+待交付不可超过原数量需求)
                </div>
              </bk-form-item>
            </bk-form>
          </CommonCard>
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
                        <span class='fontweight'>VPC : {order.value.model.spec.vpc}</span>
                        <span class='fontweight'>子网 : {order.value.model.spec.subnet}</span>
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
                    <bk-form form-type='vertical' label-width='150' model={order.value.model.spec}>
                      <bk-form-item label='VPC'>
                        <div class='component-with-detail-container'>
                          <bk-select
                            class='item-warp-resourceType component-with-detail'
                            disabled={order.value.model.spec.zone === 'cvm_separate_campus'}
                            v-model={order.value.model.spec.vpc}
                            onChange={handleVpcChange}>
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
                            disabled={order.value.model.spec.zone === 'cvm_separate_campus'}
                            v-model={order.value.model.spec.subnet}>
                            {subnetTypes.value.map((subnet) => (
                              <bk-option
                                key={subnet.subnet_id}
                                value={subnet.subnet_id}
                                label={`${subnet.subnet_id} | ${subnet.subnet_name}`}></bk-option>
                            ))}
                          </bk-select>
                        </div>
                      </bk-form-item>
                      <bk-form-item label=''>
                        <div>
                          一般需求不需要指定 VPC 和子网，如为 BCS、ODP 等 TKE 场景母机，请提前与平台支持方确认
                          VPC、子网信息
                        </div>
                      </bk-form-item>
                    </bk-form>
                  </>
                </CommonCard>
              </>
            )}
          </>
          <div class={'buttonSubmit'}>
            <Button class='mr8' onClick={handleSubmit} theme='primary'>
              确定修改
            </Button>
            <Button
              onClick={() => {
                router.go(-1);
              }}>
              取消
            </Button>
          </div>
          {/* 查看可替代资源库存 */}

          <Sideslider
            class='common-sideslider'
            width={1300}
            isShow={cvmOneKeyApplyVisible.value}
            title='查询可替代资源库存'
            onClosed={() => {
              ARtriggerShow(false);
            }}>
            <applicationSideslider
              onOneApplication={CVMapplication}
              getform={cvmOneKeyApplyVisible.value}
              cpu={applyCpu.value}
              mem={applyMem.value}
              region={applyRegion.value}
              device={{
                filter: {
                  require_type: rawOrder.value.require_type,
                },
              }}
            />
          </Sideslider>
        </div>
      </div>
    );
  },
});
