declare module "*.vue" {
  import type { DefineComponent } from "vue";
  const component: DefineComponent<object, object, unknown>;
  export default component;
}

interface ImportMetaEnv {
  readonly BASE_URL?: string;
  readonly VITE_API_BASE?: string;
  /** Base URL for product docs (e.g. hosted GitBook). Omit to use on-page anchors. */
  readonly VITE_DOCS_URL?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
