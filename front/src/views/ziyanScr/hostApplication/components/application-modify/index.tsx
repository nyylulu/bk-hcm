import { computed, defineComponent, onMounted, ref, nextTick } from 'vue';
import { Input, Button, Sideslider, Form, Alert, Message } from 'bkui-vue';
import CommonCard from '@/components/CommonCard';
import './index.scss';
import ZoneSelector from '../ZoneSelector';
import NetworkInfoCollapsePanel from '../network-info-collapse-panel/index.vue';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import { useRouter, useRoute } from 'vue-router';
import WName from '@/components/w-name';
import apiService from '@/api/scrApi';
import applicationSideslider from '../application-sideslider/index.vue';
import { getBusinessNameById } from '@/views/ziyanScr/host-recycle/field-dictionary';
import { getTypeCn } from '@/views/ziyanScr/cvm-produce/transform';
import { getResourceTypeName } from '../transform';
import { getRegionCn, getZoneCn } from '@/views/ziyanScr/cvm-web/transform';
import { getDiskTypesName, getImageName } from '@/components/property-list/transform';
import http from '@/http';
import { getEntirePath } from '@/utils';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import DevicetypeSelector from '@/views/ziyanScr/components/devicetype-selector/index.vue';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import { MENU_SERVICE_HOST_APPLICATION } from '@/constants/menu-symbol';

export default defineComponent({
  components: {
    applicationSideslider,
  },
  setup() {
    const router = useRouter();
    const route = useRoute();
    const { isBusinessPage, getBusinessApiPath } = useWhereAmI();
    const cvmOneKeyApplyVisible = ref(false);

    const onZoneChange = () => {
      order.value.model.spec.device_type = '';
    };

    const ARtriggerShow = (isShow: boolean) => {
      cvmOneKeyApplyVisible.value = isShow;
    };

    const bizId = computed(() => Number(route.query.bk_biz_id));

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
        const { info } = await http
          .post(getEntirePath(`${getBusinessApiPath()}task/findmany/apply`), {
            bk_biz_id: [bizId.value],
            suborder_id: [route?.query?.suborder_id],
          })
          .then((res: any) => res.data);

        rawOrder.value = info[0] || { spec: {} };
        const {
          suborder_id,
          bk_biz_id,
          require_type,
          expect_time,
          bk_username,
          resource_type,
          remark,
          spec: { region, disk_type, image_id, vpc, subnet },
        } = rawOrder.value as any;

        // 回显vpc, 子网
        order.value.model.spec.vpc = vpc;
        order.value.model.spec.subnet = subnet;

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
    const CVMapplication = (data: any, val: boolean) => {
      order.value.model.spec.device_type = data.device_type;
      order.value.model.spec.zone = data.zone;
      cvmOneKeyApplyVisible.value = val;
    };

    // 机型列表
    const cvmDevicetypeParams = computed(() => {
      const { zone } = order.value.model.spec;
      const { region } = rawOrder.value.spec;
      const require_type = route?.query?.require_type;
      return {
        region,
        zone: zone !== 'cvm_separate_campus' ? zone : undefined,
        require_type: require_type ? Number(require_type) : undefined,
      };
    });

    const handleSubmit = async () => {
      await formRef.value.validate();
      const { device_type, zone, replicas, subnet, vpc } = order.value.model.spec;
      const {
        suborder_id,
        bk_username,
        spec: { region, image_id, disk_size, disk_type, network_type },
      } = rawOrder.value as any;
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
      const { code } = await http.post(getEntirePath(`${getBusinessApiPath()}task/modify/apply`), params, {
        removeEmptyFields: true,
        transformFields: true,
      });
      if (code === 0) {
        Message({ theme: 'success', message: '提交成功' });
        if (isBusinessPage) {
          router.replace({ name: 'ApplicationsManage', query: { [GLOBAL_BIZS_KEY]: bizId.value, type: 'host_apply' } });
        } else {
          router.replace({ name: MENU_SERVICE_HOST_APPLICATION });
        }
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
                  <DevicetypeSelector
                    class='w200px'
                    v-model={order.value.model.spec.device_type}
                    resourceType={rawOrder.value.resource_type === 'QCLOUDCVM' ? 'cvm' : 'idcpm'}
                    params={cvmDevicetypeParams.value}
                    disabled={order.value.model.spec.zone === ''}
                    placeholder={order.value.model.spec.zone === '' ? '请先选择园区' : '请选择机型'}
                  />
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

          <Form model={order.value.model.spec} formType='vertical' class='mt15'>
            <NetworkInfoCollapsePanel
              v-model:vpc={order.value.model.spec.vpc}
              v-model:subnet={order.value.model.spec.subnet}
              region={rawOrder.value.spec.region}
              zone={order.value.model.spec.zone}
              disabledVpc={order.value.model.spec.zone === 'cvm_separate_campus'}
              disabledSubnet={order.value.model.spec.zone === 'cvm_separate_campus'}>
              {{
                tips: () => (
                  <Alert
                    class='alert-container'
                    title='一般需求不需要指定 VPC 和子网，如为 BCS、ODP 等 TKE 场景母机，请提前与平台支持方确认 VPC、子网信息。'
                  />
                ),
              }}
            </NetworkInfoCollapsePanel>
          </Form>

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
              isShow={cvmOneKeyApplyVisible.value}
              requireType={rawOrder.value.require_type}
              bizId={bizId.value}
              initialCondition={{
                cpu: applyCpu.value,
                mem: applyMem.value,
                region: applyRegion.value,
              }}
              onApply={CVMapplication}
            />
          </Sideslider>
        </div>
      </div>
    );
  },
});
