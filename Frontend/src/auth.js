// Auth bootstrap for Supabase
// Requires you to fill src/auth.config.js with your project credentials.
// Uses ESM import of supabase-js from jsdelivr.

import { SUPABASE_URL, SUPABASE_ANON_KEY } from './auth.config.js';

// Lazy import supabase-js from CDN to avoid bundling (bundle dependencies to avoid /npm paths)
const { createClient } = await import('https://esm.sh/@supabase/supabase-js@2?bundle');

export const supabase = createClient(SUPABASE_URL, SUPABASE_ANON_KEY, {
  auth: {
    persistSession: true,
    autoRefreshToken: true,
    detectSessionInUrl: true,
  }
});

async function withRetry(fn, { retries = 2, baseDelay = 300 } = {}) {
  let attempt = 0;
  // eslint-disable-next-line no-constant-condition
  while (true) {
    try {
      const res = await fn();
      // If supabase-js returns { error }, we still return it (no retry unless it's a network error we detect below)
      return res;
    } catch (e) {
      const msg = String(e?.message || e || "");
      const transient = /timed out|incomplete envelope|ECONNRESET|network|fetch failed/i.test(msg);
      if (attempt < retries && transient) {
        await new Promise(r => setTimeout(r, baseDelay * Math.pow(2, attempt)));
        attempt++;
        continue;
      }
      throw e;
    }
  }
}

export async function getSession() {
  const { data } = await supabase.auth.getSession();
  return data.session || null;
}

export async function signUpEmail({ email, password, username, captchaToken }) {
  return await withRetry(() => supabase.auth.signUp({
    email,
    password,
    options: {
      data: { username },
      captchaToken
    }
  }));
}

export async function signInEmail({ email, password, captchaToken }) {
  return await withRetry(() => supabase.auth.signInWithPassword({
    email,
    password,
    options: { captchaToken }
  }));
}

// Resolve a user's email from a username via the profiles table
export async function resolveEmailByUsername(username) {
  const { data, error } = await withRetry(() =>
    supabase.from('profiles').select('email').eq('username', username).maybeSingle()
  );
  if (error) return { error };
  if (!data) return { data: null };
  return { data };
}

// Sign in using either username or email identifier
export async function signInIdentifier({ identifier, password, captchaToken }) {
  const looksLikeEmail = /@/.test(identifier);
  let email = identifier;
  if (!looksLikeEmail) {
    const { data, error } = await resolveEmailByUsername(identifier.trim());
    if (error) return { error };
    if (!data?.email) return { error: { message: 'user_not_found' } };
    email = data.email;
  }
  return await signInEmail({ email, password, captchaToken });
}

export async function signInWithProvider(provider, options = {}) {
  return await withRetry(() => supabase.auth.signInWithOAuth({ provider, options }));
}

export async function signOut() {
  return await supabase.auth.signOut();
}

export async function signInAnonymously({ captchaToken } = {}) {
  return await withRetry(() => supabase.auth.signInAnonymously({ options: { captchaToken } }));
}

// Passkey support detection (platform authenticator / WebAuthn)
export async function isPasskeySupported() {
  try {
    if (!('PublicKeyCredential' in window)) return false;
    if (typeof PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable !== 'function') return false;
    return await PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable();
  } catch { return false; }
}

// Placeholder enrollment toggles UI only; real WebAuthn registration requires server-side challenge handling.
export async function markPasskeyEnrolled(flag = true) {
  if (flag) localStorage.setItem('salmar_has_passkey', '1');
  else localStorage.removeItem('salmar_has_passkey');
}

export function hasLocalPasskeyFlag() {
  return localStorage.getItem('salmar_has_passkey') === '1';
}

// Complete OAuth PKCE session from a redirect/deep link URL
export async function exchangeCodeFromUrl(url) {
  try {
    return await withRetry(() => supabase.auth.exchangeCodeForSession(url));
  } catch (e) {
    return { error: e };
  }
}

// WebAuthn (Passkeys) via Supabase Edge Functions
// Requires you to deploy functions at: https://<project-ref>.functions.supabase.co/webauthn/*
function functionsBase() {
  try {
    const u = new URL(SUPABASE_URL);
    const ref = u.hostname.split('.')[0];
    return `https://${ref}.functions.supabase.co`;
  } catch {
    return '';
  }
}

async function postJson(path, body) {
  const base = functionsBase();
  if (!base) throw new Error('Invalid Supabase URL for functions');
  const res = await fetch(`${base}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body || {})
  });
  const json = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(json?.error || res.statusText);
  return json;
}

export async function registerPasskey() {
  const { startRegistration } = await import('https://esm.sh/@simplewebauthn/browser@8?bundle');
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) throw new Error('Not authenticated');
  const email = user.email || '';
  const username = user.user_metadata?.username || (email ? email.split('@')[0] : 'user');
  const options = await postJson('/webauthn/generate-registration-options', {
    user_id: user.id,
    username,
    email
  });
  const attestation = await startRegistration(options);
  await postJson('/webauthn/verify-registration', { attestation, user_id: user.id });
  return { ok: true };
}

export async function authenticatePasskey(identifier) {
  const { startAuthentication } = await import('https://esm.sh/@simplewebauthn/browser@8?bundle');
  let email = identifier || '';
  if (email && !/@/.test(email)) {
    const r = await resolveEmailByUsername(email.trim());
    if (r?.data?.email) email = r.data.email; else throw new Error('user_not_found');
  }
  const options = await postJson('/webauthn/generate-authentication-options', { email });
  const assertion = await startAuthentication(options);
  const out = await postJson('/webauthn/verify-authentication', { assertion, email });
  if (out?.access_token && out?.refresh_token) {
    await supabase.auth.setSession({ access_token: out.access_token, refresh_token: out.refresh_token });
    return { ok: true };
  }
  if (out?.action_link) {
    // Use Supabase to exchange the generated magic link for a session
    await withRetry(() => supabase.auth.exchangeCodeForSession(out.action_link));
    return { ok: true };
  }
  throw new Error('auth_failed');
}
