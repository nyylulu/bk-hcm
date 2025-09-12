<script setup lang="ts">
import { getDiskTypesName } from '@/views/ziyanScr/cvm-produce/component/property-display/transform';
import { ICvmDataDisk } from './typings';

interface IProps {
  dataDiskList?: ICvmDataDisk[];
  diskType: string;
  diskSize: number;
}

defineOptions({ name: 'cvm-data-disk-display' });

defineProps<IProps>();
</script>

<template>
  <span v-if="!dataDiskList || dataDiskList.length === 0">
    {{ diskType ? `${getDiskTypesName(diskType)} ${diskSize}G` : '--' }}
  </span>
  <div class="flex-row align-items-center">
    <template v-for="(dataDisk, index) in dataDiskList" :key="index">
      <span>{{ getDiskTypesName(dataDisk.disk_type) }}</span>
      <span class="ml2">{{ dataDisk.disk_size }}G</span>
      <span class="ml2">{{ dataDisk.disk_num }}块</span>
      <span v-if="index !== dataDiskList.length - 1">,</span>
    </template>
  </div>
</template>

<style scoped lang="scss"></style>
