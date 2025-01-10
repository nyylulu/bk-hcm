<script setup lang="ts">
import { ref, useTemplateRef, watchEffect } from 'vue';
import { Form } from 'bkui-vue';
import { DownShape, RightShape } from 'bkui-vue/lib/icon';
import Panel from '@/components/panel';
import CvmVpcSelector, { type ICvmVpc } from '@/views/ziyanScr/components/cvm-vpc-selector/index.vue';
import CvmSubnetSelector, { type ICvmSubnet } from '@/views/ziyanScr/components/cvm-subnet-selector/index.vue';
import BcsSelectTips from '../application-form/bcs-select-tips';
import { useI18n } from 'vue-i18n';

const { FormItem } = Form;

defineOptions({ name: 'InternetInfoPanel' });

const vpc = defineModel<string>('vpc');
const subnet = defineModel<string>('subnet');

const props = defineProps<{
  region: string;
  zone: string;
  disabledVpc: boolean;
  disabledSubnet: boolean;
}>();
const emit = defineEmits<{
  (e: 'changeVpc', val: ICvmVpc): void;
  (e: 'changeSubnet', val: ICvmSubnet): void;
}>();

const { t } = useI18n();

const isCollapsed = ref(true);
const selectCvmVpc = ref<ICvmVpc>(null);
const selectedCvmSubnet = ref<ICvmSubnet>(null);
const cvmVpcSelectorRef = useTemplateRef('cvm-vpc-selector');
const cvmSubnetSelectorRef = useTemplateRef('cvm-subnet-selector');

const handleCvmVpcChange = (val?: ICvmVpc) => {
  selectCvmVpc.value = val;
  selectedCvmSubnet.value = undefined;
  subnet.value = undefined;
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
</script>

<template>
  <!-- eslint-disable vue/camelcase -->
  <Panel>
    <!-- title -->
    <template #title>
      <section class="header" @click="isCollapsed = !isCollapsed">
        <div class="collapse-icon">
          <component :is="isCollapsed ? RightShape : DownShape"></component>
        </div>
        <div class="title-wrapper">
          <p class="title">{{ t('网络信息') }}</p>
          <template v-if="isCollapsed">
            <div class="info-item">
              <span class="info-label">{{ t('VPC') }}</span>
              <span class="info-value">
                {{ selectCvmVpc ? `${selectCvmVpc.vpc_name}（${selectCvmVpc.vpc_id}）` : t('系统自动分配') }}
              </span>
            </div>
            <div class="info-item">
              <span class="info-label">{{ t('子网') }}</span>
              <span class="info-value">
                {{
                  selectedCvmSubnet
                    ? `${selectedCvmSubnet.subnet_name}（${selectedCvmSubnet.subnet_id}）`
                    : t('系统自动分配')
                }}
              </span>
            </div>
          </template>
        </div>
      </section>
    </template>

    <!-- selector -->
    <section class="selector-wrapper" v-show="!isCollapsed">
      <FormItem label="VPC">
        <CvmVpcSelector
          class="w600"
          ref="cvm-vpc-selector"
          v-model="vpc"
          :region="props.region"
          :disabled="props.disabledVpc"
          @change="handleCvmVpcChange"
        />
        <!-- 如果选择BSC集群的VPC，提供引导提示 -->
        <BcsSelectTips v-if="/(BCS|OVERLAY)/.test(selectCvmVpc?.vpc_name)" :desc="t('所选择的VPC为容器网络')" />
      </FormItem>
      <FormItem label="子网">
        <CvmSubnetSelector
          class="w600"
          ref="cvm-subnet-selector"
          v-model="subnet"
          :region="props.region"
          :zone="props.zone"
          :vpc="vpc"
          :disabled="props.disabledSubnet"
          @change="handleCvmSubnetChange"
        />
        <!-- 如果选择BSC集群的子网，提供引导提示 -->
        <BcsSelectTips
          v-if="/(BCS|OVERLAY)/.test(selectedCvmSubnet?.subnet_name)"
          :desc="t('所选择的子网为容器子网')"
        />
      </FormItem>

      <!-- tips -->
      <div class="tips">
        <slot name="tips"></slot>
      </div>
    </section>
  </Panel>
</template>

<style scoped lang="scss">
.header {
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  cursor: pointer;

  .collapse-icon,
  .title {
    user-select: none;
  }

  .title-wrapper {
    margin-left: 5px;
    display: flex;
    align-items: center;

    .title {
      margin-right: 50px;
      font-size: 14px;
      color: #313238;
      font-weight: 700;
    }

    .info-item {
      margin-right: 24px;
      font-size: 12px;

      .info-label {
        &::after {
          content: ':';
          margin: 0 5px;
        }
      }
    }
  }
}

.selector-wrapper {
  padding: 0 28px;
}

.tips {
  margin-top: 12px;
  font-size: 12px;
}

.w600 {
  width: 600px;
}
</style>
