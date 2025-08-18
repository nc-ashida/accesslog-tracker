# API仕様書

## 1. 概要

### 1.1 ベースURL
```
http://localhost:8080/v1 (開発環境)
https://api.access-log-tracker.com/v1 (本番環境予定)
```

### 1.2 インフラ構成（実装版）
- **Docker Compose**: 開発環境の統合管理 ✅ **実装完了**
- **Go + Gin**: 軽量APIサーバー（直接書き込み処理） ✅ **実装完了**
- **PostgreSQL**: 管理されたデータベース（直接書き込み） ✅ **実装完了**
- **Redis**: キャッシュ・セッション管理 ✅ **実装完了**

### 1.3 認証方式
- API Key認証（ヘッダー: `X-API-Key`） ✅ **実装完了**
- レート制限: 1000 req/min per API Key ✅ **実装完了**

### 1.4 レスポンス形式
```json
{
  "success": true,
  "data": {},
  "error": {
    "code": "ERROR_CODE",
    "message": "Error message",
    "details": "Error details"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 1.5 パフォーマンス最適化（実装版）
- **直接書き込み**: Go APIからPostgreSQLへの直接書き込み ✅ **実装完了**
- **コネクションプール**: PostgreSQL接続の効率化 ✅ **実装完了**
- **軽量処理**: シンプルな構成による高速処理 ✅ **実装完了**
- **安定性**: シンプルな構成による高安定性 ✅ **実装完了**
- **メモリ効率**: 低メモリ使用量による高スケーラビリティ ✅ **実装完了**

## 2. エンドポイント一覧

### 2.1 トラッキングデータ送信

#### POST /v1/tracking/track
アクセスログデータを送信するエンドポイント ✅ **実装完了**

**リクエストヘッダー**
```
Content-Type: application/json
X-API-Key: {api_key}
```

**リクエストボディ**
```json
{
  "app_id": "string (required)",
  "user_agent": "string (required)",
  "url": "string (optional)",
  "ip_address": "string (optional)",
  "session_id": "string (optional)",
  "referrer": "string (optional)",
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
  "timestamp": "2024-01-01T00:00:00Z"
}
```

**エラーレスポンス**
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid app_id",
    "details": "app_id is required"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 2.2 ヘルスチェック

#### GET /health
システムの健全性を確認するエンドポイント ✅ **実装完了**

**レスポンス**
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "timestamp": "2024-01-01T00:00:00Z",
    "services": {
      "database": "healthy",
      "redis": "healthy"
    }
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

#### GET /ready
アプリケーションの準備完了状態をチェック ✅ **実装完了**

#### GET /live
アプリケーションの生存状態をチェック ✅ **実装完了**

### 2.3 アプリケーション管理

#### GET /v1/applications
アプリケーション一覧を取得 ✅ **実装完了**

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
        "app_id": "string",
        "name": "string",
        "description": "string",
        "domain": "string",
        "api_key": "string",
        "is_active": true,
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
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

#### POST /v1/applications
新しいアプリケーションを登録 ✅ **実装完了**

**リクエストボディ**
```json
{
  "name": "string (required)",
  "description": "string (optional)",
  "domain": "string (required)"
}
```

#### GET /v1/applications/{id}
アプリケーション情報を取得 ✅ **実装完了**

#### PUT /v1/applications/{id}
アプリケーション情報を更新 ✅ **実装完了**

#### DELETE /v1/applications/{id}
アプリケーションを削除（論理削除） ✅ **実装完了**

### 2.4 ビーコン関連

#### GET /tracker.js
JavaScriptビーコンを配信 ✅ **実装完了**

**レスポンス**
```javascript
(function() {
    'use strict';
    
    // 設定
    var config = {
        endpoint: 'https://api.access-log-tracker.com/v1/track',
        version: '1.0.0',
        debug: false,
        customParams: {}
    };
    
    // データ収集と送信ロジック
    // ...
})();
```

#### GET /tracker.min.js
圧縮版JavaScriptビーコンを配信 ✅ **実装完了**

#### GET /tracker/{app_id}.js
カスタム設定のビーコンを配信 ✅ **実装完了**

#### GET /v1/beacon/generate
1x1ピクセルGIFビーコンを生成 ✅ **実装完了**

**クエリパラメータ**
- `app_id`: アプリケーションID（必須）
- `session_id`: セッションID（オプション）
- `url`: URL（オプション）
- `referrer`: リファラー（オプション）

#### POST /v1/beacon/generate
カスタム設定でビーコンを生成 ✅ **実装完了**

#### GET /v1/beacon/health
ビーコンサービスの健全性を確認 ✅ **実装完了**

### 2.5 統計情報

#### GET /v1/tracking/statistics
統計情報を取得 ✅ **実装完了**

**クエリパラメータ**
- `app_id`: アプリケーションID
- `start_date`: 開始日（YYYY-MM-DD）
- `end_date`: 終了日（YYYY-MM-DD）
- `group_by`: グループ化（hour, day, month）

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
        "checkout": 100000
      }
    }
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## 3. エラーコード

### 3.1 HTTPステータスコード
- `200`: 成功 ✅ **実装完了**
- `201`: 作成成功 ✅ **実装完了**
- `400`: バリデーションエラー ✅ **実装完了**
- `401`: 認証エラー ✅ **実装完了**
- `403`: 権限エラー ✅ **実装完了**
- `404`: リソースが見つからない ✅ **実装完了**
- `429`: レート制限超過 ✅ **実装完了**
- `500`: サーバーエラー ✅ **実装完了**

### 3.2 エラーコード詳細
- `VALIDATION_ERROR`: 入力値検証エラー ✅ **実装完了**
- `AUTHENTICATION_ERROR`: 認証エラー ✅ **実装完了**
- `RATE_LIMIT_EXCEEDED`: レート制限超過 ✅ **実装完了**
- `APPLICATION_NOT_FOUND`: アプリケーションが見つからない ✅ **実装完了**
- `INVALID_API_KEY`: 無効なAPIキー ✅ **実装完了**
- `BEACON_GENERATION_ERROR`: ビーコン生成エラー ✅ **実装完了**

## 4. レート制限

### 4.1 制限値
- **トラッキングAPI**: 1000 req/min per API Key ✅ **実装完了**
- **管理API**: 100 req/min per API Key ✅ **実装完了**
- **ビーコンAPI**: 500 req/min per API Key ✅ **実装完了**

### 4.2 レスポンスヘッダー
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
```

## 5. 認証・認可

### 5.1 API Key認証
- ヘッダー: `X-API-Key: {api_key}` ✅ **実装完了**
- アプリケーションごとに一意のAPIキー ✅ **実装完了**
- APIキーの自動生成機能 ✅ **実装完了**

### 5.2 認証ミドルウェア
- 必須認証: `/v1/tracking/*` ✅ **実装完了**
- オプショナル認証: `/v1/applications/*` ✅ **実装完了**
- 認証不要: `/health`, `/ready`, `/live`, `/tracker.js` ✅ **実装完了**

## 6. セキュリティ

### 6.1 CORS設定
```go
// 実装済みCORS設定
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization,X-Requested-With
CORS_EXPOSED_HEADERS=Content-Length
CORS_ALLOW_CREDENTIALS=true
CORS_MAX_AGE=86400
```

### 6.2 入力値検証
- リクエストボディのバリデーション ✅ **実装完了**
- SQLインジェクション対策 ✅ **実装完了**
- XSS攻撃対策 ✅ **実装完了**
- レート制限によるDoS攻撃対策 ✅ **実装完了**

## 7. 実装状況

### 7.1 完了済みエンドポイント
- ✅ **トラッキングAPI**: `/v1/tracking/track`, `/v1/tracking/statistics`
- ✅ **アプリケーションAPI**: `/v1/applications/*`
- ✅ **ビーコンAPI**: `/v1/beacon/*`, `/tracker.js`, `/tracker.min.js`, `/tracker/{app_id}.js`
- ✅ **ヘルスチェックAPI**: `/health`, `/ready`, `/live`

### 7.2 実装済み機能
- ✅ **認証・認可**: API Key認証、ミドルウェア
- ✅ **レート制限**: Redisベースのレート制限
- ✅ **CORS**: クロスオリジンリクエスト対応
- ✅ **エラーハンドリング**: 統一されたエラーレスポンス
- ✅ **ログ機能**: 構造化ログ出力
- ✅ **バリデーション**: リクエストデータ検証

### 7.3 テスト状況
- **API統合テスト**: 100%成功 ✅ **完了**
- **認証テスト**: 100%成功 ✅ **完了**
- **レート制限テスト**: 100%成功 ✅ **完了**
- **エラーハンドリングテスト**: 100%成功 ✅ **完了**

## 8. 使用例

### 8.1 トラッキングデータ送信
```bash
curl -X POST http://localhost:8080/v1/tracking/track \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your_api_key" \
  -d '{
    "app_id": "test_app_123",
    "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
    "url": "https://example.com/product/123",
    "custom_params": {
      "page_type": "product_detail",
      "product_id": "PROD_12345",
      "product_name": "Wireless Headphones"
    }
  }'
```

### 8.2 アプリケーション作成
```bash
curl -X POST http://localhost:8080/v1/applications \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Application",
    "description": "Test application for API testing",
    "domain": "test.example.com"
  }'
```

### 8.3 ビーコン配信
```html
<!-- HTMLでの埋め込み例 -->
<script>
(function() {
    var script = document.createElement('script');
    script.async = true;
    script.src = 'http://localhost:8080/tracker.js';
    script.setAttribute('data-app-id', 'test_app_123');
    var firstScript = document.getElementsByTagName('script')[0];
    firstScript.parentNode.insertBefore(script, firstScript);
})();
</script>
```

## 9. 次のステップ

### 9.1 本番環境対応
1. **HTTPS対応**: SSL/TLS証明書の設定
2. **ドメイン設定**: 本番ドメインの設定
3. **ロードバランサー**: ALB/ELBの設定
4. **CDN**: CloudFrontの設定

### 9.2 機能拡張
1. **Webhook機能**: 外部システム連携
2. **バッチ処理**: 大量データ処理
3. **統計ダッシュボード**: リアルタイム統計表示
4. **データエクスポート**: CSV/JSON形式でのデータ出力 