<script setup lang="ts">
import http from '@/http';
import { Select } from 'bkui-vue';
import { computed, ref, watch } from 'vue';

export interface ICvmVpc {
  id: number;
  region: string;
  vpc_id: string;
  vpc_name: string;
}
type ICvmVpcList = Array<ICvmVpc>;

const { Option } = Select;

defineOptions({ name: 'CvmVpcSelector' });

const model = defineModel<string>();
const props = defineProps<{ region: string; disabled: boolean }>();
const emit = defineEmits<(e: 'change', val: ICvmVpc) => void>();

const optionList = ref<ICvmVpcList>([]);

const selectedId = computed({
  get() {
    // 接口入参vpc_id, select值为id, 此处做转换
    const selectedItem = optionList.value.find((item) => item.vpc_id === model.value);
    return selectedItem?.id || undefined;
  },
  set(val) {
    const selectedItem = optionList.value.find((item) => item.id === val);
    emit('change', selectedItem);
    // 接口入参vpc_id, select值为id, 此处做转换
    model.value = selectedItem?.vpc_id;
  },
});

const findCvmVpcByVpcId = (vpc_id: string) => {
  return optionList.value.find((item) => item.vpc_id === vpc_id);
};

const getOptionList = async (region: string) => {
  const res = await http.post('/api/v1/woa/config/findmany/config/cvm/vpc', { region });
  optionList.value = res.data.info;
};

watch(
  () => props.region,
  (val) => {
    val && getOptionList(val);
  },
  { immediate: true },
);

defineExpose({ findCvmVpcByVpcId });
</script>

<template>
  <Select class="w600" v-model="selectedId" :disabled="props.disabled">
    <Option
      v-for="{ id, vpc_id: vpcId, vpc_name: vpcName } in optionList"
      :key="id"
      :id="id"
      :name="`${vpcId} | ${vpcName}`"
    />
  </Select>
</template>
