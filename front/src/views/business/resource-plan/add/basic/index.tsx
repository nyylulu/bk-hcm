import { defineComponent, type PropType, watch, ref, onBeforeMount } from 'vue';
import { useI18n } from 'vue-i18n';
import { useResourcePlanStore } from '@/store/resourcePlan';
import Panel from '@/components/panel';
import BusinessSelector from '@/components/business-selector/index.vue';
import cssModule from './index.module.scss';

import type { IPlanTicket, IBizOrgRelation } from '@/typings/resourcePlan';

export default defineComponent({
  props: {
    modelValue: Object as PropType<IPlanTicket>,
  },

  emits: ['update:modelValue'],

  setup(props, { emit, expose }) {
    const { t } = useI18n();
    const resourcePlanStore = useResourcePlanStore();

    const isLoadingDemandClasses = ref(false);
    const demandClasses = ref([]);
    const productName = ref('');
    const planProductName = ref('');
    const deptName = ref('');
    const formRef = ref();

    const handleUpdateModelValue = (key: string, value: unknown) => {
      emit('update:modelValue', {
        ...props.modelValue,
        [key]: value,
      });
    };

    const handleInitDemandClassList = () => {
      isLoadingDemandClasses.value = true;
      resourcePlanStore
        .getDemandClasses()
        .then((data: { data: { details: string[] } }) => {
          demandClasses.value = data?.data?.details || [];
        })
        .finally(() => {
          isLoadingDemandClasses.value = false;
        });
    };

    const validate = () => {
      return formRef.value.validate();
    };

    watch(
      () => props.modelValue.bk_biz_id,
      () => {
        if (props.modelValue.bk_biz_id) {
          resourcePlanStore.getBizOrgRelation(props.modelValue.bk_biz_id).then((data: { data: IBizOrgRelation }) => {
            productName.value = data?.data?.bk_product_name;
            planProductName.value = data?.data.plan_product_name;
            deptName.value = data?.data?.virtual_dept_name;
          });
        } else {
          productName.value = '';
          planProductName.value = '';
          deptName.value = '';
        }
      },
      {
        immediate: true,
      },
    );

    onBeforeMount(handleInitDemandClassList);

    expose({
      validate,
    });

    return () => (
      <Panel title={t('基本信息')}>
        <bk-form form-type='vertical' ref={formRef} model={props.modelValue} class={cssModule.home}>
          <bk-form-item label={t('业务')} property='bk_biz_id' required>
            <BusinessSelector
              authed={true}
              modelValue={props.modelValue.bk_biz_id}
              onUpdate:modelValue={(biz: number) => handleUpdateModelValue('bk_biz_id', biz)}></BusinessSelector>
          </bk-form-item>
          <bk-form-item label={t('运营产品')}>
            <span class={cssModule.text}>{productName.value || '--'}</span>
          </bk-form-item>
          <bk-form-item label={t('规划产品')}>
            <span class={cssModule.text}>{planProductName.value || '--'}</span>
          </bk-form-item>
          <bk-form-item label={t('部门')}>
            <span class={cssModule.text}>{deptName.value || '--'}</span>
          </bk-form-item>
          <bk-form-item label={t('预测类型')} property='demand_class' required>
            <bk-select
              clearable
              loading={isLoadingDemandClasses.value}
              modelValue={props.modelValue.demand_class}
              onChange={(val: string) => handleUpdateModelValue('demand_class', val)}>
              {demandClasses.value.map((demandClass) => (
                <bk-option id={demandClass} name={demandClass}></bk-option>
              ))}
            </bk-select>
          </bk-form-item>
        </bk-form>
      </Panel>
    );
  },
});
