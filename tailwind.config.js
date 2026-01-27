module.exports = {
  // Files to scan for Tailwind classes
  content: [
    './internal/generator/page.html',
    './target/build/**/*.html'
  ],

  // Tailwind plugins
  plugins: [
    require('@tailwindcss/typography')
  ],
}
