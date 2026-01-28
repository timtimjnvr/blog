module.exports = {
  // Files to scan for Tailwind classes
  content: [
    './internal/generator/page.html',
    './target/build/**/*.html'
  ],

  // Dark mode controlled by 'dark' class on <html>
  darkMode: 'class',

  // Tailwind plugins
  plugins: [
    require('@tailwindcss/typography')
  ],
}
