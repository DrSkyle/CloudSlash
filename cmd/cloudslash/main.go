package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/DrSkyle/cloudslash/internal/aws"
	"github.com/DrSkyle/cloudslash/internal/graph"
	"github.com/DrSkyle/cloudslash/internal/heuristics"
	"github.com/DrSkyle/cloudslash/internal/license"
	"github.com/DrSkyle/cloudslash/internal/notifier"
	"github.com/DrSkyle/cloudslash/internal/pricing"
	"github.com/DrSkyle/cloudslash/internal/remediation"
	"github.com/DrSkyle/cloudslash/internal/report"
	"github.com/DrSkyle/cloudslash/internal/swarm"
	"github.com/DrSkyle/cloudslash/internal/tf"
	"github.com/DrSkyle/cloudslash/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	licenseKey := flag.String("license", "", "License Key")
	region := flag.String("region", "us-east-1", "AWS Region")
	tfStatePath := flag.String("tfstate", "terraform.tfstate", "Path to terraform.tfstate")
	mockMode := flag.Bool("mock", false, "Run in Mock Mode (Simulated Data)")
	allProfiles := flag.Bool("all-profiles", false, "Scan all available AWS profiles in ~/.aws/config")
	requiredTags := flag.String("required-tags", "", "Comma-separated list of required tags (e.g. Owner,CostCenter)")
	slackWebhook := flag.String("slack-webhook", "", "Slack Webhook URL for reporting")
	flag.Parse()

	// 1. License Check (Fail-Open / Trial Mode)
	isTrial := false
	if *licenseKey == "" {
		fmt.Println("No license key provided. Running in TRIAL MODE.")
		fmt.Println("Resource IDs will be redacted and no output files will be generated.")
		isTrial = true
	} else {
		if err := license.Check(*licenseKey); err != nil {
			fmt.Printf("License check failed: %v\n", err)
			fmt.Println("Falling back to TRIAL MODE.")
			isTrial = true
		}
	}

	// 2. Initialize Components
	ctx := context.Background()
	var g *graph.Graph
	var engine *swarm.Engine
	var cwClient *aws.CloudWatchClient
	var iamClient *aws.IAMClient

	g = graph.NewGraph()
	engine = swarm.NewEngine()
	engine.Start(ctx)

	if *mockMode {
		fmt.Println("Running in MOCK MODE. Simulating AWS environment...")
		mockScanner := aws.NewMockScanner(g)
		mockScanner.Scan(ctx)

		// Synchronous Heuristics for Demo
		heuristicEngine := heuristics.NewEngine()
		heuristicEngine.Register(&heuristics.ZombieEBSHeuristic{})
		heuristicEngine.Register(&heuristics.S3MultipartHeuristic{})

		if err := heuristicEngine.Run(ctx, g); err != nil {
			fmt.Printf("Heuristic run failed: %v\n", err)
		}

		os.Mkdir("cloudslash-out", 0755)
		if err := report.GenerateHTML(g, "cloudslash-out/dashboard.html"); err != nil {
			fmt.Printf("Failed to generate mock dashboard: %v\n", err)
		}
	} else {
		// Real AWS Mode
		profiles := []string{""}
		if *allProfiles {
			var err error
			profiles, err = aws.ListProfiles()
			if err != nil {
				fmt.Printf("Failed to list profiles: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Deep Scanning enabled. Found %d profiles.\n", len(profiles))
		}

		var pricingClient *pricing.Client
		if !isTrial {
			var err error
			pricingClient, err = pricing.NewClient(ctx)
			if err != nil {
				fmt.Printf("Warning: Failed to initialize Pricing Client: %v\n", err)
			}
		}

		var scanWg sync.WaitGroup

		for _, profile := range profiles {
			if *allProfiles {
				fmt.Printf(">>> Scanning Profile: %s\n", profile)
			}

			client, err := runScanForProfile(ctx, *region, profile, g, engine, &scanWg)
			if err != nil {
				fmt.Printf("Scan failed for profile %s: %v\n", profile, err)
				continue
			}

			if client != nil {
				cwClient = aws.NewCloudWatchClient(client.Config)
				iamClient = aws.NewIAMClient(client.Config)
			}
		}

		// Background Analysis & Reporting
		go func() {
			scanWg.Wait() // Wait for ingestion

			// Shadow State Reconciliation
			if _, err := os.Stat(*tfStatePath); err == nil {
				state, err := tf.ParseStateFile(*tfStatePath)
				if err == nil {
					detector := tf.NewDriftDetector(g, state)
					detector.ScanForDrift()
				}
			}

			// Run Genius Heuristic Engine
			hEngine := heuristics.NewEngine()

			// Register Capability-Based Heuristics
			hEngine.Register(&heuristics.ElasticIPHeuristic{})
			hEngine.Register(&heuristics.S3MultipartHeuristic{})

			if cwClient != nil {
				hEngine.Register(&heuristics.NATGatewayHeuristic{CW: cwClient})
				hEngine.Register(&heuristics.RDSHeuristic{CW: cwClient})
				hEngine.Register(&heuristics.ELBHeuristic{CW: cwClient})
				if pricingClient != nil {
					hEngine.Register(&heuristics.UnderutilizedInstanceHeuristic{CW: cwClient, Pricing: pricingClient})
				}
			}

			if pricingClient != nil {
				hEngine.Register(&heuristics.ZombieEBSHeuristic{Pricing: pricingClient})
			} else {
				hEngine.Register(&heuristics.ZombieEBSHeuristic{})
			}

			if *requiredTags != "" {
				hEngine.Register(&heuristics.TagComplianceHeuristic{RequiredTags: strings.Split(*requiredTags, ",")})
			}

			if iamClient != nil {
				hEngine.Register(&heuristics.IAMHeuristic{IAM: iamClient})
			}

			// Execute Forensics
			if err := hEngine.Run(ctx, g); err != nil {
				fmt.Printf("Deep Analysis failed: %v\n", err)
			}

			// Generate Output
			if !isTrial {
				os.Mkdir("cloudslash-out", 0755)
				gen := tf.NewGenerator(g)
				gen.GenerateWasteTF("cloudslash-out/waste.tf")
				gen.GenerateImportScript("cloudslash-out/import.sh")
				gen.GenerateDestroyPlan("cloudslash-out/destroy_plan.out")

				remGen := remediation.NewGenerator(g)
				remGen.GenerateSafeDeleteScript("cloudslash-out/safe_cleanup.sh")
				os.Chmod("cloudslash-out/safe_cleanup.sh", 0755)

				if err := report.GenerateHTML(g, "cloudslash-out/dashboard.html"); err != nil {
					fmt.Printf("Failed to generate dashboard: %v\n", err)
				}

				if *slackWebhook != "" {
					if err := notifier.SendSlackReport(*slackWebhook, g); err != nil {
						fmt.Printf("Failed to send Slack report: %v\n", err)
					}
				}
			}
		}()
	}

	// 3. Start TUI
	model := ui.NewModel(engine, g, isTrial)
	p := tea.NewProgram(model)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func runScanForProfile(ctx context.Context, region, profile string, g *graph.Graph, engine *swarm.Engine, scanWg *sync.WaitGroup) (*aws.Client, error) {
	awsClient, err := aws.NewClient(ctx, region, profile)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS client: %v", err)
	}

	// Verify Identity
	identity, err := awsClient.VerifyIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to verify identity: %v", err)
	}
	fmt.Printf(" [Profile: %s] Connected to AWS Account: %s\n", profile, identity)

	// Scanners
	ec2Scanner := aws.NewEC2Scanner(awsClient.Config, g)
	s3Scanner := aws.NewS3Scanner(awsClient.Config, g)
	rdsScanner := aws.NewRDSScanner(awsClient.Config, g)
	elbScanner := aws.NewELBScanner(awsClient.Config, g)

	submitTask := func(task func(ctx context.Context) error) {
		scanWg.Add(1)
		engine.Submit(func(ctx context.Context) error {
			defer scanWg.Done()
			return task(ctx)
		})
	}

	submitTask(func(ctx context.Context) error { return ec2Scanner.ScanInstances(ctx) })
	submitTask(func(ctx context.Context) error { return ec2Scanner.ScanVolumes(ctx) })
	submitTask(func(ctx context.Context) error { return ec2Scanner.ScanNatGateways(ctx) })
	submitTask(func(ctx context.Context) error { return ec2Scanner.ScanAddresses(ctx) })
	submitTask(func(ctx context.Context) error { return s3Scanner.ScanBuckets(ctx) })
	submitTask(func(ctx context.Context) error { return rdsScanner.ScanInstances(ctx) })
	submitTask(func(ctx context.Context) error { return elbScanner.ScanLoadBalancers(ctx) })

	return awsClient, nil
}
