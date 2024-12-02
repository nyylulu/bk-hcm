<script setup lang="ts">
import { computed, Ref, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';
import useFormModel from '@/hooks/useFormModel';
import type { ModelPropertyColumn } from '@/model/typings';
import type { ICvmOperateTableView } from '../../../typings';

interface IFormModel {
  hosts: ICvmOperateTableView[];
  agree: boolean;
}

interface ISubmitHostItem {
  id: string;
  bk_asset_id: string;
  device_type: string;
  image_name_old: string;
  cloud_image_id: string;
  image_name: string;
  image_type: string;
}

interface Exposes {
  formModel: IFormModel;
  submitHosts: Ref<ISubmitHostItem[]>;
}

const props = defineProps<{ list: ICvmOperateTableView[] }>();

const { t } = useI18n();

const columns = [
  { id: 'account_id', name: t('云账号'), type: 'user' },
  { id: 'type', name: t('类型'), type: 'string' },
  { id: 'count', name: t('数量'), type: 'string' },
  { id: 'region', name: t('地域'), type: 'region' },
  { id: 'image_name', name: t('重装后镜像名称'), type: 'string' },
] as ModelPropertyColumn[];

const { formModel } = useFormModel<IFormModel>({
  hosts: [],
  agree: false,
});

const expandRowColumns: ModelPropertyColumn[] = [
  { id: 'bk_asset_id', name: t('固资号'), type: 'string', width: 150 },
  { id: 'private_ip_address', name: t('内网IP'), type: 'string' },
  { id: 'public_ip_address', name: t('外网IP'), type: 'string' },
  { id: 'bk_host_name', name: t('主机名称'), type: 'string' },
  { id: 'region', name: t('地域'), type: 'region' },
  { id: 'zone', name: t('可用区'), type: 'string' },
  { id: 'device_type', name: t('机型'), type: 'string', width: 150 },
  { id: 'image_name_old', name: t('原镜像名称'), type: 'string', width: 250 },
];

// 用于提交的 hosts
const submitHosts = computed(() => {
  return formModel.hosts
    ? formModel.hosts
        .flatMap((curr) => curr.list)
        .map(({ id, bk_asset_id, device_type, image_name_old, cloud_image_id, image_name, image_type }) => ({
          id,
          bk_asset_id,
          device_type,
          image_name_old,
          cloud_image_id,
          image_name,
          image_type,
        }))
    : [];
});

watchEffect(() => {
  formModel.hosts = props.list;
});

defineExpose<Exposes>({
  formModel,
  submitHosts,
});
</script>

<template>
  <bk-form class="i-third-step-container" form-type="vertical" :model="formModel">
    <bk-form-item :label="t('操作系统')" :required="true">
      <bk-table row-hover="auto" :data="formModel.hosts" show-overflow-tooltip row-key="id">
        <bk-table-column type="expand" min-width="50" />
        <bk-table-column v-for="(column, index) in columns" :key="index" :prop="column.id" :label="column.name">
          <template #default="{ row }">
            <display-value
              :property="column"
              :value="row[column.id]"
              :display="column?.meta?.display"
              :vendor="row.vendor"
            />
          </template>
        </bk-table-column>

        <template #expandRow="rowData">
          <div class="expand-row-wrap">
            <bk-table
              :data="rowData.list"
              show-overflow-tooltip
              max-height="300px"
              min-height="auto"
              cell-class="expand-table-cell"
              row-key="id"
            >
              <bk-table-column
                v-for="(column, index) in expandRowColumns"
                :key="index"
                :prop="column.id"
                :label="column.name"
                :width="column.width"
              >
                <template #default="{ row }">
                  <display-value
                    :property="column"
                    :value="row[column.id]"
                    :display="column?.meta?.display"
                    :vendor="row.vendor"
                  />
                </template>
              </bk-table-column>
            </bk-table>
          </div>
        </template>
      </bk-table>
    </bk-form-item>
    <bk-form-item property="agree">
      <bk-checkbox v-model="formModel.agree">
        {{ t('我已确认重装的风险和重装后的操作步骤，点击提交后重装') }}
      </bk-checkbox>
    </bk-form-item>
  </bk-form>
</template>

<style scoped lang="scss">
.i-third-step-container {
  .expand-row-wrap {
    padding: 0 24px;
    background: #fafbfd;

    :deep(.expand-table-cell) {
      background: #fafbfd;
    }
  }
}
</style>
