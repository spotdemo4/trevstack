/**
 * A color for `username` that stays the same across streams and page loads, so a user's
 * cursors are recognizable on every pad.
 */
export function userColor(username: string): string {
  let hash = 0;
  for (const char of username) hash = (Math.imul(hash, 31) + char.codePointAt(0)!) | 0;
  // The golden angle spreads similar hashes around the color wheel.
  const hue = Math.round(((((hash >>> 0) * 137.508) % 360) + 360) % 360);
  return `hsl(${hue} 75% 55%)`;
}
