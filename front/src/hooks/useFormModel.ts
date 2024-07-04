import { reactive, UnwrapRef } from 'vue';

function useFormModel<T extends object>(initialState: T) {
  const formModel = reactive({ ...initialState }) as UnwrapRef<T>;

  function resetForm(defaults = {}) {
    Object.assign(formModel, initialState, defaults);
  }
  function deepClear(obj: any, skipKey: string) {
    if (Array.isArray(obj)) {
      // 如果是数组，直接清空数组
      obj.length = 0;
    } else if (typeof obj === 'object' && obj !== null) {
      // 处理对象
      Object.keys(obj).forEach((key) => {
        if (key === skipKey) {
          return; // 跳过第一个属性
        }
        if (typeof obj[key] === 'object' && obj[key] !== null) {
          deepClear(obj[key], skipKey); // 递归处理嵌套对象和数组
        } else {
          obj[key] = '';
        }
      });
    }
  }
  function forceClear() {
    const keys = Object.keys(formModel);
    if (keys.length > 0) {
      const firstKey = keys[0];
      deepClear(formModel, firstKey); // 传递第一个键名以跳过该属性
    }
  }

  function setFormValues(values: Partial<T>) {
    Object.assign(formModel, values);
  }

  return {
    formModel,
    resetForm,
    forceClear,
    setFormValues,
  };
}

export default useFormModel;
