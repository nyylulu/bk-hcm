const capacityLevel = (capacityFlag: string | number) => {
  if (capacityFlag === 4) {
    return {
      class: 'c-success',
      text: '库存充足 (50 以上)',
    };
  }

  if (capacityFlag === 3) {
    return {
      class: 'c-warning',
      text: '少量库存 (11~50)',
    };
  }

  if (capacityFlag === 2) {
    return {
      class: 'c-danger',
      text: '库存紧张 (1~10)',
    };
  }

  if (capacityFlag === 0) {
    return {
      class: 'c-disabled',
      text: '无库存',
    };
  }

  return {
    class: '',
    text: '-',
  };
};
export { capacityLevel };
