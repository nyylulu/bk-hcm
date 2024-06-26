import ClipboardJs from 'clipboard';
import { Message } from 'bkui-vue';
export default {
  mounted(el, binding) {
    if (binding.arg === 'success') {
      el._clipboard_success = binding.value;
    } else {
      const copy = new ClipboardJs(el, {
        text() {
          // 复制空字符 报错 默认取值空格
          return binding.value || ' ';
        },
        action() {
          return binding.arg === 'cut' ? 'cut' : 'copy';
        },
      });
      copy.on('success', (e) => {
        const callback = el._clipboard_success;
        if (callback) {
          callback(e);
          return;
        }
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
    }
  },
  updated(el, binding) {
    if (binding.arg === 'success') {
      el._clipboard_success = binding.value;
    } else {
      el._clipboard.text = () => {
        return binding.value || ' ';
      };
    }
  },
  unmounted(el, binding) {
    if (binding.arg === 'success') {
      delete el._clipboard_success;
    } else {
      el._clipboard?.destroy();
    }
  },
};
