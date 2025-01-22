# Raiden support mode

## Rule
- only supported in `self_hosted`
- available mode are `bff` and `svc`
- direct connect to `postrest` server
- direct connect to `supabae/postgres-meta` server

## BFF Mode
### Config
- `ACCESS_TOKEN` : supabase access token
- `ANON_KEY` : supabase anon key
- `SERVICE_KEY` : supabase service key
- `SUPABASE_API_URL` : supabase api public url
- `SUPABASE_API_BASE_PATH` : api base path
- `SUPABASE_PUBLIC_URL` : pg-meta public url

### Controller
support all raiden controller type

### Route
- include all default supabase route (auth, realtime, storage, function and rpc) 
- support custom route from custom controller

## SVC Mode
### Config
- `PG_META_URL` : pg meta base url
- `POSTGREST_URL` : postrest base url

## Controller
only support custom controller

## Route
only support custom route from custom controller
