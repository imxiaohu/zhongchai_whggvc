import { getTabBarHeightFromSystemInfo, normalizeNavHeightReadyPayload } from './system.js';

describe('normalizeNavHeightReadyPayload', () => {
  test('supports number payload', () => {
    expect(normalizeNavHeightReadyPayload(88)).toBe(88);
  });

  test('supports px string payload', () => {
    expect(normalizeNavHeightReadyPayload('90px')).toBe(90);
    expect(normalizeNavHeightReadyPayload('  90px ')).toBe(90);
  });

  test('supports object payload with height', () => {
    expect(normalizeNavHeightReadyPayload({ height: 92, heightPx: '92px' })).toBe(92);
  });

  test('supports object payload with heightPx only', () => {
    expect(normalizeNavHeightReadyPayload({ heightPx: '96px' })).toBe(96);
  });

  test('returns 0 for invalid payload', () => {
    expect(normalizeNavHeightReadyPayload(null)).toBe(0);
    expect(normalizeNavHeightReadyPayload(undefined)).toBe(0);
    expect(normalizeNavHeightReadyPayload({})).toBe(0);
    expect(normalizeNavHeightReadyPayload('abc')).toBe(0);
  });
});

describe('getTabBarHeightFromSystemInfo', () => {
  test('uses safeAreaInsets.bottom when available', () => {
    const h = getTabBarHeightFromSystemInfo({ safeAreaInsets: { bottom: 34 } }, 50);
    expect(h).toBe(84);
  });

  test('derives bottom inset from safeArea and screenHeight', () => {
    const h = getTabBarHeightFromSystemInfo({ screenHeight: 812, safeArea: { bottom: 778 } }, 50);
    expect(h).toBe(84);
  });

  test('falls back to base height when no safe area info', () => {
    const h = getTabBarHeightFromSystemInfo({}, 50);
    expect(h).toBe(50);
  });
});

