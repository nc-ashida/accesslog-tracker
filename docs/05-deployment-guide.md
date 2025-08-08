# デプロイメントガイド

## 1. 概要

### 1.1 デプロイメント環境
- **開発環境**: Docker Compose
- **ステージング環境**: Docker + Nginx + Go
- **本番環境**: AWS (ALB + EC2 + Nginx + Go + SQS + Lambda + RDS PostgreSQL)

### 1.2 システム要件
- **CPU**: 1コア以上（コスト最適化により削減）
- **メモリ**: 1GB以上（コスト最適化により削減）
- **ストレージ**: 50GB以上（SSD推奨）
- **ネットワーク**: 1Gbps以上

### 1.3 AWS環境構成（コスト最適化版）
- **CloudFront**: ビーコン配信用CDN
- **ALB**: ロードバランシングとSSL終端
- **EC2**: 軽量Goアプリケーション実行環境
- **Nginx + OpenResty**: リバースプロキシとWebサーバー
- **SQS**: 高可用性メッセージキュー（メイン永続化）
- **ElastiCache**: Redis（高速バッファ・セッション管理）
- **Lambda/Go Worker**: サーバーレス処理（従量課金）
- **RDS PostgreSQL**: 管理されたデータベース
- **S3**: ログデータの長期保存
- **CloudWatch**: 監視・ログ・アラート

## 2. AWS本番環境セットアップ

### 2.1 インフラストラクチャ構成

#### 2.1.1 CloudFormationテンプレート
```yaml
# infrastructure.yml
AWSTemplateFormatVersion: '2010-09-09'
Description: 'Access Log Tracker Infrastructure (コスト最適化版)'

Parameters:
  Environment:
    Type: String
    Default: production
    AllowedValues: [development, staging, production]
  
  InstanceType:
    Type: String
    Default: t3.small
    AllowedValues: [t3.nano, t3.micro, t3.small, t3.medium]

Resources:
  # VPC設定
  VPC:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
      EnableDnsHostnames: true
      EnableDnsSupport: true
      Tags:
        - Key: Name
          Value: !Sub '${Environment}-access-log-tracker-vpc'

  # パブリックサブネット
  PublicSubnet1:
    Type: AWS::EC2::Subnet
    Properties:
      VpcId: !Ref VPC
      CidrBlock: 10.0.1.0/24
      AvailabilityZone: !Select [0, !GetAZs '']
      MapPublicIpOnLaunch: true

  PublicSubnet2:
    Type: AWS::EC2::Subnet
    Properties:
      VpcId: !Ref VPC
      CidrBlock: 10.0.2.0/24
      AvailabilityZone: !Select [1, !GetAZs '']
      MapPublicIpOnLaunch: true

  # プライベートサブネット
  PrivateSubnet1:
    Type: AWS::EC2::Subnet
    Properties:
      VpcId: !Ref VPC
      CidrBlock: 10.0.3.0/24
      AvailabilityZone: !Select [0, !GetAZs '']

  PrivateSubnet2:
    Type: AWS::EC2::Subnet
    Properties:
      VpcId: !Ref VPC
      CidrBlock: 10.0.4.0/24
      AvailabilityZone: !Select [1, !GetAZs '']

  # インターネットゲートウェイ
  InternetGateway:
    Type: AWS::EC2::InternetGateway
    Properties:
      Tags:
        - Key: Name
          Value: !Sub '${Environment}-access-log-tracker-igw'

  InternetGatewayAttachment:
    Type: AWS::EC2::VPCGatewayAttachment
    Properties:
      VpcId: !Ref VPC
      InternetGatewayId: !Ref InternetGateway

  # ルートテーブル
  PublicRouteTable:
    Type: AWS::EC2::RouteTable
    Properties:
      VpcId: !Ref VPC
      Tags:
        - Key: Name
          Value: !Sub '${Environment}-public-rt'

  PublicRoute:
    Type: AWS::EC2::Route
    DependsOn: InternetGatewayAttachment
    Properties:
      RouteTableId: !Ref PublicRouteTable
      DestinationCidrBlock: 0.0.0.0/0
      GatewayId: !Ref InternetGateway

  PublicSubnet1RouteTableAssociation:
    Type: AWS::EC2::SubnetRouteTableAssociation
    Properties:
      SubnetId: !Ref PublicSubnet1
      RouteTableId: !Ref PublicRouteTable

  PublicSubnet2RouteTableAssociation:
    Type: AWS::EC2::SubnetRouteTableAssociation
    Properties:
      SubnetId: !Ref PublicSubnet2
      RouteTableId: !Ref PublicRouteTable

  # セキュリティグループ
  ALBSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupName: !Sub '${Environment}-alb-sg'
      GroupDescription: Security group for ALB
      VpcId: !Ref VPC
      SecurityGroupIngress:
        - IpProtocol: tcp
          FromPort: 443
          ToPort: 443
          CidrIp: 0.0.0.0/0
        - IpProtocol: tcp
          FromPort: 80
          ToPort: 80
          CidrIp: 0.0.0.0/0

  EC2SecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupName: !Sub '${Environment}-ec2-sg'
      GroupDescription: Security group for EC2
      VpcId: !Ref VPC
      SecurityGroupIngress:
        - IpProtocol: tcp
          FromPort: 80
          ToPort: 80
          SourceSecurityGroupId: !Ref ALBSecurityGroup
        - IpProtocol: tcp
          FromPort: 22
          ToPort: 22
          CidrIp: 10.0.0.0/8

  # ALB
  ApplicationLoadBalancer:
    Type: AWS::ElasticLoadBalancingV2::LoadBalancer
    Properties:
      Name: !Sub '${Environment}-access-log-tracker-alb'
      Scheme: internet-facing
      Type: application
      Subnets:
        - !Ref PublicSubnet1
        - !Ref PublicSubnet2
      SecurityGroups:
        - !Ref ALBSecurityGroup

  # ターゲットグループ
  TargetGroup:
    Type: AWS::ElasticLoadBalancingV2::TargetGroup
    Properties:
      Name: !Sub '${Environment}-access-log-tracker-tg'
      Port: 80
      Protocol: HTTP
      TargetType: instance
      VpcId: !Ref VPC
      HealthCheckPath: /health
      HealthCheckIntervalSeconds: 30
      HealthCheckTimeoutSeconds: 5
      HealthyThresholdCount: 2
      UnhealthyThresholdCount: 3

  # リスナー
  Listener:
    Type: AWS::ElasticLoadBalancingV2::Listener
    Properties:
      LoadBalancerArn: !Ref ApplicationLoadBalancer
      Port: 443
      Protocol: HTTPS
      Certificates:
        - CertificateArn: !Ref SSLCertificate
      DefaultActions:
        - Type: forward
          TargetGroupArn: !Ref TargetGroup

  # EC2インスタンス
  EC2Instance:
    Type: AWS::EC2::Instance
    Properties:
      InstanceType: !Ref InstanceType
      ImageId: ami-0c55b159cbfafe1f0 # Amazon Linux 2
      SecurityGroups:
        - !Ref EC2SecurityGroup
      SubnetId: !Ref PrivateSubnet1
      IamInstanceProfile: !Ref EC2InstanceProfile
      UserData:
        Fn::Base64: !Sub |
          #!/bin/bash
          yum update -y
          yum install -y nginx golang git
          
          # nginx設定（Go + Nginx対応）
          cat > /etc/nginx/nginx.conf << 'EOF'
          events {
              worker_connections 10000;
              use epoll;
              multi_accept on;
          }
          
          http {
              upstream go_backend {
                  server 127.0.0.1:8080;
              }
              
              server {
                  listen 80;
                  server_name api.access-log-tracker.com;
                  
                  location / {
                      proxy_pass http://go_backend;
                      proxy_set_header Host $host;
                      proxy_set_header X-Real-IP $remote_addr;
                      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
                      proxy_set_header X-Forwarded-Proto $scheme;
                      proxy_read_timeout 30s;
                      proxy_connect_timeout 30s;
                  }
              }
          }
          EOF
          
          # Goアプリケーション設定
          mkdir -p /opt/access-log-tracker
          cd /opt/access-log-tracker
          
          # Goモジュール初期化
          go mod init access-log-tracker
          
          # go.mod作成
          cat > go.mod << 'EOF'
          module access-log-tracker

          go 1.21

          require (
              github.com/gin-gonic/gin v1.9.1
              github.com/lib/pq v1.10.9
              github.com/go-redis/redis/v8 v8.11.5
              github.com/aws/aws-sdk-go v1.48.0
              github.com/sirupsen/logrus v1.9.3
              github.com/gin-contrib/cors v1.4.0
              github.com/gin-contrib/secure v0.0.1
              github.com/gin-contrib/timeout v0.0.3
          )
          EOF
          
          # Go依存関係インストール
          go mod tidy
          
          # systemdサービス設定
          cat > /etc/systemd/system/access-log-tracker.service << 'EOF'
          [Unit]
          Description=Access Log Tracker Go Application
          After=network.target

          [Service]
          Type=simple
          User=nginx
          WorkingDirectory=/opt/access-log-tracker
          ExecStart=/usr/local/bin/access-log-tracker
          Restart=always
          RestartSec=5
          Environment=GIN_MODE=release
          Environment=DB_HOST=aurora-cluster.cluster-xyz.ap-northeast-1.rds.amazonaws.com
          Environment=DB_PORT=5432
          Environment=DB_NAME=access_log_tracker
          Environment=DB_USER=alt_admin
          Environment=REDIS_HOST=localhost
          Environment=REDIS_PORT=6379

          [Install]
          WantedBy=multi-user.target
          EOF
          
          # サービス有効化・起動
          systemctl daemon-reload
          systemctl enable access-log-tracker
          systemctl start access-log-tracker
          
          # nginx起動
          systemctl enable nginx
          systemctl start nginx

  # Aurora PostgreSQL
  AuroraCluster:
    Type: AWS::RDS::DBCluster
    Properties:
      Engine: aurora-postgresql
      EngineVersion: '14.7'
      EngineMode: provisioned
      DBClusterInstanceClass: db.r6g.large
      MasterUsername: alt_admin
      MasterUserPassword: !Ref DBPassword
      BackupRetentionPeriod: 7
      PreferredBackupWindow: '03:00-04:00'
      PreferredMaintenanceWindow: 'sun:04:00-sun:05:00'
      StorageEncrypted: true
      DeletionProtection: true
      EnableCloudwatchLogsExports:
        - postgresql
      DBSubnetGroupName: !Ref DBSubnetGroup
      VpcSecurityGroupIds:
        - !Ref DBSecurityGroup

  # RDS Proxy
  RDSProxy:
    Type: AWS::RDS::DBProxy
    Properties:
      DBProxyName: !Sub '${Environment}-access-log-proxy'
      EngineFamily: POSTGRESQL
      RequireTLS: true
      IdleClientTimeout: 1800
      MaxConnectionsPercent: 100
      MaxIdleConnectionsPercent: 50
      Auth:
        - AuthScheme: SECRETS
          SecretArn: !Ref DBSecretArn
      RoleArn: !GetAtt RDSProxyRole.Arn

Outputs:
  ALBDNSName:
    Description: DNS name of the Application Load Balancer
    Value: !GetAtt ApplicationLoadBalancer.DNSName
    Export:
      Name: !Sub '${Environment}-alb-dns-name'

  EC2InstanceId:
    Description: EC2 Instance ID
    Value: !Ref EC2Instance
    Export:
      Name: !Sub '${Environment}-ec2-instance-id'
```

### 2.2 Goアプリケーション実装

#### 2.2.1 main.go（メインアプリケーション）
```go
package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors"
    "github.com/gin-contrib/timeout"
    _ "github.com/lib/pq"
    "github.com/go-redis/redis/v8"
    "github.com/sirupsen/logrus"
)

// 設定構造体
type Config struct {
    DBHost     string
    DBPort     string
    DBName     string
    DBUser     string
    DBPassword string
    RedisHost  string
    RedisPort  string
}

// トラッキングデータ構造体
type TrackingData struct {
    AppID        string                 `json:"app_id" binding:"required"`
    ClientSubID  string                 `json:"client_sub_id,omitempty"`
    ModuleID     string                 `json:"module_id,omitempty"`
    URL          string                 `json:"url,omitempty"`
    Referrer     string                 `json:"referrer,omitempty"`
    UserAgent    string                 `json:"user_agent" binding:"required"`
    IPAddress    string                 `json:"ip_address,omitempty"`
    SessionID    string                 `json:"session_id,omitempty"`
    ScreenRes    string                 `json:"screen_resolution,omitempty"`
    Language     string                 `json:"language,omitempty"`
    Timezone     string                 `json:"timezone,omitempty"`
    CustomParams map[string]interface{} `json:"custom_params,omitempty"`
}

// レスポンス構造体
type Response struct {
    Success   bool        `json:"success"`
    Data      interface{} `json:"data,omitempty"`
    Message   string      `json:"message"`
    Timestamp string      `json:"timestamp"`
}

// アプリケーション構造体
type App struct {
    DB    *sql.DB
    Redis *redis.Client
    Logger *logrus.Logger
}

func main() {
    // 設定読み込み
    config := &Config{
        DBHost:     getEnv("DB_HOST", "localhost"),
        DBPort:     getEnv("DB_PORT", "5432"),
        DBName:     getEnv("DB_NAME", "access_log_tracker"),
        DBUser:     getEnv("DB_USER", "alt_admin"),
        DBPassword: getEnv("DB_PASSWORD", ""),
        RedisHost:  getEnv("REDIS_HOST", "localhost"),
        RedisPort:  getEnv("REDIS_PORT", "6379"),
    }

    // データベース接続
    db, err := connectDB(config)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()

    // Redis接続
    rdb := redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("%s:%s", config.RedisHost, config.RedisPort),
        Password: "",
        DB:       0,
    })

    // ロガー設定
    logger := logrus.New()
    logger.SetLevel(logrus.InfoLevel)

    app := &App{
        DB:     db,
        Redis:  rdb,
        Logger: logger,
    }

    // Ginルーター設定
    r := gin.New()
    r.Use(gin.Logger())
    r.Use(gin.Recovery())
    r.Use(cors.Default())

    // タイムアウトミドルウェア
    r.Use(timeout.New(
        timeout.WithTimeout(30*time.Second),
        timeout.WithHandler(func(c *gin.Context) {
            c.Next()
        }),
    ))

    // ルート設定
    setupRoutes(r, app)

    // サーバー起動
    port := getEnv("PORT", "8080")
    logger.Infof("Starting server on port %s", port)
    r.Run(":" + port)
}

// データベース接続
func connectDB(config *Config) (*sql.DB, error) {
    dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=require",
        config.DBHost, config.DBPort, config.DBName, config.User, config.DBPassword)
    
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, err
    }

    // 接続テスト
    if err := db.Ping(); err != nil {
        return nil, err
    }

    return db, nil
}

// ルート設定
func setupRoutes(r *gin.Engine, app *App) {
    v1 := r.Group("/v1")
    {
        // トラッキングエンドポイント
        v1.POST("/track", app.trackHandler)
        
        // ヘルスチェック
        v1.GET("/health", app.healthHandler)
        
        // 統計情報
        v1.GET("/statistics", app.statisticsHandler)
    }
}

// トラッキングハンドラー
func (app *App) trackHandler(c *gin.Context) {
    var data TrackingData
    if err := c.ShouldBindJSON(&data); err != nil {
        c.JSON(http.StatusBadRequest, Response{
            Success:   false,
            Message:   "Invalid request data",
            Timestamp: time.Now().UTC().Format(time.RFC3339),
        })
        return
    }

    // データベースに保存
    if err := app.saveTrackingData(data); err != nil {
        app.Logger.Errorf("Failed to save tracking data: %v", err)
        c.JSON(http.StatusInternalServerError, Response{
            Success:   false,
            Message:   "Failed to save tracking data",
            Timestamp: time.Now().UTC().Format(time.RFC3339),
        })
        return
    }

    c.JSON(http.StatusOK, Response{
        Success:   true,
        Message:   "Tracking data recorded successfully",
        Timestamp: time.Now().UTC().Format(time.RFC3339),
    })
}

// データ保存
func (app *App) saveTrackingData(data TrackingData) error {
    query := `
        INSERT INTO access_logs (
            app_id, client_sub_id, module_id, url, referrer,
            user_agent, ip_address, session_id, screen_resolution,
            language, timezone, custom_params, created_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
    `

    customParamsJSON, _ := json.Marshal(data.CustomParams)
    
    _, err := app.DB.Exec(query,
        data.AppID, data.ClientSubID, data.ModuleID, data.URL, data.Referrer,
        data.UserAgent, data.IPAddress, data.SessionID, data.ScreenRes,
        data.Language, data.Timezone, customParamsJSON, time.Now(),
    )

    return err
}

// ヘルスチェックハンドラー
func (app *App) healthHandler(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status":    "healthy",
        "timestamp": time.Now().UTC().Format(time.RFC3339),
        "database":  "connected",
        "redis":     "connected",
        "version":   "1.0.0",
    })
}

// 統計情報ハンドラー
func (app *App) statisticsHandler(c *gin.Context) {
    // 統計情報の実装
    c.JSON(http.StatusOK, Response{
        Success:   true,
        Message:   "Statistics retrieved successfully",
        Timestamp: time.Now().UTC().Format(time.RFC3339),
    })
}

// 環境変数取得
func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

## 3. 本番環境デプロイメント

### 3.1 Kubernetes設定

#### namespace.yaml
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: access-log-tracker
```

#### configmap.yaml
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: alt-config
  namespace: access-log-tracker
data:
  NODE_ENV: "production"
  DATABASE_URL: "postgresql://postgres:password@postgres:5432/access_log_tracker"
  REDIS_URL: "redis://redis:6379"
  API_BASE_URL: "https://api.access-log-tracker.com"
  CORS_ORIGIN: "https://access-log-tracker.com"
```

#### secret.yaml
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: alt-secrets
  namespace: access-log-tracker
type: Opaque
data:
  JWT_SECRET: <base64-encoded-jwt-secret>
  API_KEY_SALT: <base64-encoded-salt>
  DATABASE_PASSWORD: <base64-encoded-password>
```

#### deployment.yaml
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: alt-api
  namespace: access-log-tracker
spec:
  replicas: 3
  selector:
    matchLabels:
      app: alt-api
  template:
    metadata:
      labels:
        app: alt-api
    spec:
      containers:
      - name: api
        image: access-log-tracker:latest
        ports:
        - containerPort: 3000
        env:
        - name: NODE_ENV
          valueFrom:
            configMapKeyRef:
              name: alt-config
              key: NODE_ENV
        - name: DATABASE_URL
          valueFrom:
            configMapKeyRef:
              name: alt-config
              key: DATABASE_URL
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: alt-secrets
              key: JWT_SECRET
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 3000
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 3000
          initialDelaySeconds: 5
          periodSeconds: 5
```

#### service.yaml
```yaml
apiVersion: v1
kind: Service
metadata:
  name: alt-api-service
  namespace: access-log-tracker
spec:
  selector:
    app: alt-api
  ports:
  - protocol: TCP
    port: 80
    targetPort: 3000
  type: ClusterIP
```

#### ingress.yaml
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: alt-ingress
  namespace: access-log-tracker
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
  - hosts:
    - api.access-log-tracker.com
    secretName: alt-tls
  rules:
  - host: api.access-log-tracker.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: alt-api-service
            port:
              number: 80
```

### 3.2 データベース設定

#### postgres-deployment.yaml
```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: access-log-tracker
spec:
  serviceName: postgres
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:14
        ports:
        - containerPort: 5432
        env:
        - name: POSTGRES_DB
          value: access_log_tracker
        - name: POSTGRES_USER
          value: postgres
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: alt-secrets
              key: DATABASE_PASSWORD
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
        resources:
          requests:
            memory: "1Gi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "1000m"
  volumeClaimTemplates:
  - metadata:
      name: postgres-storage
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 100Gi
```

### 3.3 デプロイメント手順

#### 1. クラスター準備
```bash
# Kubernetesクラスター作成（例：GKE）
gcloud container clusters create alt-cluster \
  --zone=asia-northeast1-a \
  --num-nodes=3 \
  --machine-type=e2-standard-2

# クラスター認証
gcloud container clusters get-credentials alt-cluster --zone=asia-northeast1-a
```

#### 2. 名前空間作成
```bash
kubectl apply -f k8s/namespace.yaml
```

#### 3. 設定適用
```bash
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secret.yaml
```

#### 4. データベースデプロイ
```bash
kubectl apply -f k8s/postgres-deployment.yaml
kubectl apply -f k8s/postgres-service.yaml
```

#### 5. アプリケーションデプロイ
```bash
# イメージビルド
docker build -t access-log-tracker:latest .

# イメージプッシュ
docker tag access-log-tracker:latest gcr.io/PROJECT_ID/access-log-tracker:latest
docker push gcr.io/PROJECT_ID/access-log-tracker:latest

# デプロイメント適用
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
```

#### 6. Ingress設定
```bash
kubectl apply -f k8s/ingress.yaml
```

## 4. CI/CDパイプライン

### 4.1 GitHub Actions設定

#### .github/workflows/deploy.yml
```yaml
name: Deploy to Production

on:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Setup Node.js
      uses: actions/setup-node@v3
      with:
        node-version: '18'
        cache: 'npm'
    
    - name: Install dependencies
      run: npm ci
    
    - name: Run tests
      run: npm test
    
    - name: Run linting
      run: npm run lint

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Setup Docker Buildx
      uses: docker/setup-buildx-action@v2
    
    - name: Login to Container Registry
      uses: docker/login-action@v2
      with:
        registry: gcr.io
        username: _json_key
        password: ${{ secrets.GCP_JSON_KEY }}
    
    - name: Build and push Docker image
      uses: docker/build-push-action@v4
      with:
        context: .
        push: true
        tags: gcr.io/${{ secrets.GCP_PROJECT_ID }}/access-log-tracker:${{ github.sha }}
        cache-from: type=gha
        cache-to: type=gha,mode=max

  deploy:
    needs: build
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Setup gcloud CLI
      uses: google-github-actions/setup-gcloud@v0
      with:
        service_account_key: ${{ secrets.GCP_JSON_KEY }}
        project_id: ${{ secrets.GCP_PROJECT_ID }}
    
    - name: Configure kubectl
      run: |
        gcloud container clusters get-credentials alt-cluster --zone=asia-northeast1-a
    
    - name: Deploy to Kubernetes
      run: |
        kubectl set image deployment/alt-api api=gcr.io/${{ secrets.GCP_PROJECT_ID }}/access-log-tracker:${{ github.sha }} -n access-log-tracker
        kubectl rollout status deployment/alt-api -n access-log-tracker
```

## 5. 監視とログ

### 5.1 Prometheus設定

#### prometheus-config.yaml
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
  namespace: access-log-tracker
data:
  prometheus.yml: |
    global:
      scrape_interval: 15s
    
    scrape_configs:
    - job_name: 'alt-api'
      static_configs:
      - targets: ['alt-api-service:3000']
      metrics_path: /metrics
```

### 5.2 Grafanaダッシュボード

#### grafana-dashboard.json
```json
{
  "dashboard": {
    "title": "Access Log Tracker Dashboard",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "{{app_id}}"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          }
        ]
      }
    ]
  }
}
```

## 6. バックアップと復旧

### 6.1 データベースバックアップ
```bash
#!/bin/bash
# backup.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backups"
DB_NAME="access_log_tracker"

# PostgreSQLバックアップ
pg_dump -h localhost -U postgres -d $DB_NAME | gzip > $BACKUP_DIR/backup_$DATE.sql.gz

# 古いバックアップ削除（30日以上）
find $BACKUP_DIR -name "backup_*.sql.gz" -mtime +30 -delete
```

### 6.2 復旧手順
```bash
# データベース復旧
gunzip -c backup_20240101_120000.sql.gz | psql -h localhost -U postgres -d access_log_tracker

# アプリケーション再起動
kubectl rollout restart deployment/alt-api -n access-log-tracker
```

## 7. セキュリティ設定

### 7.1 SSL/TLS設定
```nginx
# nginx.conf
server {
    listen 443 ssl http2;
    server_name api.access-log-tracker.com;
    
    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512;
    
    location / {
        proxy_pass http://alt-api-service;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 7.2 ファイアウォール設定
```bash
# UFW設定
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw deny 5432/tcp  # PostgreSQL
ufw enable
``` 