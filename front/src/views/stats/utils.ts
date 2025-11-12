import dayjs from 'dayjs';

export function getRecentMonths(months = 6, includeCurrent = true) {
  const end = dayjs().endOf('month');
  const startMonthOffset = includeCurrent ? months - 1 : months;
  const start = dayjs().subtract(startMonthOffset, 'month').startOf('month');

  return {
    startDate: start.toDate(),
    endDate: end.toDate(),
  };
}

export function getMonthRange(date: Date, format = 'YYYY-MM-DD') {
  return {
    startTime: dayjs(date).startOf('month').format(format),
    endTime: dayjs(date).endOf('month').format(format),
  };
}
