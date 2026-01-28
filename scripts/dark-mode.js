(function () {
  const STORAGE_KEY = 'theme';
  const DARK = 'dark';
  const LIGHT = 'light';
  const AUTO = 'auto';

  // Get system preference
  function getSystemTheme() {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? DARK : LIGHT;
  }

  // Get stored preference (defaults to auto)
  function getStoredPreference() {
    return localStorage.getItem(STORAGE_KEY) || AUTO;
  }

  // Apply theme to document
  function applyTheme(theme) {
    if (theme === DARK) {
      document.documentElement.classList.add(DARK);
    } else {
      document.documentElement.classList.remove(DARK);
    }
  }

  // Update theme based on current preference
  function updateTheme() {
    const preference = getStoredPreference();
    if (preference === AUTO) {
      applyTheme(getSystemTheme());
    } else {
      applyTheme(preference);
    }
  }

  // Set theme preference
  function setTheme(preference) {
    if (preference === AUTO) {
      localStorage.removeItem(STORAGE_KEY);
    } else {
      localStorage.setItem(STORAGE_KEY, preference);
    }
    console.log('setTheme:', preference, 'system:', getSystemTheme(), 'stored:', getStoredPreference());
    updateTheme();
  }

  // Expose functions globally
  window.setTheme = setTheme;
  window.getThemePreference = getStoredPreference;

  // Initialize
  updateTheme();

  // Listen for system preference changes
  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', function () {
    if (getStoredPreference() === AUTO) {
      updateTheme();
    }
  });
})();
