import { vi } from 'vitest';

// Mock window.$message
vi.stubGlobal('$message', {
  success: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
  info: vi.fn()
});

// Mock window.$loading
vi.stubGlobal('$loading', {
  start: vi.fn(),
  finish: vi.fn()
});

// Mock window.$dialog
vi.stubGlobal('$dialog', {
  confirm: vi.fn(),
  warning: vi.fn()
});

// Mock window.$notification
vi.stubGlobal('$notification', {
  success: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
  info: vi.fn()
});
