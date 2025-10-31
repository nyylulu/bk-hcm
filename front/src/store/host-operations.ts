import { defineStore } from 'pinia';

export const useHostOperationsStore = defineStore('hostOperations', {
  state: () => ({
    operationType: '',
  }),
  actions: {
    setOperationType(type: string) {
      this.operationType = type;
    },
  },
});
