import { createContext, useContext, useEffect, useMemo, useState } from 'react'

const RouterContext = createContext(null)

function matchRoute(pattern, pathname) {
  const patternParts = pattern.split('/').filter(Boolean)
  const pathParts = pathname.split('/').filter(Boolean)

  if (patternParts.length != pathParts.length) return null

  const params = {}
  for (let index = 0; index < patternParts.length; index += 1) {
    const patternPart = patternParts[index]
    const pathPart = pathParts[index]
    if (patternPart.startsWith(':')) {
      params[patternPart.slice(1)] = decodeURIComponent(pathPart)
      continue
    }
    if (patternPart !== pathPart) return null
  }

  return params
}

export function AppRouter({ children, initialPath }) {
  const [pathname, setPathname] = useState(() => initialPath ?? window.location.pathname)

  useEffect(() => {
    if (initialPath) return undefined
    function handlePopState() {
      setPathname(window.location.pathname)
    }
    window.addEventListener('popstate', handlePopState)
    return () => window.removeEventListener('popstate', handlePopState)
  }, [initialPath])

  function navigate(to, options = {}) {
    setPathname(to)
    if (initialPath) return
    if (options.replace) {
      window.history.replaceState({}, '', to)
      return
    }
    window.history.pushState({}, '', to)
  }

  const value = useMemo(() => ({ pathname, navigate }), [pathname])

  return <RouterContext.Provider value={value}>{children}</RouterContext.Provider>
}

export function Link({ to, className, children, ...props }) {
  const router = useRouter()

  function handleClick(event) {
    if (props.onClick) props.onClick(event)
    if (event.defaultPrevented || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return
    event.preventDefault()
    router.navigate(to)
  }

  return (
    <a href={to} className={className} onClick={handleClick} {...props}>
      {children}
    </a>
  )
}

export function NavLink({ to, className, children, ...props }) {
  const router = useRouter()
  const isActive = router.pathname === to
  const resolvedClassName = typeof className === 'function' ? className({ isActive }) : className

  return (
    <Link to={to} className={resolvedClassName} {...props}>
      {typeof children === 'function' ? children({ isActive }) : children}
    </Link>
  )
}

export function useNavigate() {
  return useRouter().navigate
}

export function useParams(pattern) {
  const router = useRouter()
  return useMemo(() => matchRoute(pattern, router.pathname) ?? {}, [pattern, router.pathname])
}

export function usePathname() {
  return useRouter().pathname
}

function useRouter() {
  const value = useContext(RouterContext)
  if (!value) {
    throw new Error('Router context is not available')
  }
  return value
}
