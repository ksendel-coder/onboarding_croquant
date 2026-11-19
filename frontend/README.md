# Виджет онбординга

Опубликуйте файл `public/onboarding-widget.js` как статический ресурс. Vite
также копирует его в `dist/onboarding-widget.js` при сборке админки.
Подключение на HTML-странице:

```html
<script type="module">
  import { mountOnboardingWidget } from "/onboarding-widget.js";

  mountOnboardingWidget({
    apiUrl: "https://api.example.com/v1",
    appKey: "PUBLIC_APP_KEY",
  });
</script>
```

Укажите публичный ключ приложения и добавьте origin страницы в
`allowed_origins`. Страницу нужно открыть через HTTPS или localhost.
В SPA вызовите функцию очистки, которую возвращает `mountOnboardingWidget`,
при смене маршрута, а затем подключите виджет на новой странице.

Для сценария на нескольких страницах выберите входной путь сценария и
укажите путь страницы в первом шаге после каждого перехода, например
`/widget-test/profile`. Шаги с пустым полем продолжаются на предыдущей
странице. Виджет должен быть подключён на каждой странице с одинаковым
публичным ключом. Переходы поддерживаются в пределах одного origin.

## React + TypeScript + Vite

This template provides a minimal setup to get React working in Vite with HMR and some ESLint rules.

Currently, two official plugins are available:

- [@vitejs/plugin-react](https://github.com/vitejs/vite-plugin-react/blob/main/packages/plugin-react) uses [Oxc](https://oxc.rs)
- [@vitejs/plugin-react-swc](https://github.com/vitejs/vite-plugin-react/blob/main/packages/plugin-react-swc) uses [SWC](https://swc.rs/)

## React Compiler

The React Compiler is not enabled on this template because of its impact on dev & build performances. To add it, see [this documentation](https://react.dev/learn/react-compiler/installation).

## Expanding the ESLint configuration

If you are developing a production application, we recommend updating the configuration to enable type-aware lint rules:

```js
export default defineConfig([
  globalIgnores(['dist']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      // Other configs...

      // Remove tseslint.configs.recommended and replace with this
      tseslint.configs.recommendedTypeChecked,
      // Alternatively, use this for stricter rules
      tseslint.configs.strictTypeChecked,
      // Optionally, add this for stylistic rules
      tseslint.configs.stylisticTypeChecked,

      // Other configs...
    ],
    languageOptions: {
      parserOptions: {
        project: ['./tsconfig.node.json', './tsconfig.app.json'],
        tsconfigRootDir: import.meta.dirname,
      },
      // other options...
    },
  },
])

```

You can also install [eslint-plugin-react-x](https://npmx.dev/package/eslint-plugin-react-x) and [eslint-plugin-react-dom](https://npmx.dev/package/eslint-plugin-react-dom) for React-specific lint rules:

```js
// eslint.config.js
import reactX from 'eslint-plugin-react-x'
import reactDom from 'eslint-plugin-react-dom'

export default defineConfig([
  globalIgnores(['dist']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      // Other configs...
      // Enable lint rules for React
      reactX.configs['recommended-typescript'],
      // Enable lint rules for React DOM
      reactDom.configs.recommended,
    ],
    languageOptions: {
      parserOptions: {
        project: ['./tsconfig.node.json', './tsconfig.app.json'],
        tsconfigRootDir: import.meta.dirname,
      },
      // other options...
    },
  },
])

```
