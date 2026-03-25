import { cva } from 'class-variance-authority'
import React from 'react'

import { cn } from '@/lib/utils'

const buttonVariants = cva(
  'inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-xl text-sm font-semibold transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 ring-offset-background',
  {
    variants: {
      variant: {
        default: 'bg-gradient-to-r from-primary to-secondary text-primary-foreground shadow-md shadow-primary/30 hover:scale-[1.01] hover:brightness-105',
        outline: 'border border-primary/20 bg-white text-foreground hover:border-primary/40 hover:bg-primary/5',
        secondary: 'bg-gradient-to-r from-fuchsia-500 to-rose-500 text-white shadow-md shadow-fuchsia-500/30 hover:brightness-105',
        ghost: 'text-foreground hover:bg-accent hover:text-accent-foreground',
      },
      size: {
        default: 'h-10 px-4 py-2',
        sm: 'h-9 rounded-lg px-3',
        lg: 'h-11 rounded-xl px-8',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'default',
    },
  },
)

const Button = React.forwardRef(({ asChild = false, children, className, variant, size, ...props }, ref) => {
  const classes = cn(buttonVariants({ variant, size, className }))

  if (asChild && React.isValidElement(children)) {
    return React.cloneElement(children, {
      ...props,
      className: cn(classes, children.props.className),
    })
  }

  return (
    <button className={classes} ref={ref} {...props}>
      {children}
    </button>
  )
})
Button.displayName = 'Button'

export { Button, buttonVariants }
