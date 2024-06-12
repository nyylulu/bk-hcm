import isArray from 'lodash/isArray';
import isString from 'lodash/isString';
import isNumber from 'lodash/isNumber';

const DEBUG_PREFIX = '[simple-query-builder]';

const throwError = (msg) => {
  throw new Error(`${DEBUG_PREFIX} ${msg}`);
};
const consoleLog = (msg) => {
  console.log(`${DEBUG_PREFIX} ${msg}`);
};

/**
 * 操作符
 */
const AND = 'AND';
const OR = 'OR';
const EQUAL = 'equal';
const NOT_EQUAL = 'not_equal';
const GREATER = 'greater';
const LESS = 'less';
const LESS_OR_EQUAL = 'less_or_equal';
const GREATER_OR_EQUAL = 'greater_or_equal';
const DATETIME_LESS = 'datetime_less';
const DATETIME_GREATER = 'datetime_greater';
const DATETIME_LESS_OR_EQUAL = 'datetime_less_or_equal';
const DATETIME_GREATER_OR_EQUAL = 'datetime_greater_or_equal';

const abbrOperators = new Map();

abbrOperators.set('=', EQUAL);
abbrOperators.set('!=', NOT_EQUAL);
abbrOperators.set('>', GREATER);
abbrOperators.set('<', LESS);
abbrOperators.set('<=', LESS_OR_EQUAL);
abbrOperators.set('>=', GREATER_OR_EQUAL);
abbrOperators.set('d<', DATETIME_LESS);
abbrOperators.set('d>', DATETIME_GREATER);
abbrOperators.set('d>=', DATETIME_GREATER_OR_EQUAL);
abbrOperators.set('d<=', DATETIME_LESS_OR_EQUAL);

const conditions = [AND, OR];

const commonOperators = [EQUAL, NOT_EQUAL];

const arrayOperators = ['in', 'not_in', 'is_empty', 'is_not_empty'];

const timeOperators = [DATETIME_LESS, DATETIME_LESS_OR_EQUAL, DATETIME_GREATER, DATETIME_GREATER_OR_EQUAL];

const stringOperators = ['begins_with', 'not_begins_with', 'contains', 'not_contains', 'ends_with', 'not_ends_with'];

const nullOperators = ['is_null', 'is_not_null'];

const numberOperators = [LESS, LESS_OR_EQUAL, GREATER, GREATER_OR_EQUAL];

const existOperators = ['exist', 'not_exist'];

const allOperators = [
  ...commonOperators,
  ...arrayOperators,
  ...timeOperators,
  ...stringOperators,
  ...nullOperators,
  ...numberOperators,
  ...existOperators,
];

/**
 * 对 query builder 做校验和清理；简化 query builder 的使用
 * @param {Array} simpleConditions 简单版的查询条件
 * @returns 复杂版的查询条件
 */
export const transferSimpleConditions = (simpleConditions) => {
  if (!isArray(simpleConditions)) throwError('simpleConditions must be an array');

  const queryConditions = {
    condition: '',
    rules: [],
  };

  const emptyFields = [];

  const buildQueryConditions = (targetCondition, sourceCondition) => {
    sourceCondition.forEach((item, index) => {
      if (index === 0) {
        if (!isString(item)) throwError(`'${JSON.stringify(item)}' is not a string, condition must be a array`);

        if (!conditions.includes(item))
          throwError(`'${item}' is not a valid condition, condition must one of '${conditions.join(',')}'`);

        targetCondition.condition = item;
      }

      if (index > 0) {
        if (!isArray(item)) throwError(`'${JSON.stringify(item)}' is not an array, rule item must be a array`);

        // condition validation
        if (isString(item[0]) && conditions.includes(item[0])) {
          const newConditions = {
            condition: item[0],
            rules: [],
          };

          buildQueryConditions(newConditions, item);

          targetCondition.rules.push(newConditions);
        } else {
          // rule validation
          const [field, operator, value] = item;

          if (!isString(field)) throwError(`field '${field}' must be a string`);

          if (!isString(operator)) throwError(`operator '${operator}' must be a string`);

          if (!allOperators.includes(operator) && !abbrOperators.has(operator))
            throwError(`operator '${operator}' is not a valid operator, must one of '${allOperators.join(',')}'`);

          if (
            value === undefined ||
            value === null ||
            (isString(value) && value.trim() === '') ||
            (isArray(value) && value.length === 0) ||
            (isNumber(value) && isNaN(value))
          ) {
            emptyFields.push(field);
            return false;
          }

          if (arrayOperators.includes(operator) && !isArray(value)) throwError(`'${field}' value must be a array`);

          if (numberOperators.includes(operator) && !isNumber(value)) throwError(`'${field}' value must be a number`);

          if (stringOperators.includes(operator) && !isString(value)) throwError(`'${field}' value must be a string`);

          targetCondition.rules.push({
            field,
            operator: abbrOperators.get(operator) || operator,
            value,
          });
        }
      }
    });
  };

  buildQueryConditions(queryConditions, simpleConditions);

  if (emptyFields.length > 0) {
    consoleLog(`the value of fields '${emptyFields.join(', ')}' is empty, will be ignored`);
  }

  return queryConditions;
};
