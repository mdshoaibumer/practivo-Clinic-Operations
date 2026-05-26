import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useAuthStore } from './authStore'

// Mock window.go
const mockLogin = vi.fn()
const mockLogout = vi.fn()
const mockGetCurrentUser = vi.fn()
const mockChangePassword = vi.fn()

beforeEach(() => {
  vi.clearAllMocks()
  // Reset store state
  useAuthStore.setState({
    user: null,
    isAuthenticated: false,
    isLoading: false,
  })
  ;(window as any).go = {
    handler: {
      AuthHandler: {
        Login: mockLogin,
        Logout: mockLogout,
        GetCurrentUser: mockGetCurrentUser,
        ChangePassword: mockChangePassword,
      },
    },
  }
})

describe('useAuthStore', () => {
  it('has correct initial state', () => {
    const state = useAuthStore.getState()
    expect(state.user).toBeNull()
    expect(state.isAuthenticated).toBe(false)
    expect(state.isLoading).toBe(false)
  })

  it('login sets user and isAuthenticated on success', async () => {
    mockLogin.mockResolvedValue({
      user: { id: '1', username: 'admin', fullName: 'Admin', role: 'admin' },
      loggedIn: true,
    })

    await useAuthStore.getState().login('admin', 'Password1')

    const state = useAuthStore.getState()
    expect(state.isAuthenticated).toBe(true)
    expect(state.user?.username).toBe('admin')
    expect(state.isLoading).toBe(false)
  })

  it('login throws and resets loading on failure', async () => {
    mockLogin.mockRejectedValue('[UNAUTHORIZED] Invalid credentials')

    await expect(useAuthStore.getState().login('admin', 'wrong')).rejects.toEqual({
      code: 'UNAUTHORIZED',
      message: 'Invalid credentials',
    })

    const state = useAuthStore.getState()
    expect(state.isAuthenticated).toBe(false)
    expect(state.isLoading).toBe(false)
  })

  it('logout clears user state', async () => {
    // Set up authenticated state
    useAuthStore.setState({
      user: { id: '1', username: 'admin', fullName: 'Admin', role: 'admin' },
      isAuthenticated: true,
    })

    mockLogout.mockResolvedValue(undefined)

    await useAuthStore.getState().logout()

    const state = useAuthStore.getState()
    expect(state.user).toBeNull()
    expect(state.isAuthenticated).toBe(false)
  })

  it('logout clears state even if backend call fails', async () => {
    useAuthStore.setState({
      user: { id: '1', username: 'admin', fullName: 'Admin', role: 'admin' },
      isAuthenticated: true,
    })

    mockLogout.mockRejectedValue(new Error('network error'))

    // logout() may propagate the error, but state should still be cleared
    try {
      await useAuthStore.getState().logout()
    } catch {
      // expected — logout doesn't suppress backend errors
    }

    const state = useAuthStore.getState()
    expect(state.user).toBeNull()
    expect(state.isAuthenticated).toBe(false)
  })

  it('checkSession restores authenticated state', async () => {
    mockGetCurrentUser.mockResolvedValue({
      user: { id: '1', username: 'admin', fullName: 'Admin', role: 'admin' },
      loggedIn: true,
    })

    await useAuthStore.getState().checkSession()

    const state = useAuthStore.getState()
    expect(state.isAuthenticated).toBe(true)
    expect(state.user?.username).toBe('admin')
  })

  it('checkSession clears state when not logged in', async () => {
    useAuthStore.setState({
      user: { id: '1', username: 'admin', fullName: 'Admin', role: 'admin' },
      isAuthenticated: true,
    })

    mockGetCurrentUser.mockResolvedValue({ user: null, loggedIn: false })

    await useAuthStore.getState().checkSession()

    const state = useAuthStore.getState()
    expect(state.isAuthenticated).toBe(false)
    expect(state.user).toBeNull()
  })

  it('checkSession clears state on error', async () => {
    useAuthStore.setState({ isAuthenticated: true })
    mockGetCurrentUser.mockRejectedValue('error')

    await useAuthStore.getState().checkSession()

    expect(useAuthStore.getState().isAuthenticated).toBe(false)
  })

  it('changePassword calls backend', async () => {
    mockChangePassword.mockResolvedValue(undefined)

    await useAuthStore.getState().changePassword('oldPass1', 'NewPass1x')

    expect(mockChangePassword).toHaveBeenCalledWith('oldPass1', 'NewPass1x')
  })

  it('changePassword throws parsed error on failure', async () => {
    mockChangePassword.mockRejectedValue('[VALIDATION_ERROR] Password too weak')

    await expect(
      useAuthStore.getState().changePassword('old', 'weak')
    ).rejects.toEqual({
      code: 'VALIDATION_ERROR',
      message: 'Password too weak',
    })
  })
})
