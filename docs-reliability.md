# User synchronization, reporting and Hy2 isolation

- Failed user-list application invalidates its ETag, including failures in the
  later alive-list request. Successful per-user changes are recorded locally
  before proceeding, so a retry does not repeat a deletion which already
  succeeded. Node metadata fetches similarly invalidate their cache after an
  incomplete update or decode failure.
- Traffic snapshots are accumulated separately from the immutable in-flight
  batch. Each batch contains at most 1000 users and keeps a cryptographically
  random 32-hex report id until acknowledged. Fresh counters continue to feed
  dynamic-limit accounting while a pending report retries. Unsent data is
  aggregated by user, not appended once per failed cycle.
- Deploy the companion Xboard PR and its receipt-table migration first. New
  panels acknowledge `X-Traffic-Report-ID` with `X-Traffic-Report-Ack`. A missing
  acknowledgement is accepted for old-panel compatibility and logged once;
  retry deduplication requires the upgraded panel. A mismatched acknowledgement
  is an error. Legacy `ReportUserTraffic` callers remain supported.
- Pending batches and queued deltas are in memory. An abrupt process exit or
  controller reload may lose unreported traffic; durable spooling is not
  introduced here. This is not an exactly-once guarantee across node crashes.
- Dynamic speed limits require valid positive Periodic (seconds), Traffic
  (bytes), SpeedLimit (Mbps), and ExpireTime (minutes). The timer is started and
  fresh bytes are counted once, including during panel outages. Expiry clears
  the temporary limit without deleting the user. Xray/sing connections retain
  a shared bucket even when initially unlimited, enabling live limit changes.
  Hy2 transport bandwidth shaping continues to use its native core settings;
  this change does not add a new Hy2 per-user transport shaper.
- Every Hy2 node owns its own authenticator and UUID-to-user map. Updating or
  revoking a user on one tag does not mutate another tag's authorization.

Regression tests cover a failed alive request after a changed user list,
partial apply followed by retry, a lost traffic acknowledgement, dynamic task
activation/expiry, and independent Hy2 node authentication. Run the full short
suite with the same build tags as CI; external ACME/DNS smoke tests remain
excluded. Run the changed packages with the race detector as well.
