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
    handleClose: Function,
  },
  setup(props) {
    const activeName = ref('ApplicationList');
    const tabs = [
      {
        key: 'common',
        label: '常规资源池',
        component: () => <CommonResource formModelData={props.data} handleClose={props.handleClose} />,
      },
      {
        key: 'computed',
        label: '算力资源池',
        component: () => <ComputedResource formModelData={props.data} handleClose={props.handleClose} />,
      },
    ];

    return () => (
      <div class={'host-application-container'}>
        <Tab v-model:active={activeName.value} type='unborder-card' class={'tab-wrapper'}>
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
