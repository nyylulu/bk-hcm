import { defineComponent, type PropType } from 'vue';
import { useI18n } from 'vue-i18n';

import Panel from '@/components/panel';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';

import type { TicketDemands } from '@/typings/resourcePlan';

export default defineComponent({
  props: {
    demands: {
      type: Array as PropType<TicketDemands[]>,
    },
  },

  setup(props) {
    const { t } = useI18n();
    const { columns, settings } = useColumns('forecastDemandDetail');

    return () => (
      <Panel title={t('资源预测')}>
        <bk-table
          row-hover='auto'
          show-overflow-tooltip
          settings={settings.value}
          columns={columns}
          data={props.demands}
        />
      </Panel>
    );
  },
});
