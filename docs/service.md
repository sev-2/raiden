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

### Controller
only support custom controller

### Route
only support custom route from custom controller

## RUN
### Docker Compose
```
services:
  postgres:
    image: postgres:17
    container_name: postgres
    environment:
      POSTGRES_USER: developer
      POSTGRES_PASSWORD: devsecret
      POSTGRES_DB: boilerplate
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    networks:
      - backend

  postgrest:
    image: postgrest/postgrest
    container_name: postgrest
    environment:
      PGRST_DB_URI: "postgres://developer:devsecret@postgres:5432/boilerplate"
      PGRST_DB_SCHEMAS: 'public'
      PGRST_DB_ANON_ROLE: anon
      PGRST_JWT_SECRET: "z7F$8mP!wXq4V@kN#Ld9RbYp2Tc&JgM0"
      PGRST_DB_USE_LEGACY_GUCS: "false"
    depends_on:
      - postgres
    ports:
      - "3000:3000"
    networks:
      - backend

  pg-meta:
    container_name: pg-meta
    image: supabase/postgres-meta:v0.84.2
    environment:
      PG_META_HOST: "0.0.0.0"
      PG_META_PORT: 8080
      PG_META_DB_HOST: "postgres"
      PG_META_DB_NAME: "boilerplate"
      PG_META_DB_USER: "developer"
      PG_META_DB_PORT: 5432
      PG_META_DB_PASSWORD: "devsecret"
    depends_on:
      - postgres
    ports:
      - "8080:8080"
    networks:
      - backend

volumes:
  postgres_data:

networks:
  backend:
    driver: bridge
```

postrest available on `http://localhost:3000`
pg-meta available on `http://localhost:8080`

### Init Postrest

run this query to create and grant access permission to anon role
```
CREATE ROLE anon NOLOGIN;

GRANT USAGE ON SCHEMA public TO anon;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO anon;

```