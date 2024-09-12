export default function useCvmChargeType() {
  const cvmChargeTypes = {
    PREPAID: 'PREPAID',
    POSTPAID_BY_HOUR: 'POSTPAID_BY_HOUR',
  };

  const cvmChargeTypeNames = {
    [cvmChargeTypes.PREPAID]: '包年包月',
    [cvmChargeTypes.POSTPAID_BY_HOUR]: '按量计费',
  };

  const cvmChargeTypeTips = {
    [cvmChargeTypes.PREPAID]:
      '默认为3年，按梯度折扣分别为1-3月150%，4-6月130%，7-11月120%，1年110%，2年105%，3年100%，4年95%，',
    [cvmChargeTypes.POSTPAID_BY_HOUR]:
      '使用小于1月折扣为170%，在提交预测单后，满3月后可在腾讯云控制台转换为包年包月，无预测单不可转包年包月。计费折扣，使用1-3月150%，满3月后转4-6月130%，7-11月120%，1年110%，2年105%，3年100%，4年95%，',
  };

  // cvm购买时长选项
  const cvmChargeMonths = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 24, 36, 48];
  const getMonthName = (m: number) => {
    if (m === 6) {
      return '半年';
    }
    if (m >= 12) {
      return `${m / 12}年`;
    }
    return `${m}月`;
  };
  const cvmChargeMonthOptions = cvmChargeMonths.map((month) => ({ id: month, name: getMonthName(month) }));

  return {
    cvmChargeTypes,
    cvmChargeTypeNames,
    cvmChargeTypeTips,
    cvmChargeMonthOptions,
  };
}
