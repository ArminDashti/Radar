# Routes and API Calls

## Frontend routes

| Path | Purpose | Auth required? |
|---|---|---|
| `/` | Redirect to host grid | No after redirect |
| `/hosts` | Host latency grid; hide/show hosts persisted in browser (`localStorage` `radar-visible-host-ids`) | No |
| `/endpoints` | Redirect to `/hosts` | No after redirect |
| `/probes` | Probe latency grid | Yes |
| `/request-host` | Submit host add requests | Yes |
| `/admin` | Redirect to `/admin/hosts` | Yes (admin) |
| `/admin/hosts` | Manage hosts (edit, test, delete) | Yes (admin) |
| `/admin/probes` | Rename probes | Yes (admin) |
| `/admin/requests` | Approve or reject host requests | Yes (admin) |
| `/about-me` | About Armin Dashti | No |
| `/login` | Obtain and store JWT | No |
| `/signup` | Create account and store JWT | No |

## API calls

| Method | Path | Purpose | Auth required? |
|---|---|---|---|
| POST | `/api/auth/login` | Exchange credentials for JWT | No |
| POST | `/api/auth/signup` | Create account (role `user`) | No |
| GET | `/api/me/host-prefs` | Legacy server host visibility (dashboard uses browser storage) | Yes |
| PUT | `/api/me/host-prefs` | Legacy server host visibility (dashboard uses browser storage) | Yes |
| GET | `/api/host-requests` | List host requests | Yes |
| POST | `/api/host-requests` | Submit host request | Yes |
| POST | `/api/host-requests/:id/approve` | Approve request | Yes (admin) |
| POST | `/api/host-requests/:id/reject` | Reject request | Yes (admin) |
| GET | `/api/probes` | Populate probes and filters | No |
| GET | `/api/hosts` | List configured hosts | Yes (admin) |
| POST | `/api/hosts` | Create a host | Yes (admin) |
| PUT | `/api/hosts/:id` | Update a host | Yes (admin) |
| DELETE | `/api/hosts/:id` | Delete a host | Yes (admin) |
| POST | `/api/hosts/:id/test` | On-demand HTTP/ICMP test | Yes (admin) |
| GET | `/api/admin/stats` | DB status and counts | Yes (admin) |
| GET | `/api/grid/hosts` | Fetch host latency grid | No |
| GET | `/api/grid/probes` | Fetch probe latency grid | No |
