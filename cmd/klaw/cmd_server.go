package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/kudig-io/klaw/internal/api"
	"github.com/kudig-io/klaw/internal/config"
	"github.com/kudig-io/klaw/internal/events"
	"github.com/kudig-io/klaw/internal/kubernetes"
	"github.com/kudig-io/klaw/internal/messaging"
	"github.com/kudig-io/klaw/internal/messaging/dingtalk"
	"github.com/kudig-io/klaw/internal/messaging/feishu"
	"github.com/kudig-io/klaw/internal/monitoring"
	"github.com/kudig-io/klaw/internal/openclaw"
	"github.com/kudig-io/klaw/internal/ops"
)

var serverPort int

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "启动 Web API + ChatOps 服务",
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runServer(cmd.Context())
	},
}

func init() {
	serverCmd.Flags().IntVar(&serverPort, "port", 0, "Override server port")
	rootCmd.AddCommand(serverCmd)
}

func runServer(_ context.Context) error {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	if serverPort > 0 {
		cfg.Server.Port = serverPort
	}

	k8sManager, err := kubernetes.NewManager(cfg.Kubernetes)
	if err != nil {
		log.Fatalf("Failed to initialize Kubernetes manager: %v", err)
	}

	monitoringService := monitoring.NewService(k8sManager)
	opsHandler := ops.NewHandler(k8sManager, monitoringService)
	commandRouter := ops.NewCommandRouter(opsHandler)

	commRegistry := messaging.NewCommunicatorRegistry()
	commManager := messaging.NewManager(commRegistry)
	commManager.RegisterGlobalHandler(commandRouter.HandleFunc())

	if cfg.Messaging.DingTalk.Enabled {
		dingtalkPlugin := dingtalk.NewPlugin(dingtalk.Config{
			Enabled:     cfg.Messaging.DingTalk.Enabled,
			AppKey:      cfg.Messaging.DingTalk.AppKey,
			AppSecret:   cfg.Messaging.DingTalk.AppSecret,
			Webhook:     cfg.Messaging.DingTalk.Webhook,
			Secret:      cfg.Messaging.DingTalk.Secret,
			WebhookPort: cfg.Messaging.DingTalk.WebhookPort,
		})
		commManager.RegisterCommunicator("dingtalk", dingtalkPlugin)
		dingtalkClient, _ := dingtalk.NewClient(cfg.Messaging.DingTalk)
		monitoringService.SetDingTalkClient(dingtalkClient)
		opsHandler.SetDingTalkClient(dingtalkClient)
		fmt.Println("✓ DingTalk plugin registered (bidirectional)")
	}

	if cfg.Messaging.Feishu.Enabled {
		feishuClient, err := feishu.NewClient(cfg.Messaging.Feishu)
		if err != nil {
			log.Fatalf("Failed to initialize Feishu client: %v", err)
		}
		go feishuClient.Start()
		monitoringService.SetFeishuClient(feishuClient)
		opsHandler.SetFeishuClient(feishuClient)
		fmt.Println("✓ Feishu client started")
	}

	if err := commManager.StartAll(); err != nil {
		log.Fatalf("Failed to start communication platforms: %v", err)
	}

	var eventManager *events.Manager
	var eventNotifier *events.Notifier

	if cfg.Events.Enabled {
		eventManager = events.NewManager()
		eventNotifier = events.NewNotifier(commManager, eventManager)
		for _, cluster := range cfg.Kubernetes.Clusters {
			source, err := events.NewKubernetesSource(cluster.Name, k8sManager)
			if err != nil {
				fmt.Printf("⚠ Failed to create event source for cluster %s: %v\n", cluster.Name, err)
				continue
			}
			filter := buildEventFilter(&cfg.Events)
			source.SetFilter(filter)
			eventManager.Register(source)
			if len(cfg.Events.Channels) > 0 {
				eventNotifier.SubscribeSource(source.Name(), cfg.Events.Channels...)
			} else {
				eventNotifier.SubscribeSource(source.Name())
			}
			fmt.Printf("✓ Event source registered for cluster: %s\n", cluster.Name)
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		if err := eventManager.StartAll(ctx); err != nil {
			log.Fatalf("Failed to start event sources: %v", err)
		}
		fmt.Println("✓ Event monitoring started (Watch mode)")
	} else {
		go monitoringService.Start()
		fmt.Println("✓ Monitoring service started (Polling mode)")
	}

	if cfg.OpenClaw.Enabled {
		openclawManager, err := openclaw.NewManager(cfg.OpenClaw, k8sManager)
		if err != nil {
			log.Fatalf("Failed to initialize OpenClaw manager: %v", err)
		}
		go openclawManager.Start()
		fmt.Println("✓ OpenClaw manager started")
	}

	apiServer, err := api.NewServer(k8sManager, monitoringService, cfg.Server, cfg.SOS)
	if err != nil {
		log.Fatalf("Failed to create API server: %v", err)
	}
	go func() {
		if err := apiServer.Start(cfg.Server.Port); err != nil {
			log.Fatalf("Failed to start API server: %v", err)
		}
	}()
	fmt.Printf("✓ Web UI server started on port %d\n", cfg.Server.Port)

	fmt.Println("\n🦞 Klaw started successfully. Press Ctrl+C to exit.")
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	fmt.Println("\nShutting down...")
	if eventManager != nil {
		eventManager.StopAll()
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("API server shutdown error: %v", err)
	}
	return nil
}

func buildEventFilter(cfg *config.EventConfig) *events.FilterConfig {
	filter := &events.FilterConfig{
		Namespaces:     cfg.Namespaces,
		ExcludeReasons: cfg.ExcludeReasons,
	}
	for _, rt := range cfg.WatchTypes {
		filter.ResourceTypes = append(filter.ResourceTypes, events.ResourceType(rt))
	}
	for _, et := range cfg.EventTypes {
		filter.EventTypes = append(filter.EventTypes, events.EventType(et))
	}
	if cfg.MinSeverity != "" {
		filter.MinSeverity = events.Severity(cfg.MinSeverity)
	}
	return filter
}
