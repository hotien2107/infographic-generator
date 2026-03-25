export default {
  darkMode: ['class'],
  content: ['./index.html', './src/**/*.{js,jsx}'],
  theme: {
    extend: {
      colors: {
        border: 'hsl(241 83% 92%)',
        input: 'hsl(241 83% 92%)',
        ring: 'hsl(262 90% 62%)',
        background: 'hsl(222 100% 98%)',
        foreground: 'hsl(239 46% 16%)',
        primary: {
          DEFAULT: 'hsl(262 90% 62%)',
          foreground: 'hsl(0 0% 100%)',
        },
        secondary: {
          DEFAULT: 'hsl(190 95% 58%)',
          foreground: 'hsl(223 58% 18%)',
        },
        muted: {
          DEFAULT: 'hsl(231 74% 96%)',
          foreground: 'hsl(231 22% 45%)',
        },
        accent: {
          DEFAULT: 'hsl(325 100% 95%)',
          foreground: 'hsl(327 72% 32%)',
        },
        destructive: {
          DEFAULT: 'hsl(0 86% 61%)',
          foreground: 'hsl(0 0% 100%)',
        },
        card: {
          DEFAULT: 'hsl(0 0% 100%)',
          foreground: 'hsl(239 46% 16%)',
        },
      },
      borderRadius: {
        lg: '0.75rem',
        md: '0.5rem',
        sm: '0.375rem',
      },
      boxShadow: {
        soft: '0 24px 80px -32px rgba(67, 56, 202, 0.35)',
      },
    },
  },
  plugins: [],
}
