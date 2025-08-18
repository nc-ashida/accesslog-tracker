#!/bin/bash

# ビーコンエンドポイントを修正するスクリプト
# GETリクエストをPOSTリクエストに変更し、APIキーを追加

sed -i '' 's|beaconURL := fmt.Sprintf("%s/v1/tracking/track?app_id=%s&session_id=%s&url=/concurrent-test", baseURL, app.AppID, sessionID)|beaconURL := fmt.Sprintf("%s/v1/tracking/track", baseURL)\n\t\t\treq, err := http.NewRequest("POST", beaconURL, nil)\n\t\t\tvar resp *http.Response\n\t\t\tif err == nil {\n\t\t\t\tq := req.URL.Query()\n\t\t\t\tq.Add("app_id", app.AppID)\n\t\t\t\tq.Add("session_id", sessionID)\n\t\t\t\tq.Add("url", "/concurrent-test")\n\t\t\t\treq.URL.RawQuery = q.Encode()\n\t\t\t\treq.Header.Set("X-API-Key", app.APIKey)\n\t\t\t\tresp, err = http.DefaultClient.Do(req)\n\t\t\t}|g' tests/performance/beacon_performance_test.go

sed -i '' 's|beaconURL := fmt.Sprintf("%s/v1/tracking/track?app_id=%s&session_id=%s&url=/throughput-test", baseURL, app.AppID, sessionID)|beaconURL := fmt.Sprintf("%s/v1/tracking/track", baseURL)\n\t\t\treq, err := http.NewRequest("POST", beaconURL, nil)\n\t\t\tvar resp *http.Response\n\t\t\tif err == nil {\n\t\t\t\tq := req.URL.Query()\n\t\t\t\tq.Add("app_id", app.AppID)\n\t\t\t\tq.Add("session_id", sessionID)\n\t\t\t\tq.Add("url", "/throughput-test")\n\t\t\t\treq.URL.RawQuery = q.Encode()\n\t\t\t\treq.Header.Set("X-API-Key", app.APIKey)\n\t\t\t\tresp, err = http.DefaultClient.Do(req)\n\t\t\t}|g' tests/performance/beacon_performance_test.go

echo "ビーコンエンドポイントの修正が完了しました"
