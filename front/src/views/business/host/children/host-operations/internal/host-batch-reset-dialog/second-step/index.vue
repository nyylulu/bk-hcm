<script setup lang="ts">
import { useTemplateRef, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';

import type { ICvmListRestStatus } from '@/store/cvm/reset';
import useFormModel from '@/hooks/useFormModel';
import { IDC_SVC_SOURCE_TYPE_IDS, IMAGE_TYPE_NAME } from '../constants';
import type { ModelPropertyColumn } from '@/model/typings';
import type { ITableModel } from '../typings';

import { Form } from 'bkui-vue';
import ImageSelector, { IImageItem } from './image-selector.vue';
import PwdInput from '@/views/service/service-apply/cvm/children/PwdInput';

interface IFormModel {
  hosts: ITableModel[];
  pwd: string;
  pwd_confirm: string;
}
interface Exposes {
  formModel: IFormModel;
  validateEmpty: () => boolean;
  validateForm: () => Promise<boolean>;
}

const props = defineProps<{ list: ICvmListRestStatus[] }>();

const { t } = useI18n();

const columns = [
  { id: 'account_id', name: t('云账号'), type: 'user' },
  { id: 'type', name: t('类型'), type: 'string' },
  { id: 'count', name: t('数量'), type: 'string' },
  { id: 'region', name: t('地域'), type: 'region' },
] as ModelPropertyColumn[];

const formRef = useTemplateRef<typeof Form>('form');
const { formModel } = useFormModel<IFormModel>({ hosts: [], pwd: '', pwd_confirm: '' });
const rules = {
  pwd: [
    {
      validator: (value: string) => value.length >= 8 && value.length <= 30,
      message: t('密码长度需要在8-30个字符之间'),
      trigger: 'blur',
    },
    {
      validator: (value: string) => {
        const pattern = /^(?=.*[A-Z])(?=.*[a-z])(?=.*\d|.*[!@$%^\-_=+[{}\]:,./?])[A-Za-z\d!@$%^\-_=+[{}\]:,./?]+$/;
        return pattern.test(value);
      },
      message: t('密码不符合复杂度要求'),
      trigger: 'blur',
    },
    {
      validator: (value: string) => {
        if (formModel.pwd_confirm.length) {
          return value === formModel.pwd_confirm;
        }
        return true;
      },
      message: t('两次输入的密码不一致'),
      trigger: 'blur',
    },
  ],
  pwd_confirm: [
    {
      validator: (value: string) => value.length >= 8 && value.length <= 30,
      message: t('密码长度需要在8-30个字符之间'),
      trigger: 'blur',
    },
    {
      validator: (value: string) => {
        const pattern = /^(?=.*[A-Z])(?=.*[a-z])(?=.*\d|.*[!@$%^\-_=+[{}\]:,./?])[A-Za-z\d!@$%^\-_=+[{}\]:,./?]+$/;
        return pattern.test(value);
      },
      message: t('密码不符合复杂度要求'),
      trigger: 'blur',
    },
    {
      validator: (value: string) => formModel.pwd.length && value === formModel.pwd,
      message: t('两次输入的密码不一致'),
      trigger: 'blur',
    },
  ],
};

const handleImageNameChange = (image: IImageItem, row: any) => {
  const { name, type, cloud_id } = image || {};
  row.image_name = name;

  // 为聚合条下的每个主机添加批量重装参数
  row.list = row.list.map((item: ICvmListRestStatus) => {
    return {
      ...item,
      image_name_old: item.bk_os_name,
      cloud_image_id: cloud_id,
      image_name: name,
      image_type: type,
    };
  });
};

const validateEmpty = () => {
  const v1 = formModel.hosts.some((item) => item.image_type === '' || item.image_name === '');
  const v2 = formModel.pwd === '' || formModel.pwd_confirm === '';
  return v1 || v2;
};

const validateForm = async () => formRef.value.validate();

watchEffect(() => {
  // 将重装列表根据『云账号』+『类型』+『地域』进行分组
  const initialMap = props.list.reduce((prev, curr) => {
    const { account_id, bk_svr_source_type_id, region, vendor } = curr;
    const type = IDC_SVC_SOURCE_TYPE_IDS.includes(bk_svr_source_type_id) ? t('物理机') : t('云主机');

    const key = `${account_id}-${type}-${region}`;
    if (!prev.has(key)) {
      prev.set(key, { account_id, type, count: 0, region, vendor, image_type: '', image_name: '', list: [] });
    }

    const target = prev.get(key);
    Object.assign(target, { count: target.count + 1, list: [...target.list, curr] });

    return prev;
  }, new Map<string, ITableModel>());

  formModel.hosts = initialMap.values().toArray();
});

defineExpose<Exposes>({
  formModel,
  validateEmpty,
  validateForm,
});
</script>

<template>
  <bk-form class="i-second-step-container" form-type="vertical" :model="formModel" :rules="rules" ref="form">
    <bk-form-item :label="t('操作系统')" property="hosts" :required="true">
      <bk-table row-hover="auto" :data="formModel.hosts" pagination show-overflow-tooltip row-key="id">
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

        <bk-table-column :label="t('镜像类型')">
          <template #default="{ row }">
            <hcm-form-enum
              v-model="row.image_type"
              :option="IMAGE_TYPE_NAME"
              :display="{ on: 'cell' }"
              style="height: 42px"
              @change="handleImageNameChange(null, row)"
            />
          </template>
        </bk-table-column>

        <bk-table-column :label="t('镜像名称')" width="300">
          <template #default="{ row }">
            <image-selector
              :model-value="row.image_name"
              :account-id="row.account_id"
              :region="row.region"
              :vendor="row.vendor"
              :image-type="row.image_type"
              @change="(image) => handleImageNameChange(image, row)"
            />
          </template>
        </bk-table-column>
      </bk-table>
    </bk-form-item>

    <bk-form-item
      :label="t('密码')"
      property="pwd"
      :description="t('密码必须包含3种组合：1.大写字母，2.小写字母，3. 数字或特殊字符（!@$%^-_=+[{}]:,./?）')"
      :required="true"
    >
      <pwd-input v-model="formModel.pwd" class="pwd-input" />
    </bk-form-item>

    <bk-form-item :label="t('确认密码')" property="pwd_confirm" :required="true">
      <bk-input v-model="formModel.pwd_confirm" class="pwd-input" @change="formRef.validate('pwd')" />
    </bk-form-item>
  </bk-form>
</template>

<style scoped lang="scss">
.i-second-step-container {
  .pwd-input {
    width: 420px;
  }
}
</style>
