import { Button } from 'bkui-vue';
import { defineComponent } from 'vue';

export default defineComponent({
  props: {
    name: { type: String, required: true },
    alias: String,
  },
  setup(props) {
    return () => (
      <Button
        text
        theme='primary'
        onClick={() => {
          window.open(`wxwork://message?username=${props.name}`);
        }}>
        {props.alias || props.name}
      </Button>
    );
  },
});
