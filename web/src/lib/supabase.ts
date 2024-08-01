import { createClient } from "@supabase/supabase-js";

export const supabase = createClient(
  import.meta.env.SUPABASE_API_URL,
  import.meta.env.SUPABASE_API_KEY,
);
