<script setup lang="ts">
import { CVM_SYSTEM_DISK_INFO } from './constants';
import { ICvmSystemDisk } from './typings';

interface IProps {
  systemDisk: ICvmSystemDisk;
}

defineProps<IProps>();
</script>

<template>
  <span v-if="!systemDisk || !systemDisk.disk_type">--</span>
  <div v-else class="flex-row align-items-center">
    <span>{{ CVM_SYSTEM_DISK_INFO[systemDisk.disk_type].disk_name }}</span>
    <span class="ml2">{{ systemDisk.disk_size }}G</span>
  </div>
</template>

<style scoped lang="scss"></style>
