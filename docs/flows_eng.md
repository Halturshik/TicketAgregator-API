# Main Flows

## Registration

1. The client submits a profile, email, and password to `POST /api/auth/register`.
2. The API normalizes and validates the fields and checks whether the email is already in use.
3. The password is hashed with bcrypt before temporary registration data is stored in Redis.
4. A four-digit code is stored for five minutes and sent through the console mailer.
5. The client confirms the code through `POST /api/auth/register/confirm`.
6. The user is created in PostgreSQL and the temporary data is deleted.
7. The access JWT is returned in JSON and the refresh JWT is set in an HttpOnly cookie.

Code requests are limited to once every two minutes. Five incorrect attempts trigger a five-minute cooldown. Temporary registration data expires after ten minutes.

## Login and Refresh

Login also has two steps: the API first checks the email and password, then verifies a one-time code. The access token lasts 15 minutes and the refresh token lasts 14 days.

On `POST /api/auth/refresh`, the old refresh token is atomically replaced with a new one. Reuse of a rotated token is rejected. The user's token version is stored in PostgreSQL; changing the password increments it and invalidates earlier JWTs.

Password recovery does not reveal whether an email is registered. After the code is confirmed, the new password is saved, previous refresh sessions are deleted, and a new token pair is issued.

## Searching Offers

Search is split by transport type:

- `POST /api/search/air`
- `POST /api/search/railway`
- `POST /api/search/bus`

The client provides cities, a departure date, an optional return date, and the number of passengers. The search accepts 1–12 passengers and dates from today up to 12 months ahead.

The number of generated offers depends on how far away the trip is:

| Trip date | Number of offers |
| --- | ---: |
| 0–3 months | 52–88 |
| 4–6 months | 31–45 |
| 7–9 months | 19–28 |
| 10–12 months | 6–16 |

### Supplier Requests

1. The API loads the route and reference data from PostgreSQL.
2. It sends concurrent gRPC requests to `atlas`, `nexus`, and `vertex`.
3. The supplier service generates a shared schedule catalog and applies supplier-specific settings: available carriers, fare type, and price coefficient.
4. The API merges matching `schedule_id + fare_type` offers and keeps the lower price.
5. The result is balanced across `non_refundable` (about 40%), `standard` (35%), and `flexible` (25%) fares, then sorted by price.
6. The result is stored in Redis for 10 minutes.
7. The client receives the first page and requests more pages using `search_id`.

By default, a page contains 10 offers, with a maximum of 50. Search creation is limited to 10 requests per minute for guests and 20 for authenticated users. Reading cached pages is limited to 60 requests per minute.

For international air travel, the chance of a connection is 35%. A hub is selected from the three closest suitable options; if no suitable airport is available, a direct flight is generated. Each segment gets a separate flight number. The return trip uses the same carrier as the outbound trip in 75% of cases.

An offer's total price includes all selected passengers. `price_per_passenger` contains the per-passenger price. Authenticated users see the points they can earn; guests see the same amount with a registration prompt.

## Creating an Order

The client sends `search_id`, `trip_option_id`, passenger data, and the number of points to spend to `POST /api/orders/`.

The API retrieves the selected offer from the cache and does not accept the price, route, or supplier from the client. This prevents a client from changing offer details before creating an order.

Each passenger has a source:

- `self` — an authenticated client identifies the passenger as the account owner. Submitted name, birth date, and citizenship become an order snapshot and do not update the profile; a new document is saved to the user's documents after successful payment.
- `saved` — uses a passenger saved by the account owner. Data can be overridden, and `save_changes` applies those changes after successful payment.
- `new` — a new passenger who is saved automatically only after successful payment.

Guests can use only `new` and must provide an email address. A guest order receives a secret `guest_payment_token`.

### Documents

The required document is selected based on citizenship, age, and whether the trip is international:

| Passenger and route | Document |
| --- | --- |
| Russian citizen, 14 or older, domestic | `internal_passport` |
| Russian citizen, under 14, domestic | `birth_certificate` |
| Non-Russian citizen, domestic | `foreign_passport` |
| Russian citizen, international | `international_passport` |
| Non-Russian citizen, international | `foreign_passport` |

The API checks format, eligibility, and document expiration on the departure date. Birth certificates and foreign passports are automatically marked `verified`. Mock verification rejects Russian passports in 5% of cases; the result is deterministic for a document fingerprint and verification month. Existing documents are checked again for each new order, so a previous `verified` status is not permanent approval.

Each passenger and direction gets a separate ticket. A round trip for two passengers creates four tickets. An air connection remains one ticket with multiple segments.

An order receives a public number such as `ABC-12345` and expires after 15 minutes. Unpaid orders and unreferenced trip records are removed by background cleanup after their retention period.

## Payment

`POST /api/payments/mock` accepts the internal order ID and, for guests, the `guest_payment_token`.

1. The order is locked with `FOR UPDATE`.
2. Ownership, `created` status, and `expires_at` are checked.
3. The mock provider declines the payment in about 3% of cases.
4. The payment result is stored in PostgreSQL.
5. On success, the order and tickets move to `paid`.
6. For registered users, points are spent and earned.
7. New passengers and documents are saved automatically; permitted changes to existing passengers are applied.
8. The whole transaction commits or rolls back.

Points earned equal `floor(total_price * 0.02)`. The client can spend no more than half of the order price or their current balance. Guests cannot use loyalty points.

## Order History

`GET /api/my/orders/` returns only the current user's `paid`, `partially_refunded`, and `refunded` orders. `created` and `expired` orders are excluded.

Each page contains up to five orders. For a “Load more” interface, the client increases `offset` by five and appends the new results to those already displayed. The `transport=avia|rail|bus` filter is available.

History uses snapshots: changes to a saved passenger or document do not rewrite past tickets.

## Refunds

The client can first request a refund quote and then submit the refund. It can select specific tickets or all remaining tickets in the order.

Fare rules:

- `non_refundable` — no refund is allowed.
- `standard` — 90% is refundable when requested at least 72 hours before departure.
- `flexible` — 70% is refundable 24–72 hours before departure and 90% earlier than that.

Execution requires a UUID in the `Idempotency-Key` header. Repeating the same request with the same key returns the existing operation. Reusing a key with a different ticket set returns `409 Conflict`.

### Distributed Execution

1. The API creates a local operation with `processing` status and moves tickets to `refund_pending`.
2. It calls the original supplier's `ExecuteRefund` method over gRPC.
3. The supplier stores the result in its own database under the idempotency key.
4. On success, the API records `success`, marks tickets `refunded`, and updates the order status to `partially_refunded` or `refunded`.
5. On an explicit rejection, the operation becomes `failed` and tickets return to `paid`.
6. If the response times out, the API returns `processing` and a background goroutine retries later.
7. After eight attempts or 24 hours, the operation moves to `requires_review`.

`Retry-After` indicates when the client should check an unfinished `processing` operation again.

### Money and Loyalty Points

`current_total_price` is reduced by the full original price of the refunded tickets, regardless of the supplier's fee. The fare policy determines the cash amount the supplier agrees to refund.

Spent points are recalculated so that they still cover no more than 50% of the remaining order price. Any excess is restored to the points balance and deducted from the cash refund. Earned points are recalculated from the order's current price.

If points earned from the order have already been spent, the balance does not become negative. The shortfall is recorded as `bonus_debt` and repaid from future accruals.

## Guest Booking Access

1. `POST /api/bookings/lookup` accepts a ticket or order number.
2. The API returns a public summary only: route, times, transport, ticket/passenger counts, and a masked passenger name for ticket lookups.
3. The guest submits the same ticket or order number and email to `/api/bookings/access/request`.
4. If the email belongs to the guest order, a verification code is sent and a `challenge_id` is returned.
5. `/api/bookings/access/confirm` verifies the code and sets an HttpOnly `booking_access_token` cookie that lasts 15 minutes.
6. The cookie grants access to booking details, refund quotes, and refund submission by public order number.

The name remains abbreviated in the details, and only the last four document characters are shown. An authenticated user can access only their own order. Unauthorized access returns `404` to make order number enumeration harder.

Public lookup methods are limited to 20 requests per minute per IP address.

## Public Trip Lookup

`POST /api/trips/lookup` supports two modes:

- `ticket_number` returns all segments for a specific ticket.
- `route_number + date` returns trips with that route number on the given day.

Only trips linked to paid or refunded orders are returned. The response does not include personal data, documents, prices, or the order owner.
