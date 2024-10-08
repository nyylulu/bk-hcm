import { ITimeRange } from '@/typings/plan';
import dayjs from 'dayjs';

// 用于计算两个日期范围的交集
export const getDateRangeIntersectionWithMonth = (start: string, end: string, month: number): ITimeRange => {
  const startDate = dayjs(start);
  const endDate = dayjs(end);

  const startYear = startDate.year();
  const endYear = endDate.year();

  let monthStart;
  let monthEnd;

  // 处理日期范围跨年和指定月份跨年的情况
  if (startDate.month() + 1 <= month) {
    monthStart = dayjs(`${startYear}-${month}-01`).startOf('month');
    monthEnd = dayjs(`${startYear}-${month}-01`).endOf('month');
  } else {
    monthStart = dayjs(`${endYear}-${month}-01`).startOf('month');
    monthEnd = dayjs(`${endYear}-${month}-01`).endOf('month');
  }

  const maxStart = startDate.isAfter(monthStart) ? startDate : monthStart;
  const minEnd = endDate.isBefore(monthEnd) ? endDate : monthEnd;

  // 确保返回的范围在指定月份内
  if (maxStart.isAfter(minEnd)) {
    return { start: monthStart.format('YYYY-MM-DD'), end: monthEnd.format('YYYY-MM-DD') };
  }

  return {
    start: maxStart.format('YYYY-MM-DD'),
    end: minEnd.format('YYYY-MM-DD'),
  };
};

// 判断当前日期是否在时间范围内
export const isDateInRange = (time: string, range: ITimeRange): boolean => {
  const timeDate = dayjs(time);
  const startDate = dayjs(range.start);
  const endDate = dayjs(range.end);

  return (
    (timeDate.isAfter(startDate) && timeDate.isBefore(endDate)) ||
    timeDate.isSame(startDate) ||
    timeDate.isSame(endDate)
  );
};
