package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"accesslog-tracker/internal/beacon/generator"
	"accesslog-tracker/internal/utils/logger"
	"encoding/json"
	"time"
	"github.com/sirupsen/logrus"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
	GoVersion = "unknown"
)

func main() {
	// コマンドライン引数の定義
	var (
		appID     = flag.Int("app-id", 1, "Application ID")
		sessionID = flag.String("session-id", "", "Session ID (auto-generated if empty)")
		url       = flag.String("url", "/", "URL to track")
		referrer  = flag.String("referrer", "", "Referrer URL")
		userAgent = flag.String("user-agent", "BeaconGenerator/1.0", "User Agent")
		ipAddress = flag.String("ip", "127.0.0.1", "IP Address")
		count     = flag.Int("count", 1, "Number of beacons to generate")
		interval  = flag.Duration("interval", 0, "Interval between beacons")
		output    = flag.String("output", "console", "Output format: console, json, csv")
		help      = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// ロガーの初期化
	logger := logger.NewLogger()
	logger.WithFields(logrus.Fields{
		"version":    Version,
		"buildTime":  BuildTime,
		"goVersion":  GoVersion,
	}).Info("Starting Beacon Generator")

	// ビーコン生成器の初期化
	generator := generator.NewBeaconGenerator()

	// ビーコンの生成と出力
	for i := 0; i < *count; i++ {
		beacon := generator.GenerateBeacon(*appID, *sessionID, *url, *referrer, *userAgent, *ipAddress)
		
		switch *output {
		case "console":
			fmt.Printf("Generated Beacon %d:\n", i+1)
			fmt.Printf("  App ID: %d\n", beacon.AppID)
			fmt.Printf("  Session ID: %s\n", beacon.SessionID)
			fmt.Printf("  URL: %s\n", beacon.URL)
			fmt.Printf("  Referrer: %s\n", beacon.Referrer)
			fmt.Printf("  User Agent: %s\n", beacon.UserAgent)
			fmt.Printf("  IP Address: %s\n", beacon.IPAddress)
			fmt.Printf("  Timestamp: %s\n", beacon.Timestamp)
			fmt.Println()
		case "json":
			jsonData, err := json.Marshal(beacon)
			if err != nil {
				logger.WithError(err).Error("Failed to marshal beacon to JSON")
				continue
			}
			fmt.Println(string(jsonData))
		case "csv":
			if i == 0 {
				// CSVヘッダーを出力
				fmt.Println("app_id,session_id,url,referrer,user_agent,ip_address,timestamp")
			}
			fmt.Printf("%d,%s,%s,%s,%s,%s,%s\n",
				beacon.AppID,
				beacon.SessionID,
				beacon.URL,
				beacon.Referrer,
				beacon.UserAgent,
				beacon.IPAddress,
				beacon.Timestamp,
			)
		default:
			logger.WithField("output", *output).Error("Invalid output format")
			os.Exit(1)
		}

		// 間隔を待機（最後のビーコン以外）
		if *interval > 0 && i < *count-1 {
			time.Sleep(*interval)
		}
	}

	logger.Info("Beacon generation completed")
}

func showHelp() {
	fmt.Println("Beacon Generator - Access Log Tracker")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  beacon-generator [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -app-id int")
	fmt.Println("        Application ID (default 1)")
	fmt.Println("  -session-id string")
	fmt.Println("        Session ID (auto-generated if empty)")
	fmt.Println("  -url string")
	fmt.Println("        URL to track (default \"/\")")
	fmt.Println("  -referrer string")
	fmt.Println("        Referrer URL")
	fmt.Println("  -user-agent string")
	fmt.Println("        User Agent (default \"BeaconGenerator/1.0\")")
	fmt.Println("  -ip string")
	fmt.Println("        IP Address (default \"127.0.0.1\")")
	fmt.Println("  -count int")
	fmt.Println("        Number of beacons to generate (default 1)")
	fmt.Println("  -interval duration")
	fmt.Println("        Interval between beacons (e.g., 1s, 100ms)")
	fmt.Println("  -output string")
	fmt.Println("        Output format: console, json, csv (default \"console\")")
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  beacon-generator -app-id 1 -url /home -count 5")
	fmt.Println("  beacon-generator -app-id 2 -session-id test123 -interval 1s -count 10")
	fmt.Println("  beacon-generator -app-id 1 -output json -count 3")
	fmt.Println("  beacon-generator -app-id 1 -output csv -count 100 > beacons.csv")
}
