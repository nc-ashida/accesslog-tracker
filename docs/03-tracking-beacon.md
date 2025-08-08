# トラッキングビーコン仕様書

## 1. 概要

### 1.1 目的
クライアントサイトに埋め込まれる軽量なJavaScriptトラッキングビーコンを提供する。
ページの読み込み速度に影響を与えず、安全かつ効率的にアクセスログを収集する。

### 1.2 基本仕様
- **ファイルサイズ**: 5KB以下（gzip圧縮後）
- **読み込み方式**: 非同期読み込み
- **ブラウザ対応**: IE11以上、Chrome、Firefox、Safari、Edge
- **セキュリティ**: クロスサイト干渉防止
- **配信方式**: CloudFront + S3による高速配信

### 1.3 配信インフラ
- **CloudFront**: グローバルCDN配信
- **S3**: ビーコンファイル保存
- **Lambda@Edge**: 配信最適化処理
- **エッジロケーション**: 世界各地のエッジサーバー

## 2. 実装仕様

### 2.1 基本実装

#### HTML埋め込みコード
```html
<!-- 基本実装 -->
<script>
(function() {
    var script = document.createElement('script');
    script.async = true;
    script.src = 'https://d1234567890.cloudfront.net/tracker.js'; // CloudFront URL
    script.setAttribute('data-app-id', 'YOUR_APP_ID');
    script.setAttribute('data-client-sub-id', 'OPTIONAL_SUB_ID');
    script.setAttribute('data-module-id', 'OPTIONAL_MODULE_ID');
    var firstScript = document.getElementsByTagName('script')[0];
    firstScript.parentNode.insertBefore(script, firstScript);
})();
</script>
```

#### 高度な実装
```html
<!-- カスタム設定付き実装 -->
<script>
window.ALT_CONFIG = {
    app_id: 'YOUR_APP_ID',
    client_sub_id: 'OPTIONAL_SUB_ID',
    module_id: 'OPTIONAL_MODULE_ID',
    endpoint: 'https://api.access-log-tracker.com/v1/track', // ALBエンドポイント
    debug: false,
    respect_dnt: true,
    session_timeout: 1800000 // 30分
};
</script>
<script async src="https://d1234567890.cloudfront.net/tracker.js"></script>
```

### 2.2 CloudFront配信設定

#### 2.2.1 S3バケット設定
```yaml
# CloudFormation テンプレート
Resources:
  BeaconBucket:
    Type: AWS::S3::Bucket
    Properties:
      BucketName: access-log-tracker-beacon
      PublicAccessBlockConfiguration:
        BlockPublicAcls: false
        BlockPublicPolicy: false
        IgnorePublicAcls: false
        RestrictPublicBuckets: false
      CorsConfiguration:
        CorsRules:
          - AllowedHeaders: ['*']
            AllowedMethods: [GET]
            AllowedOrigins: ['*']
            MaxAge: 3000

  BeaconBucketPolicy:
    Type: AWS::S3::BucketPolicy
    Properties:
      Bucket: !Ref BeaconBucket
      PolicyDocument:
        Statement:
          - Effect: Allow
            Principal: '*'
            Action: s3:GetObject
            Resource: !Sub '${BeaconBucket}/*'
```

#### 2.2.2 CloudFrontディストリビューション
```yaml
  BeaconDistribution:
    Type: AWS::CloudFront::Distribution
    Properties:
      DistributionConfig:
        Origins:
          - Id: S3Origin
            DomainName: !GetAtt BeaconBucket.RegionalDomainName
            S3OriginConfig:
              OriginAccessIdentity: ''
        DefaultCacheBehavior:
          TargetOriginId: S3Origin
          ViewerProtocolPolicy: redirect-to-https
          CachePolicyId: 4135ea2d-6df8-44a3-9df3-4b5a84be39ad # CachingOptimized
          OriginRequestPolicyId: 88a5eaf4-2fd4-4709-b370-b4c650ea3fcf # CORS-S3Origin
        Enabled: true
        Comment: Access Log Tracker Beacon Distribution
```

#### 2.2.3 Lambda@Edge最適化
```javascript
// Lambda@Edge - Viewer Request
exports.handler = async (event) => {
  const request = event.Records[0].cf.request;
  
  // ビーコンファイルの配信最適化
  if (request.uri.endsWith('.js')) {
    // キャッシュヘッダー設定
    request.headers['cache-control'] = [{
      key: 'Cache-Control',
      value: 'public, max-age=3600, s-maxage=86400'
    }];
    
    // 圧縮対応
    const acceptEncoding = request.headers['accept-encoding']?.[0]?.value || '';
    if (acceptEncoding.includes('gzip')) {
      request.uri = request.uri.replace('.js', '.js.gz');
    }
  }
  
  return request;
};
```

### 2.2 ページごとカスタムパラメータ対応

#### 2.2.1 データ属性による設定
```html
<!-- ページ固有のカスタムパラメータ -->
<script>
window.ALT_CONFIG = {
    app_id: 'YOUR_APP_ID',
    // ページ固有のカスタムパラメータ
    custom_params: {
        page_type: 'product_detail',
        product_id: '12345',
        category: 'electronics',
        price_range: '1000-5000',
        user_segment: 'premium'
    }
};
</script>
<script async src="https://cdn.access-log-tracker.com/tracker.js"></script>
```

#### 2.2.2 動的パラメータ設定
```html
<!-- 動的にパラメータを設定 -->
<script>
// ページ読み込み後に動的にパラメータを設定
window.ALT_CONFIG = {
    app_id: 'YOUR_APP_ID'
};

// ページ固有の情報を動的に取得
document.addEventListener('DOMContentLoaded', function() {
    // 商品ページの場合
    if (window.location.pathname.includes('/product/')) {
        window.ALT_CONFIG.custom_params = {
            page_type: 'product_detail',
            product_id: document.querySelector('[data-product-id]')?.dataset.productId,
            product_name: document.querySelector('[data-product-name]')?.dataset.productName,
            product_price: document.querySelector('[data-product-price]')?.dataset.productPrice,
            product_category: document.querySelector('[data-product-category]')?.dataset.productCategory
        };
    }
    
    // カテゴリページの場合
    else if (window.location.pathname.includes('/category/')) {
        window.ALT_CONFIG.custom_params = {
            page_type: 'category_list',
            category_id: document.querySelector('[data-category-id]')?.dataset.categoryId,
            category_name: document.querySelector('[data-category-name]')?.dataset.categoryName,
            product_count: document.querySelectorAll('.product-item').length
        };
    }
    
    // 検索結果ページの場合
    else if (window.location.pathname.includes('/search')) {
        const urlParams = new URLSearchParams(window.location.search);
        window.ALT_CONFIG.custom_params = {
            page_type: 'search_results',
            search_query: urlParams.get('q'),
            search_results_count: document.querySelectorAll('.search-result').length,
            search_filters: urlParams.get('filters')
        };
    }
});
</script>
<script async src="https://cdn.access-log-tracker.com/tracker.js"></script>
```

#### 2.2.3 イベントベースのパラメータ更新
```javascript
// ユーザーアクションに基づくパラメータ更新
document.addEventListener('click', function(e) {
    // 商品クリック時
    if (e.target.closest('.product-item')) {
        const productItem = e.target.closest('.product-item');
        ALT.updateCustomParams({
            action_type: 'product_click',
            product_id: productItem.dataset.productId,
            product_position: productItem.dataset.position,
            click_element: e.target.tagName.toLowerCase()
        });
    }
    
    // カテゴリクリック時
    if (e.target.closest('.category-link')) {
        const categoryLink = e.target.closest('.category-link');
        ALT.updateCustomParams({
            action_type: 'category_click',
            category_id: categoryLink.dataset.categoryId,
            category_name: categoryLink.dataset.categoryName
        });
    }
});

// フォーム送信時
document.addEventListener('submit', function(e) {
    if (e.target.classList.contains('search-form')) {
        const searchInput = e.target.querySelector('input[name="q"]');
        ALT.updateCustomParams({
            action_type: 'search_submit',
            search_query: searchInput.value,
            search_form_id: e.target.id
        });
    }
});
```

### 2.3 トラッキングビーコン（tracker.js）

#### 基本構造（簡素化版）
```javascript
(function() {
    'use strict';
    
    // 名前空間の分離
    var ALT = window.ALT || {};
    
    // 設定（簡素化版）
    var config = {
        endpoint: 'https://api.access-log-tracker.com/v1/track',
        session_timeout: 1800000, // 30分
        respect_dnt: true,
        debug: false,
        custom_params: {},
        // 簡素化による最適化
        retry_attempts: 2,
        retry_delay: 1000,
        timeout: 5000 // 5秒タイムアウト
    };
    
    // 初期化
    function init() {
        if (shouldTrack()) {
            loadConfig();
            setupSession();
            trackPageView();
        }
    }
    
    // トラッキング判定
    function shouldTrack() {
        // DNT（Do Not Track）チェック
        if (config.respect_dnt && navigator.doNotTrack === '1') {
            return false;
        }
        
        // クローラー検出
        if (isCrawler()) {
            return false;
        }
        
        return true;
    }
    
    // クローラー検出
    function isCrawler() {
        var userAgent = navigator.userAgent.toLowerCase();
        var crawlerPatterns = [
            /bot/, /crawler/, /spider/, /scraper/,
            /googlebot/, /bingbot/, /slurp/, /duckduckbot/,
            /baiduspider/, /yandexbot/, /facebookexternalhit/,
            /twitterbot/, /linkedinbot/, /whatsapp/
        ];
        
        return crawlerPatterns.some(function(pattern) {
            return pattern.test(userAgent);
        });
    }
    
    // 設定読み込み
    function loadConfig() {
        // スクリプトタグから設定を読み込み
        var script = document.currentScript || 
                    document.querySelector('script[src*="tracker.js"]');
        
        if (script) {
            config.app_id = script.getAttribute('data-app-id');
            config.client_sub_id = script.getAttribute('data-client-sub-id');
            config.module_id = script.getAttribute('data-module-id');
        }
        
        // グローバル設定をマージ
        if (window.ALT_CONFIG) {
            Object.assign(config, window.ALT_CONFIG);
        }
    }
    
    // セッション管理
    function setupSession() {
        var sessionId = getSessionId();
        var lastActivity = sessionStorage.getItem('alt_last_activity');
        var now = Date.now();
        
        // セッションタイムアウトチェック
        if (lastActivity && (now - parseInt(lastActivity)) > config.session_timeout) {
            sessionId = generateSessionId();
        }
        
        sessionStorage.setItem('alt_session_id', sessionId);
        sessionStorage.setItem('alt_last_activity', now.toString());
        
        return sessionId;
    }
    
    // セッションID生成
    function generateSessionId() {
        return 'alt_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
    }
    
    // セッションID取得
    function getSessionId() {
        return sessionStorage.getItem('alt_session_id') || generateSessionId();
    }
    
    // ページビュートラッキング（簡素化版）
    function trackPageView() {
        var trackingData = {
            app_id: config.app_id,
            client_sub_id: config.client_sub_id,
            module_id: config.module_id,
            url: window.location.href,
            referrer: document.referrer,
            user_agent: navigator.userAgent,
            session_id: getSessionId(),
            screen_resolution: screen.width + 'x' + screen.height,
            language: navigator.language || navigator.userLanguage,
            timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
            custom_params: config.custom_params
        };
        
        // 簡素化版の送信処理
        sendTrackingData(trackingData);
    }
    
    // データ送信（簡素化版）
    function sendTrackingData(data) {
        var xhr = new XMLHttpRequest();
        var retryCount = 0;
        
        function attemptSend() {
            xhr.open('POST', config.endpoint, true);
            xhr.setRequestHeader('Content-Type', 'application/json');
            xhr.setRequestHeader('X-API-Key', config.app_id);
            xhr.timeout = config.timeout;
            
            xhr.onreadystatechange = function() {
                if (xhr.readyState === 4) {
                    if (xhr.status === 200) {
                        // 成功（直接書き込み）
                        if (config.debug) {
                            console.log('ALT: Tracking data saved successfully');
                        }
                    } else if (xhr.status >= 500 && retryCount < config.retry_attempts) {
                        // サーバーエラー時はリトライ
                        retryCount++;
                        setTimeout(attemptSend, config.retry_delay * retryCount);
                    } else {
                        // エラー
                        if (config.debug) {
                            console.error('ALT: Failed to send tracking data', xhr.status);
                        }
                    }
                }
            };
            
            xhr.onerror = function() {
                if (retryCount < config.retry_attempts) {
                    retryCount++;
                    setTimeout(attemptSend, config.retry_delay * retryCount);
                }
            };
            
            xhr.ontimeout = function() {
                if (retryCount < config.retry_attempts) {
                    retryCount++;
                    setTimeout(attemptSend, config.retry_delay * retryCount);
                }
            };
            
            xhr.send(JSON.stringify(data));
        }
        
        attemptSend();
    }
    
    // カスタムパラメータ更新
    ALT.updateCustomParams = function(newParams) {
        Object.assign(config.custom_params, newParams);
    };
    
    // カスタムイベント送信
    ALT.trackEvent = function(eventType, eventData) {
        var eventTrackingData = {
            app_id: config.app_id,
            client_sub_id: config.client_sub_id,
            module_id: config.module_id,
            url: window.location.href,
            user_agent: navigator.userAgent,
            session_id: getSessionId(),
            custom_params: Object.assign({}, config.custom_params, {
                event_type: eventType,
                event_data: eventData
            })
        };
        
        sendTrackingData(eventTrackingData);
    };
    
    // 初期化実行
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
    
    // グローバルに公開
    window.ALT = ALT;
})();
```

## 3. 送信情報仕様

### 3.1 基本送信パラメータ

| キー                | 型     | 必須 | 説明                 | 例                                  |
| ------------------- | ------ | ---- | -------------------- | ----------------------------------- |
| `app_id`            | string | ○    | アプリケーションID   | `"my-app-001"`                      |
| `client_sub_id`     | string | ×    | クライアントサブID   | `"client-123"`                      |
| `module_id`         | string | ×    | モジュールID         | `"product-catalog"`                 |
| `url`               | string | ○    | 現在のページURL      | `"https://example.com/product/123"` |
| `referrer`          | string | ×    | リファラーページURL  | `"https://google.com"`              |
| `user_agent`        | string | ○    | ブラウザのUser-Agent | `"Mozilla/5.0..."`                  |
| `session_id`        | string | ○    | セッションID         | `"alt_1703123456789_abc123def"`     |
| `screen_resolution` | string | ×    | 画面解像度           | `"1920x1080"`                       |
| `language`          | string | ×    | ブラウザ言語設定     | `"ja-JP"`                           |
| `timezone`          | string | ×    | タイムゾーン         | `"Asia/Tokyo"`                      |
| `custom_params`     | object | ×    | カスタムパラメータ   | `{"page_type": "product_detail"}`   |

### 3.2 カスタムパラメータ例

#### Eコマースサイト
| キー                   | 型     | 説明               | 例                                         |
| ---------------------- | ------ | ------------------ | ------------------------------------------ |
| `page_type`            | string | ページタイプ       | `"product_detail"`, `"cart"`, `"checkout"` |
| `product_id`           | string | 商品ID             | `"PROD-12345"`                             |
| `product_name`         | string | 商品名             | `"iPhone 15 Pro"`                          |
| `product_category`     | string | 商品カテゴリ       | `"electronics"`                            |
| `product_price`        | number | 商品価格           | `129800`                                   |
| `product_brand`        | string | ブランド名         | `"Apple"`                                  |
| `product_availability` | string | 在庫状況           | `"in_stock"`, `"out_of_stock"`             |
| `product_rating`       | number | 評価（1-5）        | `4.5`                                      |
| `product_review_count` | number | レビュー数         | `128`                                      |
| `cart_total`           | number | カート合計金額     | `258000`                                   |
| `cart_item_count`      | number | カート内商品数     | `3`                                        |
| `user_segment`         | string | ユーザーセグメント | `"premium"`, `"regular"`                   |

#### ニュースサイト
| キー                    | 型     | 説明           | 例                                    |
| ----------------------- | ------ | -------------- | ------------------------------------- |
| `page_type`             | string | ページタイプ   | `"article"`, `"category"`, `"search"` |
| `article_id`            | string | 記事ID         | `"ART-67890"`                         |
| `article_title`         | string | 記事タイトル   | `"最新技術トレンド"`                  |
| `article_category`      | string | 記事カテゴリ   | `"technology"`                        |
| `article_author`        | string | 著者名         | `"田中太郎"`                          |
| `article_publish_date`  | string | 公開日         | `"2024-01-15"`                        |
| `article_read_time`     | number | 読了時間（分） | `5`                                   |
| `article_tags`          | array  | タグ一覧       | `["AI", "機械学習", "DX"]`            |
| `article_word_count`    | number | 文字数         | `2500`                                |
| `article_comment_count` | number | コメント数     | `42`                                  |

#### 企業サイト
| キー                | 型     | 説明             | 例                                    |
| ------------------- | ------ | ---------------- | ------------------------------------- |
| `page_type`         | string | ページタイプ     | `"company"`, `"service"`, `"contact"` |
| `company_section`   | string | 企業セクション   | `"about"`, `"services"`, `"careers"`  |
| `service_name`      | string | サービス名       | `"クラウドソリューション"`            |
| `contact_form_type` | string | 問い合わせ種別   | `"general"`, `"sales"`, `"support"`   |
| `download_item`     | string | ダウンロード項目 | `"whitepaper"`, `"case_study"`        |

### 3.3 イベント送信パラメータ

#### カスタムイベント
| キー         | 型     | 必須 | 説明               | 例                                |
| ------------ | ------ | ---- | ------------------ | --------------------------------- |
| `event_type` | string | ○    | イベントタイプ     | `"button_click"`, `"form_submit"` |
| `event_data` | object | ×    | イベント詳細データ | `{"button_id": "cta-button"}`     |

#### ユーザーアクション例
| イベントタイプ  | 説明                 | 追加パラメータ例                                     |
| --------------- | -------------------- | ---------------------------------------------------- |
| `product_click` | 商品クリック         | `{"product_id": "123", "position": 5}`               |
| `add_to_cart`   | カート追加           | `{"product_id": "123", "quantity": 2}`               |
| `search_submit` | 検索実行             | `{"query": "iPhone", "results_count": 25}`           |
| `form_submit`   | フォーム送信         | `{"form_id": "contact", "field_count": 5}`           |
| `video_play`    | 動画再生             | `{"video_id": "intro", "duration": 120}`             |
| `download`      | ファイルダウンロード | `{"file_name": "whitepaper.pdf", "file_size": 2048}` |

### 3.4 セッション管理パラメータ

| キー                 | 型      | 説明                                 | 例              |
| -------------------- | ------- | ------------------------------------ | --------------- |
| `session_start_time` | number  | セッション開始時刻（Unix timestamp） | `1703123456789` |
| `session_duration`   | number  | セッション継続時間（ミリ秒）         | `1800000`       |
| `page_views`         | number  | ページビュー数                       | `5`             |
| `is_new_session`     | boolean | 新規セッション判定                   | `true`          |

### 3.5 エラー情報パラメータ

| キー              | 型     | 説明                     | 例                                      |
| ----------------- | ------ | ------------------------ | --------------------------------------- |
| `error_type`      | string | エラータイプ             | `"network_error"`, `"validation_error"` |
| `error_message`   | string | エラーメッセージ         | `"Network timeout"`                     |
| `retry_count`     | number | リトライ回数             | `2`                                     |
| `response_status` | number | HTTPレスポンスステータス | `500`                                   |