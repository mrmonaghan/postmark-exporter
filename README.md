# postmark-exporter

This application collects stats from the Postmark API and emits them as Prometheus metrics.

## Metrics

| Metric Name                            | Type  | Description                                         |
|----------------------------------------|-------|-----------------------------------------------------|
| `postmark_bounced_total`               | Gauge | Total number of bounced emails                      |
| `postmark_hard_bounced_total`          | Gauge | Total number of hard bounced emails                 |
| `postmark_soft_bounced_total`          | Gauge | Total number of soft bounced emails                 |
| `postmark_transient_bounced_total`     | Gauge | Total number of transient bounced emails            |
| `postmark_sent_total`                  | Gauge | Total number of sent emails                         |
| `postmark_smtp_api_errors_total`       | Gauge | Total number of SMTP API errors                     |
| `postmark_bounce_rate`                 | Gauge | Percentage of sent emails that have bounced         |
| `postmark_spam_complaints_total`       | Gauge | Total number of spam complaints                     |
| `postmark_spam_complaints_rate`        | Gauge | Rate of spam complaints                             |
| `postmark_opens_total`                 | Gauge | Total number of email opens                         |
| `postmark_unique_opens_total`          | Gauge | Total number of unique email opens                  |
| `postmark_tracked_total`               | Gauge | Total number of tracked emails                      |
| `postmark_with_link_tracking_total`    | Gauge | Total number of emails with link tracking enabled   |
| `postmark_with_open_tracking_total`    | Gauge | Total number of emails with open tracking enabled   |
| `postmark_total_tracked_links_sent`    | Gauge | Total number of tracked links sent in emails        |
| `postmark_total_clicks`                | Gauge | Total number of clicks on tracked links in emails   |

## Configuration

This application is configured using the following environment variables:

| Name | Description | Required | Default |
|---   |---          |---       |---      |
| `POSTMARK_SERVER_TOKEN` | The Server API token for your Postmark server | yes | `-` |
| `POSTMARK_POLLING_INTERVAL` | The interval at which to poll the Postmark API for new data. See [API](#api-usage) Usage below for details on API requests made by the exporter. | no | `15s` |

## API Usage

The `postmark-exporter` makes two API calls to fetch Postmark data:

- [one fetch general statistics](./internal/postmark/postmark.go#L56)
- [one to fetch bounce statistics](./internal/postmark/postmark.go#L83)

By default the `POSTMARK_POLLING_INTERVAL` is `15s` to ensure metrics remain reasonably current. You can adjust thus value using any `time.Duration`-parsable value.