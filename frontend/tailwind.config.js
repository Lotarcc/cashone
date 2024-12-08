/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // Main colors
        primary: '#7F3DFF',  // Purple accent
        background: {
          DEFAULT: '#1A1C1E',  // Main dark background
          light: '#242731',    // Lighter card background
        },
        // Status colors
        success: '#00A86B',    // Green for income
        danger: '#FD3C4A',     // Red for expenses
        warning: '#FACC15',    // Yellow for warnings/alerts
        // Text colors
        text: {
          DEFAULT: '#FFFFFF',   // Primary text
          secondary: '#E5E7EB', // Secondary text
          muted: '#9CA3AF',    // Muted text
        },
        // Additional UI colors
        border: '#374151',     // Border color
        divider: '#1F2937',    // Divider lines
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
      },
      // Custom box shadows for dark theme
      boxShadow: {
        card: '0 4px 6px -1px rgba(0, 0, 0, 0.5)',
        dropdown: '0 10px 15px -3px rgba(0, 0, 0, 0.5)',
      },
    },
  },
  plugins: [
    require('@tailwindcss/forms'),
  ],
}
