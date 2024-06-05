import ClipboardJs from 'clipboard';
import { Message } from 'bkui-vue';
export default {
  mounted(el, binding) {
    const copy = new ClipboardJs(el, {
      text() {
        return binding.value;
      },
      action() {
        return binding.arg === 'cut' ? 'cut' : 'copy';
      },
    });
    copy.on('success', () => {
      Message({
        message: '复制成功',
        theme: 'success',
        duration: 1500,
      });
    });
    copy.on('error', () => {
      Message({
        message: '复制失败',
        theme: 'error',
        duration: 1500,
      });
    });
    el._clipboard = copy;
  },
  updated(el, binding) {
    el._clipboard.text = () => {
      return binding.value;
    };
  },
  unmounted(el) {
    el._clipboard?.destroy();
  },
};
