# Running the App

## First-time setup

**1. Install dependencies**
```bash
npm install
```

**2. Start the database**
```bash
docker run -d --name klapp-postgres \
  -e POSTGRES_USER=klapp \
  -e POSTGRES_PASSWORD=klapp \
  -e POSTGRES_DB=klapp \
  -p 5432:5432 \
  postgres:16-alpine

docker exec klapp-postgres psql -U klapp -c "CREATE DATABASE klapp_test;"
```

**3. Create your `.env` file**
```bash
cp .env.example .env
```

Then open `.env` and set:
```
DATABASE_URL="postgresql://klapp:klapp@localhost:5432/klapp?schema=public"
DATABASE_URL_TEST="postgresql://klapp:klapp@localhost:5432/klapp_test?schema=public"
AUTH_SECRET="any-random-32-char-string-here"
```

**4. Apply the database schema**
```bash
npx prisma migrate dev
```

**5. Seed test data**
```bash
npm run db:seed
```

**6. Start the app**
```bash
npm run dev
```

Open **http://localhost:3000**

---

## Returning after a restart

The Docker container stops when your machine restarts. Just run:
```bash
docker start klapp-postgres
npm run dev
```

---

## Test credentials

| Role     | Email                  | Password        |
|----------|------------------------|-----------------|
| Admin    | admin@klaus.test       | admin12345      |
| Employee | manthan@klaus.test     | employee12345   |
| Employee | thomas@klaus.test      | employee12345   |

**Admin** can see all jobs, all hours, and manage customers, employees, job types, and materials.

**Employee** can log jobs and view their own hours.
