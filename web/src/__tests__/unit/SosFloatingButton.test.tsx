import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { MemoryRouter } from 'react-router-dom'
import SosFloatingButton from '../../components/SosFloatingButton'

describe('SosFloatingButton', () => {
  it('渲染 SOS 入口链接', () => {
    render(
      <MemoryRouter initialEntries={['/pods']}>
        <SosFloatingButton />
      </MemoryRouter>,
    )
    expect(screen.getByRole('link', { name: /sos/i })).toBeInTheDocument()
  })

  it('通话页内隐藏', () => {
    render(
      <MemoryRouter initialEntries={['/sos']}>
        <SosFloatingButton />
      </MemoryRouter>,
    )
    expect(screen.queryByRole('link', { name: /sos/i })).not.toBeInTheDocument()
  })
})
