# API仕様書

## 1. 概要

### 1.1 ベースURL
```
https://api.access-log-tracker.com/v1
```

### 1.2 インフラ構成（簡素化版）
- **ALB**: ロードバランシングとSSL終端
- **Nginx + OpenResty**: リバースプロキシとWebサーバー
- **Go + Gin**: 軽量APIサーバー（直接書き込み処理）
- **RDS PostgreSQL**: 管理されたデータベース（直接書き込み）

### 1.3 認証方式
- API Key認証（ヘッダー: `X-API-Key`）
- レート制限: 5000 req/min per API Key（簡素化による最適化）

### 1.4 レスポンス形式
```json
{
  "success": true,
  "data": {},
  "message": "Success",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 1.5 パフォーマンス最適化（簡素化版）
- **直接書き込み**: Go APIからPostgreSQLへの直接書き込み
- **コネクションプール**: PostgreSQL接続の効率化
- **軽量処理**: シンプルな構成による高速処理
- **安定性**: シンプルな構成による高安定性
- **メモリ効率**: 低メモリ使用量による高スケーラビリティ

## 2. エンドポイント一覧

### 2.1 トラッキングデータ送信

#### POST /track
アクセスログデータを送信するエンドポイント

**リクエストヘッダー**
```
Content-Type: application/json
X-API-Key: {api_key}
```

**リクエストボディ**
```json
{
  "app_id": "string (required)",
  "client_sub_id": "string (optional)",
  "module_id": "string (optional)",
  "url": "string (optional)",
  "referrer": "string (optional)",
  "user_agent": "string (required)",
  "ip_address": "string (optional)",
  "session_id": "string (optional)",
  "screen_resolution": "string (optional)",
  "language": "string (optional)",
  "timezone": "string (optional)",
  "custom_params": {
    "page_type": "string (optional)",
    "product_id": "string (optional)",
    "product_name": "string (optional)",
    "product_category": "string (optional)",
    "product_price": "number (optional)",
    "product_brand": "string (optional)",
    "product_availability": "string (optional)",
    "product_rating": "number (optional)",
    "product_review_count": "number (optional)",
    "cart_total": "number (optional)",
    "cart_item_count": "number (optional)",
    "user_segment": "string (optional)",
    "article_id": "string (optional)",
    "article_title": "string (optional)",
    "article_category": "string (optional)",
    "article_author": "string (optional)",
    "article_publish_date": "string (optional)",
    "article_read_time": "number (optional)",
    "article_tags": ["string"] (optional),
    "article_word_count": "number (optional)",
    "article_comment_count": "number (optional)",
    "search_query": "string (optional)",
    "search_results_count": "number (optional)",
    "search_filters": "string (optional)",
    "action_type": "string (optional)",
    "click_element": "string (optional)",
    "form_id": "string (optional)",
    "form_data": "object (optional)"
  }
}
```

**レスポンス**
```json
{
  "success": true,
  "data": {
    "tracking_id": "uuid",
    "timestamp": "2024-01-01T00:00:00Z"
  },
  "message": "Tracking data recorded successfully"
}
```

**エラーレスポンス**
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid app_id",
    "details": ["app_id is required"]
  }
}
```

### 2.2 ヘルスチェック

#### GET /health
システムの健全性を確認するエンドポイント

**レスポンス**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z",
  "database": "connected",
  "redis": "connected",
  "version": "1.0.0"
}
```

### 2.2 アプリケーション管理

#### GET /applications
アプリケーション一覧を取得

**リクエストヘッダー**
```
X-API-Key: {api_key}
```

**クエリパラメータ**
- `page`: ページ番号（デフォルト: 1）
- `limit`: 取得件数（デフォルト: 20, 最大: 100）
- `status`: ステータスフィルター（active, inactive）

**レスポンス**
```json
{
  "success": true,
  "data": {
    "applications": [
      {
        "id": "uuid",
        "app_id": "string",
        "name": "string",
        "description": "string",
        "status": "active",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 100,
      "total_pages": 5
    }
  }
}
```

#### POST /applications
新しいアプリケーションを登録

**リクエストボディ**
```json
{
  "name": "string (required)",
  "description": "string (optional)",
  "domain": "string (optional)"
}
```

#### PUT /applications/{id}
アプリケーション情報を更新

#### DELETE /applications/{id}
アプリケーションを削除（論理削除）

### 2.3 カスタムパラメータ管理

#### GET /applications/{id}/custom-parameters
アプリケーションのカスタムパラメータ定義一覧を取得

**リクエストヘッダー**
```
X-API-Key: {api_key}
```

**レスポンス**
```json
{
  "success": true,
  "data": {
    "custom_parameters": [
      {
        "id": "uuid",
        "parameter_key": "page_type",
        "parameter_name": "ページタイプ",
        "parameter_type": "string",
        "description": "ページの種類（product_detail, cart, checkout等）",
        "is_required": false,
        "default_value": null,
        "validation_rules": {
          "allowed_values": ["product_detail", "cart", "checkout", "category", "search"]
        },
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ]
  }
}
```

#### POST /applications/{id}/custom-parameters
新しいカスタムパラメータ定義を追加

**リクエストボディ**
```json
{
  "parameter_key": "string (required)",
  "parameter_name": "string (required)",
  "parameter_type": "string (required) - string, number, boolean, array, object",
  "description": "string (optional)",
  "is_required": "boolean (optional, default: false)",
  "default_value": "string (optional)",
  "validation_rules": {
    "allowed_values": ["value1", "value2"],
    "min_value": 0,
    "max_value": 100,
    "pattern": "^[a-zA-Z0-9_-]+$"
  }
}
```

#### PUT /applications/{id}/custom-parameters/{parameter_id}
カスタムパラメータ定義を更新

#### DELETE /applications/{id}/custom-parameters/{parameter_id}
カスタムパラメータ定義を削除

### 2.4 統計情報

#### GET /statistics
統計情報を取得

**クエリパラメータ**
- `app_id`: アプリケーションID
- `start_date`: 開始日（YYYY-MM-DD）
- `end_date`: 終了日（YYYY-MM-DD）
- `group_by`: グループ化（hour, day, month）
- `custom_params`: カスタムパラメータフィルター（JSON形式）

**カスタムパラメータフィルター例**
```
custom_params={"page_type": "product_detail", "product_category": "Electronics"}
```

**レスポンス**
```json
{
  "success": true,
  "data": {
    "total_requests": 1000000,
    "unique_visitors": 50000,
    "unique_sessions": 75000,
    "requests_by_date": [
      {
        "date": "2024-01-01",
        "requests": 1000,
        "unique_visitors": 500
      }
    ],
    "custom_param_breakdown": {
      "page_type": {
        "product_detail": 400000,
        "cart": 200000,
        "checkout": 100000,
        "category": 150000,
        "search": 150000
      },
      "product_category": {
        "Electronics": 250000,
        "Clothing": 200000,
        "Books": 100000,
        "Home": 150000,
        "Sports": 100000
      },
      "user_segment": {
        "premium": 300000,
        "regular": 500000,
        "guest": 200000
      }
    }
  }
}
```

#### GET /statistics/custom-parameters
カスタムパラメータ別の詳細統計を取得

**クエリパラメータ**
- `app_id`: アプリケーションID
- `parameter_key`: パラメータキー
- `start_date`: 開始日（YYYY-MM-DD）
- `end_date`: 終了日（YYYY-MM-DD）
- `limit`: 取得件数（デフォルト: 10）

**レスポンス**
```json
{
  "success": true,
  "data": {
    "parameter_key": "product_id",
    "parameter_name": "商品ID",
    "breakdown": [
      {
        "value": "PROD_12345",
        "count": 15000,
        "unique_visitors": 8000,
        "unique_sessions": 12000
      },
      {
        "value": "PROD_67890",
        "count": 12000,
        "unique_visitors": 6500,
        "unique_sessions": 9500
      }
    ]
  }
}
```

### 2.5 データエクスポート

#### GET /export
トラッキングデータをエクスポート

**クエリパラメータ**
- `app_id`: アプリケーションID
- `start_date`: 開始日（YYYY-MM-DD）
- `end_date`: 終了日（YYYY-MM-DD）
- `format`: 出力形式（json, csv, excel）
- `custom_params`: カスタムパラメータフィルター（JSON形式）
- `include_custom_params`: カスタムパラメータを含めるか（boolean, デフォルト: true）

**レスポンス**
```json
{
  "success": true,
  "data": {
    "export_id": "uuid",
    "download_url": "https://api.access-log-tracker.com/v1/exports/{export_id}/download",
    "expires_at": "2024-01-02T00:00:00Z"
  }
}
```

#### GET /exports/{export_id}/download
エクスポートファイルをダウンロード

## 3. エラーコード

### 3.1 HTTPステータスコード
- `200`: 成功
- `201`: 作成成功
- `400`: バリデーションエラー
- `401`: 認証エラー
- `403`: 権限エラー
- `404`: リソースが見つからない
- `429`: レート制限超過
- `500`: サーバーエラー

### 3.2 エラーコード詳細
- `VALIDATION_ERROR`: 入力値検証エラー
- `AUTHENTICATION_ERROR`: 認証エラー
- `RATE_LIMIT_EXCEEDED`: レート制限超過
- `APPLICATION_NOT_FOUND`: アプリケーションが見つからない
- `INVALID_API_KEY`: 無効なAPIキー
- `CUSTOM_PARAMETER_ERROR`: カスタムパラメータエラー
- `EXPORT_ERROR`: エクスポートエラー

## 4. レート制限

### 4.1 制限値
- **トラッキングAPI**: 1000 req/min per API Key
- **管理API**: 100 req/min per API Key
- **統計API**: 60 req/min per API Key
- **エクスポートAPI**: 10 req/min per API Key

### 4.2 レスポンスヘッダー
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
```

## 5. バッチ処理

### 5.1 バッチ送信エンドポイント

#### POST /track/batch
複数のトラッキングデータを一括送信

**リクエストボディ**
```json
{
  "events": [
    {
      "app_id": "string",
      "client_sub_id": "string",
      "module_id": "string",
      "url": "string",
      "referrer": "string",
      "user_agent": "string",
      "ip_address": "string",
      "session_id": "string",
      "custom_params": {
        "page_type": "product_detail",
        "product_id": "PROD_12345",
        "product_name": "Wireless Headphones",
        "product_category": "Electronics",
        "product_price": 299.99
      },
      "timestamp": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**制限**
- 最大100件まで一括送信可能
- タイムスタンプは過去24時間以内

## 6. Webhook

### 6.1 Webhook設定

#### POST /webhooks
Webhookエンドポイントを設定

**リクエストボディ**
```json
{
  "url": "string (required)",
  "events": ["tracking.created", "application.updated", "custom_parameter.updated"],
  "secret": "string (optional)"
}
```

### 6.2 Webhookペイロード例
```json
{
  "event": "tracking.created",
  "timestamp": "2024-01-01T00:00:00Z",
  "data": {
    "tracking_id": "uuid",
    "app_id": "string",
    "url": "string",
    "custom_params": {
      "page_type": "product_detail",
      "product_id": "PROD_12345",
      "product_name": "Wireless Headphones"
    }
  }
}
```

## 7. カスタムパラメータの活用例

### 7.1 Eコマースサイトでの活用
```javascript
// 商品詳細ページのトラッキング
const trackingData = {
  app_id: 'ecommerce_app',
  user_agent: navigator.userAgent,
  url: window.location.href,
  custom_params: {
    page_type: 'product_detail',
    product_id: 'PROD_12345',
    product_name: 'Wireless Headphones',
    product_category: 'Electronics',
    product_price: 299.99,
    product_brand: 'TechBrand',
    product_availability: 'in_stock',
    product_rating: 4.5,
    product_review_count: 128
  }
};

fetch('https://api.access-log-tracker.com/v1/track', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': 'your_api_key'
  },
  body: JSON.stringify(trackingData)
});
```

### 7.2 ニュースサイトでの活用
```javascript
// 記事ページのトラッキング
const trackingData = {
  app_id: 'news_app',
  user_agent: navigator.userAgent,
  url: window.location.href,
  custom_params: {
    page_type: 'article',
    article_id: 'ART_001',
    article_title: 'Breaking News: Technology Advancements',
    article_category: 'Technology',
    article_author: 'John Doe',
    article_publish_date: '2024-01-15',
    article_read_time: 5,
    article_tags: ['technology', 'innovation', 'AI']
  }
};
```

### 7.3 検索結果ページでの活用
```javascript
// 検索結果ページのトラッキング
const urlParams = new URLSearchParams(window.location.search);
const trackingData = {
  app_id: 'ecommerce_app',
  user_agent: navigator.userAgent,
  url: window.location.href,
  custom_params: {
    page_type: 'search_results',
    search_query: urlParams.get('q'),
    search_results_count: document.querySelectorAll('.search-result').length,
    search_filters: urlParams.get('filters')
  }
};
```

## 8. データ分析クエリ例

### 8.1 ページタイプ別アクセス数
```bash
curl -X GET "https://api.access-log-tracker.com/v1/statistics?app_id=ecommerce_app&start_date=2024-01-01&end_date=2024-01-31" \
  -H "X-API-Key: your_api_key"
```

### 8.2 商品別アクセス数
```bash
curl -X GET "https://api.access-log-tracker.com/v1/statistics/custom-parameters?app_id=ecommerce_app&parameter_key=product_id&start_date=2024-01-01&end_date=2024-01-31&limit=10" \
  -H "X-API-Key: your_api_key"
```

### 8.3 カスタムパラメータフィルター付き統計
```bash
curl -X GET "https://api.access-log-tracker.com/v1/statistics?app_id=ecommerce_app&start_date=2024-01-01&end_date=2024-01-31&custom_params={\"page_type\":\"product_detail\",\"product_category\":\"Electronics\"}" \
  -H "X-API-Key: your_api_key"
``` 