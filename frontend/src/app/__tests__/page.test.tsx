import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import Page from '../page'

describe('Page', () => {
  it('renders the main page', () => {
    render(<Page />)
    
    // Check if the page renders without crashing
    expect(document.body).toBeInTheDocument()
  })

  it('renders without errors', () => {
    const consoleSpy = jest.spyOn(console, 'error').mockImplementation(() => {})
    
    render(<Page />)
    
    expect(consoleSpy).not.toHaveBeenCalled()
    consoleSpy.mockRestore()
  })
})