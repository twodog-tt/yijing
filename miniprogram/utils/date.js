const CHINA_TIME_OFFSET_MS = 8 * 60 * 60 * 1000;

function pad(value) {
  return String(value).padStart(2, "0");
}

function toChinaClock(date) {
  return new Date(date.getTime() + CHINA_TIME_OFFSET_MS);
}

/** 无论用户设备位于哪个时区，都按 UTC+8 返回 YYYY-MM-DD。 */
function getChinaTodayDate(now = new Date()) {
  const chinaClock = toChinaClock(now);
  return [
    chinaClock.getUTCFullYear(),
    pad(chinaClock.getUTCMonth() + 1),
    pad(chinaClock.getUTCDate()),
  ].join("-");
}

/** 将可解析时间格式化为 UTC+8 的 YYYY-MM-DD HH:mm；无效输入返回空字符串。 */
function formatDateTime(value) {
  const date = value instanceof Date ? value : new Date(value);
  if (Number.isNaN(date.getTime())) return "";

  const chinaClock = toChinaClock(date);
  const datePart = [
    chinaClock.getUTCFullYear(),
    pad(chinaClock.getUTCMonth() + 1),
    pad(chinaClock.getUTCDate()),
  ].join("-");
  const timePart = [
    pad(chinaClock.getUTCHours()),
    pad(chinaClock.getUTCMinutes()),
  ].join(":");
  return `${datePart} ${timePart}`;
}

module.exports = {
  formatDateTime,
  getChinaTodayDate,
};
