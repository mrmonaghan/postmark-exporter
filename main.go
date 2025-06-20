package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mrmonaghan/postmark-exporter/internal/postmark"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	bounced = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "postmark_bounced_total",
		Help: "Total number of bounced emails",
	})
	hardBounced = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "postmark_hard_bounced_total",
		Help: "Total number of hard bounced emails",
	})
	softBounced = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "postmark_soft_bounced_total",
		Help: "Total number of soft bounced emails",
	})
	transientBounced = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "postmark_transient_bounced_total",
		Help: "Total number of transient bounced emails",
	})
	sent = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "postmark_sent_total",
		Help: "Total number of sent emails",
	})
	smtpApiErrors = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "postmark_smtp_api_errors_total",
		Help: "Total number of SMTP API errors",
	})
	bounceRate = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "postmark_bounce_rate",
		Help: "Bounce rate of sent emails",
	})
	spamComplaints = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "postmark_spam_complaints_total",
		Help: "Total number of spam complaints",
	})
	spamComplaintsRate = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "postmark_spam_complaints_rate",
		Help: "Rate of spam complaints",
	})
	opens = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "postmark_opens_total",
		Help: "Total number of email opens",
	})
	uniqueOpens = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "postmark_unique_opens_total",
		Help: "Total number of unique email opens",
	})
	tracked = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "postmark_tracked_total",
		Help: "Total number of tracked emails",
	})
	withLinkTracking = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "postmark_with_link_tracking_total",
		Help: "Total number of emails with link tracking enabled",
	})
	withOpenTracking = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "postmark_with_open_tracking_total",
		Help: "Total number of emails with open tracking enabled",
	})
	totalTrackedLinksSent = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "postmark_total_tracked_links_sent",
		Help: "Total number of tracked links sent in emails",
	})
	totalClicks = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "postmark_total_clicks",
		Help: "Total number of clicks on tracked links in emails",
	})
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("initializing...")

	// init postmark client
	token := getEnv("POSTMARK_SERVER_TOKEN", "")
	if token == "" {
		logger.Error("POSTMARK_SERVER_TOKEN environment variable is not set")
		os.Exit(1)
	}
	pm := postmark.New(token)
	logger.Debug("postmark client initialized")

	// setup polling resources
	duration, err := time.ParseDuration(getEnv("POSTMARK_POLLING_INTERVAL", "15s"))
	if err != nil {
		logger.Error("failed to parse POSTMARK_POLLING_INTERVAL", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errs := make(chan error, 1)

	// Start the postmark poll routine
	go func() {
		logger.Debug("starting postmark poll routine", "interval", duration)
		tick := time.NewTicker(duration)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				logger.Info("stopping poll routine")
				return
			case <-tick.C:
				s, err := pm.GetOutboundStats()
				if err != nil {
					logger.Error("failed to get outbound stats", "error", err)
					select {
					case errs <- err:
					default:
					}
					return
				}

				b, err := pm.GetBounceStats()
				if err != nil {
					logger.Error("failed to get bounce stats", "error", err)
					select {
					case errs <- err:
					default:
					}
					return
				}
				updateMetrics(s, b)
				logger.Info("updated metrics")
			}
		}
	}()

	// Start metrics server
	server := &http.Server{Addr: ":8080"}
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		logger.Info("starting metrics server", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("failed to start HTTP server", "error", err)
			select {
			case errs <- err:
			default:
			}
		}
	}()

	// Handle graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	select {
	case err := <-errs:
		logger.Error("child process encountered an error", "error", err)
		cancel()
		// create a new context with timeout for graceful shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		server.Shutdown(shutdownCtx)
		os.Exit(1)
	case <-sigs:
		logger.Info("received shutdown signal, exiting")
		cancel()
		// create a new context with timeout for graceful shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		server.Shutdown(shutdownCtx)
		os.Exit(0)
	}
}

func updateMetrics(s postmark.OutboundStats, b postmark.BounceStats) {
	bounced.Set(float64(s.Bounced))
	sent.Set(float64(s.Sent))
	smtpApiErrors.Set(float64(s.SMTPApiErrors))
	bounceRate.Set(s.BounceRate)
	spamComplaints.Set(float64(s.SpamComplaints))
	spamComplaintsRate.Set(s.SpamComplaintsRate)
	opens.Set(float64(s.Opens))
	uniqueOpens.Set(float64(s.UniqueOpens))
	tracked.Set(float64(s.Tracked))
	withLinkTracking.Set(float64(s.WithLinkTracking))
	withOpenTracking.Set(float64(s.WithOpenTracking))
	totalTrackedLinksSent.Set(float64(s.TotalTrackedLinksSent))
	totalClicks.Set(float64(s.TotalClicks))
	hardBounced.Set(float64(b.HardBounce))
	softBounced.Set(float64(b.SoftBounce))
	transientBounced.Set(float64(b.Transient))
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
