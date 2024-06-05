import dayjs from 'dayjs';
const EMPTY_PLACEHOLDER = '-';
export function dateTimeTransform(value) {
  if (value) {
    return dayjs(value).format('YYYY-MM-DD HH:mm:ss');
  }
  return value || EMPTY_PLACEHOLDER;
}
