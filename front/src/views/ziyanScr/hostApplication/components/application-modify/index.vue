<script setup lang="ts">
import { computed, h, onBeforeMount, reactive, ref, useTemplateRef, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useZiyanScrStore } from '@/store';
import usePlanStore from '@/store/usePlanStore';
import { GLOBAL_BIZS_KEY, INSTANCE_CHARGE_MAP } from '@/common/constant';
import { ModelPropertyDisplay } from '@/model/typings';
import { isNil } from 'lodash';
import { CvmDeviceType, IdcpmDeviceType, SelectionType } from '@/views/ziyanScr/components/devicetype-selector/types';
import type { IApplyOrderItem } from '@/typings/ziyanScr';
import type { IDemandVerification } from '@/typings/plan';
import { MENU_SERVICE_HOST_APPLICATION } from '@/constants/menu-symbol';
import { RequirementType } from '@/store/config/requirement';

// TODO: 这些翻译项，后续都要通过display组件进行优化
import { getDiskTypesName, getImageName } from '@/views/ziyanScr/cvm-produce/component/property-display/transform';
import { getZoneCn } from '@/views/ziyanScr/cvm-web/transform';

import { Form, Message } from 'bkui-vue';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import GridDisplay from './grid-display.vue';
import CollapsePanelGrid from './collapse-panel-grid.vue';
import Panel from '@/components/panel';
import ZoneSelector from '../ZoneSelector';
import DevicetypeSelector from '@/views/ziyanScr/components/devicetype-selector/index.vue';
import NetworkInfoCollapsePanel from '../network-info-collapse-panel/index.vue';
import CvmQuickApplyForm from '../application-sideslider/index.vue';

const route = useRoute();
const router = useRouter();
const ziyanScrStore = useZiyanScrStore();
const planStore = usePlanStore();

const businessId = computed(() => Number(route.query.bk_biz_id));
const suborderId = computed(() => route.query.suborder_id as string);

const formRef = useTemplateRef<typeof Form>('adjust-form');
const formModel = reactive({ zone: '', device_type: '', replicas: 0, vpc: '', subnet: '' });

const details = ref<IApplyOrderItem>();
const unProductNum = computed(() => (!details.value ? 0 : details.value.origin_num - details.value.product_num));
const getDetails = async () => {
  const list = await ziyanScrStore.getApplyOrderList({
    bk_biz_id: [businessId.value],
    suborder_id: [suborderId.value],
    get_product: true,
  });
  // suborder_id请求回来的只会是一个单据数据
  [details.value] = list;
};
onBeforeMount(async () => {
  await getDetails();
  // 初始化表单
  const { origin_num, product_num } = details.value || {};
  const { zone, device_type, vpc, subnet } = details.value?.spec || {};
  Object.assign(formModel, {
    zone,
    device_type,
    replicas: origin_num - product_num,
    vpc,
    subnet,
  });
});

const RESOURCE_TYPE_NAME_MAP: Record<string, string> = {
  QCLOUDCVM: '腾讯云_CVM',
  IDCPM: 'IDC物理机',
  QCLOUDDVM: '腾讯云_DockerVM',
  IDCDVM: 'IDC_DockerVM',
};

const baseInfoFields: ModelPropertyDisplay[] = [
  { id: 'suborder_id', name: '子单据ID', type: 'string' },
  { id: 'bk_biz_id', name: '业务', type: 'business' },
  { id: 'bk_username', name: '提单人', type: 'user' },
  { id: 'expect_time', name: '期望交付时间', type: 'datetime' },
];
const originDemandFields: ModelPropertyDisplay[] = [
  { id: 'require_type', name: '项目类型', type: 'req-type' },
  { id: 'resource_type', name: '资源类型', type: 'enum', option: RESOURCE_TYPE_NAME_MAP },
  { id: 'spec.device_type', name: '机型', type: 'string' },
  { id: 'spec.charge_type', name: '计费模式', type: 'enum', option: INSTANCE_CHARGE_MAP },
  { id: 'spec.region', name: '地域', type: 'region' },
  {
    id: 'spec.zone',
    name: '园区',
    type: 'string',
    meta: {
      display: {
        format: (val) => {
          return isNil(val) ? '--' : `${getZoneCn(val)}`;
        },
      },
    },
  },
  {
    id: 'spec.image_id',
    name: '镜像',
    type: 'string',
    meta: {
      display: {
        format: (val) => {
          return isNil(val) ? '--' : `${getImageName(val)}`;
        },
      },
    },
  },
  { id: 'spec.vpc', name: '所属VPC', type: 'string' },
  {
    id: 'spec.disk_type',
    name: '数据盘类型',
    type: 'string',
    render: (details) => {
      const cell = details.spec.disk_type;
      return isNil(cell) ? '--' : `${getDiskTypesName(cell)}(${details.spec.disk_size}G)`;
    },
  },
  { id: 'spec.subnet', name: '所属子网', type: 'string' },
];
const productionFields: ModelPropertyDisplay[] = [
  { id: 'origin_num', name: '需求总数', type: 'number' },
  { id: 'product_num', name: '已生产数', type: 'number' },
  {
    id: 'un_product_num',
    name: '未生产数',
    type: 'number',
    render: () => h('span', { class: 'text-warning' }, unProductNum.value),
  },
];

const handleZoneChange = () => {
  formModel.device_type = '';
};
const selectedDevicetypeInfo = reactive({ cpu: 0, mem: 0 });
const handleDevicetypeChange = (devicetype: SelectionType) => {
  if (!devicetype) {
    Object.assign(selectedDevicetypeInfo, { cpu: 0, mem: 0 });
    return;
  }
  if (details.value.resource_type === 'QCLOUDCVM') {
    const { cpu_amount, ram_amount } = devicetype as CvmDeviceType;
    Object.assign(selectedDevicetypeInfo, { cpu: cpu_amount, mem: ram_amount });
  } else {
    const { cpu, mem } = devicetype as IdcpmDeviceType;
    Object.assign(selectedDevicetypeInfo, { cpu, mem });
  }
};

const cvmDevicetypeParams = computed(() => {
  if (!details.value) return {};

  const { require_type } = details.value;
  const { region, device_group, device_size } = details.value.spec;
  const { zone } = formModel;

  return {
    require_type,
    region,
    zone: zone !== 'cvm_separate_campus' ? zone : undefined,
    device_group: device_group ? [device_group] : undefined,
    device_size,
  };
});

const isNeedVerify = computed(() => {
  return (
    ![RequirementType.RollServer, RequirementType.GreenChannel].includes(details.value?.require_type) &&
    !verifyResult.value
  );
});
const isVerifyLoading = ref(false);
const verifyResult = ref<IDemandVerification>();
const handleVerify = async () => {
  await formRef.value.validate();
  const { zone, device_type, vpc, subnet, replicas } = formModel;
  const spec = Object.assign({}, details.value.spec, { zone, device_type, vpc, subnet });
  isVerifyLoading.value = true;
  try {
    const res = await planStore.verify_resource_demand({
      bk_biz_id: businessId.value,
      require_type: details.value.require_type,
      suborders: [{ ...details.value, replicas, spec }],
    });
    [verifyResult.value] = res.data?.verifications ?? [];
  } catch (error) {
    verifyResult.value = null;
    return Promise.reject(error);
  } finally {
    isVerifyLoading.value = false;
  }
};
watch(formModel, () => {
  verifyResult.value = null;
});

const handleSubmit = async () => {
  await formRef.value.validate();
  const { suborder_id, bk_username } = details.value;
  const { zone, device_type, vpc, subnet, replicas } = formModel;
  const spec = Object.assign({}, details.value.spec, { zone, device_type, vpc, subnet });
  const res = await ziyanScrStore.modifyApplyOrder({ suborder_id, bk_username, replicas, spec });
  if (res.code === 0) {
    Message({ theme: 'success', message: '提交成功' });
    handleBack();
  }
};
const handleBack = () => {
  const globalBusinessId = route.query[GLOBAL_BIZS_KEY];
  if (globalBusinessId) {
    router.replace({ name: 'ApplicationsManage', query: { [GLOBAL_BIZS_KEY]: globalBusinessId, type: 'host_apply' } });
  } else {
    router.replace({ name: MENU_SERVICE_HOST_APPLICATION });
  }
};

const cvmQuickApplySidesliderState = reactive({
  isShow: false,
  isHidden: false,
  initialCondition: { region: [], device_families: [] },
});
const handleSearchAvailable = () => {
  const { region, device_group } = details.value.spec;

  Object.assign(cvmQuickApplySidesliderState.initialCondition, {
    region: region ? [region] : undefined,
    device_families: device_group ? [device_group] : undefined, // application-sideslider\index.vue 组件中机型族的key为device_families
  });

  cvmQuickApplySidesliderState.isShow = true;
  cvmQuickApplySidesliderState.isHidden = false;
};
const handleCvmQuickApply = (data: any) => {
  const { device_type, zone } = data;
  Object.assign(formModel, { device_type, zone });
  cvmQuickApplySidesliderState.isShow = false;
  cvmQuickApplySidesliderState.isHidden = true;
};
</script>

<template>
  <detail-header>修改申请</detail-header>
  <div class="container">
    <panel class="panel" title="基本信息">
      <grid-display :fields="baseInfoFields" :details="details" />
    </panel>
    <collapse-panel-grid class="panel" title="原始需求" :fields="originDemandFields" :details="details" />
    <panel class="panel" title="生产情况">
      <grid-display :fields="productionFields" :details="details" />
    </panel>
    <bk-form ref="adjust-form" :model="formModel" form-type="vertical">
      <panel title="未生产需求调整">
        <div class="adjust-form-content">
          <bk-form-item label="园区" required property="zone">
            <zone-selector
              class="form-control"
              v-model="formModel.zone"
              :params="{ resourceType: details?.resource_type, region: details?.spec.region }"
              @change="handleZoneChange"
            />
            <div class="tips">原始值：{{ getZoneCn(details?.spec.zone) }}</div>
          </bk-form-item>
          <bk-form-item label="机型" required property="device_type">
            <div class="flex-row align-items-center">
              <devicetype-selector
                class="form-control"
                v-model="formModel.device_type"
                :resource-type="details?.resource_type === 'QCLOUDCVM' ? 'cvm' : 'idcpm'"
                :params="cvmDevicetypeParams"
                :disabled="formModel.zone === ''"
                :placeholder="formModel.zone === '' ? '请先选择园区' : '请选择机型'"
                @change="handleDevicetypeChange"
              />
              <bk-button class="ml8" @click="handleSearchAvailable">查询可替代资源库存</bk-button>
            </div>
            <div class="tips">
              原始值：{{ details?.spec.device_type }}
              <template v-if="selectedDevicetypeInfo">
                <br />
                所选机型为{{ formModel.device_type }}，CPU为 {{ selectedDevicetypeInfo.cpu }} 核，内存为
                {{ selectedDevicetypeInfo.mem }} G
              </template>
            </div>
          </bk-form-item>
          <bk-form-item label="剩余生产数量" required property="replicas">
            <bk-input
              class="form-control"
              v-model.number="formModel.replicas"
              type="number"
              :min="1"
              :max="unProductNum"
            />
            <div class="tips">
              所需CPU总核心数为 {{ selectedDevicetypeInfo.cpu * formModel.replicas }} 核 ({{
                `${selectedDevicetypeInfo.cpu}*${formModel.replicas}`
              }})
              <br />
              <span class="text-danger">注意：</span>
              已生产 {{ details?.product_num }}，剩余生产数量为
              <span class="text-danger">{{ formModel.replicas }}</span>
              ，将共计生产
              <span class="text-danger">{{ details?.product_num + formModel.replicas }}</span>
              后（原单据需求数为
              <span class="text-danger">{{ details?.origin_num }}</span>
              ），该单据会自动结单，不可以再重试修改
            </div>
          </bk-form-item>
        </div>
        <network-info-collapse-panel
          v-model:vpc="formModel.vpc"
          v-model:subnet="formModel.subnet"
          vpc-property="vpc"
          subnet-property="subnet"
          :region="details?.spec.region"
          :zone="formModel.zone"
          :disabled-vpc="formModel.zone === 'cvm_separate_campus'"
          :disabled-subnet="formModel.zone === 'cvm_separate_campus'"
        >
          <template #tips>
            <bk-alert
              title="一般需求不需要指定 VPC 和子网，如为 BCS、ODP 等 TKE 场景母机，请提前与平台支持方确认 VPC、子网信息。"
            />
          </template>
        </network-info-collapse-panel>
      </panel>
    </bk-form>
    <panel v-if="verifyResult?.verify_result === 'FAILED'" title="需求预检">
      <bk-alert theme="danger" :title="verifyResult.reason" />
    </panel>
    <div class="footer">
      <bk-button v-if="isNeedVerify" theme="primary" :loading="isVerifyLoading" @click="handleVerify">
        需求校验
      </bk-button>
      <bk-button
        v-else
        theme="primary"
        :loading="ziyanScrStore.modifyApplyOrderLoading"
        :disabled="isNeedVerify && verifyResult.verify_result !== 'PASS'"
        @click="handleSubmit"
      >
        提交修改
      </bk-button>
      <bk-button :disabled="isVerifyLoading || ziyanScrStore.modifyApplyOrderLoading" @click="handleBack">
        取消
      </bk-button>
    </div>

    <template v-if="!cvmQuickApplySidesliderState.isHidden">
      <bk-sideslider
        v-model:is-show="cvmQuickApplySidesliderState.isShow"
        width="60vw"
        title="查询可替代资源库存"
        @hidden="cvmQuickApplySidesliderState.isHidden = true"
      >
        <cvm-quick-apply-form
          :is-show="cvmQuickApplySidesliderState.isShow"
          :require-type="details?.require_type"
          :biz-id="businessId"
          :initial-condition="cvmQuickApplySidesliderState.initialCondition"
          @apply="handleCvmQuickApply"
        />
      </bk-sideslider>
    </template>
  </div>
</template>

<style scoped lang="scss">
.container {
  margin-top: 52px;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;

  .adjust-form-content {
    padding: 0 24px;
  }

  .tips {
    margin-top: 8px;
    line-height: normal;
    font-size: 12px;
  }

  .form-control {
    width: 35%;
  }

  .footer {
    display: flex;
    align-items: center;
    gap: 8px;
  }
}
</style>
