<script setup lang="ts">
import { ref, useTemplateRef, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';
import CvmVpcSelector, { type ICvmVpc } from '@/views/ziyanScr/components/cvm-vpc-selector/index.vue';
import CvmSubnetSelector, { type ICvmSubnet } from '@/views/ziyanScr/components/cvm-subnet-selector/index.vue';
import BcsSelectTips from '../application-form/bcs-select-tips';

interface IProps {
  region: string;
  zone: string;
  disabled?: boolean;
  disabledVpc?: boolean;
  disabledSubnet?: boolean;
  vpcProperty?: string;
  subnetProperty?: string;
}

defineOptions({ name: 'internet-info-collapse-panel' });

const vpc = defineModel<string>('vpc');
const subnet = defineModel<string>('subnet');

const props = withDefaults(defineProps<IProps>(), {
  disabled: false,
  disabledVpc: false,
  disabledSubnet: false,
  vpcProperty: 'spec.vpc',
  subnetProperty: 'spec.subnet',
});
const emit = defineEmits<{
  (e: 'changeVpc', val: ICvmVpc): void;
  (e: 'changeSubnet', val: ICvmSubnet): void;
}>();

const { t } = useI18n();

const isExpand = ref(false);
const selectCvmVpc = ref<ICvmVpc>(null);
const selectedCvmSubnet = ref<ICvmSubnet>(null);
const cvmVpcSelectorRef = useTemplateRef('cvm-vpc-selector');
const cvmSubnetSelectorRef = useTemplateRef('cvm-subnet-selector');

const handleToggle = (v: boolean) => (isExpand.value = v);

const handleCvmVpcChange = (val?: ICvmVpc) => {
  selectCvmVpc.value = val;
  selectedCvmSubnet.value = null;
  subnet.value = '';
  emit('changeVpc', val);
};

const handleCvmSubnetChange = (val?: ICvmSubnet) => {
  selectedCvmSubnet.value = val;
  emit('changeSubnet', val);
};

watchEffect(() => {
  selectCvmVpc.value = cvmVpcSelectorRef.value?.findCvmVpcByVpcId(vpc.value);
});
watchEffect(() => {
  selectedCvmSubnet.value = cvmSubnetSelectorRef.value?.findCvmSubnetBySubnetId(subnet.value);
});
watchEffect(() => {
  if (props.disabled) {
    isExpand.value = false;
    vpc.value = '';
    subnet.value = '';
  }
});

defineExpose({ handleToggle });
</script>

<template>
  <bk-collapse class="home">
    <bk-collapse-panel
      :model-value="isExpand"
      :title="t('网络信息')"
      icon="right-shape"
      alone
      :disabled="disabled"
      @update:model-value="handleToggle"
    >
      <template #default v-if="!isExpand">
        {{ t('网络信息') }}
        <span class="overview">
          <span>
            {{ t('VPC：') }}
            {{ selectCvmVpc ? `${selectCvmVpc.vpc_name}（${selectCvmVpc.vpc_id}）` : t('系统自动分配') }}
          </span>
          <span>
            {{ t('子网：') }}
            {{
              selectedCvmSubnet
                ? `${selectedCvmSubnet.subnet_name}（${selectedCvmSubnet.subnet_id}）`
                : t('系统自动分配')
            }}
          </span>
        </span>
      </template>

      <template #content>
        <bk-form-item label="VPC" :property="vpcProperty">
          <cvm-vpc-selector
            class="cvm-vpc-selector"
            ref="cvm-vpc-selector"
            v-model="vpc"
            :region="props.region"
            :disabled="props.disabledVpc"
            :popover-options="{ boundary: 'parent' }"
            @change="handleCvmVpcChange"
          />
          <!-- 如果选择BSC集群的VPC，提供引导提示 -->
          <bcs-select-tips v-if="/(BCS|OVERLAY)/.test(selectCvmVpc?.vpc_name)" :desc="t('所选择的VPC为容器网络')" />
        </bk-form-item>
        <bk-form-item :label="t('子网')" :property="subnetProperty">
          <cvm-subnet-selector
            class="cvm-subnet-selector"
            ref="cvm-subnet-selector"
            v-model="subnet"
            :region="props.region"
            :zone="props.zone"
            :vpc="vpc"
            :disabled="props.disabledSubnet"
            :popover-options="{ boundary: 'parent' }"
            @change="handleCvmSubnetChange"
          />
          <!-- 如果选择BSC集群的子网，提供引导提示 -->
          <bcs-select-tips
            v-if="/(BCS|OVERLAY)/.test(selectedCvmSubnet?.subnet_name)"
            :desc="t('所选择的子网为容器子网')"
          />
        </bk-form-item>

        <!-- tips -->
        <div class="tips">
          <slot name="tips"></slot>
        </div>
      </template>
    </bk-collapse-panel>
  </bk-collapse>
</template>

<style scoped lang="scss">
.home {
  background-color: #fff;
  box-shadow: 0 2px 4px 0 #1919290d;

  :deep(.bk-collapse-title) {
    font-weight: 700;
  }

  :deep(.bk-collapse-content) {
    padding: 16px 24px;
  }

  .overview {
    margin-left: 56px;
    display: inline-flex;
    gap: 40px;
    font-weight: normal;
    color: #313238;
  }

  .tips {
    font-size: 12px;
  }
}
</style>
