import { defineComponent, type PropType, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import Panel from '@/components/panel';
import { timeFormatter } from '@/common/util';
import cssModule from './index.module.scss';

import { TicketBaseInfo } from '@/typings/resourcePlan';
export default defineComponent({
  props: {
    baseInfo: {
      type: Object as PropType<TicketBaseInfo>,
    },
    isBiz: {
      type: Boolean,
      required: true,
    },
  },
  setup(props) {
    const { t } = useI18n();

    const baseList = computed(() => [
      {
        label: t('需求类型'),
        value: props.baseInfo?.type_name,
      },
      {
        label: t('业务名称'),
        value: props.baseInfo?.bk_biz_name,
      },
      {
        label: t('部门'),
        value: props.baseInfo?.virtual_dept_name,
      },
      {
        label: t('运营产品'),
        value: props.baseInfo?.op_product_name,
      },
      {
        label: t('规划产品'),
        value: props.baseInfo?.plan_product_name,
      },
      {
        label: t('提单人'),
        value: props.baseInfo?.applicant,
      },
      {
        label: t('提单时间'),
        value: timeFormatter(props.baseInfo?.submitted_at, 'YYYY-MM-DD'),
      },
      {
        label: t('预测说明'),
        value: props.baseInfo?.remark,
      },
    ]);

    return () => (
      <Panel title={t('基本信息')}>
        <ul class={cssModule.home}>
          {baseList.value.map((item) => (
            <li>
              <span class={cssModule.label}>{item.label}：</span>
              <span class={cssModule.value} title={item.value}>
                {item.value || '--'}
              </span>
            </li>
          ))}
        </ul>
      </Panel>
    );
  },
});
