# PLAN-014: Admin Frontend Pages

| | |
|-|-|
| **Status** | Not Started |
| **Date** | 2026-04-16 |
| **Depends on** | [PLAN-013](plan-013-jwt-auth.md) |

## Goal

Implement all admin panel pages with real data from the backend API.
Each page is a TanStack Query-powered React component using shadcn/ui primitives.

## Pages

### Dashboard `/dashboard`

Overview of the system at a glance.

**Widgets:**
- Accounts by status (Pending / Active / Frozen) — stat cards
- Total events in the event store — stat card
- Projector lag (events behind) — stat card with warning color if lag > 10
- Recent events table (last 20 global events, auto-refreshes every 5s)

```
┌──────────────┬──────────────┬──────────────┐
│  Total: 142  │  Active: 125 │  Lag: 3      │
│  Accounts    │  Accounts    │  Events      │
└──────────────┴──────────────┴──────────────┘

Recent events
┌────────────────────────┬──────────┬──────────────────┐
│ Time                   │ Type     │ Account ID       │
├────────────────────────┼──────────┼──────────────────┤
│ 2026-04-16 10:05:00   │ Deposit  │ 01960f...        │
└────────────────────────┴──────────┴──────────────────┘
```

### Accounts List `/accounts`

Searchable, sortable, paginated table of all accounts.

**Columns:** Account ID, Customer ID, Status (badge), Balance, Currency, Version
**Features:**
- Client-side search by account/customer ID (filter from loaded page)
- Status filter dropdown (All / Pending / Active / Frozen)
- Pagination (50 per page)
- Click row → navigate to account detail

### Account Detail `/accounts/:id`

Everything about one account.

**Sections:**
1. **Summary card** — ID, customer, status badge, balance, currency, version
2. **Event timeline** — vertical list of all domain events in chronological order.
   Each event shows: type (colored badge), version, occurred_at, expanded payload on click.
3. **Actions** — (read-only for now, can add commands later)

Event type colors:
- `AccountOpened` → blue
- `MoneyDeposited` → green
- `MoneyWithdrawn` → orange
- `AccountActivated` → teal
- `AccountFrozen` → red

### Projector Status `/projector`

Shows projector health.

**Content:**
- Projector name
- Last processed seq
- Latest global seq
- Lag (highlighted red if > 50)
- Auto-refreshes every 3s

## Data Fetching

All queries via TanStack Query hooks in `src/hooks/`:

```ts
// src/hooks/useAdminDashboard.ts
export const useAdminDashboard = () =>
  useQuery({ queryKey: ['admin', 'dashboard'], queryFn: fetchDashboard, refetchInterval: 5000 })

// src/hooks/useAccounts.ts
export const useAccounts = (page: number, limit: number) =>
  useQuery({ queryKey: ['admin', 'accounts', page, limit], queryFn: () => fetchAccounts(page, limit) })

// src/hooks/useAccountEvents.ts
export const useAccountEvents = (id: string) =>
  useQuery({ queryKey: ['admin', 'accounts', id, 'events'], queryFn: () => fetchAccountEvents(id) })

// src/hooks/useProjectorStatus.ts
export const useProjectorStatus = () =>
  useQuery({ queryKey: ['admin', 'projector'], queryFn: fetchProjectorStatus, refetchInterval: 3000 })
```

## Component Structure

```
src/
  components/
    ui/               ← shadcn/ui generated components
    layout/
      Sidebar.tsx     ← nav links
      AppShell.tsx    ← sidebar + main area
    accounts/
      AccountStatusBadge.tsx
      AccountTable.tsx
    events/
      EventTimeline.tsx
      EventTypeBadge.tsx
    dashboard/
      StatCard.tsx
      RecentEventsTable.tsx
    projector/
      ProjectorStatusCard.tsx
  pages/
    _authenticated/
      dashboard.tsx
      accounts/
        index.tsx
        $accountId.tsx
      projector.tsx
```

## Acceptance Criteria

- [ ] Dashboard loads and shows correct counts matching DB state
- [ ] Dashboard auto-refreshes stats every 5s without full page reload
- [ ] Accounts list shows all accounts with correct pagination (50/page)
- [ ] Status filter correctly filters the account list
- [ ] Account detail shows all events in version-ascending order
- [ ] Event payload expands on click (accordion / collapsible)
- [ ] Projector status shows correct lag and highlights red when lag > 50
- [ ] Projector status auto-refreshes every 3s
- [ ] Navigation between pages works without full reload
- [ ] All pages show loading skeletons while data is fetching
- [ ] All pages show error state if API call fails
- [ ] `pnpm build` passes with no TS errors

## Tasks

- [ ] Generate openapi-fetch client from admin OpenAPI spec (`openapi-typescript`)
- [ ] Implement TanStack Query hooks for all 4 data sources
- [ ] `AppShell` + `Sidebar` layout with nav links
- [ ] `StatCard` component
- [ ] Dashboard page — stat cards + recent events table with auto-refresh
- [ ] `AccountTable` component with pagination + status filter
- [ ] Accounts list page
- [ ] `EventTimeline` + `EventTypeBadge` components
- [ ] Account detail page — summary card + event timeline
- [ ] Projector status page with auto-refresh
- [ ] Loading skeleton states for all pages
- [ ] Error boundary / error state components
- [ ] Update docs
