import { defineComponent, ref, computed, watch, onMounted, nextTick } from 'vue';
import './index.scss';
export default defineComponent({
  props: {
    modelValue: {
      type: [Array, String],
    },
    placeholder: {
      type: String,
      default: '请输入 IP 地址，多行换行分割',
    },
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    const isShowInput = ref(false);
    const realInputVal = ref('');
    const showInputVal = ref('');
    watch(
      () => props.modelValue,
      (val) => {
        let inputVal = val;
        let showVal = val;
        if (Array.isArray(val)) {
          inputVal = val.join('\n');
          showVal = val.join(' ');
        } else {
          showVal = val?.replace(/\\n/g, ' ') || val;
        }
        realInputVal.value = inputVal;
        showInputVal.value = showVal;
      },
      { immediate: true },
    );
    const inputValSum = computed(() => {
      if (realInputVal.value) {
        return getValArr(realInputVal.value).length;
      }
      return props.modelValue.length;
    });
    const getValArr = (inputVal) => {
      if (!inputVal) return [];
      return inputVal.split('\n').filter((item) => item);
    };
    const updateRealValue = (value) => {
      emit('update:modelValue', getValArr(value));
    };
    const textAreaRef = ref(null);
    const switchInput = (boolVal) => {
      isShowInput.value = boolVal;
      if (boolVal) {
        nextTick(() => {
          textAreaRef.value.focus();
        });
      }
    };
    const clearRealVal = () => {
      realInputVal.value = '';
      updateRealValue('');
    };
    onMounted(() => {});
    return () => {
      if (!isShowInput.value) {
        return (
          <div class='virtual-input'>
            <bk-input
              key='input1'
              v-model={showInputVal.value}
              placeholder={props.placeholder}
              clearable
              onFocus={() => switchInput(true)}
              onClear={clearRealVal}>
              {{
                suffix: () => {
                  return <span class='num-style'>{inputValSum.value}</span>;
                },
              }}
            </bk-input>
          </div>
        );
      }
      return (
        <div class='real-input'>
          <bk-input
            ref={textAreaRef}
            key='input2'
            v-model={realInputVal.value}
            onChange={updateRealValue}
            autosize={{ minRows: 4, maxRows: 6 }}
            placeholder={props.placeholder}
            onBlur={() => switchInput(false)}
            type='textarea'
          />
        </div>
      );
    };
  },
});
