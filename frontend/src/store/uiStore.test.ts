import { describe, it, expect } from 'vitest'
import { useUIStore } from './uiStore'

describe('useUIStore', () => {
  it('has correct initial state', () => {
    const state = useUIStore.getState()
    expect(state.sidebarCollapsed).toBe(false)
    expect(state.notifications).toEqual([])
  })

  it('toggleSidebar flips collapsed state', () => {
    expect(useUIStore.getState().sidebarCollapsed).toBe(false)

    useUIStore.getState().toggleSidebar()
    expect(useUIStore.getState().sidebarCollapsed).toBe(true)

    useUIStore.getState().toggleSidebar()
    expect(useUIStore.getState().sidebarCollapsed).toBe(false)
  })

  it('addNotification adds to the list', () => {
    useUIStore.setState({ notifications: [] })

    useUIStore.getState().addNotification({
      type: 'success',
      title: 'Patient created',
      message: 'Ramesh Kumar has been registered',
    })

    const notifications = useUIStore.getState().notifications
    expect(notifications).toHaveLength(1)
    expect(notifications[0].type).toBe('success')
    expect(notifications[0].title).toBe('Patient created')
    expect(notifications[0].id).toBeDefined()
  })

  it('removeNotification removes by id', () => {
    useUIStore.setState({
      notifications: [
        { id: '1', type: 'info', title: 'Test 1' },
        { id: '2', type: 'error', title: 'Test 2' },
      ],
    })

    useUIStore.getState().removeNotification('1')

    const notifications = useUIStore.getState().notifications
    expect(notifications).toHaveLength(1)
    expect(notifications[0].id).toBe('2')
  })

  it('addNotification generates unique ids', () => {
    useUIStore.setState({ notifications: [] })

    useUIStore.getState().addNotification({ type: 'info', title: 'A' })
    useUIStore.getState().addNotification({ type: 'info', title: 'B' })

    const notifications = useUIStore.getState().notifications
    expect(notifications[0].id).not.toBe(notifications[1].id)
  })
})
