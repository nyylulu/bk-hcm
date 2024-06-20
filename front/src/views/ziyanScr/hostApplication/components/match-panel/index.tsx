import { defineComponent, ref } from 'vue';
import './index.scss';
import ComputedResource from './components/computed-resource';
import CommonResource from './components/common-resource';
import { Tab } from 'bkui-vue';
import { BkTabPanel } from 'bkui-vue/lib/tab';

export default defineComponent({
  props: {
    data: {
      type: Object,
      require: true,
    },
  },
  setup(props) {
    const activeName = ref('ApplicationList');
    const tabs = [
      {
        key: 'common',
        label: '常规资源池',
        component: () => <CommonResource />,
      },
      {
        key: 'computed',
        label: '算力资源池',
        component: () => <ComputedResource formModelData={props.data} />,
      },
    ];

    return () => (
      <div class={'host-application-container'}>
        <Tab v-model:active={activeName.value} type='card-grid' class={'tab-wrapper'}>
          {tabs.map(({ key, label, component }) => (
            <BkTabPanel key={key} label={label} name={key} renderDirective='if'>
              {component()}
            </BkTabPanel>
          ))}
        </Tab>
      </div>
    );
  },
});
