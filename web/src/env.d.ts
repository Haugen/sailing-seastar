/// <reference types="astro/client" />

interface ImportMetaEnv {
  readonly SUPABASE_API_URL: string;
  readonly SUPABASE_API_KEY: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

declare const L: any;
